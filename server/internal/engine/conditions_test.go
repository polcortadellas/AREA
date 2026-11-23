package engine

import (
	"slices"
	"testing"

	"github.com/polcortadellas/AREA/server/internal/node"
)

func TestEvaluate(t *testing.T) {
	cases := []struct {
		left, op, right string
		want            bool
		wantErr         bool
	}{
		{"urgent", "equal", "urgent", true, false},
		{"urgent", "equal", "later", false, false},
		{"urgent", "not_equal", "later", true, false},
		{"urgent", "not_equal", "urgent", false, false},
		{"10", "greater_than", "9", true, false},
		{"9", "greater_than", "10", false, false},
		{"2.5", "less_than", "3", true, false},
		{"abc", "greater_than", "1", false, true},
		{"1", "less_than", "", false, true},
		{"a", "equals", "a", false, true},
	}
	for _, c := range cases {
		got, err := evaluate(node.Values{"left": c.left, "op": c.op, "right": c.right})
		if (err != nil) != c.wantErr || got != c.want {
			t.Errorf("%q %s %q = %v, %v; want %v, error %v", c.left, c.op, c.right, got, err, c.want, c.wantErr)
		}
	}
}

// Every operator the validator accepts must be one the engine can evaluate.
func TestEvaluateKnowsEveryOperator(t *testing.T) {
	for _, op := range node.ConditionOps {
		if _, err := evaluate(node.Values{"left": "1", "op": op, "right": "1"}); err != nil {
			t.Errorf("operator %q accepted by the validator but not evaluated: %v", op, err)
		}
	}
	if !slices.Contains(node.ConditionOps, "equal") {
		t.Fatal("ConditionOps is empty or renamed")
	}
}
