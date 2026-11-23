package oauth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	stateTTL    = 10 * time.Minute
	randomBytes = 32
	// bindingCookie ties a linking attempt to the browser that started it.
	bindingCookie = "area_oauth_binding"
)

var errInvalidState = errors.New("missing, unknown or expired state")

// pendingLink is what authorize remembers until the provider calls back.
type pendingLink struct {
	UserID   uuid.UUID `json:"user_id"`
	Service  string    `json:"service"`
	Verifier string    `json:"verifier"`
	// BindingHash is the SHA-256 of the binding cookie's value.
	BindingHash string `json:"binding_hash"`
}

// Only a hash of the state is used as the key, so the Redis contents alone
// cannot be replayed against the callback.
func stateKey(state string) string {
	return "oauth:state:" + sha256Hex(state)
}

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func randomToken() string {
	b := make([]byte, randomBytes)
	_, _ = rand.Read(b) // never fails since Go 1.24
	return base64.RawURLEncoding.EncodeToString(b)
}

func (s *Service) savePending(ctx context.Context, state string, p pendingLink) error {
	b, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("encode state: %w", err)
	}
	if err := s.states.Set(ctx, stateKey(state), b, stateTTL).Err(); err != nil {
		return fmt.Errorf("save state: %w", err)
	}
	return nil
}

// takePending deletes the state as it reads it, so each state works once.
func (s *Service) takePending(ctx context.Context, state string) (pendingLink, error) {
	if state == "" {
		return pendingLink{}, errInvalidState
	}
	b, err := s.states.GetDel(ctx, stateKey(state)).Bytes()
	if errors.Is(err, redis.Nil) {
		return pendingLink{}, errInvalidState
	}
	if err != nil {
		return pendingLink{}, fmt.Errorf("load state: %w", err)
	}
	var p pendingLink
	if err := json.Unmarshal(b, &p); err != nil {
		return pendingLink{}, fmt.Errorf("decode state: %w", err)
	}
	return p, nil
}

// setBinding is Lax, not Strict like the session cookie: the browser must send
// it on the top-level redirect coming back from the provider's site.
func (s *Service) setBinding(w http.ResponseWriter, service, value string, maxAge int) {
	// Secure follows the API's scheme: plain-HTTP dev and LAN setups would lose the cookie.
	http.SetCookie(w, &http.Cookie{ //nolint:gosec
		Name:     bindingCookie,
		Value:    value,
		Path:     callbackPath(service),
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   s.secureCookie,
		SameSite: http.SameSiteLaxMode,
	})
}

func bindingMatches(r *http.Request, wantHash string) bool {
	c, err := r.Cookie(bindingCookie)
	if err != nil || c.Value == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(sha256Hex(c.Value)), []byte(wantHash)) == 1
}
