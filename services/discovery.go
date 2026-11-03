package services

import "github.com/openbyte-os/sdk-go/translation"

// DiscoveryPath is appended to a provider registration's Path. GET this endpoint
// with verified workspace and user context to list its configured instances.
const DiscoveryPath = "/_discover"

// Instance identifies one workspace-specific configuration of an advertised
// provider. Key is a stable, non-secret identifier unique within that provider
// and workspace. Never put API credentials in Key or Name.
type Instance struct {
	Key  string           `json:"key"`
	Name translation.Text `json:"name,omitzero"`
}

// DiscoveryResponse is returned with HTTP 200 and application/json. Services
// must be an array (empty when none are available), with non-empty unique keys.
// Only return instances available to the verified workspace and user.
type DiscoveryResponse struct {
	Services []Instance `json:"services"`
}
