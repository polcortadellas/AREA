package oauth

import (
	"net/url"
	"testing"
)

func TestProvidersFromEnvOnlyIncludesConfiguredProviders(t *testing.T) {
	env := map[string]string{
		"GITHUB_CLIENT_ID":     "gh-id",
		"GITHUB_CLIENT_SECRET": "gh-secret",
		"GOOGLE_CLIENT_ID":     "google-id",
	}

	got := ProvidersFromEnv(func(k string) string { return env[k] }, "http://localhost:8080/")

	if len(got) != 1 {
		t.Fatalf("got %d providers, want only github", len(got))
	}
	gh := got["github"].Config
	if gh.ClientID != "gh-id" || gh.ClientSecret != "gh-secret" || gh.RedirectURL != "http://localhost:8080/services/github/callback" {
		t.Errorf("github config = %+v", gh)
	}
}

func TestGoogleAsksForARefreshToken(t *testing.T) {
	env := map[string]string{"GOOGLE_CLIENT_ID": "id", "GOOGLE_CLIENT_SECRET": "secret"}
	google := ProvidersFromEnv(func(k string) string { return env[k] }, "http://localhost:8080")["google"]

	u, err := url.Parse(google.Config.AuthCodeURL("s", google.AuthParams...))

	if err != nil || u.Query().Get("access_type") != "offline" || u.Query().Get("prompt") != "consent" {
		t.Errorf("consent URL %v lacks access_type=offline and prompt=consent", u)
	}
}

func TestNormalizeScopes(t *testing.T) {
	tests := []struct{ in, fallback, want string }{
		{"repo,read:user", "", "repo read:user"},
		{" openid  email ", "", "openid email"},
		{"", "repo read:user", "repo read:user"},
	}
	for _, tt := range tests {
		if got := normalizeScopes(tt.in, tt.fallback); got != tt.want {
			t.Errorf("normalizeScopes(%q, %q) = %q, want %q", tt.in, tt.fallback, got, tt.want)
		}
	}
}
