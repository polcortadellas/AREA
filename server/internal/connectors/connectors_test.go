package connectors

import (
	"testing"

	"github.com/polcortadellas/AREA/server/internal/node"
)

func TestRegisterBuildsTheCatalogue(t *testing.T) {
	r := node.NewRegistry()
	Register(r)

	if _, ok := r.Trigger("dummy.counter"); !ok {
		t.Error("dummy.counter is not registered")
	}
	for _, id := range []string{"dummy.echo", "discord.send_message", "discord.create_forum_thread"} {
		if _, ok := r.Reaction(id); !ok {
			t.Errorf("%s is not registered", id)
		}
	}
}

func TestEveryNodeIsDescribed(t *testing.T) {
	r := node.NewRegistry()
	Register(r)

	for _, s := range r.Services() {
		if s.Description == "" {
			t.Errorf("service %s has no description", s.Name)
		}
		var nodes []node.Node
		for _, n := range r.Triggers(s.Name) {
			nodes = append(nodes, n)
		}
		for _, n := range r.Reactions(s.Name) {
			nodes = append(nodes, n)
		}
		for _, n := range nodes {
			if n.Metadata().Description == "" {
				t.Errorf("node %s has no description", n.Metadata().ID())
			}
		}
	}
}
