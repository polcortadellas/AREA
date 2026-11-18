package node

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
)

type fakeTrigger struct {
	NoWebhook
	meta Metadata
	in   Schema
	out  Schema
}

func (f fakeTrigger) Metadata() Metadata            { return f.meta }
func (f fakeTrigger) InputSchema() Schema           { return f.in }
func (f fakeTrigger) OutputSchema() Schema          { return f.out }
func (f fakeTrigger) ValidateConfig(c Values) error { return f.in.Validate(c) }
func (fakeTrigger) Poll(context.Context, Values, string, Credentials) ([]Event, string, error) {
	return nil, "", nil
}

type fakeReaction struct {
	meta Metadata
}

func (f fakeReaction) Metadata() Metadata        { return f.meta }
func (fakeReaction) InputSchema() Schema         { return nil }
func (fakeReaction) OutputSchema() Schema        { return nil }
func (fakeReaction) ValidateConfig(Values) error { return nil }
func (fakeReaction) Execute(context.Context, Values, Credentials) (Values, error) {
	return Values{}, nil
}

var svc = Service{Name: "svc", Description: "test service"}

func trigger(name string) fakeTrigger {
	return fakeTrigger{meta: Metadata{Service: "svc", Name: name}}
}

func reaction(name string) fakeReaction {
	return fakeReaction{meta: Metadata{Service: "svc", Name: name}}
}

func TestRegisterAndLookup(t *testing.T) {
	r := NewRegistry()
	r.RegisterService(svc)
	r.RegisterTrigger(trigger("new_item"))
	r.RegisterReaction(reaction("send"))

	if _, ok := r.Trigger("svc.new_item"); !ok {
		t.Error("trigger svc.new_item not found")
	}
	if _, ok := r.Reaction("svc.send"); !ok {
		t.Error("reaction svc.send not found")
	}
	if _, ok := r.Trigger("svc.send"); ok {
		t.Error("a reaction was returned as a trigger")
	}
	if got := len(r.Triggers("svc")); got != 1 {
		t.Errorf("Triggers(svc) has %d nodes, want 1", got)
	}
	if got := len(r.Reactions("other")); got != 0 {
		t.Errorf("Reactions(other) has %d nodes, want 0", got)
	}
}

func TestListsAreSorted(t *testing.T) {
	r := NewRegistry()
	r.RegisterService(Service{Name: "zeta"})
	r.RegisterService(svc)
	r.RegisterTrigger(trigger("b_event"))
	r.RegisterTrigger(trigger("a_event"))

	var services []string
	for _, s := range r.Services() {
		services = append(services, s.Name)
	}
	if got := strings.Join(services, ","); got != "svc,zeta" {
		t.Errorf("Services() = %s, want svc,zeta", got)
	}
	triggers := r.Triggers("svc")
	if triggers[0].Metadata().Name != "a_event" {
		t.Errorf("Triggers(svc) starts with %s, want a_event", triggers[0].Metadata().Name)
	}
}

func TestRegisterPanicsOnMistakes(t *testing.T) {
	badKey := trigger("bad_key")
	badKey.out = Schema{{Key: "Title", Type: TypeString}}
	dupKey := trigger("dup_key")
	dupKey.in = Schema{{Key: "repo", Type: TypeString}, {Key: "repo", Type: TypeString}}
	badType := trigger("bad_type")
	badType.in = Schema{{Key: "when", Type: "date"}}

	cases := []struct {
		name string
		want string
		fn   func(r *Registry)
	}{
		{"service twice", "registered twice", func(r *Registry) { r.RegisterService(svc) }},
		{"bad service name", "invalid service name", func(r *Registry) { r.RegisterService(Service{Name: "Git Hub"}) }},
		{"unknown service", "is not registered", func(r *Registry) {
			r.RegisterReaction(fakeReaction{meta: Metadata{Service: "nope", Name: "send"}})
		}},
		{"bad node name", "invalid node name", func(r *Registry) { r.RegisterTrigger(trigger("new-item")) }},
		{"same ID twice", "registered twice", func(r *Registry) {
			r.RegisterTrigger(trigger("item"))
			r.RegisterReaction(reaction("item"))
		}},
		{"bad field key", "invalid field key", func(r *Registry) { r.RegisterTrigger(badKey) }},
		{"duplicate field key", "duplicate field key", func(r *Registry) { r.RegisterTrigger(dupKey) }},
		{"unknown field type", "unknown type", func(r *Registry) { r.RegisterTrigger(badType) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := NewRegistry()
			r.RegisterService(svc)
			msg := panicMessage(func() { tc.fn(r) })
			if !strings.Contains(msg, tc.want) {
				t.Fatalf("panic message %q does not contain %q", msg, tc.want)
			}
		})
	}
}

func TestNoWebhook(t *testing.T) {
	_, err := trigger("x").HandleWebhook(context.Background(), nil, nil)
	if !errors.Is(err, ErrWebhookNotSupported) {
		t.Fatalf("HandleWebhook error = %v, want ErrWebhookNotSupported", err)
	}
}

func TestID(t *testing.T) {
	if got := (Metadata{Service: "github", Name: "new_issue"}).ID(); got != "github.new_issue" {
		t.Fatalf("ID() = %q, want github.new_issue", got)
	}
}

func panicMessage(fn func()) (msg string) {
	defer func() {
		if v := recover(); v != nil {
			msg = fmt.Sprint(v)
		}
	}()
	fn()
	return ""
}
