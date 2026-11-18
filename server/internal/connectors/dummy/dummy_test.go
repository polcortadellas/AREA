package dummy

import (
	"context"
	"errors"
	"testing"

	"github.com/polcortadellas/AREA/server/internal/node"
)

func TestCounterReturnsOneNewEventPerPoll(t *testing.T) {
	ctx := context.Background()
	cursor := ""
	for want := 1; want <= 3; want++ {
		events, next, err := Counter{}.Poll(ctx, nil, cursor, node.Credentials{})
		if err != nil {
			t.Fatalf("poll %d: %v", want, err)
		}
		if len(events) != 1 {
			t.Fatalf("poll %d returned %d events, want 1", want, len(events))
		}
		if got := events[0].Outputs["count"]; got != float64(want) {
			t.Fatalf("poll %d: count = %v, want %d", want, got, want)
		}
		if events[0].ID == cursor {
			t.Fatalf("poll %d repeated the previous event ID %q", want, cursor)
		}
		cursor = next
	}
}

func TestCounterRejectsInvalidCursor(t *testing.T) {
	if _, _, err := (Counter{}).Poll(context.Background(), nil, "abc", node.Credentials{}); err == nil {
		t.Fatal("Poll accepted the cursor \"abc\"")
	}
}

func TestCounterHasNoParameters(t *testing.T) {
	if err := (Counter{}).ValidateConfig(node.Values{"x": "y"}); !errors.Is(err, node.ErrInvalidConfig) {
		t.Fatalf("ValidateConfig = %v, want an ErrInvalidConfig error", err)
	}
}

func TestEchoReturnsItsMessage(t *testing.T) {
	out, err := Echo{}.Execute(context.Background(), node.Values{"message": "hello"}, node.Credentials{})
	if err != nil {
		t.Fatal(err)
	}
	if out["message"] != "hello" {
		t.Fatalf("message = %v, want hello", out["message"])
	}
}

func TestEchoRequiresAMessage(t *testing.T) {
	if err := (Echo{}).ValidateConfig(node.Values{}); !errors.Is(err, node.ErrInvalidConfig) {
		t.Fatalf("ValidateConfig = %v, want an ErrInvalidConfig error", err)
	}
	if err := (Echo{}).ValidateConfig(node.Values{"message": "{{ trigger.count }}"}); err != nil {
		t.Fatalf("ValidateConfig rejected a {{ }} reference: %v", err)
	}
}
