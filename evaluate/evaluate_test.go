package evaluate_test

import (
	"testing"
	"time"

	"github.com/flagmint/flagmint-go/evaluate"
)

// ─── helpers ─────────────────────────────────────────────────────────────────

func ptr[T any](v T) *T { return &v }

func boolConfig(key string, isActive bool, defVal bool, variations []evaluate.Variation, rules []evaluate.TargetingRule) *evaluate.FlagConfig {
	c := &evaluate.FlagConfig{
		Key:            key,
		Type:           evaluate.FlagTypeBoolean,
		IsActive:       isActive,
		DefaultValue:   defVal,
		TargetingRules: rules,
		Variations:     variations,
	}
	c.HydrateVariations()
	return c
}

func stringConfig(key string, isActive bool, defVal string, variations []evaluate.Variation, rules []evaluate.TargetingRule) *evaluate.FlagConfig {
	c := &evaluate.FlagConfig{
		Key:            key,
		Type:           evaluate.FlagTypeString,
		IsActive:       isActive,
		DefaultValue:   defVal,
		TargetingRules: rules,
		Variations:     variations,
	}
	c.HydrateVariations()
	return c
}

var ev = evaluate.NewEvaluator()

func eval(config *evaluate.FlagConfig, ctx map[string]any) (any, error) {
	return ev.Evaluate(config, ctx)
}

// ─── Hash parity ─────────────────────────────────────────────────────────────

func TestStringHash_KnownValues(t *testing.T) {
	// Expected values verified against the npm string-hash package output.
	cases := map[string]uint32{
		"":            5381,
		"hello":       261238937,
		"test":        2090756197,
		"a":           177670,
		"z":           177695,
		"abc":         193485963,
		"user123salt": 240410286,
	}
	for input, want := range cases {
		got := evaluate.StringHash(input)
		if got != want {
			t.Errorf("StringHash(%q) = %d, want %d", input, got, want)
		}
	}
}

func TestStringHash_Deterministic(t *testing.T) {
	for _, s := range []string{"", "hello", "user-1", "org.abc123", "verylongstring_with_unicode_äöü"} {
		h1 := evaluate.StringHash(s)
		h2 := evaluate.StringHash(s)
		if h1 != h2 {
			t.Errorf("StringHash(%q) not deterministic", s)
		}
	}
}

func TestHashPercent_Range(t *testing.T) {
	for _, s := range []string{"alice", "bob", "carol", "dave", "", "x"} {
		p := evaluate.HashPercent(s)
		if p < 0 || p > 99 {
			t.Errorf("HashPercent(%q) = %d, want [0,99]", s, p)
		}
	}
}

func TestStableHash_Deterministic(t *testing.T) {
	h1 := evaluate.StableHash("hello")
	h2 := evaluate.StableHash("hello")
	if h1 != h2 {
		t.Errorf("StableHash not deterministic: %d != %d", h1, h2)
	}
}

func TestStableHash_Different(t *testing.T) {
	if evaluate.StableHash("a") == evaluate.StableHash("b") {
		t.Error("unexpected hash collision for 'a' and 'b'")
	}
}

// ─── Coercion ─────────────────────────────────────────────────────────────────

func TestCoerceString(t *testing.T) {
	cases := []struct {
		in   any
		want string
	}{
		{"hello", "hello"},
		{true, "true"},
		{false, "false"},
		{float64(3.14), "3.14"},
		{float64(0), "0"},
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
	trueInputs := []any{true, "true", "1", float64(1), float64(3.14)}
	for _, in := range trueInputs {
		v, err := evaluate.CoerceBool(in)
		if err != nil || !v {
			t.Errorf("CoerceBool(%v) should be true (got %v, err %v)", in, v, err)
		}
	}
	falseInputs := []any{false, "false", "0", float64(0)}
	for _, in := range falseInputs {
		v, err := evaluate.CoerceBool(in)
		if err != nil || v {
			t.Errorf("CoerceBool(%v) should be false (got %v, err %v)", in, v, err)
		}
	}
	_, err := evaluate.CoerceBool(nil)
	if err == nil {
		t.Error("CoerceBool(nil) should return error")
	}
}

func TestCoerceFloat64(t *testing.T) {
	cases := []struct {
		in   any
		want float64
	}{
		{float64(3.14), 3.14},
		{42, 42},
		{"2.5", 2.5},
		{true, 1},
		{false, 0},
	}
	for _, tc := range cases {
		got, err := evaluate.CoerceFloat64(tc.in)
		if err != nil {
			t.Errorf("CoerceFloat64(%v): %v", tc.in, err)
		}
		if got != tc.want {
			t.Errorf("CoerceFloat64(%v) = %f, want %f", tc.in, got, tc.want)
		}
	}
}

// ─── Condition operators ──────────────────────────────────────────────────────

func TestCondition_Eq(t *testing.T) {
	attrs := map[string]any{"plan": "pro", "age": float64(30), "active": true}

	match(t, "eq string", attrs, evaluate.Condition{Attribute: "plan", Operator: evaluate.OpEquals, Value: "pro"}, true)
	match(t, "eq string miss", attrs, evaluate.Condition{Attribute: "plan", Operator: evaluate.OpEquals, Value: "free"}, false)
	match(t, "eq bool vs string true", attrs, evaluate.Condition{Attribute: "active", Operator: evaluate.OpEquals, Value: "true"}, true)
	match(t, "eq bool vs string false", attrs, evaluate.Condition{Attribute: "active", Operator: evaluate.OpEquals, Value: "false"}, false)
	match(t, "eq number vs string", attrs, evaluate.Condition{Attribute: "age", Operator: evaluate.OpEquals, Value: "30"}, true)
	match(t, "eq number vs string miss", attrs, evaluate.Condition{Attribute: "age", Operator: evaluate.OpEquals, Value: "31"}, false)
}

func TestCondition_Neq(t *testing.T) {
	attrs := map[string]any{"plan": "pro", "active": true}
	match(t, "neq miss", attrs, evaluate.Condition{Attribute: "plan", Operator: evaluate.OpNotEquals, Value: "free"}, true)
	match(t, "neq hit", attrs, evaluate.Condition{Attribute: "plan", Operator: evaluate.OpNotEquals, Value: "pro"}, false)
	match(t, "neq bool vs string", attrs, evaluate.Condition{Attribute: "active", Operator: evaluate.OpNotEquals, Value: "false"}, true)
}

func TestCondition_In(t *testing.T) {
	attrs := map[string]any{"plan": "pro"}
	match(t, "in hit", attrs, evaluate.Condition{Attribute: "plan", Operator: evaluate.OpIn, Value: []any{"free", "pro", "enterprise"}}, true)
	match(t, "in miss", attrs, evaluate.Condition{Attribute: "plan", Operator: evaluate.OpIn, Value: []any{"free", "starter"}}, false)
	match(t, "in nil value", attrs, evaluate.Condition{Attribute: "plan", Operator: evaluate.OpIn, Value: nil}, false)
}

func TestCondition_Nin(t *testing.T) {
	attrs := map[string]any{"plan": "pro"}
	match(t, "nin hit (not in list)", attrs, evaluate.Condition{Attribute: "plan", Operator: evaluate.OpNin, Value: []any{"free", "starter"}}, true)
	match(t, "nin miss (in list)", attrs, evaluate.Condition{Attribute: "plan", Operator: evaluate.OpNin, Value: []any{"free", "pro"}}, false)
	match(t, "nin nil value", attrs, evaluate.Condition{Attribute: "plan", Operator: evaluate.OpNin, Value: nil}, true)
}

func TestCondition_Gt_Lt(t *testing.T) {
	attrs := map[string]any{"age": float64(25)}
	match(t, "gt true", attrs, evaluate.Condition{Attribute: "age", Operator: evaluate.OpGreaterThan, Value: float64(18)}, true)
	match(t, "gt false", attrs, evaluate.Condition{Attribute: "age", Operator: evaluate.OpGreaterThan, Value: float64(30)}, false)
	match(t, "gt equal", attrs, evaluate.Condition{Attribute: "age", Operator: evaluate.OpGreaterThan, Value: float64(25)}, false)
	match(t, "lt true", attrs, evaluate.Condition{Attribute: "age", Operator: evaluate.OpLessThan, Value: float64(30)}, true)
	match(t, "lt false", attrs, evaluate.Condition{Attribute: "age", Operator: evaluate.OpLessThan, Value: float64(18)}, false)
}

func TestCondition_Gt_NaN(t *testing.T) {
	attrs := map[string]any{"age": "not-a-number"}
	match(t, "gt NaN attr", attrs, evaluate.Condition{Attribute: "age", Operator: evaluate.OpGreaterThan, Value: float64(18)}, false)
	attrs2 := map[string]any{"age": float64(25)}
	match(t, "gt NaN value", attrs2, evaluate.Condition{Attribute: "age", Operator: evaluate.OpGreaterThan, Value: "nope"}, false)
}

func TestCondition_Contains(t *testing.T) {
	attrs := map[string]any{"email": "alice@example.com"}
	match(t, "contains hit", attrs, evaluate.Condition{Attribute: "email", Operator: evaluate.OpContains, Value: "@example"}, true)
	match(t, "contains miss", attrs, evaluate.Condition{Attribute: "email", Operator: evaluate.OpContains, Value: "@other"}, false)
	match(t, "not_contains hit", attrs, evaluate.Condition{Attribute: "email", Operator: evaluate.OpNotContains, Value: "@other"}, true)
	match(t, "not_contains miss", attrs, evaluate.Condition{Attribute: "email", Operator: evaluate.OpNotContains, Value: "@example"}, false)
}

func TestCondition_StartsWith_EndsWith(t *testing.T) {
	attrs := map[string]any{"email": "alice@example.com"}
	match(t, "startsWith hit", attrs, evaluate.Condition{Attribute: "email", Operator: evaluate.OpStartsWith, Value: "alice"}, true)
	match(t, "startsWith miss", attrs, evaluate.Condition{Attribute: "email", Operator: evaluate.OpStartsWith, Value: "bob"}, false)
	match(t, "endsWith hit", attrs, evaluate.Condition{Attribute: "email", Operator: evaluate.OpEndsWith, Value: ".com"}, true)
	match(t, "endsWith miss", attrs, evaluate.Condition{Attribute: "email", Operator: evaluate.OpEndsWith, Value: ".net"}, false)
}

func TestCondition_Exists_NotExists(t *testing.T) {
	attrs := map[string]any{"plan": "pro", "score": nil}
	match(t, "exists present non-nil", attrs, evaluate.Condition{Attribute: "plan", Operator: evaluate.OpExists}, true)
	match(t, "exists nil value", attrs, evaluate.Condition{Attribute: "score", Operator: evaluate.OpExists}, false)
	match(t, "exists absent", attrs, evaluate.Condition{Attribute: "missing", Operator: evaluate.OpExists}, false)
	match(t, "not_exists absent", attrs, evaluate.Condition{Attribute: "missing", Operator: evaluate.OpNotExists}, true)
	match(t, "not_exists nil value", attrs, evaluate.Condition{Attribute: "score", Operator: evaluate.OpNotExists}, true)
	match(t, "not_exists present", attrs, evaluate.Condition{Attribute: "plan", Operator: evaluate.OpNotExists}, false)
}

func TestCondition_MissingAttribute(t *testing.T) {
	attrs := map[string]any{"plan": "pro"}
	match(t, "eq missing attr", attrs, evaluate.Condition{Attribute: "country", Operator: evaluate.OpEquals, Value: "US"}, false)
}

// ─── Custom rules (AND / OR / NOT) ───────────────────────────────────────────

func TestCustomRule_AND(t *testing.T) {
	rule := evaluate.TargetingRule{
		Kind:      "custom",
		LogicalOp: evaluate.LogicalAND,
		Conditions: []evaluate.Condition{
			{Attribute: "plan", Operator: evaluate.OpEquals, Value: "pro"},
			{Attribute: "age", Operator: evaluate.OpGreaterThan, Value: float64(18)},
		},
		VariationID: ptr("var-true"),
	}
	trueVar := evaluate.Variation{ID: "var-true", Value: true}

	cfg := boolConfig("f", true, false, []evaluate.Variation{trueVar}, []evaluate.TargetingRule{rule})
	v, _ := eval(cfg, map[string]any{"plan": "pro", "age": float64(25)})
	if v != true {
		t.Errorf("AND rule should match: got %v", v)
	}
	v, _ = eval(cfg, map[string]any{"plan": "pro", "age": float64(10)})
	if v != false {
		t.Errorf("AND rule should not match (age too low): got %v", v)
	}
}

func TestCustomRule_OR(t *testing.T) {
	rule := evaluate.TargetingRule{
		Kind:      "custom",
		LogicalOp: evaluate.LogicalOR,
		Conditions: []evaluate.Condition{
			{Attribute: "plan", Operator: evaluate.OpEquals, Value: "pro"},
			{Attribute: "plan", Operator: evaluate.OpEquals, Value: "enterprise"},
		},
		VariationID: ptr("var-true"),
	}
	trueVar := evaluate.Variation{ID: "var-true", Value: true}

	cfg := boolConfig("f", true, false, []evaluate.Variation{trueVar}, []evaluate.TargetingRule{rule})
	v, _ := eval(cfg, map[string]any{"plan": "pro"})
	if v != true {
		t.Errorf("OR rule should match 'pro': got %v", v)
	}
	v, _ = eval(cfg, map[string]any{"plan": "enterprise"})
	if v != true {
		t.Errorf("OR rule should match 'enterprise': got %v", v)
	}
	v, _ = eval(cfg, map[string]any{"plan": "free"})
	if v != false {
		t.Errorf("OR rule should not match 'free': got %v", v)
	}
}

func TestCustomRule_NOT(t *testing.T) {
	rule := evaluate.TargetingRule{
		Kind:      "custom",
		LogicalOp: evaluate.LogicalNOT,
		Conditions: []evaluate.Condition{
			{Attribute: "plan", Operator: evaluate.OpEquals, Value: "free"},
		},
		VariationID: ptr("var-true"),
	}
	trueVar := evaluate.Variation{ID: "var-true", Value: true}

	cfg := boolConfig("f", true, false, []evaluate.Variation{trueVar}, []evaluate.TargetingRule{rule})
	v, _ := eval(cfg, map[string]any{"plan": "pro"})
	if v != true {
		t.Errorf("NOT rule should match non-'free' plan: got %v", v)
	}
	v, _ = eval(cfg, map[string]any{"plan": "free"})
	if v != false {
		t.Errorf("NOT rule should not match 'free' plan: got %v", v)
	}
}

func TestCustomRule_EmptyConditions(t *testing.T) {
	// Empty conditions with AND → all conditions satisfied (vacuous truth) → match.
	rule := evaluate.TargetingRule{
		Kind:        "custom",
		LogicalOp:   evaluate.LogicalAND,
		Conditions:  []evaluate.Condition{},
		VariationID: ptr("var-true"),
	}
	cfg := boolConfig("f", true, false, []evaluate.Variation{{ID: "var-true", Value: true}}, []evaluate.TargetingRule{rule})
	v, _ := eval(cfg, map[string]any{})
	if v != true {
		t.Errorf("empty AND should vacuously match: got %v", v)
	}
}

// ─── Kill switch ─────────────────────────────────────────────────────────────

func TestKillSwitch_InactiveWithOffVariation(t *testing.T) {
	cfg := &evaluate.FlagConfig{
		Key:            "f",
		Type:           evaluate.FlagTypeBoolean,
		IsActive:       false,
		DefaultValue:   false,
		OffVariationID: ptr("var-off"),
		Variations:     []evaluate.Variation{{ID: "var-off", Value: false}},
	}
	cfg.HydrateVariations()
	v, err := eval(cfg, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != false {
		t.Errorf("kill switch: got %v, want false", v)
	}
}

func TestKillSwitch_InactiveNoOffVariation(t *testing.T) {
	cfg := &evaluate.FlagConfig{
		Key:          "f",
		Type:         evaluate.FlagTypeBoolean,
		IsActive:     false,
		DefaultValue: false,
	}
	cfg.HydrateVariations()
	v, _ := eval(cfg, nil)
	if v != false {
		t.Errorf("kill switch (no off var): got %v, want false", v)
	}
}

func TestKillSwitch_InactiveUnknownOffVariation(t *testing.T) {
	cfg := &evaluate.FlagConfig{
		Key:            "f",
		Type:           evaluate.FlagTypeBoolean,
		IsActive:       false,
		DefaultValue:   false,
		OffVariationID: ptr("var-missing"),
	}
	cfg.HydrateVariations()
	_, err := eval(cfg, nil)
	if err == nil {
		t.Error("expected ErrMissingVariation for unknown off_variation_id")
	}
}

func TestKillSwitch_ActiveFlag(t *testing.T) {
	cfg := boolConfig("f", true, false,
		[]evaluate.Variation{{ID: "v-true", Value: true}},
		nil,
	)
	v, _ := eval(cfg, nil)
	if v != true {
		t.Errorf("active boolean flag: got %v, want true", v)
	}
}

// ─── On-variation check ───────────────────────────────────────────────────────

func TestOnVariation_BoolWithTrueVariation(t *testing.T) {
	cfg := boolConfig("f", true, false,
		[]evaluate.Variation{{ID: "v-false", Value: false}, {ID: "v-true", Value: true}},
		nil,
	)
	v, _ := eval(cfg, map[string]any{})
	if v != true {
		t.Errorf("on-variation: got %v, want true", v)
	}
}

func TestOnVariation_BoolNoTrueVariation(t *testing.T) {
	cfg := boolConfig("f", true, false,
		[]evaluate.Variation{{ID: "v-false", Value: false}},
		nil,
	)
	v, _ := eval(cfg, map[string]any{})
	if v != false {
		t.Errorf("on-variation (no true): got %v, want false", v)
	}
}

func TestOnVariation_NonBooleanFlag(t *testing.T) {
	cfg := stringConfig("f", true, "default",
		[]evaluate.Variation{{ID: "v1", Value: "hello"}},
		nil,
	)
	// Non-boolean: on-variation check is skipped; no rules → default value.
	v, _ := eval(cfg, map[string]any{})
	if v != "default" {
		t.Errorf("non-boolean on-variation: got %v, want 'default'", v)
	}
}

// ─── Targeting rule ordering ──────────────────────────────────────────────────

func TestTargetingRules_Order(t *testing.T) {
	// Rule with order_index=2 comes first in slice but should lose to order_index=1.
	r1 := evaluate.TargetingRule{
		ID:          "r1",
		Kind:        "custom",
		OrderIndex:  1,
		LogicalOp:   evaluate.LogicalAND,
		Conditions:  []evaluate.Condition{{Attribute: "plan", Operator: evaluate.OpEquals, Value: "pro"}},
		VariationID: ptr("v-pro"),
	}
	r2 := evaluate.TargetingRule{
		ID:          "r2",
		Kind:        "custom",
		OrderIndex:  2,
		LogicalOp:   evaluate.LogicalAND,
		Conditions:  []evaluate.Condition{{Attribute: "plan", Operator: evaluate.OpEquals, Value: "pro"}},
		VariationID: ptr("v-second"),
	}
	cfg := &evaluate.FlagConfig{
		Key:          "f",
		Type:         evaluate.FlagTypeString,
		IsActive:     true,
		DefaultValue: "default",
		TargetingRules: []evaluate.TargetingRule{r2, r1}, // reversed
		Variations: []evaluate.Variation{
			{ID: "v-pro", Value: "first"},
			{ID: "v-second", Value: "second"},
		},
	}
	cfg.HydrateVariations()
	v, _ := eval(cfg, map[string]any{"plan": "pro"})
	if v != "first" {
		t.Errorf("rule order: got %v, want 'first'", v)
	}
}

// ─── Segment rules ─────────────────────────────────────────────────────────

func TestSegment_AND(t *testing.T) {
	seg := &evaluate.Segment{
		ID:        "seg-1",
		LogicalOp: evaluate.LogicalAND,
		Rules: []evaluate.Condition{
			{Attribute: "plan", Operator: evaluate.OpEquals, Value: "pro"},
			{Attribute: "country", Operator: evaluate.OpEquals, Value: "DE"},
		},
	}
	rule := evaluate.TargetingRule{
		Kind:        "segment",
		SegmentID:   "seg-1",
		OrderIndex:  1,
		VariationID: ptr("v-true"),
	}
	cfg := boolConfig("f", true, false,
		[]evaluate.Variation{{ID: "v-true", Value: true}},
		[]evaluate.TargetingRule{rule},
	)
	cfg.SegmentsByID = map[string]*evaluate.Segment{"seg-1": seg}

	v, _ := eval(cfg, map[string]any{"plan": "pro", "country": "DE"})
	if v != true {
		t.Errorf("segment AND match: got %v, want true", v)
	}
	v, _ = eval(cfg, map[string]any{"plan": "pro", "country": "US"})
	if v != false {
		t.Errorf("segment AND no match: got %v, want false", v)
	}
}

func TestSegment_Force(t *testing.T) {
	seg := &evaluate.Segment{
		ID:        "seg-force",
		LogicalOp: evaluate.LogicalAND,
		Rules: []evaluate.Condition{
			{Attribute: "plan", Operator: evaluate.OpEquals, Value: "pro"},
		},
		Force: true,
	}
	rule := evaluate.TargetingRule{
		Kind:        "segment",
		SegmentID:   "seg-force",
		OrderIndex:  1,
		VariationID: ptr("v-true"),
	}
	cfg := boolConfig("f", true, false,
		[]evaluate.Variation{{ID: "v-true", Value: true}},
		[]evaluate.TargetingRule{rule},
	)
	cfg.SegmentsByID = map[string]*evaluate.Segment{"seg-force": seg}

	// force=true means the segment matches even if conditions don't.
	v, _ := eval(cfg, map[string]any{"plan": "free"})
	if v != true {
		t.Errorf("segment force: got %v, want true", v)
	}
}

func TestSegment_Missing(t *testing.T) {
	rule := evaluate.TargetingRule{
		Kind:        "segment",
		SegmentID:   "seg-missing",
		OrderIndex:  1,
		VariationID: ptr("v-true"),
	}
	cfg := boolConfig("f", true, false,
		[]evaluate.Variation{{ID: "v-true", Value: true}},
		[]evaluate.TargetingRule{rule},
	)
	// No SegmentsByID → should fail-open (skip rule, return default).
	v, _ := eval(cfg, map[string]any{"plan": "pro"})
	if v != false {
		t.Errorf("missing segment fail-open: got %v, want false", v)
	}
}

// ─── Percentage rollout ───────────────────────────────────────────────────────

func rolloutConfig(flagType evaluate.FlagType, defVal any, rollout *evaluate.Rollout, rolloutID string) *evaluate.FlagConfig {
	c := &evaluate.FlagConfig{
		Key:          "f",
		Type:         flagType,
		IsActive:     true,
		DefaultValue: defVal,
		TargetingRules: []evaluate.TargetingRule{
			{
				Kind:       "custom",
				LogicalOp:  evaluate.LogicalAND,
				Conditions: []evaluate.Condition{},
				OrderIndex: 1,
				RolloutID:  ptr(rolloutID),
			},
		},
		Rollouts: map[string]*evaluate.Rollout{rolloutID: rollout},
	}
	c.HydrateVariations()
	return c
}

func TestPercentageRollout_0(t *testing.T) {
	cfg := rolloutConfig(evaluate.FlagTypeBoolean, false, &evaluate.Rollout{
		Strategy:   evaluate.RolloutPercentage,
		Percentage: 0,
		Salt:       "s",
	}, "r1")
	v, _ := eval(cfg, map[string]any{"kind": "user", "key": "alice"})
	if v != false {
		t.Errorf("0%% rollout: got %v, want false", v)
	}
}

func TestPercentageRollout_100(t *testing.T) {
	cfg := rolloutConfig(evaluate.FlagTypeBoolean, false, &evaluate.Rollout{
		Strategy:   evaluate.RolloutPercentage,
		Percentage: 100,
		Salt:       "s",
	}, "r1")
	v, _ := eval(cfg, map[string]any{"kind": "user", "key": "alice"})
	if v != true {
		t.Errorf("100%% rollout: got %v, want true", v)
	}
}

func TestPercentageRollout_50_Stable(t *testing.T) {
	cfg := rolloutConfig(evaluate.FlagTypeBoolean, false, &evaluate.Rollout{
		Strategy:   evaluate.RolloutPercentage,
		Percentage: 50,
		Salt:       "salt",
	}, "r1")
	ctx := map[string]any{"kind": "user", "key": "stable-user"}
	v1, _ := eval(cfg, ctx)
	v2, _ := eval(cfg, ctx)
	if v1 != v2 {
		t.Error("percentage rollout must be stable across calls")
	}
}

func TestPercentageRollout_NonBoolean(t *testing.T) {
	cfg := rolloutConfig(evaluate.FlagTypeString, "default", &evaluate.Rollout{
		Strategy:   evaluate.RolloutPercentage,
		Percentage: 50,
		Salt:       "s",
	}, "r1")
	_, err := eval(cfg, map[string]any{"kind": "user", "key": "alice"})
	if err == nil {
		t.Error("percentage rollout on non-boolean should return error")
	}
}

func TestPercentageRollout_MissingKey(t *testing.T) {
	cfg := rolloutConfig(evaluate.FlagTypeBoolean, false, &evaluate.Rollout{
		Strategy:   evaluate.RolloutPercentage,
		Percentage: 50,
		Salt:       "s",
	}, "r1")
	_, err := eval(cfg, map[string]any{}) // no kind/key
	if err == nil {
		t.Error("percentage rollout without context key should return error")
	}
}

// ─── Variant rollout ──────────────────────────────────────────────────────────

func TestVariantRollout_TwoVariants(t *testing.T) {
	variations := []evaluate.Variation{
		{ID: "v-a", Value: "A"},
		{ID: "v-b", Value: "B"},
	}
	cfg := &evaluate.FlagConfig{
		Key:          "f",
		Type:         evaluate.FlagTypeString,
		IsActive:     true,
		DefaultValue: "default",
		Variations:   variations,
		TargetingRules: []evaluate.TargetingRule{
			{Kind: "custom", LogicalOp: evaluate.LogicalAND, Conditions: []evaluate.Condition{}, OrderIndex: 1, RolloutID: ptr("r1")},
		},
		Rollouts: map[string]*evaluate.Rollout{
			"r1": {
				Strategy: evaluate.RolloutVariant,
				Salt:     "s",
				Variants: []evaluate.VariantWeight{
					{VariationID: "v-a", Weight: 50},
					{VariationID: "v-b", Weight: 50},
				},
			},
		},
	}
	cfg.HydrateVariations()

	ctx := map[string]any{"kind": "user", "key": "alice"}
	v, err := eval(cfg, ctx)
	if err != nil {
		t.Fatalf("variant rollout error: %v", err)
	}
	if v != "A" && v != "B" {
		t.Errorf("variant rollout: got %v, want A or B", v)
	}
	// Stability check.
	v2, _ := eval(cfg, ctx)
	if v != v2 {
		t.Error("variant rollout must be stable across calls")
	}
}

func TestVariantRollout_InvalidWeights(t *testing.T) {
	cfg := &evaluate.FlagConfig{
		Key:          "f",
		Type:         evaluate.FlagTypeString,
		IsActive:     true,
		DefaultValue: "default",
		Variations:   []evaluate.Variation{{ID: "v-a", Value: "A"}},
		TargetingRules: []evaluate.TargetingRule{
			{Kind: "custom", LogicalOp: evaluate.LogicalAND, Conditions: []evaluate.Condition{}, OrderIndex: 1, RolloutID: ptr("r1")},
		},
		Rollouts: map[string]*evaluate.Rollout{
			"r1": {
				Strategy: evaluate.RolloutVariant,
				Salt:     "s",
				Variants: []evaluate.VariantWeight{
					{VariationID: "v-a", Weight: 60},
				},
			},
		},
	}
	cfg.HydrateVariations()
	_, err := eval(cfg, map[string]any{"kind": "user", "key": "alice"})
	if err == nil {
		t.Error("variant rollout with weight != 100 should return error")
	}
}

func TestVariantRollout_MissingVariation(t *testing.T) {
	cfg := &evaluate.FlagConfig{
		Key:          "f",
		Type:         evaluate.FlagTypeString,
		IsActive:     true,
		DefaultValue: "default",
		Variations:   []evaluate.Variation{},
		TargetingRules: []evaluate.TargetingRule{
			{Kind: "custom", LogicalOp: evaluate.LogicalAND, Conditions: []evaluate.Condition{}, OrderIndex: 1, RolloutID: ptr("r1")},
		},
		Rollouts: map[string]*evaluate.Rollout{
			"r1": {
				Strategy: evaluate.RolloutVariant,
				Salt:     "s",
				Variants: []evaluate.VariantWeight{
					{VariationID: "v-missing", Weight: 100},
				},
			},
		},
	}
	cfg.HydrateVariations()
	_, err := eval(cfg, map[string]any{"kind": "user", "key": "alice"})
	if err == nil {
		t.Error("variant rollout with missing variation should return error")
	}
}

// ─── Gradual rollout ──────────────────────────────────────────────────────────

func TestGradualRollout_NotStarted(t *testing.T) {
	// Use a fixed future timestamp (year 2099) to ensure the rollout hasn't started.
	future := time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)
	cfg := rolloutConfig(evaluate.FlagTypeBoolean, false, &evaluate.Rollout{
		Strategy:         evaluate.RolloutGradual,
		Salt:             "s",
		TargetPercentage: 100,
		Increment:        10,
		IntervalHours:    1,
		StartAt:          &future,
	}, "r1")
	v, _ := eval(cfg, map[string]any{"kind": "user", "key": "alice"})
	if v != false {
		t.Errorf("gradual not started: got %v, want false", v)
	}
}

func TestGradualRollout_Completed(t *testing.T) {
	// Use a fixed past timestamp that is far enough back (200h) to reach 100% target.
	// start=2000-01-01, interval=1h, increment=10%/h, target=100%: after 10h it's done.
	past := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	cfg := rolloutConfig(evaluate.FlagTypeBoolean, false, &evaluate.Rollout{
		Strategy:         evaluate.RolloutGradual,
		Salt:             "s",
		TargetPercentage: 100,
		Increment:        10,
		IntervalHours:    1,
		StartAt:          &past,
	}, "r1")
	v, _ := eval(cfg, map[string]any{"kind": "user", "key": "alice"})
	if v != true {
		t.Errorf("gradual completed (100%%): got %v, want true", v)
	}
}

func TestGradualRollout_NonBoolean(t *testing.T) {
	past := time.Now().Add(-time.Hour)
	cfg := rolloutConfig(evaluate.FlagTypeString, "default", &evaluate.Rollout{
		Strategy:         evaluate.RolloutGradual,
		Salt:             "s",
		TargetPercentage: 100,
		Increment:        10,
		IntervalHours:    1,
		StartAt:          &past,
	}, "r1")
	_, err := eval(cfg, map[string]any{"kind": "user", "key": "alice"})
	if err == nil {
		t.Error("gradual rollout on non-boolean should return error")
	}
}

// ─── Context flattening (via rollout key extraction) ─────────────────────────

func TestGetStableRolloutKey_KindKey(t *testing.T) {
	// Indirect: build a 50% rollout and verify the same user always buckets the same.
	cfg := rolloutConfig(evaluate.FlagTypeBoolean, false, &evaluate.Rollout{
		Strategy:   evaluate.RolloutPercentage,
		Percentage: 50,
		Salt:       "salt",
	}, "r1")
	ctx := map[string]any{"kind": "user", "key": "deterministic-key"}
	v1, _ := eval(cfg, ctx)
	v2, _ := eval(cfg, ctx)
	if v1 != v2 {
		t.Error("rollout key extraction must produce stable results")
	}
}

// ─── Missing rollout / variation errors ───────────────────────────────────────

func TestMissingRollout(t *testing.T) {
	cfg := &evaluate.FlagConfig{
		Key:          "f",
		Type:         evaluate.FlagTypeBoolean,
		IsActive:     true,
		DefaultValue: false,
		TargetingRules: []evaluate.TargetingRule{
			{Kind: "custom", LogicalOp: evaluate.LogicalAND, Conditions: []evaluate.Condition{}, OrderIndex: 1, RolloutID: ptr("r-missing")},
		},
		Rollouts: map[string]*evaluate.Rollout{},
	}
	cfg.HydrateVariations()
	_, err := eval(cfg, map[string]any{"kind": "user", "key": "alice"})
	if err == nil {
		t.Error("missing rollout should return error")
	}
}

func TestMissingVariation(t *testing.T) {
	cfg := &evaluate.FlagConfig{
		Key:          "f",
		Type:         evaluate.FlagTypeBoolean,
		IsActive:     true,
		DefaultValue: false,
		TargetingRules: []evaluate.TargetingRule{
			{Kind: "custom", LogicalOp: evaluate.LogicalAND, Conditions: []evaluate.Condition{}, OrderIndex: 1, VariationID: ptr("v-missing")},
		},
		Variations: []evaluate.Variation{},
	}
	cfg.HydrateVariations()
	_, err := eval(cfg, map[string]any{})
	if err == nil {
		t.Error("missing variation should return error")
	}
}

// ─── InRollout / VariantForKey (backward compat) ──────────────────────────────

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

// ─── MatchAll backward compat ─────────────────────────────────────────────────

func TestMatchAll(t *testing.T) {
	attrs := map[string]any{"plan": "pro", "age": float64(30)}
	rules := []evaluate.Condition{
		{Attribute: "plan", Operator: evaluate.OpEquals, Value: "pro"},
		{Attribute: "age", Operator: evaluate.OpGreaterThan, Value: float64(18)},
	}
	if !evaluate.MatchAll(rules, attrs) {
		t.Error("MatchAll: expected true")
	}
	rules[0].Value = "free"
	if evaluate.MatchAll(rules, attrs) {
		t.Error("MatchAll: expected false")
	}
}

// ─── NilConfig guard ─────────────────────────────────────────────────────────

func TestEvaluate_NilConfig(t *testing.T) {
	_, err := ev.Evaluate(nil, nil)
	if err == nil {
		t.Error("Evaluate(nil, nil) should return an error")
	}
}

// ─── helper ──────────────────────────────────────────────────────────────────

func match(t *testing.T, name string, attrs map[string]any, cond evaluate.Condition, want bool) {
	t.Helper()
	got := cond.Match(attrs)
	if got != want {
		t.Errorf("%s: Match() = %v, want %v", name, got, want)
	}
}
