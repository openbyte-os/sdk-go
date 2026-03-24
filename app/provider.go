package app

// ProviderRegistration declares that an app provides a service of a given type.
type ProviderRegistration struct {
	Type    string   `json:"type"`              // e.g. "ai", "email", "storage"
	Key     string   `json:"key"`               // unique within app for same type, e.g. "claude", "openai"
	Labels  []string `json:"labels,omitempty"`  // freeform tags for routing preference
	Actions []string `json:"actions,omitempty"` // subset of valid actions for this type
}

// ProviderDependency declares that an app needs access to a provider type.
type ProviderDependency struct {
	Type            string   `json:"type"`                      // e.g. "ai", "email", "storage"
	Required        bool     `json:"required"`                  // if true, workspace must have this provider
	Reason          string   `json:"reason,omitempty"`          // human-readable explanation
	RequiredActions []string `json:"requiredActions,omitempty"` // actions the consumer needs
}
