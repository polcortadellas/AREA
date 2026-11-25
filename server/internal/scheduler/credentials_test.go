package scheduler

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/polcortadellas/AREA/server/internal/node"
)

// inbox is a trigger of an OAuth2 service that records the token it polled with.
type inbox struct {
	node.NoWebhook
	got *[]string
}

func (inbox) Metadata() node.Metadata {
	return node.Metadata{Service: "acme", Name: "inbox", Description: "Records its token."}
}
func (inbox) InputSchema() node.Schema         { return nil }
func (inbox) OutputSchema() node.Schema        { return nil }
func (inbox) ValidateConfig(node.Values) error { return nil }
func (i inbox) Poll(_ context.Context, _ node.Values, _ string, creds node.Credentials) ([]node.Event, string, error) {
	*i.got = append(*i.got, creds.AccessToken)
	return nil, "", nil
}

type ownerTokens map[uuid.UUID]string

func (o ownerTokens) Credentials(_ context.Context, userID uuid.UUID, _ string) (node.Credentials, error) {
	if token, ok := o[userID]; ok {
		return node.Credentials{AccessToken: token}, nil
	}
	return node.Credentials{}, errors.New("service not linked")
}

func TestTriggerPollsWithTheOwnersToken(t *testing.T) {
	ctx := t.Context()
	pool := testPool(t)
	s, _ := newScheduler(pool)
	t.Cleanup(s.resign)
	var got []string
	s.registry.RegisterService(node.Service{Name: "acme", Description: "Test service with OAuth2.", OAuth2: true})
	s.registry.RegisterTrigger(inbox{got: &got})

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
	if _, err := pool.Exec(ctx,
		`INSERT INTO nodes (workflow_id, kind, type) VALUES ($1, 'trigger', 'acme.inbox')`, workflow,
	); err != nil {
		t.Fatal(err)
	}
	if !s.tryLead(ctx) {
		t.Fatal("scheduler should take the free lock")
	}

	s.Credentials = ownerTokens{owner: "owner-token"}
	s.pollAll(ctx)
	s.Credentials = ownerTokens{} // unlinked: the trigger is skipped, not polled without a token
	s.pollAll(ctx)

	if len(got) != 1 || got[0] != "owner-token" {
		t.Errorf("polled with %v, want once with the owner's token", got)
	}
}
