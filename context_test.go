package flagmint_test

import (
	"testing"

	flagmint "github.com/flagmint/flagmint-go"
)

func TestEvaluationContextFlatten_Simple(t *testing.T) {
	ctx := flagmint.EvaluationContext{
		Kind: "user",
		Key:  "u1",
		Attributes: map[string]any{
			"plan": "pro",
		},
	}
	flat := ctx.Flatten()

	if flat["kind"] != "user" {
		t.Errorf("kind: got %v, want user", flat["kind"])
	}
	if flat["key"] != "u1" {
		t.Errorf("key: got %v, want u1", flat["key"])
	}
	if flat["plan"] != "pro" {
		t.Errorf("plan: got %v, want pro", flat["plan"])
	}
}

func TestEvaluationContextFlatten_Multi(t *testing.T) {
	ctx := flagmint.EvaluationContext{
		Kind: "multi",
		Key:  "multi-1",
		User: &flagmint.ContextEntity{
			Key:        "u2",
			Attributes: map[string]any{"role": "admin"},
		},
		Organization: &flagmint.ContextEntity{
			Key:        "org-42",
			Attributes: map[string]any{"tier": "enterprise"},
		},
	}
	flat := ctx.Flatten()

	if flat["user.key"] != "u2" {
		t.Errorf("user.key: got %v, want u2", flat["user.key"])
	}
	if flat["user.role"] != "admin" {
		t.Errorf("user.role: got %v, want admin", flat["user.role"])
	}
	if flat["organization.key"] != "org-42" {
		t.Errorf("organization.key: got %v, want org-42", flat["organization.key"])
	}
	if flat["organization.tier"] != "enterprise" {
		t.Errorf("organization.tier: got %v, want enterprise", flat["organization.tier"])
	}
}
