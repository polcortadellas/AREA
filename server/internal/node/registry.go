package node

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
)

// Registry is filled once at start-up and only read afterwards, so it has no
// lock. Register* panic on mistakes so the server never starts with a broken
// catalogue.
type Registry struct {
	services  map[string]Service
	triggers  map[string]TriggerNode
	reactions map[string]ReactionNode
}

func NewRegistry() *Registry {
	return &Registry{
		services:  make(map[string]Service),
		triggers:  make(map[string]TriggerNode),
		reactions: make(map[string]ReactionNode),
	}
}

func (r *Registry) RegisterService(s Service) {
	if !identRe.MatchString(s.Name) {
		panic(fmt.Sprintf("node: invalid service name %q", s.Name))
	}
	if _, dup := r.services[s.Name]; dup {
		panic(fmt.Sprintf("node: service %q registered twice", s.Name))
	}
	r.services[s.Name] = s
}

func (r *Registry) RegisterTrigger(t TriggerNode) {
	r.triggers[r.checkNode(t)] = t
}

func (r *Registry) RegisterReaction(n ReactionNode) {
	r.reactions[r.checkNode(n)] = n
}

// Triggers and reactions share one ID space, like nodes.type.
func (r *Registry) checkNode(n Node) string {
	m := n.Metadata()
	id := m.ID()
	if _, ok := r.services[m.Service]; !ok {
		panic(fmt.Sprintf("node: %s: service %q is not registered", id, m.Service))
	}
	if !identRe.MatchString(m.Name) {
		panic(fmt.Sprintf("node: %s: invalid node name %q", id, m.Name))
	}
	_, isTrigger := r.triggers[id]
	_, isReaction := r.reactions[id]
	if isTrigger || isReaction {
		panic(fmt.Sprintf("node: %s registered twice", id))
	}
	if err := n.InputSchema().check(); err != nil {
		panic(fmt.Sprintf("node: %s: input schema: %v", id, err))
	}
	if err := n.OutputSchema().check(); err != nil {
		panic(fmt.Sprintf("node: %s: output schema: %v", id, err))
	}
	return id
}

func (r *Registry) Trigger(id string) (TriggerNode, bool) {
	t, ok := r.triggers[id]
	return t, ok
}

func (r *Registry) Reaction(id string) (ReactionNode, bool) {
	n, ok := r.reactions[id]
	return n, ok
}

func (r *Registry) Services() []Service {
	return slices.SortedFunc(maps.Values(r.services), func(a, b Service) int {
		return cmp.Compare(a.Name, b.Name)
	})
}

func (r *Registry) Triggers(service string) []TriggerNode {
	return nodesOf(r.triggers, service)
}

func (r *Registry) Reactions(service string) []ReactionNode {
	return nodesOf(r.reactions, service)
}

func nodesOf[N Node](nodes map[string]N, service string) []N {
	var out []N
	for _, id := range slices.Sorted(maps.Keys(nodes)) {
		if n := nodes[id]; n.Metadata().Service == service {
			out = append(out, n)
		}
	}
	return out
}
