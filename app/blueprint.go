package app

// BlueprintDefinition is the vendor-published blueprint document.
// This is what gets JSON-serialized into blueprint_versions.definition.
type BlueprintDefinition struct {
	ID          string `json:"id"` // "vendor/blueprint-name"
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Version     string `json:"version"` // Semantic version

	Apps         []BlueprintApp         `json:"apps"`
	Roles        []BlueprintRole        `json:"roles"`
	Integrations []BlueprintIntegration `json:"integrations"`
}

type BlueprintApp struct {
	VendorID       string            `json:"vendorID"`
	AppID          string            `json:"appID"`
	ReleaseChannel string            `json:"releaseChannel,omitempty"`
	Required       bool              `json:"required"`
	Settings       map[string]string `json:"settings,omitempty"`
}

type BlueprintRole struct {
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}

type BlueprintIntegration struct {
	SourceVendor string   `json:"sourceVendor"`
	SourceApp    string   `json:"sourceApp"`
	TargetVendor string   `json:"targetVendor"`
	TargetApp    string   `json:"targetApp"`
	Operations   []string `json:"operations"`
}
