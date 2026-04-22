package app

type AccessRequest struct {
	RequestPermissions  []ScopedKey `json:"requestPermissions,omitempty"`  // Permissions that should be sent to this path
	RequiredPermissions []ScopedKey `json:"requiredPermissions,omitempty"` // Permissions that must be set for the user to call this page

	BuiltInResources []BuiltInResource `json:"builtInResources,omitempty"`
	RequestConfig    []ScopedKey       `json:"requestConfig,omitempty"` // Configuration values that should be sent to this path

	// ReadOnly indicates the Operation is safe for AI assistants to invoke
	// without confirmation — it does not create, modify, or delete state.
	// Consumed by Obi's tool-catalog filter in Rubix (/obi/tools). Apps that
	// haven't opted in get filtered by a fallback HTTP-method heuristic.
	ReadOnly bool `json:"readOnly,omitempty"`
}
