package auth

import (
	"encoding/json"
	"errors"
	"log/slog"
	"mime"
	"net/http"
	"time"

	"github.com/polcortadellas/AREA/server/internal/httpjson"
)

const maxBodyBytes = 1 << 16

func (s *Service) Routes(mux *http.ServeMux) {
	mux.HandleFunc("POST /auth/register", s.handleRegister)
	mux.HandleFunc("POST /auth/verify-email", s.handleVerifyEmail)
	mux.HandleFunc("POST /auth/login", s.handleLogin)
	mux.HandleFunc("POST /auth/logout", s.handleLogout)
	mux.Handle("GET /auth/me", s.tokens.RequireAuth(http.HandlerFunc(s.handleMe)))
}

type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

func (s *Service) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req credentials
	if !decode(w, r, &req) {
		return
	}
	switch err := s.Register(r.Context(), req.Email, req.Password); {
	case errors.Is(err, ErrInvalidEmail):
		httpjson.Error(w, http.StatusBadRequest, "invalid_email")
	case errors.Is(err, ErrWeakPassword):
		httpjson.Error(w, http.StatusBadRequest, "weak_password")
	case err != nil:
		internalError(w, "register", err)
	default:
		httpjson.Write(w, http.StatusAccepted, map[string]string{"status": "verification_email_sent"})
	}
}

func (s *Service) handleVerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
	}
	if !decode(w, r, &req) {
		return
	}
	token, ttl, err := s.VerifyEmail(r.Context(), req.Token)
	switch {
	case errors.Is(err, ErrInvalidVerifyToken):
		httpjson.Error(w, http.StatusBadRequest, "invalid_token")
	case err != nil:
		internalError(w, "verify email", err)
	default:
		s.writeToken(w, token, ttl)
	}
}

func (s *Service) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req credentials
	if !decode(w, r, &req) {
		return
	}
	token, ttl, err := s.Login(r.Context(), req.Email, req.Password)
	switch {
	case errors.Is(err, ErrInvalidCredentials):
		httpjson.Error(w, http.StatusUnauthorized, "invalid_credentials")
	case errors.Is(err, ErrEmailNotVerified):
		httpjson.Error(w, http.StatusForbidden, "email_not_verified")
	case err != nil:
		internalError(w, "login", err)
	default:
		s.writeToken(w, token, ttl)
	}
}

func (s *Service) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get(CSRFHeader) == "" {
		httpjson.Error(w, http.StatusForbidden, "csrf_header_required")
		return
	}
	s.setSessionCookie(w, "", -1)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Service) handleMe(w http.ResponseWriter, r *http.Request) {
	id, _ := UserID(r.Context())
	user, err := s.User(r.Context(), id)
	if errors.Is(err, ErrUserNotFound) {
		unauthorized(w)
		return
	}
	if err != nil {
		internalError(w, "me", err)
		return
	}
	httpjson.Write(w, http.StatusOK, map[string]any{"id": user.ID, "email": user.Email})
}

// decode only accepts application/json: a cross-site form can post text/plain
// without a CORS preflight, which would allow login CSRF.
func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	if mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mt != "application/json" {
		httpjson.Error(w, http.StatusBadRequest, "invalid_body")
		return false
	}
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid_body")
		return false
	}
	return true
}

// writeToken returns the token in the body for the mobile client and sets it
// as an HttpOnly cookie for the web client.
func (s *Service) writeToken(w http.ResponseWriter, token string, ttl time.Duration) {
	s.setSessionCookie(w, token, int(ttl.Seconds()))
	w.Header().Set("Cache-Control", "no-store")
	httpjson.Write(w, http.StatusOK, tokenResponse{AccessToken: token, TokenType: "Bearer", ExpiresIn: int(ttl.Seconds())})
}

// setSessionCookie with a negative maxAge deletes the cookie.
func (s *Service) setSessionCookie(w http.ResponseWriter, token string, maxAge int) {
	// Secure follows the web client's scheme: plain-HTTP dev and LAN setups would lose the cookie.
	http.SetCookie(w, &http.Cookie{ //nolint:gosec
		Name:     SessionCookie,
		Value:    token,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   s.secureCookie,
		SameSite: http.SameSiteStrictMode,
	})
}

func internalError(w http.ResponseWriter, op string, err error) {
	slog.Error("auth: "+op, "err", err)
	httpjson.Error(w, http.StatusInternalServerError, "internal_error")
}
