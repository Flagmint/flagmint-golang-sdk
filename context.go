package flagmint

// Flatten returns a flat map of all attributes for the context, suitable for
// rule evaluation. For a "multi" kind context the user and organization
// sub-contexts are merged under "user.<key>" and "organization.<key>" prefixes.
func (e EvaluationContext) Flatten() map[string]any {
	result := make(map[string]any, len(e.Attributes)+2)

	for k, v := range e.Attributes {
		result[k] = v
	}

	result["kind"] = e.Kind
	result["key"] = e.Key

	if e.User != nil {
		result["user.key"] = e.User.Key
		for k, v := range e.User.Attributes {
			result["user."+k] = v
		}
	}

	if e.Organization != nil {
		result["organization.key"] = e.Organization.Key
		for k, v := range e.Organization.Attributes {
			result["organization."+k] = v
		}
	}

	return result
}
