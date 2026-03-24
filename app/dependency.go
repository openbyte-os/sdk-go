package app

// Dependency declares a dependency on another app's Operation endpoints.
type Dependency struct {
	App                 GlobalAppID `json:"app"`
	RequiredOperations  []string    `json:"requiredOperations,omitempty"`  // OpKeys
	RequestedOperations []string    `json:"requestedOperations,omitempty"` // OpKeys
	Required            bool        `json:"required"`
	Reason              string      `json:"reason,omitempty"`
}
