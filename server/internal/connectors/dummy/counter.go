package dummy

import (
	"context"
	"fmt"
	"strconv"

	"github.com/polcortadellas/AREA/server/internal/node"
)

// Counter fires once per poll; the cursor is the number of the last event.
type Counter struct {
	node.NoWebhook
}

func (Counter) Metadata() node.Metadata {
	return node.Metadata{
		Service:     Service.Name,
		Name:        "counter",
		Description: "Fires on every poll with an increasing number.",
	}
}

func (Counter) InputSchema() node.Schema { return nil }

func (Counter) OutputSchema() node.Schema {
	return node.Schema{
		{Key: "count", Type: node.TypeNumber, Required: true, Description: "Number of the event, starting at 1."},
	}
}

func (c Counter) ValidateConfig(config node.Values) error {
	return c.InputSchema().Validate(config)
}

func (Counter) Poll(_ context.Context, _ node.Values, cursor string, _ node.Credentials) ([]node.Event, string, error) {
	last := 0
	if cursor != "" {
		n, err := strconv.Atoi(cursor)
		if err != nil {
			return nil, "", fmt.Errorf("dummy.counter: invalid cursor %q: %w", cursor, err)
		}
		last = n
	}
	next := strconv.Itoa(last + 1)
	event := node.Event{ID: next, Outputs: node.Values{"count": float64(last + 1)}}
	return []node.Event{event}, next, nil
}
