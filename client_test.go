package flagmint_test

import (
	"testing"

	flagmint "github.com/flagmint/flagmint-go"
)

func TestNewClient_EmptyAPIKey(t *testing.T) {
	_, err := flagmint.NewClient("")
	if err == nil {
		t.Fatal("expected error for empty API key, got nil")
	}
}

func TestNewClient_DeferInit(t *testing.T) {
	c, err := flagmint.NewClient("test-key",
		flagmint.WithDeferInit(),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer c.Close() //nolint:errcheck

	// No flags should be available before Initialize.
	if flags := c.GetFlags(); len(flags) != 0 {
		t.Fatalf("expected empty flags, got %d", len(flags))
	}

	if _, ok := c.GetFlag("absent"); ok {
		t.Fatal("expected missing flag to return ok=false")
	}
}

func TestSetContext(t *testing.T) {
	c, err := flagmint.NewClient("test-key", flagmint.WithDeferInit())
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close() //nolint:errcheck

	// Should not panic.
	c.SetContext(flagmint.EvaluationContext{Kind: "user", Key: "u999"})
}

func TestWithContext_Option(t *testing.T) {
	ctx := flagmint.EvaluationContext{Kind: "user", Key: "u1"}
	c, err := flagmint.NewClient("test-key",
		flagmint.WithContext(ctx),
		flagmint.WithDeferInit(),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close() //nolint:errcheck
}
