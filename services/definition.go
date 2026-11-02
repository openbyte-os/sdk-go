// Package services defines versioned JSON HTTP contracts shared by apps.
// Providers implement the same contract regardless of their underlying vendor.
package services

// Definition describes a standard service contract, independently of any app
// that implements it. Version is the positive major version of the wire contract.
// Breaking changes require a new major version.
type Definition struct {
	Type        string     `json:"type"`
	Version     int        `json:"version"`
	Description string     `json:"description,omitempty"`
	Endpoints   []Endpoint `json:"endpoints"`
}

// Endpoint describes an action within a service. Path is relative to the
// provider's advertised base path. Request and response types are defined by
// each service; JSON bodies use Content-Type: application/json.
type Endpoint struct {
	Action      string `json:"action"`
	Method      string `json:"method"`
	Path        string `json:"path"`
	Description string `json:"description,omitempty"`
}

// ErrorResponse is the JSON body returned with a non-2xx HTTP status.
// HTTP status determines the outcome; Code is a stable machine-readable reason.
type ErrorResponse struct {
	Error Error `json:"error"`
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e Error) Error() string {
	return e.Code + ": " + e.Message
}

const (
	ErrorCodeInvalidRequest = "invalid_request"
	ErrorCodeUnauthorized   = "unauthorized"
	ErrorCodeForbidden      = "forbidden"
	ErrorCodeNotFound       = "not_found"
	ErrorCodeUnsupported    = "unsupported"
	ErrorCodeRateLimited    = "rate_limited"
	ErrorCodeInternal       = "internal_error"
	ErrorCodeUnavailable    = "unavailable"
)
