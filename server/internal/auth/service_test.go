package auth

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/polcortadellas/AREA/server/internal/db"
)

// testPool creates a throwaway database, so these tests never touch tables
// that other packages' tests (e.g. the migration test) wipe in parallel.
func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	base := os.Getenv("DATABASE_URL")
	if base == "" {
		t.Skip("DATABASE_URL not set (use a throwaway PostgreSQL)")
	}
	ctx := context.Background()
	admin, err := pgx.Connect(ctx, base)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close(ctx)

	b := make([]byte, 6)
	_, _ = rand.Read(b)
	name := "area_auth_test_" + hex.EncodeToString(b)
	if _, err := admin.Exec(ctx, "CREATE DATABASE "+name); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		conn, err := pgx.Connect(context.Background(), base)
		if err == nil {
			_, _ = conn.Exec(context.Background(), "DROP DATABASE "+name+" WITH (FORCE)")
			_ = conn.Close(context.Background())
		}
	})

	u, _ := url.Parse(base)
	u.Path = "/" + name
	if err := db.Migrate(u.String()); err != nil {
		t.Fatal(err)
	}
	pool, err := pgxpool.New(ctx, u.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	return pool
}

type fakeMailer struct {
	mu   sync.Mutex
	sent []string
}

func (f *fakeMailer) Send(_ context.Context, to, _, body string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sent = append(f.sent, to+"\n"+body)
	return nil
}

var tokenInLink = regexp.MustCompile(`/verify-email\?token=([A-Za-z0-9_-]+)`)

func (f *fakeMailer) lastToken(t *testing.T) string {
	t.Helper()
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.sent) == 0 {
		t.Fatal("no e-mail sent")
	}
	m := tokenInLink.FindStringSubmatch(f.sent[len(f.sent)-1])
	if m == nil {
		t.Fatalf("no verification link in e-mail:\n%s", f.sent[len(f.sent)-1])
	}
	return m[1]
}

type testServer struct {
	t    *testing.T
	mux  *http.ServeMux
	mail *fakeMailer
}

func newTestServer(t *testing.T) *testServer {
	mail := &fakeMailer{}
	svc, err := NewService(testPool(t), newTestTokens(t), mail, "http://web.test")
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	svc.Routes(mux)
	return &testServer{t: t, mux: mux, mail: mail}
}

func (s *testServer) do(method, path, bearer string, body any) (int, map[string]any) {
	s.t.Helper()
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequestWithContext(s.t.Context(), method, path, &buf)
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	var out map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	return rec.Code, out
}

// M-08 "done when": an unconfirmed account cannot log in; the e-mailed link
// activates it and returns a JWT that the auth middleware accepts.
func TestRegisterVerifyLogin(t *testing.T) {
	s := newTestServer(t)
	creds := map[string]string{"email": "Alice@Example.com", "password": "s3cret-password"}

	if code, _ := s.do("POST", "/auth/register", "", creds); code != http.StatusAccepted {
		t.Fatalf("register: %d", code)
	}
	if code, body := s.do("POST", "/auth/login", "", creds); code != http.StatusForbidden || body["error"] != "email_not_verified" {
		t.Fatalf("login before verification: %d %v", code, body)
	}

	token := s.mail.lastToken(t)
	code, body := s.do("POST", "/auth/verify-email", "", map[string]string{"token": token})
	if code != http.StatusOK {
		t.Fatalf("verify: %d %v", code, body)
	}
	jwt, _ := body["access_token"].(string)
	if code, body := s.do("GET", "/auth/me", jwt, nil); code != http.StatusOK || body["email"] != "alice@example.com" {
		t.Fatalf("me with verification JWT: %d %v", code, body)
	}

	if code, _ := s.do("POST", "/auth/verify-email", "", map[string]string{"token": token}); code != http.StatusBadRequest {
		t.Fatalf("reused verification token: %d", code)
	}
	if code, _ := s.do("POST", "/auth/login", "", creds); code != http.StatusOK {
		t.Fatalf("login after verification: %d", code)
	}
	if code, _ := s.do("GET", "/auth/me", "", nil); code != http.StatusUnauthorized {
		t.Fatalf("me without token: %d", code)
	}
}

func TestLoginRejectsBadCredentials(t *testing.T) {
	s := newTestServer(t)
	s.do("POST", "/auth/register", "", map[string]string{"email": "bob@example.com", "password": "right-password"})
	s.do("POST", "/auth/verify-email", "", map[string]string{"token": s.mail.lastToken(t)})

	for _, c := range []map[string]string{
		{"email": "bob@example.com", "password": "wrong-password"},
		{"email": "nobody@example.com", "password": "right-password"},
	} {
		if code, body := s.do("POST", "/auth/login", "", c); code != http.StatusUnauthorized || body["error"] != "invalid_credentials" {
			t.Errorf("%v: %d %v", c, code, body)
		}
	}
}

func TestRegisterValidation(t *testing.T) {
	s := newTestServer(t)
	for _, c := range []struct {
		body map[string]string
		want string
	}{
		{map[string]string{"email": "not-an-email", "password": "long-enough"}, "invalid_email"},
		{map[string]string{"email": "c@example.com", "password": "short"}, "weak_password"},
	} {
		if code, body := s.do("POST", "/auth/register", "", c.body); code != http.StatusBadRequest || body["error"] != c.want {
			t.Errorf("%v: %d %v", c.body, code, body)
		}
	}
}

func TestRegisterExistingAccounts(t *testing.T) {
	s := newTestServer(t)
	first := map[string]string{"email": "dana@example.com", "password": "first-password"}
	s.do("POST", "/auth/register", "", first)
	oldToken := s.mail.lastToken(t)

	// Registering again before confirming replaces the password and the link.
	second := map[string]string{"email": "dana@example.com", "password": "second-password"}
	s.do("POST", "/auth/register", "", second)
	if code, _ := s.do("POST", "/auth/verify-email", "", map[string]string{"token": oldToken}); code != http.StatusBadRequest {
		t.Fatalf("old link still valid: %d", code)
	}
	s.do("POST", "/auth/verify-email", "", map[string]string{"token": s.mail.lastToken(t)})
	if code, _ := s.do("POST", "/auth/login", "", second); code != http.StatusOK {
		t.Fatalf("login with second password: %d", code)
	}

	// A verified address answers the same way but sends nothing.
	sent := len(s.mail.sent)
	if code, _ := s.do("POST", "/auth/register", "", second); code != http.StatusAccepted {
		t.Fatalf("register verified address: %d", code)
	}
	if len(s.mail.sent) != sent {
		t.Fatal("e-mail sent for an already verified address")
	}
}

func TestMeRejectsDeletedUser(t *testing.T) {
	s := newTestServer(t)
	token, _, err := newTestTokens(t).Issue(uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	if code, _ := s.do("GET", "/auth/me", token, nil); code != http.StatusUnauthorized {
		t.Fatalf("me for unknown user: status %d, want 401", code)
	}
}
