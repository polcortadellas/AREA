package node_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/polcortadellas/AREA/server/internal/connectors"
	"github.com/polcortadellas/AREA/server/internal/node"
)

func registry() *node.Registry {
	r := node.NewRegistry()
	connectors.Register(r)
	return r
}

func trig() node.WorkflowNode {
	return node.WorkflowNode{ID: "t", Kind: node.KindTrigger, Type: "dummy.counter"}
}

func echo(id, msg string) node.WorkflowNode {
	return node.WorkflowNode{ID: id, Kind: node.KindReaction, Type: "dummy.echo", Config: node.Values{"message": msg}}
}

func cond(id string) node.WorkflowNode {
	return node.WorkflowNode{ID: id, Kind: node.KindCondition, Type: "compare"}
}

// trigger -> condition -> true: a -> b, false: c
func branched() node.Workflow {
	return node.Workflow{
		Nodes: []node.WorkflowNode{trig(), cond("c1"), echo("a", "{{ t.count }}"), echo("b", "{{ a.message }}"), echo("c", "{{ trigger.count }}")},
		Edges: []node.Edge{
			{From: "t", To: "c1"},
			{From: "c1", To: "a", Branch: "true"},
			{From: "a", To: "b"},
			{From: "c1", To: "c", Branch: "false"},
		},
	}
}

func TestValidateWorkflowAccepts(t *testing.T) {
	r := registry()
	if err := r.ValidateWorkflow(branched()); err != nil {
		t.Fatalf("branched workflow: %v", err)
	}

	chain := node.Workflow{Nodes: []node.WorkflowNode{trig()}}
	prev := "t"
	for _, id := range []string{"r1", "r2", "r3", "r4", "r5"} {
		chain.Nodes = append(chain.Nodes, echo(id, "{{ "+prev+".count }}"))
		if prev != "t" {
			chain.Nodes[len(chain.Nodes)-1].Config["message"] = "{{ " + prev + ".message }}"
		}
		chain.Edges = append(chain.Edges, node.Edge{From: prev, To: id})
		prev = id
	}
	if err := r.ValidateWorkflow(chain); err != nil {
		t.Fatalf("5 reactions: %v", err)
	}
}

func TestValidateWorkflowRejects(t *testing.T) {
	cases := map[string]struct {
		edit func(w *node.Workflow)
		want string
	}{
		"two conditions": {func(w *node.Workflow) {
			w.Nodes = append(w.Nodes, cond("c2"), echo("x", "hi"), echo("y", "hi"))
			w.Edges = append(w.Edges,
				node.Edge{From: "b", To: "c2"},
				node.Edge{From: "c2", To: "x", Branch: "true"},
				node.Edge{From: "c2", To: "y", Branch: "false"})
		}, "at most one condition"},
		"empty branch": {func(w *node.Workflow) {
			w.Edges = w.Edges[:3]
			w.Nodes = w.Nodes[:4]
		}, `branch "false" is empty`},
		"reconverging branches": {func(w *node.Workflow) {
			w.Edges = append(w.Edges, node.Edge{From: "c", To: "b"})
		}, "cannot reconverge"},
		"undeclared key": {func(w *node.Workflow) {
			w.Nodes[2].Config["message"] = "{{ t.nope }}"
		}, `does not declare output "nope"`},
		"key from the other branch": {func(w *node.Workflow) {
			w.Nodes[4].Config["message"] = "{{ a.message }}"
		}, "runs before this one"},
		"unknown node type": {func(w *node.Workflow) {
			w.Nodes[2].Type = "dummy.nope"
		}, "unknown reaction"},
		"cycle cut off from trigger": {func(w *node.Workflow) {
			w.Nodes = append(w.Nodes, echo("p", "hi"), echo("q", "hi"))
			w.Edges = append(w.Edges, node.Edge{From: "p", To: "q"}, node.Edge{From: "q", To: "p"})
		}, "not connected to the trigger"},
	}
	r := registry()
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			w := branched()
			tc.edit(&w)
			err := r.ValidateWorkflow(w)
			if !errors.Is(err, node.ErrInvalidWorkflow) || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("got %v, want an ErrInvalidWorkflow containing %q", err, tc.want)
			}
		})
	}
}
