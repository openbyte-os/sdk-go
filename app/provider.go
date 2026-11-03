package app

import "github.com/openbyte-os/sdk-go/translation"

// ProviderRegistration declares that an app provides a service of a given type.
type ProviderRegistration struct {
	Type    string           `json:"type"`              // e.g. "ai", "email", "storage"
	Key     string           `json:"key"`               // Stable identifier, unique within app for the same type; not a display name.
	Name    translation.Text `json:"name,omitzero"`     // Optional display name; consumers fall back to Key when no localized name is available.
	Labels  []string         `json:"labels,omitempty"`  // freeform tags for routing preference
	Actions []string         `json:"actions,omitempty"` // subset of valid actions for this type
	Version int              `json:"version,omitempty"` // Major service contract version; omitted for legacy providers.
	Path    string           `json:"path,omitempty"`    // App-relative base path for the service's HTTP endpoints; required for versioned services.
}

// ProviderDependency declares that an app needs access to a provider type.
type ProviderDependency struct {
	Type            string   `json:"type"`                      // e.g. "ai", "email", "storage"
	Required        bool     `json:"required"`                  // if true, workspace must have this provider
	Reason          string   `json:"reason,omitempty"`          // human-readable explanation
	RequiredActions []string `json:"requiredActions,omitempty"` // actions the consumer needs
	Version         int      `json:"version,omitempty"`         // Exact major service contract version; omitted for legacy dependencies.
}
