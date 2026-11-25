// Package node defines the interfaces every Action and REAction implements
// (ENG-EXT, docs/adr/0004-typed-nodes.md).
package node

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
)

// Values decoded from JSON hold string, float64 or bool.
type Values map[string]any

type Service struct {
	Name        string
	Description string
	OAuth2      bool
}

type Metadata struct {
	Service     string
	Name        string
	Description string
}

// ID is the value stored in nodes.type, e.g. "github.new_issue".
func (m Metadata) ID() string {
	return m.Service + "." + m.Name
}

type Credentials struct {
	AccessToken string
}

// CredentialSource gives out a user's token for a linked service (internal/oauth).
type CredentialSource interface {
	Credentials(ctx context.Context, userID uuid.UUID, service string) (Credentials, error)
}

type Event struct {
	// ID must stay the same across polls: the hook engine uses it to skip
	// events it has already seen.
	ID      string
	Outputs Values
}

type Node interface {
	Metadata() Metadata
	InputSchema() Schema
	// OutputSchema lists the keys later nodes can reference with {{ node.key }}.
	OutputSchema() Schema
	ValidateConfig(config Values) error
}

type TriggerNode interface {
	Node
	// Poll returns the events after cursor, oldest first, and the next cursor.
	// An empty cursor means the first poll.
	Poll(ctx context.Context, config Values, cursor string, creds Credentials) (events []Event, next string, err error)
	HandleWebhook(ctx context.Context, config Values, r *http.Request) ([]Event, error)
}

type ReactionNode interface {
	Node
	// Execute receives inputs whose {{ node.key }} references are already resolved.
	Execute(ctx context.Context, inputs Values, creds Credentials) (Values, error)
}

// ErrInvalidConfig is answered with HTTP 422 by the API.
var ErrInvalidConfig = errors.New("invalid config")

var ErrWebhookNotSupported = errors.New("trigger does not support webhooks")

// NoWebhook is embedded by triggers that only poll.
type NoWebhook struct{}

func (NoWebhook) HandleWebhook(context.Context, Values, *http.Request) ([]Event, error) {
	return nil, ErrWebhookNotSupported
}
