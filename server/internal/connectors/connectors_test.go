package connectors

import (
	"testing"

	"github.com/polcortadellas/AREA/server/internal/node"
)

func TestRegisterBuildsTheCatalogue(t *testing.T) {
	r := node.NewRegistry()
	Register(r)

	for _, id := range []string{"dummy.counter", "github.new_issue"} {
		if _, ok := r.Trigger(id); !ok {
			t.Errorf("%s is not registered", id)
		}
	}
	for _, id := range []string{"dummy.echo", "discord.send_message", "discord.create_forum_thread", "github.create_comment"} {
		if _, ok := r.Reaction(id); !ok {
			t.Errorf("%s is not registered", id)
		}
	}
}

// The web client offers to link the services that /about.json marks as OAuth2.
func TestOAuthServicesAreMarked(t *testing.T) {
	r := node.NewRegistry()
	Register(r)

	oauth := map[string]bool{}
	for _, s := range r.Services() {
		oauth[s.Name] = s.OAuth2
	}
	if !oauth["github"] {
		t.Error("github is not registered as an OAuth2 service")
	}
	if oauth["discord"] || oauth["dummy"] {
		t.Error("only services that need a linked account should be marked OAuth2")
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
