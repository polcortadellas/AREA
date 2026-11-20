package api_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	openapi "github.com/polcortadellas/AREA/server/api"
	"github.com/polcortadellas/AREA/server/internal/api"
)

func TestOpenAPISpec(t *testing.T) {
	rec := httptest.NewRecorder()
	api.OpenAPISpec()(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/openapi.yaml", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/yaml" {
		t.Errorf("Content-Type = %q", ct)
	}
	if !bytes.HasPrefix(openapi.Spec, []byte("openapi: 3.0")) {
		t.Error("embedded spec is not an OpenAPI 3.0 document")
	}
	if !bytes.Equal(rec.Body.Bytes(), openapi.Spec) {
		t.Error("body is not the embedded spec")
	}
}

func TestDocs(t *testing.T) {
	rec := httptest.NewRecorder()
	api.Docs()(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/docs", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("Content-Type = %q", ct)
	}
	if !strings.Contains(rec.Body.String(), `url: "/openapi.yaml"`) {
		t.Error("page does not load /openapi.yaml")
	}
}
