package app

import (
	"crypto/sha256"
	"fmt"
)

// CreateSignatureKey Each app can manage signature keys in its own way, however this method allows a basic key to be generated on the workspace ID & secret
func CreateSignatureKey(workspaceID, appSecret string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(workspaceID+"\x00"+appSecret)))
}

type InstallationRequest struct {
	// PlatformID provides a unique identifier to the running platform
	PlatformID  string `json:"platformId"`
	WorkspaceID string `json:"workspaceId"`
}

type InstallationResponse struct {
	SignatureKey string `json:"signatureKey"`
}
