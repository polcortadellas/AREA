package engine_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/polcortadellas/AREA/server/internal/connectors"
	"github.com/polcortadellas/AREA/server/internal/engine"
	"github.com/polcortadellas/AREA/server/internal/node"
)

// whoami is a reaction of an OAuth2 service that returns the token it was given.
type whoami struct{}

func (whoami) Metadata() node.Metadata {
	return node.Metadata{Service: "acme", Name: "whoami", Description: "Returns its token."}
}
func (whoami) InputSchema() node.Schema { return nil }
func (whoami) OutputSchema() node.Schema {
	return node.Schema{{Key: "message", Type: node.TypeString, Required: true, Description: "The token."}}
}
func (whoami) ValidateConfig(node.Values) error { return nil }
func (whoami) Execute(_ context.Context, _ node.Values, creds node.Credentials) (node.Values, error) {
	return node.Values{"message": creds.AccessToken}, nil
}

type tokens map[uuid.UUID]string

func (t tokens) Credentials(_ context.Context, userID uuid.UUID, service string) (node.Credentials, error) {
	token, ok := t[userID]
	if !ok || service != "acme" {
		return node.Credentials{}, errors.New("service not linked")
	}
	return node.Credentials{AccessToken: token}, nil
}

func TestReactionGetsTheOwnersToken(t *testing.T) {
	pool := database(t)
	ctx := t.Context()
	reg := node.NewRegistry()
	connectors.Register(reg)
	reg.RegisterService(node.Service{Name: "acme", Description: "Test service with OAuth2.", OAuth2: true})
	reg.RegisterReaction(whoami{})

	var owner, workflow uuid.UUID
	if err := pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash) VALUES ('owner@example.com', 'x') RETURNING id`,
	).Scan(&owner); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx,
		`INSERT INTO workflows (user_id, name) VALUES ($1, 'oauth') RETURNING id`, owner,
	).Scan(&workflow); err != nil {
		t.Fatal(err)
	}
	add := func(kind, typ string) string {
		var id string
		if err := pool.QueryRow(ctx,
			`INSERT INTO nodes (workflow_id, kind, type) VALUES ($1, $2, $3) RETURNING id::text`, workflow, kind, typ,
		).Scan(&id); err != nil {
			t.Fatal(err)
		}
		return id
	}
	trigger := add(node.KindTrigger, "dummy.counter")
	reaction := add(node.KindReaction, "acme.whoami")
	if _, err := pool.Exec(ctx,
		`INSERT INTO edges (workflow_id, from_node_id, to_node_id) VALUES ($1, $2, $3)`, workflow, trigger, reaction,
	); err != nil {
		t.Fatal(err)
	}

	e := engine.New(pool, reg)
	e.Credentials = tokens{owner: "owner-token"}
	run := newRun(t, pool, workflow, node.Values{"count": 1})
	if err := e.Execute(ctx, run); err != nil {
		t.Fatal(err)
	}
	if got := steps(t, pool, run)[reaction]; got != (stepRow{"SUCCEEDED", "owner-token"}) {
		t.Errorf("reaction step = %+v, want the owner's token", got)
	}

	e.Credentials = tokens{} // the owner unlinked the service
	run = newRun(t, pool, workflow, node.Values{"count": 2})
	if err := e.Execute(ctx, run); err == nil {
		t.Fatal("want an error when the service is not linked")
	}
	if got := steps(t, pool, run)[reaction]; got.status != "FAILED" {
		t.Errorf("reaction step = %+v, want FAILED", got)
	}
}
