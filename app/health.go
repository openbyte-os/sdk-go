package app

// HealthCheckResponse should be returned as json from the health check endpoint
type HealthCheckResponse struct {
	Status         string `json:"status"`
	IsHealthy      bool   `json:"isHealthy"`
	NextRetry      int    `json:"nextRetry"` // Next retry in seconds if a non-healthy response is given
	DefinitionHash string `json:"definitionHash"`
}
