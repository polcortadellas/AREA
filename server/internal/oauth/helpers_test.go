package oauth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"

	"github.com/polcortadellas/AREA/server/internal/auth"
	"github.com/polcortadellas/AREA/server/internal/db/dbtest"
	"github.com/polcortadellas/AREA/server/internal/db/sqlc"
)

// fakeProvider is an OAuth2 token endpoint that checks PKCE and, like GitHub
// Apps, accepts each refresh token only once.
type fakeProvider struct {
	url string

	mu           sync.Mutex
	codes        map[string]string // authorization code -> PKCE challenge
	refresh      map[string]bool
	issued       int
	refreshCalls int
	unavailable  bool
}

func newFakeProvider(t *testing.T) *fakeProvider {
	t.Helper()
	f := &fakeProvider{codes: map[string]string{}, refresh: map[string]bool{}}
	srv := httptest.NewServer(http.HandlerFunc(f.token))
	t.Cleanup(srv.Close)
	f.url = srv.URL
	return f
}

func (f *fakeProvider) endpoint() oauth2.Endpoint {
	return oauth2.Endpoint{AuthURL: f.url + "/authorize", TokenURL: f.url + "/token", AuthStyle: oauth2.AuthStyleInHeader}
}

// allowCode makes code valid for the consent URL the user was sent to.
func (f *fakeProvider) allowCode(code string, consent *url.URL) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.codes[code] = consent.Query().Get("code_challenge")
}

func (f *fakeProvider) allowRefresh(token string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.refresh[token] = true
}

func (f *fakeProvider) refreshes() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.refreshCalls
}

func (f *fakeProvider) token(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()

	switch r.PostForm.Get("grant_type") {
	case "authorization_code":
		challenge, ok := f.codes[r.PostForm.Get("code")]
		sum := sha256.Sum256([]byte(r.PostForm.Get("code_verifier")))
		if !ok || base64.RawURLEncoding.EncodeToString(sum[:]) != challenge {
			oauthError(w, "invalid_grant")
			return
		}
		delete(f.codes, r.PostForm.Get("code"))
	case "refresh_token":
		if f.unavailable {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		rt := r.PostForm.Get("refresh_token")
		if !f.refresh[rt] {
			oauthError(w, "invalid_grant")
			return
		}
		delete(f.refresh, rt)
		f.refreshCalls++
	default:
		oauthError(w, "unsupported_grant_type")
		return
	}

	f.issued++
	rt := fmt.Sprintf("rt-%d", f.issued)
	f.refresh[rt] = true
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"access_token":  fmt.Sprintf("at-%d", f.issued),
		"token_type":    "bearer",
		"expires_in":    3600,
		"refresh_token": rt,
		"scope":         "repo,read:user",
	})
}

func oauthError(w http.ResponseWriter, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": code})
}

type testEnv struct {
	svc     *Service
	mux     *http.ServeMux
	pool    *pgxpool.Pool
	rdb     *redis.Client
	fake    *fakeProvider
	user    uuid.UUID
	session string
	tokens  *auth.Tokens
}

func testRedis(t *testing.T) *redis.Client {
	t.Helper()
	addr := os.Getenv("REDIS_URL")
	if addr == "" {
		t.Skip("REDIS_URL not set (use a throwaway Redis)")
	}
	opt, err := redis.ParseURL(addr)
	if err != nil {
		t.Fatal(err)
	}
	c := redis.NewClient(opt)
	t.Cleanup(func() { _ = c.Close() })
	return c
}

func setup(t *testing.T) *testEnv {
	t.Helper()
	pool := dbtest.Pool(t)
	rdb := testRedis(t)
	fake := newFakeProvider(t)
	providers := map[string]Provider{}
	for _, name := range []string{"github", "google"} {
		providers[name] = Provider{Config: oauth2.Config{
			ClientID:     "client-" + name,
			ClientSecret: "secret-" + name,
			Endpoint:     fake.endpoint(),
			RedirectURL:  "http://api.test/services/" + name + "/callback",
			Scopes:       []string{"repo", "read:user"},
		}}
	}
	tokens, err := auth.NewTokens(strings.Repeat("k", 32), time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	svc := New(pool, rdb, providers, "http://web.test", "http://api.test")
	mux := http.NewServeMux()
	svc.Routes(mux, tokens)

	e := &testEnv{svc: svc, mux: mux, pool: pool, rdb: rdb, fake: fake, tokens: tokens}
	e.user, e.session = e.newUser(t, "owner@example.com")
	return e
}

func (e *testEnv) newUser(t *testing.T, email string) (uuid.UUID, string) {
	t.Helper()
	hash := "not-a-real-hash"
	u, err := sqlc.New(e.pool).CreateUser(context.Background(), sqlc.CreateUserParams{Email: email, PasswordHash: &hash})
	if err != nil {
		t.Fatal(err)
	}
	session, _, err := e.tokens.Issue(u.ID)
	if err != nil {
		t.Fatal(err)
	}
	return u.ID, session
}

// do sends a request as the web client: session cookie plus the CSRF header.
func (e *testEnv) do(t *testing.T, method, target, session string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequestWithContext(t.Context(), method, target, nil)
	if session != "" {
		req.AddCookie(&http.Cookie{Name: auth.SessionCookie, Value: session})
		req.Header.Set(auth.CSRFHeader, "XMLHttpRequest")
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	e.mux.ServeHTTP(rec, req)
	return rec
}

// authorize starts linking and returns the consent URL and the binding cookie.
func (e *testEnv) authorize(t *testing.T, service string) (*url.URL, *http.Cookie) {
	t.Helper()
	rec := e.do(t, http.MethodPost, "/services/"+service+"/authorize", e.session)
	if rec.Code != http.StatusOK {
		t.Fatalf("authorize: status %d, body %s", rec.Code, rec.Body)
	}
	var body struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	consent, err := url.Parse(body.URL)
	if err != nil {
		t.Fatal(err)
	}
	return consent, findCookie(rec.Result().Cookies(), bindingCookie)
}

func (e *testEnv) callback(t *testing.T, service string, q url.Values, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	return e.do(t, http.MethodGet, "/services/"+service+"/callback?"+q.Encode(), "", cookies...)
}

// link runs the whole consent flow for service.
func (e *testEnv) link(t *testing.T, service string) {
	t.Helper()
	consent, binding := e.authorize(t, service)
	e.fake.allowCode("good-code", consent)
	rec := e.callback(t, service, url.Values{"state": {consent.Query().Get("state")}, "code": {"good-code"}}, binding)
	if rec.Code != http.StatusFound {
		t.Fatalf("callback: status %d, body %s", rec.Code, rec.Body)
	}
}

func (e *testEnv) row(t *testing.T, user uuid.UUID, service string) sqlc.UserService {
	t.Helper()
	var r sqlc.UserService
	err := e.pool.QueryRow(context.Background(),
		`SELECT id, user_id, service, access_token_enc, refresh_token_enc, expires_at, scopes, created_at, updated_at
		 FROM user_services WHERE user_id = $1 AND service = $2`, user, service,
	).Scan(&r.ID, &r.UserID, &r.Service, &r.AccessTokenEnc, &r.RefreshTokenEnc, &r.ExpiresAt, &r.Scopes, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// seed stores tokens directly, as if the service had been linked earlier.
func (e *testEnv) seed(t *testing.T, user uuid.UUID, service, access, refresh string, expiresAt *time.Time) {
	t.Helper()
	var rt []byte
	if refresh != "" {
		rt = []byte(refresh)
	}
	err := sqlc.New(e.pool).UpsertUserService(context.Background(), sqlc.UpsertUserServiceParams{
		UserID: user, Service: service, AccessTokenEnc: []byte(access), RefreshTokenEnc: rt, ExpiresAt: expiresAt, Scopes: "repo",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func findCookie(cookies []*http.Cookie, name string) *http.Cookie {
	for _, c := range cookies {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func at(d time.Duration) *time.Time {
	v := time.Now().Add(d)
	return &v
}
