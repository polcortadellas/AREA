package oauth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/oauth2"

	"github.com/polcortadellas/AREA/server/internal/db/sqlc"
	"github.com/polcortadellas/AREA/server/internal/node"
)

const (
	// A token this close to expiry is refreshed before it is handed out.
	expiryLeeway = time.Minute
)

var (
	ErrNotLinked = errors.New("service not linked")
	// ErrRelinkRequired means the access token expired and cannot be renewed:
	// the user has to link the service again.
	ErrRelinkRequired = errors.New("service must be linked again")
)

// sealToken and openToken are the only places tokens are converted for
// storage. F-03 replaces them with AES-GCM; until then the bytes are the token.
func sealToken(token string) []byte {
	if token == "" {
		return nil
	}
	return []byte(token)
}

func openToken(b []byte) string {
	return string(b)
}

func (s *Service) saveLink(ctx context.Context, userID uuid.UUID, service string, tok *oauth2.Token, requestedScopes string) error {
	granted, _ := tok.Extra("scope").(string)
	return s.q.UpsertUserService(ctx, sqlc.UpsertUserServiceParams{
		UserID:          userID,
		Service:         service,
		AccessTokenEnc:  sealToken(tok.AccessToken),
		RefreshTokenEnc: sealToken(tok.RefreshToken),
		ExpiresAt:       expiry(tok),
		Scopes:          normalizeScopes(granted, requestedScopes),
	})
}

// expiry is nil for tokens that never expire, such as GitHub OAuth App tokens.
func expiry(tok *oauth2.Token) *time.Time {
	if tok.Expiry.IsZero() {
		return nil
	}
	e := tok.Expiry
	return &e
}

// Credentials returns a usable access token for the user's linked service,
// refreshing it first when it is about to expire. The row stays locked during
// the refresh, so concurrent callers wait for it instead of spending the same
// single-use refresh token twice.
func (s *Service) Credentials(ctx context.Context, userID uuid.UUID, service string) (node.Credentials, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return node.Credentials{}, fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	row, err := s.q.WithTx(tx).LockUserService(ctx, sqlc.LockUserServiceParams{UserID: userID, Service: service})
	if errors.Is(err, pgx.ErrNoRows) {
		return node.Credentials{}, ErrNotLinked
	}
	if err != nil {
		return node.Credentials{}, fmt.Errorf("load %s token: %w", service, err)
	}
	if !expiresBefore(row, time.Now().Add(expiryLeeway)) {
		return node.Credentials{AccessToken: openToken(row.AccessTokenEnc)}, nil
	}
	access, err := s.refresh(ctx, tx, row)
	if err := endRefresh(ctx, tx, err); err != nil {
		return node.Credentials{}, err
	}
	return node.Credentials{AccessToken: access}, nil
}

func expiresBefore(row sqlc.UserService, t time.Time) bool {
	return row.ExpiresAt != nil && row.ExpiresAt.Before(t)
}

// refresh trades the locked row's refresh token for new tokens and saves them
// in tx. A refresh token the provider rejects for good is dropped, so it is
// not tried again, and ErrRelinkRequired is returned.
func (s *Service) refresh(ctx context.Context, tx pgx.Tx, row sqlc.UserService) (string, error) {
	if row.RefreshTokenEnc == nil {
		return "", ErrRelinkRequired
	}
	p, ok := s.providers[row.Service]
	if !ok {
		return "", fmt.Errorf("refresh %s: provider not configured", row.Service)
	}
	oldRefresh := openToken(row.RefreshTokenEnc)
	tok, err := p.Config.TokenSource(s.providerCtx(ctx), &oauth2.Token{RefreshToken: oldRefresh}).Token()
	q := s.q.WithTx(tx)
	if isInvalidGrant(err) {
		if err := q.UpdateUserServiceTokens(ctx, sqlc.UpdateUserServiceTokensParams{
			ID: row.ID, AccessTokenEnc: row.AccessTokenEnc, RefreshTokenEnc: nil, ExpiresAt: row.ExpiresAt,
		}); err != nil {
			return "", fmt.Errorf("drop rejected refresh token: %w", err)
		}
		return "", ErrRelinkRequired
	}
	if err != nil {
		return "", fmt.Errorf("refresh %s: %s", row.Service, describe(err))
	}

	newRefresh := tok.RefreshToken
	if newRefresh == "" {
		newRefresh = oldRefresh // providers that do not rotate refresh tokens omit it
	}
	if err := q.UpdateUserServiceTokens(ctx, sqlc.UpdateUserServiceTokensParams{
		ID: row.ID, AccessTokenEnc: sealToken(tok.AccessToken), RefreshTokenEnc: sealToken(newRefresh), ExpiresAt: expiry(tok),
	}); err != nil {
		return "", fmt.Errorf("save refreshed %s token: %w", row.Service, err)
	}
	return tok.AccessToken, nil
}

// endRefresh commits a refresh, including a dropped refresh token, and rolls
// back anything else. It returns the refresh error, or the commit error.
func endRefresh(ctx context.Context, tx pgx.Tx, err error) error {
	if err != nil && !errors.Is(err, ErrRelinkRequired) {
		_ = tx.Rollback(ctx)
		return err
	}
	if cerr := tx.Commit(ctx); cerr != nil {
		return fmt.Errorf("commit refresh: %w", cerr)
	}
	return err
}

func isInvalidGrant(err error) bool {
	var re *oauth2.RetrieveError
	return errors.As(err, &re) && re.ErrorCode == "invalid_grant"
}

// describe keeps provider response bodies out of the logs: a body could hold
// a token.
func describe(err error) string {
	var re *oauth2.RetrieveError
	if !errors.As(err, &re) {
		return err.Error()
	}
	status := 0
	if re.Response != nil {
		status = re.Response.StatusCode
	}
	return fmt.Sprintf("provider answered %d %s", status, re.ErrorCode)
}
