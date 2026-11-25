package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/polcortadellas/AREA/server/internal/api"
	"github.com/polcortadellas/AREA/server/internal/auth"
	"github.com/polcortadellas/AREA/server/internal/connectors"
	"github.com/polcortadellas/AREA/server/internal/engine"
	"github.com/polcortadellas/AREA/server/internal/node"
)

func registry() *node.Registry {
	reg := node.NewRegistry()
	connectors.Register(reg)
	return reg
}

// Rejected bodies are answered before the database is touched, so no pool is needed.
func TestCreateWorkflowRejectsBadBodies(t *testing.T) {
	h := api.CreateWorkflow(nil, registry())

	cases := map[string]struct {
		body string
		want int
	}{
		"undeclared key": {`{"name":"x","nodes":[{"id":"t","kind":"trigger","type":"dummy.counter"},
			{"id":"a","kind":"reaction","type":"dummy.echo","config":{"message":"{{ t.nope }}"}}],
			"edges":[{"from":"t","to":"a"}]}`, http.StatusUnprocessableEntity},
		"missing name": {`{"nodes":[{"id":"t","kind":"trigger","type":"dummy.counter"}],"edges":[]}`, http.StatusBadRequest},
		"blank name":   {`{"name":"  ","nodes":[],"edges":[]}`, http.StatusBadRequest},
		"long name":    {`{"name":"` + strings.Repeat("a", 101) + `","nodes":[],"edges":[]}`, http.StatusBadRequest},
		"broken JSON":  {`{"nodes":`, http.StatusBadRequest},
		"too big":      {`{"nodes":[` + strings.Repeat(" ", 2<<20) + `]}`, http.StatusBadRequest},
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
	h := api.CreateWorkflow(nil, registry())

	body := `{"name":"x","nodes":[{"id":"t","kind":"trigger","type":"dummy.counter"},
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

// start -> cond (count > 3) ? true: big : false: small
const branched = `{"name":"Demo","nodes":[
	{"id":"start","kind":"trigger","type":"dummy.counter","config":{}},
	{"id":"cond","kind":"condition","type":"compare","config":{"left":"{{ trigger.count }}","op":"greater_than","right":"3"}},
	{"id":"big","kind":"reaction","type":"dummy.echo","config":{"message":"big {{ start.count }}"}},
	{"id":"small","kind":"reaction","type":"dummy.echo","config":{"message":"small"}}],
	"edges":[{"from":"start","to":"cond"},{"from":"cond","to":"big","branch":"true"},{"from":"cond","to":"small","branch":"false"}]}`

type workflowBody struct {
	ID      string              `json:"id"`
	Name    string              `json:"name"`
	Enabled bool                `json:"enabled"`
	Nodes   []node.WorkflowNode `json:"nodes"`
	Edges   []node.Edge         `json:"edges"`
}

type client struct {
	t    *testing.T
	mux  *http.ServeMux
	toks *auth.Tokens
	user uuid.UUID
}

func newClient(t *testing.T, pool *pgxpool.Pool, user uuid.UUID) *client {
	t.Helper()
	toks, err := auth.NewTokens(strings.Repeat("k", auth.MinSecretLen), time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	reg := registry()
	mux := http.NewServeMux()
	for pattern, h := range map[string]http.Handler{
		"GET /workflows":                   api.ListWorkflows(pool),
		"POST /workflows":                  api.CreateWorkflow(pool, reg),
		"GET /workflows/{workflowId}":      api.GetWorkflow(pool),
		"PUT /workflows/{workflowId}":      api.ReplaceWorkflow(pool, reg),
		"PATCH /workflows/{workflowId}":    api.UpdateWorkflow(pool),
		"DELETE /workflows/{workflowId}":   api.DeleteWorkflow(pool),
		"GET /workflows/{workflowId}/runs": api.ListRuns(pool),
		"GET /runs/{runId}":                api.GetRun(pool),
	} {
		mux.Handle(pattern, toks.RequireAuth(h))
	}
	return &client{t: t, mux: mux, toks: toks, user: user}
}

// do sends the request and decodes a JSON answer into out when out is not nil.
func (c *client) do(method, path, body string, out any) int {
	c.t.Helper()
	req := httptest.NewRequestWithContext(c.t.Context(), method, path, strings.NewReader(body))
	token, _, err := c.toks.Issue(c.user)
	if err != nil {
		c.t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	c.mux.ServeHTTP(rec, req)
	if out != nil && rec.Code < 300 {
		if err := json.Unmarshal(rec.Body.Bytes(), out); err != nil {
			c.t.Fatalf("%s %s: body is not JSON: %v (%s)", method, path, err, rec.Body)
		}
	}
	return rec.Code
}

func newUser(t *testing.T, pool *pgxpool.Pool) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	if err := pool.QueryRow(t.Context(),
		`INSERT INTO users (email, password_hash) VALUES ($1, 'x') RETURNING id`, uuid.NewString()+"@example.com",
	).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}

func TestWorkflowLifecycle(t *testing.T) {
	pool := triggerPool(t)
	c := newClient(t, pool, newUser(t, pool))

	var created workflowBody
	if code := c.do(http.MethodPost, "/workflows", branched, &created); code != http.StatusCreated {
		t.Fatalf("create: %d, want 201", code)
	}
	if created.Name != "Demo" || !created.Enabled || len(created.Nodes) != 4 || len(created.Edges) != 3 {
		t.Fatalf("created %+v, want the 4 nodes and 3 edges sent", created)
	}
	var branches int
	for _, e := range created.Edges {
		if e.From == "cond" && (e.Branch == "true" || e.Branch == "false") {
			branches++
		}
	}
	if branches != 2 {
		t.Fatalf("edges %+v, want both branches of cond with their keys", created.Edges)
	}

	path := "/workflows/" + created.ID
	var got workflowBody
	if code := c.do(http.MethodGet, path, "", &got); code != http.StatusOK || got.ID != created.ID || len(got.Nodes) != 4 {
		t.Fatalf("get: %d %+v", code, got)
	}

	var list []workflowBody
	if code := c.do(http.MethodGet, "/workflows", "", &list); code != http.StatusOK || len(list) != 1 || list[0].ID != created.ID {
		t.Fatalf("list: %d %+v, want the created workflow", code, list)
	}

	var bigRow uuid.UUID
	if err := pool.QueryRow(t.Context(),
		`SELECT id FROM nodes WHERE workflow_id = $1 AND key = 'big'`, created.ID).Scan(&bigRow); err != nil {
		t.Fatal(err)
	}
	replaced := strings.Replace(branched, `"message":"small"`, `"message":"tiny"`, 1)
	replaced = strings.Replace(replaced, `"name":"Demo"`, `"name":"Demo 2"`, 1)
	var after workflowBody
	if code := c.do(http.MethodPut, path, replaced, &after); code != http.StatusOK || after.Name != "Demo 2" {
		t.Fatalf("replace: %d %+v", code, after)
	}
	for _, n := range after.Nodes {
		if n.ID == "small" && n.Config["message"] != "tiny" {
			t.Fatalf("small config %v, want the replaced message", n.Config)
		}
	}
	var bigAfter uuid.UUID
	if err := pool.QueryRow(t.Context(),
		`SELECT id FROM nodes WHERE workflow_id = $1 AND key = 'big'`, created.ID).Scan(&bigAfter); err != nil {
		t.Fatal(err)
	}
	if bigAfter != bigRow {
		t.Fatal("replace recreated an unchanged node, which would drop its run history")
	}

	var paused workflowBody
	if code := c.do(http.MethodPatch, path, `{"enabled":false}`, &paused); code != http.StatusOK || paused.Enabled || paused.Name != "Demo 2" {
		t.Fatalf("pause: %d %+v", code, paused)
	}

	if code := c.do(http.MethodDelete, path, "", nil); code != http.StatusNoContent {
		t.Fatalf("delete: %d, want 204", code)
	}
	if code := c.do(http.MethodGet, path, "", nil); code != http.StatusNotFound {
		t.Fatalf("get after delete: %d, want 404", code)
	}
}

func TestWorkflowsOfAnotherUserAre404(t *testing.T) {
	pool := triggerPool(t)
	owner := newClient(t, pool, newUser(t, pool))
	var wf workflowBody
	if code := owner.do(http.MethodPost, "/workflows", branched, &wf); code != http.StatusCreated {
		t.Fatalf("create: %d", code)
	}

	other := newClient(t, pool, newUser(t, pool))
	path := "/workflows/" + wf.ID
	for _, req := range []struct{ method, path, body string }{
		{http.MethodGet, path, ""},
		{http.MethodPut, path, branched},
		{http.MethodPatch, path, `{"enabled":false}`},
		{http.MethodDelete, path, ""},
		{http.MethodGet, path + "/runs", ""},
		{http.MethodGet, "/workflows/not-a-uuid", ""},
	} {
		if code := other.do(req.method, req.path, req.body, nil); code != http.StatusNotFound {
			t.Errorf("%s %s by another user: %d, want 404", req.method, req.path, code)
		}
	}
	var list []workflowBody
	if code := other.do(http.MethodGet, "/workflows", "", &list); code != http.StatusOK || len(list) != 0 {
		t.Fatalf("other user's list: %d %+v, want []", code, list)
	}

	var still workflowBody
	if code := owner.do(http.MethodGet, path, "", &still); code != http.StatusOK || !still.Enabled {
		t.Fatalf("owner's workflow changed: %d %+v", code, still)
	}
}

func TestInvalidWorkflowIsNotSaved(t *testing.T) {
	pool := triggerPool(t)
	c := newClient(t, pool, newUser(t, pool))
	var wf workflowBody
	if code := c.do(http.MethodPost, "/workflows", branched, &wf); code != http.StatusCreated {
		t.Fatalf("create: %d", code)
	}

	empty := strings.Replace(branched, `,{"from":"cond","to":"small","branch":"false"}`, ``, 1)
	if code := c.do(http.MethodPost, "/workflows", empty, nil); code != http.StatusUnprocessableEntity {
		t.Fatalf("create with an empty branch: %d, want 422", code)
	}
	if code := c.do(http.MethodPut, "/workflows/"+wf.ID, empty, nil); code != http.StatusUnprocessableEntity {
		t.Fatalf("replace with an empty branch: %d, want 422", code)
	}
	var list []workflowBody
	if c.do(http.MethodGet, "/workflows", "", &list); len(list) != 1 {
		t.Fatalf("list %+v, want only the valid workflow", list)
	}
	var got workflowBody
	if c.do(http.MethodGet, "/workflows/"+wf.ID, "", &got); len(got.Edges) != 3 {
		t.Fatalf("edges %+v, want the 3 saved before the rejected replace", got.Edges)
	}
}

func TestUpdateWorkflowRejectsBadPatches(t *testing.T) {
	pool := triggerPool(t)
	c := newClient(t, pool, newUser(t, pool))
	var wf workflowBody
	if code := c.do(http.MethodPost, "/workflows", branched, &wf); code != http.StatusCreated {
		t.Fatalf("create: %d", code)
	}
	for _, body := range []string{`{}`, `{"nodes":[]}`, `{"name":""}`, `not json`} {
		if code := c.do(http.MethodPatch, "/workflows/"+wf.ID, body, nil); code != http.StatusBadRequest {
			t.Errorf("patch %s: %d, want 400", body, code)
		}
	}
}

func TestReplacingTheTriggerResetsThePollCursor(t *testing.T) {
	pool := triggerPool(t)
	c := newClient(t, pool, newUser(t, pool))
	var wf workflowBody
	if code := c.do(http.MethodPost, "/workflows", branched, &wf); code != http.StatusCreated {
		t.Fatalf("create: %d", code)
	}
	cursor := func() *string {
		var s *string
		if err := pool.QueryRow(t.Context(), `SELECT poll_cursor #>> '{}' FROM workflows WHERE id = $1`, wf.ID).Scan(&s); err != nil {
			t.Fatal(err)
		}
		return s
	}
	setCursor := func() {
		if _, err := pool.Exec(t.Context(), `UPDATE workflows SET poll_cursor = '"7"' WHERE id = $1`, wf.ID); err != nil {
			t.Fatal(err)
		}
	}

	setCursor()
	c.do(http.MethodPut, "/workflows/"+wf.ID, strings.Replace(branched, `"Demo"`, `"Renamed"`, 1), nil)
	if got := cursor(); got == nil || *got != "7" {
		t.Fatalf("cursor %v after an edit that kept the trigger, want 7", got)
	}

	// The only registered trigger takes no config, so the stored one is made
	// different and the PUT changes it back.
	if _, err := pool.Exec(t.Context(),
		`UPDATE nodes SET config = '{"repo":"a/b"}' WHERE workflow_id = $1 AND kind = 'trigger'`, wf.ID); err != nil {
		t.Fatal(err)
	}
	if code := c.do(http.MethodPut, "/workflows/"+wf.ID, branched, nil); code != http.StatusOK {
		t.Fatalf("replace: %d", code)
	}
	if got := cursor(); got != nil {
		t.Fatalf("cursor %q after the trigger changed, want reset", *got)
	}
}

func TestRunHistoryUsesNodeKeys(t *testing.T) {
	pool := triggerPool(t)
	c := newClient(t, pool, newUser(t, pool))
	var wf workflowBody
	if code := c.do(http.MethodPost, "/workflows", branched, &wf); code != http.StatusCreated {
		t.Fatalf("create: %d", code)
	}
	var runID uuid.UUID
	if err := pool.QueryRow(t.Context(),
		`INSERT INTO runs (workflow_id, external_event_id, trigger_payload) VALUES ($1, 'e1', '{"count": 5}') RETURNING id`, wf.ID,
	).Scan(&runID); err != nil {
		t.Fatal(err)
	}
	if err := engine.New(pool, registry()).Execute(t.Context(), runID); err != nil {
		t.Fatalf("execute a workflow saved through the API: %v", err)
	}

	var runs []struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if code := c.do(http.MethodGet, "/workflows/"+wf.ID+"/runs?limit=5", "", &runs); code != http.StatusOK || len(runs) != 1 || runs[0].ID != runID.String() {
		t.Fatalf("runs: %d %+v", code, runs)
	}
	if code := c.do(http.MethodGet, "/workflows/"+wf.ID+"/runs?limit=0", "", nil); code != http.StatusBadRequest {
		t.Fatalf("limit=0: %d, want 400", code)
	}

	var detail struct {
		Steps []struct {
			NodeID string         `json:"node_id"`
			Status string         `json:"status"`
			Output map[string]any `json:"output"`
		} `json:"steps"`
	}
	if code := c.do(http.MethodGet, "/runs/"+runID.String(), "", &detail); code != http.StatusOK {
		t.Fatalf("get run: %d", code)
	}
	status := map[string]string{}
	for _, s := range detail.Steps {
		status[s.NodeID] = s.Status
		if s.NodeID == "big" && s.Output["message"] != "big 5" {
			t.Fatalf("big output %v, want the resolved {{ start.count }}", s.Output)
		}
	}
	want := map[string]string{"start": "SUCCEEDED", "cond": "SUCCEEDED", "big": "SUCCEEDED", "small": "SKIPPED"}
	for k, v := range want {
		if status[k] != v {
			t.Fatalf("steps %v, want %v", status, want)
		}
	}

	other := newClient(t, pool, newUser(t, pool))
	if code := other.do(http.MethodGet, "/runs/"+runID.String(), "", nil); code != http.StatusNotFound {
		t.Fatalf("another user's run: %d, want 404", code)
	}
}
