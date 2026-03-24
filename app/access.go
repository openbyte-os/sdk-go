package app

type AccessRequest struct {
	RequestPermissions  []ScopedKey `json:"requestPermissions,omitempty"`  // Permissions that should be sent to this path
	RequiredPermissions []ScopedKey `json:"requiredPermissions,omitempty"` // Permissions that must be set for the user to call this page

	BuiltInResources []BuiltInResource `json:"builtInResources,omitempty"`
	RequestConfig    []ScopedKey       `json:"requestConfig,omitempty"` // Configuration values that should be sent to this path
}
