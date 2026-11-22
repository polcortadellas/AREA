package api_test

import (
	"encoding/json"
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

func TestCreateWorkflowReturnsProblemsPerNode(t *testing.T) {
	reg := node.NewRegistry()
	connectors.Register(reg)
	h := api.CreateWorkflow(reg)

	body := `{"nodes":[{"id":"t","kind":"trigger","type":"dummy.counter"},
		{"id":"a","kind":"reaction","type":"dummy.echo","config":{"message":"{{ t.nope }}"}}],
		"edges":[{"from":"t","to":"a"}]}`
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/workflows", strings.NewReader(body)))

	var got struct {
		Error    string `json:"error"`
		Problems []struct {
			Node    string `json:"node"`
			Message string `json:"message"`
		} `json:"problems"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("body is not JSON: %v (%s)", err, rec.Body)
	}
	if got.Error != "invalid_workflow" || len(got.Problems) != 1 {
		t.Fatalf("got %+v, want invalid_workflow with one problem", got)
	}
	if got.Problems[0].Node != "a" || !strings.Contains(got.Problems[0].Message, "does not declare output") {
		t.Fatalf("problem = %+v, want node a with an undeclared output", got.Problems[0])
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("content type %q, want application/json", ct)
	}
}
