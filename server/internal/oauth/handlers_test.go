package oauth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestAuthorizeReturnsConsentURLWithStateAndPKCE(t *testing.T) {
	e := setup(t)

	rec := e.do(t, http.MethodPost, "/services/github/authorize", e.session)
	consent, binding := e.authorize(t, "github")

	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", got)
	}
	q := consent.Query()
	want := map[string]string{
		"client_id":             "client-github",
		"redirect_uri":          "http://api.test/services/github/callback",
		"response_type":         "code",
		"scope":                 "repo read:user",
		"code_challenge_method": "S256",
	}
	for k, v := range want {
		if q.Get(k) != v {
			t.Errorf("%s = %q, want %q", k, q.Get(k), v)
		}
	}
	if len(q.Get("state")) < 40 || q.Get("code_challenge") == "" {
		t.Errorf("state %q or code_challenge %q missing or too short", q.Get("state"), q.Get("code_challenge"))
	}
	if binding == nil {
		t.Fatal("no binding cookie")
	}
	if !binding.HttpOnly || binding.SameSite != http.SameSiteLaxMode || binding.Path != "/services/github/callback" {
		t.Errorf("binding cookie = %+v, want HttpOnly, SameSite=Lax, Path=/services/github/callback", binding)
	}
	if ttl := e.rdb.TTL(context.Background(), stateKey(q.Get("state"))).Val(); ttl <= 0 || ttl > stateTTL {
		t.Errorf("state TTL = %v, want within (0, %v]", ttl, stateTTL)
	}
}

func TestAuthorizeRejectsUnknownServiceAndAnonymousUser(t *testing.T) {
	e := setup(t)

	if rec := e.do(t, http.MethodPost, "/services/dropbox/authorize", e.session); rec.Code != http.StatusNotFound {
		t.Errorf("unknown service: status %d, want 404", rec.Code)
	}
	if rec := e.do(t, http.MethodPost, "/services/github/authorize", ""); rec.Code != http.StatusUnauthorized {
		t.Errorf("anonymous: status %d, want 401", rec.Code)
	}
}

func TestCallbackStoresTokensAndRedirectsToWebClient(t *testing.T) {
	e := setup(t)
	consent, binding := e.authorize(t, "github")
	e.fake.allowCode("good-code", consent)

	rec := e.callback(t, "github", url.Values{"state": {consent.Query().Get("state")}, "code": {"good-code"}}, binding)

	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "http://web.test/services?linked=github" {
		t.Fatalf("status %d, Location %q", rec.Code, rec.Header().Get("Location"))
	}
	if c := findCookie(rec.Result().Cookies(), bindingCookie); c == nil || c.MaxAge >= 0 {
		t.Errorf("binding cookie not cleared: %+v", c)
	}
	r := e.row(t, e.user, "github")
	if string(r.AccessTokenEnc) != "at-1" || string(r.RefreshTokenEnc) != "rt-1" || r.Scopes != "repo read:user" {
		t.Errorf("stored %q / %q / scopes %q", r.AccessTokenEnc, r.RefreshTokenEnc, r.Scopes)
	}
	if r.ExpiresAt == nil || time.Until(*r.ExpiresAt) < 59*time.Minute {
		t.Errorf("expires_at = %v, want about an hour from now", r.ExpiresAt)
	}
}

func TestCallbackRejectsBadState(t *testing.T) {
	e := setup(t)

	tests := []struct {
		name    string
		prepare func() (service string, q url.Values, cookies []*http.Cookie)
	}{
		{"missing state", func() (string, url.Values, []*http.Cookie) {
			_, binding := e.authorize(t, "github")
			return "github", url.Values{"code": {"good-code"}}, []*http.Cookie{binding}
		}},
		{"unknown state", func() (string, url.Values, []*http.Cookie) {
			_, binding := e.authorize(t, "github")
			return "github", url.Values{"state": {"forged"}, "code": {"good-code"}}, []*http.Cookie{binding}
		}},
		{"state of another service", func() (string, url.Values, []*http.Cookie) {
			consent, binding := e.authorize(t, "github")
			e.fake.allowCode("good-code", consent)
			return "google", url.Values{"state": {consent.Query().Get("state")}, "code": {"good-code"}}, []*http.Cookie{binding}
		}},
		// An attacker sends a victim the consent URL of the attacker's own linking
		// attempt: the victim's browser lacks the binding cookie.
		{"no binding cookie", func() (string, url.Values, []*http.Cookie) {
			consent, _ := e.authorize(t, "github")
			e.fake.allowCode("good-code", consent)
			return "github", url.Values{"state": {consent.Query().Get("state")}, "code": {"good-code"}}, nil
		}},
		{"binding cookie of another attempt", func() (string, url.Values, []*http.Cookie) {
			consent, _ := e.authorize(t, "github")
			_, other := e.authorize(t, "github")
			e.fake.allowCode("good-code", consent)
			return "github", url.Values{"state": {consent.Query().Get("state")}, "code": {"good-code"}}, []*http.Cookie{other}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, q, cookies := tt.prepare()

			rec := e.callback(t, service, q, cookies...)

			if rec.Code != http.StatusBadRequest || !hasError(rec.Body.Bytes(), "invalid_state") {
				t.Errorf("status %d, body %s; want 400 invalid_state", rec.Code, rec.Body)
			}
			assertNotLinked(t, e)
		})
	}
}

func TestCallbackStateIsSingleUse(t *testing.T) {
	e := setup(t)
	consent, binding := e.authorize(t, "github")
	e.fake.allowCode("good-code", consent)
	q := url.Values{"state": {consent.Query().Get("state")}, "code": {"good-code"}}
	e.callback(t, "github", q, binding)

	rec := e.callback(t, "github", q, binding)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("replayed callback: status %d, want 400", rec.Code)
	}
}

func TestCallbackRedirectsWithErrorWhenLinkingFails(t *testing.T) {
	tests := []struct {
		name  string
		query func(state string) url.Values
		want  string
	}{
		{"user refused", func(s string) url.Values { return url.Values{"state": {s}, "error": {"access_denied"}} }, "access_denied"},
		{"provider error", func(s string) url.Values { return url.Values{"state": {s}, "error": {"server_error"}} }, "link_failed"},
		{"no code", func(s string) url.Values { return url.Values{"state": {s}} }, "link_failed"},
		{"code rejected by provider", func(s string) url.Values { return url.Values{"state": {s}, "code": {"bad-code"}} }, "link_failed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := setup(t)
			consent, binding := e.authorize(t, "github")

			rec := e.callback(t, "github", tt.query(consent.Query().Get("state")), binding)

			if want := "http://web.test/services?error=" + tt.want; rec.Code != http.StatusFound || rec.Header().Get("Location") != want {
				t.Errorf("status %d, Location %q; want 302 to %q", rec.Code, rec.Header().Get("Location"), want)
			}
			assertNotLinked(t, e)
		})
	}
}

func TestListShowsOnlyTheUsersLinkedServices(t *testing.T) {
	e := setup(t)
	e.link(t, "github")
	other, _ := e.newUser(t, "other@example.com")
	e.seed(t, other, "google", "x", "", nil)

	rec := e.do(t, http.MethodGet, "/services", e.session)

	var got []struct {
		Service  string    `json:"service"`
		Scopes   string    `json:"scopes"`
		LinkedAt time.Time `json:"linked_at"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("status %d, body %s: %v", rec.Code, rec.Body, err)
	}
	if len(got) != 1 || got[0].Service != "github" || got[0].Scopes != "repo read:user" || got[0].LinkedAt.IsZero() {
		t.Errorf("got %+v, want only github", got)
	}
}

func TestListIsAnEmptyArrayWhenNothingIsLinked(t *testing.T) {
	e := setup(t)

	rec := e.do(t, http.MethodGet, "/services", e.session)

	if rec.Code != http.StatusOK || rec.Body.String() != "[]\n" {
		t.Errorf("status %d, body %q; want 200 []", rec.Code, rec.Body)
	}
}

func TestUnlinkDeletesTheTokens(t *testing.T) {
	e := setup(t)
	e.link(t, "github")

	first := e.do(t, http.MethodDelete, "/services/github", e.session)
	second := e.do(t, http.MethodDelete, "/services/github", e.session)

	if first.Code != http.StatusNoContent || second.Code != http.StatusNotFound {
		t.Errorf("statuses %d then %d, want 204 then 404", first.Code, second.Code)
	}
	assertNotLinked(t, e)
}

func hasError(body []byte, code string) bool {
	var v struct {
		Error string `json:"error"`
	}
	return json.Unmarshal(body, &v) == nil && v.Error == code
}

func assertNotLinked(t *testing.T, e *testEnv) {
	t.Helper()
	var n int
	if err := e.pool.QueryRow(context.Background(), `SELECT count(*) FROM user_services`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Errorf("%d linked services, want 0", n)
	}
}
