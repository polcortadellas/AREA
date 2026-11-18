package node

import (
	"errors"
	"strings"
	"testing"
)

var testSchema = Schema{
	{Key: "repo", Type: TypeString, Required: true},
	{Key: "threshold", Type: TypeNumber},
	{Key: "notify", Type: TypeBoolean},
}

func TestValidateAcceptsCorrectValues(t *testing.T) {
	cases := map[string]Values{
		"all fields":          {"repo": "area", "threshold": 30.5, "notify": true},
		"optional omitted":    {"repo": "area"},
		"int is a number":     {"repo": "area", "threshold": 30},
		"template for number": {"repo": "area", "threshold": "{{ trigger.temp }}"},
		"template for string": {"repo": "{{ trigger.repo }}"},
	}
	for name, values := range cases {
		t.Run(name, func(t *testing.T) {
			if err := testSchema.Validate(values); err != nil {
				t.Fatalf("Validate(%v) = %v, want nil", values, err)
			}
		})
	}
}

func TestValidateRejectsWrongValues(t *testing.T) {
	cases := map[string]struct {
		values Values
		want   string
	}{
		"missing required": {Values{}, `"repo" is required`},
		"empty required":   {Values{"repo": ""}, `"repo" is required`},
		"wrong type":       {Values{"repo": "area", "notify": "yes"}, `"notify" must be a boolean`},
		"unknown key":      {Values{"repo": "area", "colour": "red"}, `"colour" is not a parameter`},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := testSchema.Validate(tc.values)
			if !errors.Is(err, ErrInvalidConfig) {
				t.Fatalf("Validate(%v) = %v, want an ErrInvalidConfig error", tc.values, err)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error %q does not mention %q", err, tc.want)
			}
		})
	}
}

func TestValidateListsEveryProblem(t *testing.T) {
	err := testSchema.Validate(Values{"threshold": "hot", "extra": 1})
	for _, want := range []string{`"repo" is required`, `"threshold" must be a number`, `"extra" is not a parameter`} {
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("error %v does not mention %q", err, want)
		}
	}
}

func TestLookup(t *testing.T) {
	if f, ok := testSchema.Lookup("threshold"); !ok || f.Type != TypeNumber {
		t.Fatalf("Lookup(threshold) = %v, %v", f, ok)
	}
	if _, ok := testSchema.Lookup("missing"); ok {
		t.Fatal("Lookup(missing) found a field")
	}
}
