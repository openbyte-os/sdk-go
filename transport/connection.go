package transport

import "github.com/openbyte-os/sdk-go/app"

// Connection is one vendor link on a request. Token is the vendor's own access
// token: validating it is what proves the link, ID and Name are a convenience.
type Connection struct {
	Vendor    string              `json:"vendor"`
	Provider  string              `json:"provider"`
	Scope     app.ActivationScope `json:"scope"`
	ID        string              `json:"id"` // the vendor's workspace or user ID
	Name      string              `json:"name,omitempty"`
	Token     string              `json:"token"`
	TokenType string              `json:"tokenType,omitempty"`
	Expires   int64               `json:"expires,omitempty"` // unix seconds; 0 when the token does not expire
	Scopes    []string            `json:"scopes,omitempty"`
	// Brand is the workspace brand this account is for, on a provider a
	// workspace links an account per brand to.
	Brand string `json:"brand,omitempty"`
}

// Connection returns the link to provider at scope; an empty vendor matches
// any vendor.
func (r *RequestContext) Connection(vendor, provider string, scope app.ActivationScope) *Connection {
	for i := range r.Connections {
		c := &r.Connections[i]
		if c.Provider == provider && c.Scope == scope && (vendor == "" || c.Vendor == vendor) {
			return c
		}
	}
	return nil
}

// ConnectionForBrand returns the workspace's account on provider for brand,
// else the one linked for no brand in particular. An empty vendor matches any
// vendor.
func (r *RequestContext) ConnectionForBrand(vendor, provider, brand string) *Connection {
	var unassigned *Connection
	for i := range r.Connections {
		c := &r.Connections[i]
		if c.Provider != provider || c.Scope != app.ActivationScopeWorkspace || (vendor != "" && c.Vendor != vendor) {
			continue
		}
		if c.Brand == brand {
			return c
		}
		if c.Brand == "" && unassigned == nil {
			unassigned = c
		}
	}
	return unassigned
}

// ConnectionBrand is a workspace brand on a provider that takes an account
// per brand, and where its account stands.
type ConnectionBrand struct {
	Vendor   string `json:"vendor"`
	Provider string `json:"provider"`
	ID       string `json:"id"` // the brand's
	Name     string `json:"name"`
	// Account is the ID of the brand's linked account, the ID of its
	// Connection; empty when the brand has none.
	Account string `json:"account,omitempty"`
	// Request is the ID the vendor gave the account request outstanding for
	// the brand; empty when it has none or its account is linked.
	Request string `json:"request,omitempty"`
	// Link is the console path where the user links or requests the brand's
	// account, for a user who may; an app sends the browser there with
	// ResponseForwardUri.
	Link string `json:"link,omitempty"`
}
