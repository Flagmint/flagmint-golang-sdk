// Package flagmint provides the Go SDK for the Flagmint feature flag service.
package flagmint

// FlagType enumerates the supported flag value types.
type FlagType string

const (
	FlagTypeBoolean FlagType = "boolean"
	FlagTypeString  FlagType = "string"
	FlagTypeNumber  FlagType = "number"
	FlagTypeJSON    FlagType = "json"
)

// FlagValue represents the evaluated value of a single flag.
// The concrete type depends on FlagType:
//
//	boolean → bool, string → string, number → float64, json → map[string]any
type FlagValue any

// FeatureFlags maps flag keys to their evaluated values.
type FeatureFlags map[string]FlagValue

// EvaluationContext is the user/org context sent to the server for evaluation.
// Mirrors EvaluationContextT from the JS SDK.
type EvaluationContext struct {
	Kind         string         `json:"kind"` // "user", "organization", "multi"
	Key          string         `json:"key"`
	Attributes   map[string]any `json:"attributes,omitempty"`
	User         *ContextEntity `json:"user,omitempty"`         // for kind="multi"
	Organization *ContextEntity `json:"organization,omitempty"` // for kind="multi"
}

// ContextEntity represents a single entity within a multi-kind context.
type ContextEntity struct {
	Key        string         `json:"key"`
	Attributes map[string]any `json:"attributes,omitempty"`
}
