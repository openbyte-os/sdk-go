package app

// Operation describes a callable API endpoint for app-to-app communication.
type Operation struct {
	AccessRequest

	Key         string `json:"key"`
	Description string `json:"description,omitempty"`

	Path   string `json:"path,omitempty"`
	Method string `json:"method,omitempty"`

	Inputs  []Property `json:"inputs,omitempty"`
	Outputs []Property `json:"outputs,omitempty"`

	UserContextRequired bool `json:"userContextRequired,omitempty"`
}
