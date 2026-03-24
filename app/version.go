package app

type Version struct {
	Environment string `json:"environment"` // Environment name (e.g., "production", "sandbox")
	Version     string `json:"version"`     // Application version
	BuildTime   string `json:"buildTime"`   // Build date in RFC3339 format
	BuildID     string `json:"buildID"`     // Build ID
	Commit      string `json:"commit"`      // Commit hash
}
