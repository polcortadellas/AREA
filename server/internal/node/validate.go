package node

import (
	"errors"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"
)

const (
	KindTrigger   = "trigger"
	KindReaction  = "reaction"
	KindCondition = "condition"
)

type WorkflowNode struct {
	ID     string `json:"id"`
	Kind   string `json:"kind"`
	Type   string `json:"type"`
	Config Values `json:"config"`
}

type Edge struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Branch string `json:"branch,omitempty"`
}

type Workflow struct {
	Nodes []WorkflowNode `json:"nodes"`
	Edges []Edge         `json:"edges"`
}

var conditionSchema = Schema{
	{Key: "left", Type: TypeString, Required: true, Description: "Value to compare, usually a {{ node.key }} reference."},
	{Key: "op", Type: TypeString, Required: true, Description: "Comparison operator."},
	{Key: "right", Type: TypeString, Required: true, Description: "Value to compare against."},
}

var ConditionOps = []string{"equal", "not_equal", "greater_than", "less_than"}

var ErrInvalidWorkflow = errors.New("invalid workflow")

type Problem struct {
	Node    string `json:"node,omitempty"`
	Message string `json:"message"`
}

type InvalidWorkflowError struct {
	Problems []Problem
}

func (e *InvalidWorkflowError) Error() string {
	parts := make([]string, len(e.Problems))
	for i, p := range e.Problems {
		if p.Node != "" {
			parts[i] = fmt.Sprintf("node %q: %s", p.Node, p.Message)
		} else {
			parts[i] = p.Message
		}
	}
	return ErrInvalidWorkflow.Error() + ": " + strings.Join(parts, "; ")
}

func (e *InvalidWorkflowError) Is(target error) bool { return target == ErrInvalidWorkflow }

// RefRe matches {{ node.key }}: group 1 is the node, group 2 the key.
var RefRe = regexp.MustCompile(`\{\{\s*([^{}\s.]+)\.([a-z][a-z0-9_]*)\s*\}\}`)

// One parent per node keeps the graph a tree: branches cannot reconverge.
func (r *Registry) ValidateWorkflow(w Workflow) error {
	var problems []Problem
	bad := func(node, format string, args ...any) {
		problems = append(problems, Problem{Node: node, Message: fmt.Sprintf(format, args...)})
	}

	nodes := make(map[string]WorkflowNode, len(w.Nodes))
	var triggerID string
	conditions := 0
	for _, n := range w.Nodes {
		if n.ID == "" || n.ID == "trigger" {
			bad("", "node id %q is not allowed", n.ID)
			continue
		}
		if _, dup := nodes[n.ID]; dup {
			bad(n.ID, "node appears twice")
			continue
		}
		nodes[n.ID] = n

		switch n.Kind {
		case KindTrigger:
			if triggerID != "" {
				bad(n.ID, "a workflow has exactly one trigger")
			}
			triggerID = n.ID
			if t, ok := r.Trigger(n.Type); !ok {
				bad(n.ID, "unknown trigger %q", n.Type)
			} else if err := t.ValidateConfig(n.Config); err != nil {
				bad(n.ID, "%v", err)
			}
		case KindReaction:
			if re, ok := r.Reaction(n.Type); !ok {
				bad(n.ID, "unknown reaction %q", n.Type)
			} else if err := re.ValidateConfig(n.Config); err != nil {
				bad(n.ID, "%v", err)
			}
		case KindCondition:
			conditions++
			if err := conditionSchema.Validate(n.Config); err != nil {
				bad(n.ID, "%v", err)
			} else if op, _ := n.Config["op"].(string); !slices.Contains(ConditionOps, op) {
				bad(n.ID, "unknown operator %q, want one of %s", op, strings.Join(ConditionOps, ", "))
			}
		default:
			bad(n.ID, "unknown kind %q", n.Kind)
		}
	}
	if triggerID == "" {
		bad("", "a workflow needs a trigger")
	}
	if conditions > 1 {
		bad("", "a workflow has at most one condition, found %d", conditions)
	}

	parent := make(map[string]string, len(w.Nodes))
	children := make(map[string][]string, len(w.Nodes))
	branches := make(map[string]map[string]bool)
	for _, e := range w.Edges {
		from, okFrom := nodes[e.From]
		_, okTo := nodes[e.To]
		if !okFrom || !okTo {
			bad("", "edge %s -> %s points to a missing node", e.From, e.To)
			continue
		}
		switch {
		case e.To == triggerID:
			bad(e.From, "edge %s -> %s: nothing can point to the trigger", e.From, e.To)
			continue
		case from.Kind == KindCondition && e.Branch != "true" && e.Branch != "false":
			bad(e.From, "edge %s -> %s: an edge leaving a condition needs branch \"true\" or \"false\"", e.From, e.To)
			continue
		case from.Kind != KindCondition && e.Branch != "":
			bad(e.From, "edge %s -> %s: only a condition has branches", e.From, e.To)
			continue
		}
		if _, has := parent[e.To]; has {
			bad(e.To, "several incoming edges: branches cannot reconverge")
			continue
		}
		if from.Kind == KindCondition {
			if branches[e.From] == nil {
				branches[e.From] = map[string]bool{}
			}
			if branches[e.From][e.Branch] {
				bad(e.From, "branch %q has more than one edge", e.Branch)
				continue
			}
			branches[e.From][e.Branch] = true
		}
		parent[e.To] = e.From
		children[e.From] = append(children[e.From], e.To)
	}

	for _, n := range w.Nodes {
		if n.Kind == KindCondition {
			for _, b := range []string{"true", "false"} {
				if !branches[n.ID][b] {
					bad(n.ID, "branch %q is empty", b)
				}
			}
		}
	}

	reached := map[string]bool{}
	stack := []string{triggerID}
	for len(stack) > 0 {
		id := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if id == "" || reached[id] {
			continue
		}
		reached[id] = true
		stack = append(stack, children[id]...)
	}
	for _, n := range w.Nodes {
		if !reached[n.ID] {
			bad(n.ID, "not connected to the trigger")
			continue
		}
		for _, p := range r.checkRefs(n, parent, nodes, triggerID) {
			bad(n.ID, "%s", p)
		}
	}

	if len(problems) == 0 {
		return nil
	}
	return &InvalidWorkflowError{Problems: problems}
}

func (r *Registry) checkRefs(n WorkflowNode, parent map[string]string, nodes map[string]WorkflowNode, triggerID string) []string {
	var problems []string
	for _, key := range slices.Sorted(maps.Keys(n.Config)) {
		s, ok := n.Config[key].(string)
		if !ok {
			continue
		}
		if strings.Contains(RefRe.ReplaceAllString(s, ""), "{{") {
			problems = append(problems, fmt.Sprintf("%q has a malformed {{ node.key }} reference", key))
		}
		for _, m := range RefRe.FindAllStringSubmatch(s, -1) {
			ref, field := m[1], m[2]
			if ref == "trigger" {
				ref = triggerID
			}
			if !isAncestor(ref, n.ID, parent) {
				problems = append(problems, fmt.Sprintf("%q: {{ %s.%s }} must point to a node that runs before this one", key, m[1], field))
				continue
			}
			if _, ok := r.outputSchema(nodes[ref]).Lookup(field); !ok {
				problems = append(problems, fmt.Sprintf("%q: node %q does not declare output %q", key, m[1], field))
			}
		}
	}
	return problems
}

func isAncestor(ref, id string, parent map[string]string) bool {
	seen := map[string]bool{}
	for p, ok := parent[id]; ok && !seen[p]; p, ok = parent[p] {
		if p == ref {
			return true
		}
		seen[p] = true
	}
	return false
}

func (r *Registry) outputSchema(n WorkflowNode) Schema {
	switch n.Kind {
	case KindTrigger:
		if t, ok := r.Trigger(n.Type); ok {
			return t.OutputSchema()
		}
	case KindReaction:
		if re, ok := r.Reaction(n.Type); ok {
			return re.OutputSchema()
		}
	}
	return nil
}
