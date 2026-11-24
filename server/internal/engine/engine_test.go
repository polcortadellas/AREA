package engine_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/url"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/polcortadellas/AREA/server/internal/connectors"
	"github.com/polcortadellas/AREA/server/internal/db"
	"github.com/polcortadellas/AREA/server/internal/engine"
	"github.com/polcortadellas/AREA/server/internal/node"
)

func database(t *testing.T) *pgxpool.Pool {
	t.Helper()
	base := os.Getenv("DATABASE_URL")
	if base == "" {
		t.Skip("DATABASE_URL must be set (throwaway PostgreSQL)")
	}
	ctx := context.Background()
	admin, err := pgx.Connect(ctx, base)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close(ctx)

	b := make([]byte, 6)
	_, _ = rand.Read(b)
	name := "area_engine_test_" + hex.EncodeToString(b)
	if _, err := admin.Exec(ctx, "CREATE DATABASE "+name); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		conn, err := pgx.Connect(context.Background(), base)
		if err == nil {
			_, _ = conn.Exec(context.Background(), "DROP DATABASE "+name+" WITH (FORCE)")
			_ = conn.Close(context.Background())
		}
	})
	u, _ := url.Parse(base)
	u.Path = "/" + name
	if err := db.Migrate(u.String()); err != nil {
		t.Fatal(err)
	}
	pool, err := pgxpool.New(ctx, u.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	return pool
}

type demo struct {
	workflow                         uuid.UUID
	trigger, cond, big, after, small string
}

// trigger -> count > 3 ? true: big -> after, false: small
func demoWorkflow(t *testing.T, pool *pgxpool.Pool) demo {
	t.Helper()
	ctx := t.Context()
	var d demo
	var user uuid.UUID
	if err := pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash) VALUES ($1, 'x') RETURNING id`, uuid.NewString()+"@example.com",
	).Scan(&user); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx,
		`INSERT INTO workflows (user_id, name) VALUES ($1, 'demo') RETURNING id`, user,
	).Scan(&d.workflow); err != nil {
		t.Fatal(err)
	}
	add := func(kind, typ string, config node.Values) string {
		var id string
		if err := pool.QueryRow(ctx,
			// key = id, so the test can use one string for edges, references and run_steps.
			`WITH n AS (SELECT gen_random_uuid() AS id)
			 INSERT INTO nodes (id, workflow_id, key, kind, type, config)
			 SELECT id, $1, id::text, $2, $3, $4 FROM n RETURNING id::text`,
			d.workflow, kind, typ, config,
		).Scan(&id); err != nil {
			t.Fatal(err)
		}
		return id
	}
	link := func(from, to, branch string) {
		var b *string
		if branch != "" {
			b = &branch
		}
		if _, err := pool.Exec(ctx,
			`INSERT INTO edges (workflow_id, from_node_id, to_node_id, branch) VALUES ($1, $2, $3, $4)`,
			d.workflow, from, to, b); err != nil {
			t.Fatal(err)
		}
	}
	d.trigger = add(node.KindTrigger, "dummy.counter", node.Values{})
	d.cond = add(node.KindCondition, "compare", node.Values{"left": "{{ trigger.count }}", "op": "greater_than", "right": "3"})
	d.big = add(node.KindReaction, "dummy.echo", node.Values{"message": "big {{ trigger.count }}"})
	d.after = add(node.KindReaction, "dummy.echo", node.Values{"message": "after {{ " + d.big + ".message }}"})
	d.small = add(node.KindReaction, "dummy.echo", node.Values{"message": "small"})
	link(d.trigger, d.cond, "")
	link(d.cond, d.big, "true")
	link(d.big, d.after, "")
	link(d.cond, d.small, "false")
	return d
}

func newRun(t *testing.T, pool *pgxpool.Pool, workflow uuid.UUID, payload node.Values) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	if err := pool.QueryRow(t.Context(),
		`INSERT INTO runs (workflow_id, external_event_id, trigger_payload) VALUES ($1, $2, $3) RETURNING id`,
		workflow, uuid.NewString(), payload,
	).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}

type stepRow struct {
	status  string
	message string
}

func steps(t *testing.T, pool *pgxpool.Pool, run uuid.UUID) map[string]stepRow {
	t.Helper()
	rows, err := pool.Query(t.Context(),
		`SELECT node_id::text, status, coalesce(output->>'message', '') FROM run_steps WHERE run_id = $1`, run)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]stepRow{}
	for rows.Next() {
		var id string
		var s stepRow
		if err := rows.Scan(&id, &s.status, &s.message); err != nil {
			t.Fatal(err)
		}
		got[id] = s
	}
	return got
}

func TestEngineRoutesBothBranches(t *testing.T) {
	pool := database(t)
	reg := node.NewRegistry()
	connectors.Register(reg)
	e := engine.New(pool, reg)
	d := demoWorkflow(t, pool)

	cases := map[string]struct {
		payload node.Values
		want    map[string]stepRow
	}{
		"true branch": {node.Values{"count": 5}, map[string]stepRow{
			d.trigger: {"SUCCEEDED", ""},
			d.cond:    {"SUCCEEDED", ""},
			d.big:     {"SUCCEEDED", "big 5"},
			d.after:   {"SUCCEEDED", "after big 5"},
			d.small:   {"SKIPPED", ""},
		}},
		"false branch": {node.Values{"count": 1}, map[string]stepRow{
			d.trigger: {"SUCCEEDED", ""},
			d.cond:    {"SUCCEEDED", ""},
			d.big:     {"SKIPPED", ""},
			d.after:   {"SKIPPED", ""},
			d.small:   {"SUCCEEDED", "small"},
		}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			run := newRun(t, pool, d.workflow, tc.payload)
			if err := e.Execute(t.Context(), run); err != nil {
				t.Fatal(err)
			}
			got := steps(t, pool, run)
			if len(got) != len(tc.want) {
				t.Fatalf("%d steps recorded, want %d: %v", len(got), len(tc.want), got)
			}
			for id, want := range tc.want {
				if got[id] != want {
					t.Errorf("node %s = %+v, want %+v", id, got[id], want)
				}
			}
		})
	}
}

func TestEngineFailsOnMissingRequiredKey(t *testing.T) {
	pool := database(t)
	reg := node.NewRegistry()
	connectors.Register(reg)
	d := demoWorkflow(t, pool)

	// dummy.counter declares count as required, so the condition cannot be evaluated.
	run := newRun(t, pool, d.workflow, node.Values{})
	if err := engine.New(pool, reg).Execute(t.Context(), run); err == nil {
		t.Fatal("want an error")
	}
	got := steps(t, pool, run)
	want := map[string]string{d.trigger: "SUCCEEDED", d.cond: "FAILED", d.big: "SKIPPED", d.after: "SKIPPED", d.small: "SKIPPED"}
	for id, status := range want {
		if got[id].status != status {
			t.Errorf("node %s = %s, want %s", id, got[id].status, status)
		}
	}
}
