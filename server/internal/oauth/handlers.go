package oauth

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"

	"github.com/polcortadellas/AREA/server/internal/auth"
	"github.com/polcortadellas/AREA/server/internal/db/sqlc"
	"github.com/polcortadellas/AREA/server/internal/httpjson"
)

const providerTimeout = 10 * time.Second

type Service struct {
	pool       *pgxpool.Pool
	q          *sqlc.Queries
	states     *redis.Client
	providers  map[string]Provider
	webBaseURL string
	// The binding cookie is Secure whenever the API is served over HTTPS.
	secureCookie bool
	client       *http.Client
}

// New takes the web client's URL (where the callback sends the user back) and
// the API's URL as seen from the browser.
func New(pool *pgxpool.Pool, states *redis.Client, providers map[string]Provider, webBaseURL, apiBaseURL string) *Service {
	return &Service{
		pool:         pool,
		q:            sqlc.New(pool),
		states:       states,
		providers:    providers,
		webBaseURL:   strings.TrimRight(webBaseURL, "/"),
		secureCookie: strings.HasPrefix(apiBaseURL, "https://"),
		client:       &http.Client{Timeout: providerTimeout},
	}
}

func (s *Service) Routes(mux *http.ServeMux, tokens *auth.Tokens) {
	mux.Handle("GET /services", tokens.RequireAuth(http.HandlerFunc(s.handleList)))
	mux.Handle("POST /services/{service}/authorize", tokens.RequireAuth(http.HandlerFunc(s.handleAuthorize)))
	mux.HandleFunc("GET /services/{service}/callback", s.handleCallback)
	mux.Handle("DELETE /services/{service}", tokens.RequireAuth(http.HandlerFunc(s.handleUnlink)))
}

// Configured lists the providers that have credentials, for the startup log.
func (s *Service) Configured() []string {
	names := make([]string, 0, len(s.providers))
	for name := range s.providers {
		names = append(names, name)
	}
	return names
}

// providerCtx makes x/oauth2 call the provider with a timeout.
func (s *Service) providerCtx(ctx context.Context) context.Context {
	return context.WithValue(ctx, oauth2.HTTPClient, s.client)
}

type linkedService struct {
	Service  string    `json:"service"`
	Scopes   string    `json:"scopes"`
	LinkedAt time.Time `json:"linked_at"`
}

func (s *Service) handleList(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r.Context())
	rows, err := s.q.ListUserServices(r.Context(), userID)
	if err != nil {
		internalError(w, "list services", err)
		return
	}
	out := make([]linkedService, 0, len(rows))
	for _, row := range rows {
		out = append(out, linkedService{Service: row.Service, Scopes: row.Scopes, LinkedAt: row.CreatedAt})
	}
	httpjson.Write(w, http.StatusOK, out)
}

// handleAuthorize remembers who is linking which service, with the PKCE
// verifier, under a random state, and returns the provider's consent URL.
func (s *Service) handleAuthorize(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("service")
	p, ok := s.providers[name]
	if !ok {
		httpjson.Error(w, http.StatusNotFound, "not_found")
		return
	}
	userID, _ := auth.UserID(r.Context())
	state, binding, verifier := randomToken(), randomToken(), oauth2.GenerateVerifier()
	pending := pendingLink{UserID: userID, Service: name, Verifier: verifier, BindingHash: sha256Hex(binding)}
	if err := s.savePending(r.Context(), state, pending); err != nil {
		internalError(w, "authorize", err)
		return
	}
	s.setBinding(w, name, binding, int(stateTTL.Seconds()))

	opts := append([]oauth2.AuthCodeOption{oauth2.S256ChallengeOption(verifier)}, p.AuthParams...)
	w.Header().Set("Cache-Control", "no-store")
	httpjson.Write(w, http.StatusOK, map[string]string{"url": p.Config.AuthCodeURL(state, opts...)})
}

// handleCallback is reached by the browser coming back from the provider. The
// session cookie is SameSite=Strict, so it is not sent here: the state says
// who started the link, and the binding cookie proves it is the same browser.
// The cookie is cleared only once both match, so another site sending the
// browser here with a made-up state cannot cancel a link in progress.
func (s *Service) handleCallback(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("service")
	q := r.URL.Query()
	p, known := s.providers[name]
	if !known {
		httpjson.Error(w, http.StatusBadRequest, "invalid_state")
		return
	}
	pending, err := s.takePending(r.Context(), q.Get("state"))
	if err != nil && !errors.Is(err, errInvalidState) {
		internalError(w, "callback", err)
		return
	}
	if err != nil || pending.Service != name || !bindingMatches(r, pending.BindingHash) {
		httpjson.Error(w, http.StatusBadRequest, "invalid_state")
		return
	}
	s.setBinding(w, name, "", -1)

	if e := q.Get("error"); e != "" {
		s.backToWeb(w, r, "error", callbackError(e))
		return
	}
	code := q.Get("code")
	if code == "" {
		s.backToWeb(w, r, "error", "link_failed")
		return
	}
	tok, err := p.Config.Exchange(s.providerCtx(r.Context()), code, oauth2.VerifierOption(pending.Verifier))
	if err != nil {
		slog.Warn("oauth: code exchange failed", "service", name, "user", pending.UserID, "err", describe(err))
		s.backToWeb(w, r, "error", "link_failed")
		return
	}
	if err := s.saveLink(r.Context(), pending.UserID, name, tok, strings.Join(p.Config.Scopes, " ")); err != nil {
		slog.Error("oauth: save linked service", "service", name, "user", pending.UserID, "err", err)
		s.backToWeb(w, r, "error", "link_failed")
		return
	}
	s.backToWeb(w, r, "linked", name)
}

// callbackError keeps the provider's own error codes out of the web client,
// except a refusal, which the user should be told about.
func callbackError(providerError string) string {
	if providerError == "access_denied" {
		return "access_denied"
	}
	return "link_failed"
}

func (s *Service) backToWeb(w http.ResponseWriter, r *http.Request, key, value string) {
	w.Header().Set("Cache-Control", "no-store")
	http.Redirect(w, r, s.webBaseURL+"/services?"+url.Values{key: {value}}.Encode(), http.StatusFound)
}

func (s *Service) handleUnlink(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r.Context())
	n, err := s.q.DeleteUserService(r.Context(), sqlc.DeleteUserServiceParams{UserID: userID, Service: r.PathValue("service")})
	if err != nil {
		internalError(w, "unlink", err)
		return
	}
	if n == 0 {
		httpjson.Error(w, http.StatusNotFound, "not_found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func internalError(w http.ResponseWriter, op string, err error) {
	slog.Error("oauth: "+op, "err", err)
	httpjson.Error(w, http.StatusInternalServerError, "internal_error")
}
