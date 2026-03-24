package app

type MCPCapability struct {
	AccessRequest
	Capability string `json:"capability"` // e.g. resources/templates/list
}
