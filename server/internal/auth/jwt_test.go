package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const testSecret = "0123456789abcdef0123456789abcdef"

func newTestTokens(t *testing.T) *Tokens {
	t.Helper()
	tk, err := NewTokens(testSecret, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	return tk
}

func TestNewTokensRejectsShortSecret(t *testing.T) {
	if _, err := NewTokens("too-short", time.Hour); err == nil {
		t.Fatal("expected error")
	}
}

func TestIssueAndParse(t *testing.T) {
	tk := newTestTokens(t)
	id := uuid.New()
	token, ttl, err := tk.Issue(id)
	if err != nil || ttl != time.Hour {
		t.Fatalf("issue: ttl=%v err=%v", ttl, err)
	}
	got, err := tk.Parse(token)
	if err != nil || got != id {
		t.Fatalf("parse: got=%v err=%v", got, err)
	}
}

func TestParseRejectsExpiredToken(t *testing.T) {
	tk := newTestTokens(t)
	token, _, _ := tk.Issue(uuid.New())
	tk.now = func() time.Time { return time.Now().Add(2 * time.Hour) }
	if _, err := tk.Parse(token); err == nil {
		t.Fatal("expired token accepted")
	}
}

func TestParseRejectsOtherSecret(t *testing.T) {
	token, _, _ := newTestTokens(t).Issue(uuid.New())
	other, _ := NewTokens(strings.Repeat("x", 32), time.Hour)
	if _, err := other.Parse(token); err == nil {
		t.Fatal("token signed with another secret accepted")
	}
}

func TestParseRejectsOtherAlgorithms(t *testing.T) {
	tk := newTestTokens(t)
	claims := jwt.RegisteredClaims{
		Issuer:    issuer,
		Subject:   uuid.NewString(),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}
	hs512, _ := jwt.NewWithClaims(jwt.SigningMethodHS512, claims).SignedString([]byte(testSecret))
	none, _ := jwt.NewWithClaims(jwt.SigningMethodNone, claims).SignedString(jwt.UnsafeAllowNoneSignatureType)
	for name, token := range map[string]string{"HS512": hs512, "none": none} {
		if _, err := tk.Parse(token); err == nil {
			t.Errorf("%s token accepted", name)
		}
	}
}

func TestRequireAuth(t *testing.T) {
	tk := newTestTokens(t)
	id := uuid.New()
	valid, _, _ := tk.Issue(id)

	h := tk.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, ok := UserID(r.Context())
		if !ok || got != id {
			t.Errorf("user id in context = %v, %v", got, ok)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	for header, want := range map[string]int{
		"":                 http.StatusUnauthorized,
		"Bearer not-a-jwt": http.StatusUnauthorized,
		"Basic " + valid:   http.StatusUnauthorized,
		"Bearer " + valid:  http.StatusNoContent,
	} {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
		if header != "" {
			req.Header.Set("Authorization", header)
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != want {
			t.Errorf("Authorization %q: status %d, want %d", header, rec.Code, want)
		}
	}
}
