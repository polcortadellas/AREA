package oauth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/oauth2"

	"github.com/polcortadellas/AREA/server/internal/db/sqlc"
)

// sealToken is the only place tokens are converted for storage. F-03
// replaces it with AES-GCM; until then the bytes are the token.
func sealToken(token string) []byte {
	if token == "" {
		return nil
	}
	return []byte(token)
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
