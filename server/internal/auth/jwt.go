package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const issuer = "area"

// MinSecretLen is the minimum JWT_SECRET length (32 bytes = 256 bits for HS256).
const MinSecretLen = 32

var ErrInvalidToken = errors.New("invalid token")

type Tokens struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

func NewTokens(secret string, ttl time.Duration) (*Tokens, error) {
	if len(secret) < MinSecretLen {
		return nil, errors.New("JWT_SECRET must be at least 32 characters")
	}
	if ttl <= 0 {
		return nil, errors.New("JWT TTL must be positive")
	}
	return &Tokens{secret: []byte(secret), ttl: ttl, now: time.Now}, nil
}

// Issue signs an HS256 access token for userID.
func (t *Tokens) Issue(userID uuid.UUID) (string, time.Duration, error) {
	now := t.now()
	claims := jwt.RegisteredClaims{
		Issuer:    issuer,
		Subject:   userID.String(),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(t.ttl)),
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(t.secret)
	return signed, t.ttl, err
}

// Parse validates the signature, algorithm, issuer and expiry and returns the user ID.
func (t *Tokens) Parse(token string) (uuid.UUID, error) {
	var claims jwt.RegisteredClaims
	_, err := jwt.ParseWithClaims(token, &claims,
		func(*jwt.Token) (any, error) { return t.secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(issuer),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(t.now),
	)
	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}
	id, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}
	return id, nil
}
