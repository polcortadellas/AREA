package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/polcortadellas/AREA/server/internal/api"
)

func TestCORS(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	h := api.CORS([]string{" http://localhost:8081/", "http://localhost:5173"}, next)

	tests := []struct {
		name, method, origin, preflight string
		wantCode                        int
		wantOrigin                      string
	}{
		{"allowed request", http.MethodPost, "http://localhost:8081", "", http.StatusTeapot, "http://localhost:8081"},
		{"second origin", http.MethodGet, "http://localhost:5173", "", http.StatusTeapot, "http://localhost:5173"},
		{"allowed preflight", http.MethodOptions, "http://localhost:8081", "POST", http.StatusNoContent, "http://localhost:8081"},
		{"other origin", http.MethodPost, "https://evil.example", "", http.StatusTeapot, ""},
		{"other origin preflight", http.MethodOptions, "https://evil.example", "POST", http.StatusTeapot, ""},
		{"no origin", http.MethodGet, "", "", http.StatusTeapot, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequestWithContext(t.Context(), tc.method, "/auth/login", nil)
			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}
			if tc.preflight != "" {
				req.Header.Set("Access-Control-Request-Method", tc.preflight)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if rec.Code != tc.wantCode {
				t.Errorf("status %d, want %d", rec.Code, tc.wantCode)
			}
			if got := rec.Header().Get("Access-Control-Allow-Origin"); got != tc.wantOrigin {
				t.Errorf("Allow-Origin %q, want %q", got, tc.wantOrigin)
			}
			if rec.Header().Get("Vary") != "Origin" {
				t.Error("missing Vary: Origin")
			}
			wantCreds := ""
			if tc.wantOrigin != "" {
				wantCreds = "true"
			}
			if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != wantCreds {
				t.Errorf("Allow-Credentials %q, want %q", got, wantCreds)
			}
			if tc.wantCode == http.StatusNoContent && rec.Header().Get("Access-Control-Allow-Headers") != "Authorization, Content-Type, X-Requested-With" {
				t.Errorf("Allow-Headers %q", rec.Header().Get("Access-Control-Allow-Headers"))
			}
		})
	}
}
