// Package oauth links users' third-party accounts with OAuth2 (authorization
// code flow with PKCE), stores the tokens and refreshes them before they expire.
package oauth

import (
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
)

type Provider struct {
	Config oauth2.Config
	// AuthParams are added to the consent URL.
	AuthParams []oauth2.AuthCodeOption
}

type providerSpec struct {
	name     string
	envKey   string
	endpoint oauth2.Endpoint
	scopes   []string
	params   []oauth2.AuthCodeOption
}

var providerSpecs = []providerSpec{
	{name: "github", envKey: "GITHUB", endpoint: endpoints.GitHub, scopes: []string{"repo", "read:user"}},
	// Google only returns a refresh token for offline access, and only on the
	// first consent unless asked to show the consent screen again.
	{
		name: "google", envKey: "GOOGLE", endpoint: endpoints.Google, scopes: []string{"openid", "email"},
		params: []oauth2.AuthCodeOption{oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("prompt", "consent")},
	},
}

// ProvidersFromEnv returns the providers whose <NAME>_CLIENT_ID and
// <NAME>_CLIENT_SECRET are both set. apiBaseURL is the API as seen from the
// browser; the provider redirects there after consent.
func ProvidersFromEnv(getenv func(string) string, apiBaseURL string) map[string]Provider {
	base := strings.TrimRight(apiBaseURL, "/")
	providers := make(map[string]Provider, len(providerSpecs))
	for _, spec := range providerSpecs {
		id, secret := getenv(spec.envKey+"_CLIENT_ID"), getenv(spec.envKey+"_CLIENT_SECRET")
		if id == "" || secret == "" {
			continue
		}
		providers[spec.name] = Provider{
			Config: oauth2.Config{
				ClientID:     id,
				ClientSecret: secret,
				Endpoint:     spec.endpoint,
				RedirectURL:  base + callbackPath(spec.name),
				Scopes:       spec.scopes,
			},
			AuthParams: spec.params,
		}
	}
	return providers
}

func callbackPath(service string) string {
	return "/services/" + service + "/callback"
}
