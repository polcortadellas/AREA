package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/polcortadellas/AREA/server/internal/api"
	"github.com/polcortadellas/AREA/server/internal/connectors"
	"github.com/polcortadellas/AREA/server/internal/node"
)

func TestCreateWorkflow(t *testing.T) {
	reg := node.NewRegistry()
	connectors.Register(reg)
	h := api.CreateWorkflow(reg)

	cases := map[string]struct {
		body string
		want int
	}{
		"valid": {`{"nodes":[{"id":"t","kind":"trigger","type":"dummy.counter"},
			{"id":"a","kind":"reaction","type":"dummy.echo","config":{"message":"{{ t.count }}"}}],
			"edges":[{"from":"t","to":"a"}]}`, http.StatusOK},
		"undeclared key": {`{"nodes":[{"id":"t","kind":"trigger","type":"dummy.counter"},
			{"id":"a","kind":"reaction","type":"dummy.echo","config":{"message":"{{ t.nope }}"}}],
			"edges":[{"from":"t","to":"a"}]}`, http.StatusUnprocessableEntity},
		"broken JSON": {`{"nodes":`, http.StatusBadRequest},
		"too big":     {`{"nodes":[` + strings.Repeat(" ", 2<<20) + `]}`, http.StatusBadRequest},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			h(rec, httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/workflows", strings.NewReader(tc.body)))
			if rec.Code != tc.want {
				t.Fatalf("got %d (%s), want %d", rec.Code, rec.Body, tc.want)
			}
		})
	}
}
