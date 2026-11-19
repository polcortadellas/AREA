package auth

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

const maxBodyBytes = 1 << 16

func (s *Service) Routes(mux *http.ServeMux) {
	mux.HandleFunc("POST /auth/register", s.handleRegister)
	mux.HandleFunc("POST /auth/verify-email", s.handleVerifyEmail)
	mux.HandleFunc("POST /auth/login", s.handleLogin)
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
		writeError(w, http.StatusBadRequest, "invalid_email")
	case errors.Is(err, ErrWeakPassword):
		writeError(w, http.StatusBadRequest, "weak_password")
	case err != nil:
		internalError(w, "register", err)
	default:
		writeJSON(w, http.StatusAccepted, map[string]string{"status": "verification_email_sent"})
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
		writeError(w, http.StatusBadRequest, "invalid_token")
	case err != nil:
		internalError(w, "verify email", err)
	default:
		writeToken(w, token, ttl)
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
		writeError(w, http.StatusUnauthorized, "invalid_credentials")
	case errors.Is(err, ErrEmailNotVerified):
		writeError(w, http.StatusForbidden, "email_not_verified")
	case err != nil:
		internalError(w, "login", err)
	default:
		writeToken(w, token, ttl)
	}
}

func (s *Service) handleMe(w http.ResponseWriter, r *http.Request) {
	id, _ := UserID(r.Context())
	user, err := s.User(r.Context(), id)
	if err != nil {
		internalError(w, "me", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": user.ID, "email": user.Email})
}

func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body")
		return false
	}
	return true
}

func writeToken(w http.ResponseWriter, token string, ttl time.Duration) {
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, tokenResponse{AccessToken: token, TokenType: "Bearer", ExpiresIn: int(ttl.Seconds())})
}

func internalError(w http.ResponseWriter, op string, err error) {
	slog.Error("auth: "+op, "err", err)
	writeError(w, http.StatusInternalServerError, "internal_error")
}

func writeError(w http.ResponseWriter, status int, code string) {
	writeJSON(w, status, map[string]string{"error": code})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
