package evaluate_test

import (
	"testing"

	"github.com/flagmint/flagmint-go/evaluate"
)

func TestStableHash_Deterministic(t *testing.T) {
	h1 := evaluate.StableHash("hello")
	h2 := evaluate.StableHash("hello")
	if h1 != h2 {
		t.Errorf("hash not deterministic: %d != %d", h1, h2)
	}
}

func TestStableHash_Different(t *testing.T) {
	if evaluate.StableHash("a") == evaluate.StableHash("b") {
		t.Error("unexpected hash collision for 'a' and 'b'")
	}
}

func TestInRollout_Boundaries(t *testing.T) {
	if evaluate.InRollout("any", 0) {
		t.Error("0% rollout should always return false")
	}
	if !evaluate.InRollout("any", 100) {
		t.Error("100% rollout should always return true")
	}
}

func TestVariantForKey_Stable(t *testing.T) {
	variants := []string{"control", "treatment"}
	v1 := evaluate.VariantForKey("user-1", variants)
	v2 := evaluate.VariantForKey("user-1", variants)
	if v1 != v2 {
		t.Errorf("variant not stable: %q != %q", v1, v2)
	}
	if v1 != "control" && v1 != "treatment" {
		t.Errorf("unexpected variant %q", v1)
	}
}

func TestCoerceString(t *testing.T) {
	cases := []struct {
		in   any
		want string
	}{
		{"hello", "hello"},
		{true, "true"},
		{float64(3.14), "3.14"},
		{42, "42"},
	}
	for _, tc := range cases {
		got, err := evaluate.CoerceString(tc.in)
		if err != nil {
			t.Errorf("CoerceString(%v): %v", tc.in, err)
		}
		if got != tc.want {
			t.Errorf("CoerceString(%v) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestCoerceBool(t *testing.T) {
	if v, _ := evaluate.CoerceBool(true); !v {
		t.Error("CoerceBool(true) should be true")
	}
	if v, _ := evaluate.CoerceBool("false"); v {
		t.Error(`CoerceBool("false") should be false`)
	}
	if v, _ := evaluate.CoerceBool(float64(1)); !v {
		t.Error("CoerceBool(1.0) should be true")
	}
}

func TestCondition_Match(t *testing.T) {
	attrs := map[string]any{"plan": "pro", "age": float64(30)}

	eq := evaluate.Condition{Attribute: "plan", Op: evaluate.OpEquals, Value: "pro"}
	if !eq.Match(attrs) {
		t.Error("eq.Match should be true")
	}

	neq := evaluate.Condition{Attribute: "plan", Op: evaluate.OpNotEquals, Value: "free"}
	if !neq.Match(attrs) {
		t.Error("neq.Match should be true")
	}

	gt := evaluate.Condition{Attribute: "age", Op: evaluate.OpGreaterThan, Value: float64(18)}
	if !gt.Match(attrs) {
		t.Error("gt.Match should be true")
	}
}

func TestEvaluator(t *testing.T) {
	e := evaluate.NewEvaluator()

	// No rules yet.
	if _, ok := e.Evaluate("flag", nil); ok {
		t.Error("expected miss before rules are loaded")
	}

	e.SetRules(map[string]any{"feature-x": true})
	v, ok := e.Evaluate("feature-x", nil)
	if !ok {
		t.Fatal("expected hit after SetRules")
	}
	if v != true {
		t.Errorf("got %v, want true", v)
	}
}
