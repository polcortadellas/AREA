package engine

import (
	"testing"

	"github.com/polcortadellas/AREA/server/internal/node"
)

func TestResolve(t *testing.T) {
	mail := step{
		outputs: node.Values{"subject": "urgent: server down", "count": float64(3)},
		schema: node.Schema{
			{Key: "subject", Type: node.TypeString, Required: true},
			{Key: "count", Type: node.TypeNumber, Required: true},
			{Key: "cc", Type: node.TypeString},
			{Key: "body", Type: node.TypeString, Required: true},
		},
	}
	done := map[string]step{"gmail1": mail, "trigger": mail}

	got, err := resolve(node.Values{
		"text":   "New mail: {{ gmail1.subject }} ({{trigger.count}})",
		"count":  "{{ gmail1.count }}",
		"cc":     "cc: {{ gmail1.cc }}",
		"static": true,
	}, done)
	if err != nil {
		t.Fatal(err)
	}
	want := node.Values{"text": "New mail: urgent: server down (3)", "count": float64(3), "cc": "cc: ", "static": true}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s = %#v, want %#v", k, got[k], v)
		}
	}

	for name, config := range map[string]node.Values{
		"required output missing": {"text": "{{ gmail1.body }}"},
		"node that has not run":   {"text": "hi {{ discord1.id }}"},
	} {
		if _, err := resolve(config, done); err == nil {
			t.Errorf("%s: want an error", name)
		}
	}
}
