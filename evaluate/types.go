package evaluate

import "time"

// FlagType enumerates the supported flag value types.
type FlagType string

const (
	FlagTypeBoolean FlagType = "boolean"
	FlagTypeString  FlagType = "string"
	FlagTypeNumber  FlagType = "number"
	FlagTypeJSON    FlagType = "json"
)

// LogicalOperator determines how conditions are combined.
type LogicalOperator string

const (
	LogicalAND LogicalOperator = "AND"
	LogicalOR  LogicalOperator = "OR"
	LogicalNOT LogicalOperator = "NOT"
)

// Variation is a possible flag value.
type Variation struct {
	ID    string `json:"id"`
	Value any    `json:"value"`
}

// Segment is a reusable group of conditions evaluated as a unit.
type Segment struct {
	ID        string          `json:"id"`
	Rules     []Condition     `json:"rules"`
	LogicalOp LogicalOperator `json:"logical_op"`
	// Force, when true, causes the segment to always match regardless of conditions.
	Force bool `json:"force"`
}

// VariantWeight pairs a variation with its traffic allocation weight (0–100).
type VariantWeight struct {
	VariationID string  `json:"variation_id"`
	Weight      float64 `json:"weight"`
}

// RolloutKind enumerates the supported rollout strategies.
type RolloutKind string

const (
	// RolloutOff disables the rollout; the flag returns its default value.
	RolloutOff RolloutKind = "off"
	// RolloutPercentage enables a flag for a stable percentage of users (boolean-only).
	RolloutPercentage RolloutKind = "percentage"
	// RolloutVariant distributes traffic across named variations using weighted buckets.
	RolloutVariant RolloutKind = "variant"
	// RolloutGradual slowly ramps a boolean flag from 0% to TargetPercentage over time.
	RolloutGradual RolloutKind = "gradual"
)

// Rollout defines how flag values are distributed across users.
type Rollout struct {
	ID               string          `json:"id,omitempty"`
	Strategy         RolloutKind     `json:"strategy"`
	Percentage       float64         `json:"percentage,omitempty"`
	Salt             string          `json:"salt"`
	Variants         []VariantWeight `json:"variants,omitempty"`
	TargetPercentage float64         `json:"target_percentage,omitempty"`
	Increment        float64         `json:"increment,omitempty"`
	IntervalHours    float64         `json:"interval_hours,omitempty"`
	StartAt          *time.Time      `json:"start_at,omitempty"`
}

// TargetingRule is a single rule in the targeting chain.
// Rules are evaluated in ascending OrderIndex order; the first match wins.
type TargetingRule struct {
	ID          string          `json:"id"`
	Kind        string          `json:"kind"` // "segment" | "custom"
	OrderIndex  int             `json:"order_index"`
	SegmentID   string          `json:"segment_id,omitempty"`
	Conditions  []Condition     `json:"conditions,omitempty"`
	LogicalOp   LogicalOperator `json:"logical_op,omitempty"`
	VariationID *string         `json:"variation_id,omitempty"`
	RolloutID   *string         `json:"rollout_id,omitempty"`
}

// FlagConfig is the full flag configuration used for local evaluation.
// Deserialize from the server payload, then call HydrateVariations before
// passing to Evaluator.Evaluate.
type FlagConfig struct {
	Key            string              `json:"key"`
	Type           FlagType            `json:"type"`
	IsActive       bool                `json:"is_active"`
	DefaultValue   any                 `json:"default_value"`
	OffVariationID *string             `json:"off_variation_id,omitempty"`
	TargetingRules []TargetingRule     `json:"targeting_rules,omitempty"`
	Variations     []Variation         `json:"variations,omitempty"`
	// Rollouts maps rollout ID → Rollout for targeting-rule lookups.
	Rollouts map[string]*Rollout `json:"rollouts,omitempty"`

	// VariationsByID is populated by HydrateVariations; not serialised.
	VariationsByID map[string]Variation `json:"-"`
	// SegmentsByID must be populated before calling Evaluate when any
	// targeting rule has kind="segment". Not serialised.
	SegmentsByID map[string]*Segment `json:"-"`
}

// HydrateVariations populates VariationsByID from the Variations slice.
// Call this once after deserialising a FlagConfig from JSON.
func (fc *FlagConfig) HydrateVariations() {
	fc.VariationsByID = make(map[string]Variation, len(fc.Variations))
	for _, v := range fc.Variations {
		fc.VariationsByID[v.ID] = v
	}
}
