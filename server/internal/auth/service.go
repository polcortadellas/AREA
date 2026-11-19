// Package auth handles local accounts: registration with e-mail confirmation,
// Argon2id password hashing and JWT access tokens.
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/polcortadellas/AREA/server/internal/db/sqlc"
	mailer "github.com/polcortadellas/AREA/server/internal/mail"
)

const (
	minPasswordLen    = 8
	maxPasswordLen    = 128 // bounds Argon2id work per request
	verificationTTL   = 24 * time.Hour
	verificationBytes = 32
)

var (
	ErrInvalidEmail       = errors.New("invalid email")
	ErrWeakPassword       = errors.New("password must be 8 to 128 characters")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailNotVerified   = errors.New("email not verified")
	ErrInvalidVerifyToken = errors.New("invalid or expired verification token")
)

type Service struct {
	pool       *pgxpool.Pool
	tokens     *Tokens
	mail       mailer.Sender
	webBaseURL string
	// Compared against when the e-mail is unknown, so login takes the same time.
	dummyHash string
}

func NewService(pool *pgxpool.Pool, tokens *Tokens, mail mailer.Sender, webBaseURL string) (*Service, error) {
	dummy, err := HashPassword("dummy-password")
	if err != nil {
		return nil, err
	}
	return &Service{
		pool:       pool,
		tokens:     tokens,
		mail:       mail,
		webBaseURL: strings.TrimRight(webBaseURL, "/"),
		dummyHash:  dummy,
	}, nil
}

// Register creates an unverified account and e-mails a confirmation link. It
// returns nil for an already verified address so the endpoint cannot be used to
// discover which e-mails have accounts. Registering again before confirming
// replaces the password and invalidates earlier links.
func (s *Service) Register(ctx context.Context, email, password string) error {
	email, err := normalizeEmail(email)
	if err != nil {
		return err
	}
	if n := len([]rune(password)); n < minPasswordLen || n > maxPasswordLen {
		return ErrWeakPassword
	}
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}
	token, tokenHash, err := newVerificationToken()
	if err != nil {
		return err
	}

	send := false
	err = pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		q := sqlc.New(tx)
		user, err := q.GetUserByEmail(ctx, email)
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			user, err = q.CreateUser(ctx, sqlc.CreateUserParams{Email: email, PasswordHash: &hash})
			if err != nil {
				return err
			}
		case err != nil:
			return err
		case user.EmailVerifiedAt != nil:
			return nil
		default:
			if err := q.SetPasswordHash(ctx, sqlc.SetPasswordHashParams{ID: user.ID, PasswordHash: &hash}); err != nil {
				return err
			}
			if err := q.InvalidateEmailVerifications(ctx, user.ID); err != nil {
				return err
			}
		}
		send = true
		return q.CreateEmailVerification(ctx, sqlc.CreateEmailVerificationParams{
			UserID:    user.ID,
			TokenHash: tokenHash,
			ExpiresAt: time.Now().Add(verificationTTL),
		})
	})
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return nil // concurrent registration of the same e-mail
	}
	if err != nil || !send {
		return err
	}

	link := s.webBaseURL + "/verify-email?token=" + url.QueryEscape(token)
	body := fmt.Sprintf("Welcome to AREA!\n\nConfirm your e-mail address by opening this link within 24 hours:\n\n%s\n\nIf you did not create an account, ignore this e-mail.\n", link)
	return s.mail.Send(ctx, email, "Confirm your AREA account", body)
}

// VerifyEmail consumes a confirmation token, activates the account and logs the user in.
func (s *Service) VerifyEmail(ctx context.Context, token string) (string, time.Duration, error) {
	var userID uuid.UUID
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		q := sqlc.New(tx)
		id, err := q.ConsumeEmailVerification(ctx, hashToken(token))
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInvalidVerifyToken
		}
		if err != nil {
			return err
		}
		userID = id
		return q.MarkEmailVerified(ctx, id)
	})
	if err != nil {
		return "", 0, err
	}
	return s.tokens.Issue(userID)
}

// Login checks the password first and only then the verification state, so an
// unverified account is only revealed to someone who knows its password.
func (s *Service) Login(ctx context.Context, email, password string) (string, time.Duration, error) {
	email, err := normalizeEmail(email)
	if err != nil {
		return "", 0, ErrInvalidCredentials
	}
	user, err := sqlc.New(s.pool).GetUserByEmail(ctx, email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return "", 0, err
	}
	if err != nil || user.PasswordHash == nil {
		_, _ = CheckPassword(password, s.dummyHash)
		return "", 0, ErrInvalidCredentials
	}
	ok, err := CheckPassword(password, *user.PasswordHash)
	if err != nil {
		return "", 0, err
	}
	if !ok {
		return "", 0, ErrInvalidCredentials
	}
	if user.EmailVerifiedAt == nil {
		return "", 0, ErrEmailNotVerified
	}
	return s.tokens.Issue(user.ID)
}

func (s *Service) User(ctx context.Context, id uuid.UUID) (sqlc.User, error) {
	return sqlc.New(s.pool).GetUserByID(ctx, id)
}

func normalizeEmail(raw string) (string, error) {
	addr, err := mail.ParseAddress(strings.TrimSpace(raw))
	if err != nil || addr.Name != "" || len(addr.Address) > 254 {
		return "", ErrInvalidEmail
	}
	return strings.ToLower(addr.Address), nil
}

// newVerificationToken returns the token for the link and the SHA-256 hash
// stored in the database, so a database dump cannot be used to verify accounts.
func newVerificationToken() (token, tokenHash string, err error) {
	b := make([]byte, verificationBytes)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	token = base64.RawURLEncoding.EncodeToString(b)
	return token, hashToken(token), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
