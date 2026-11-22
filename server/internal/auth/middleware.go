package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

const (
	// SessionCookie holds the access token for the web client, out of reach of scripts.
	SessionCookie = "area_session"
	// CSRFHeader must accompany cookie-authenticated requests that change state:
	// cross-site forms cannot send it, and cross-origin scripts need a CORS preflight to.
	CSRFHeader = "X-Requested-With"
)

type ctxKey struct{}

// RequireAuth accepts "Authorization: Bearer <jwt>" (mobile) or the session
// cookie (web), and stores the user ID in the request context.
func (t *Tokens) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, fromCookie := credential(r)
		id, err := t.Parse(raw)
		if err != nil {
			unauthorized(w)
			return
		}
		if fromCookie && !isSafeMethod(r.Method) && r.Header.Get(CSRFHeader) == "" {
			writeError(w, http.StatusForbidden, "csrf_header_required")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, id)))
	})
}

func UserID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(ctxKey{}).(uuid.UUID)
	return id, ok
}

func credential(r *http.Request) (token string, fromCookie bool) {
	if h := r.Header.Get("Authorization"); h != "" {
		raw, _ := strings.CutPrefix(h, "Bearer ")
		if raw == h {
			return "", false
		}
		return strings.TrimSpace(raw), false
	}
	if c, err := r.Cookie(SessionCookie); err == nil {
		return c.Value, true
	}
	return "", false
}

func isSafeMethod(m string) bool {
	return m == http.MethodGet || m == http.MethodHead || m == http.MethodOptions
}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Bearer realm="area"`)
	writeError(w, http.StatusUnauthorized, "unauthorized")
}
