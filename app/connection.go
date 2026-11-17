package app

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	"github.com/openbyte-os/sdk-go/translation"
)

// AuthProvider is an OAuth 2.0 authorization server a vendor runs. One
// declared on the vendor app (the app whose ID equals its vendor ID) serves
// the vendor's apps; one declared on any other app is that app's alone. Rubix
// is its client: it sends the browser to AuthorizeURL with PKCE, exchanges the
// code at TokenURL, then reads the linked account from IdentityURL.
//
// AuthorizeURL is opened by the browser and must be absolute. A vendor whose
// accounts each sign in at their own address declares it in Addresses and
// writes AuthorizeURL with its {key} ("{baseUrl}/oauth/authorize"). The other
// URLs are called by Rubix and may be paths relative to the declaring app's
// endpoint.
type AuthProvider struct {
	ID   string           `json:"id"` // without "/"
	Name translation.Text `json:"name,omitzero"`

	// ClientID is the client the vendor registered for Rubix. Rubix
	// authenticates at TokenURL and RevokeURL with OAuthClientSecret.
	ClientID string `json:"clientID"`

	AuthorizeURL string `json:"authorizeURL"`
	// Addresses are asked of the workspace owner before the workspace is
	// linked, and kept as the workspace's auth data for the vendor.
	Addresses []AuthAddress `json:"addresses,omitempty"`
	TokenURL  string        `json:"tokenURL"`
	RevokeURL string        `json:"revokeURL,omitempty"` // RFC 7009
	// IdentityURL answers GET with the bearer access token and returns a
	// ConnectionIdentity for the account the token was granted on.
	IdentityURL string `json:"identityURL"`
	// ApplicationURL takes a workspace owner's request for a new account, for
	// a vendor that approves accounts by hand. Rubix POSTs the request to it
	// and reads its status from ApplicationURL/{id}, authenticating as it does
	// at TokenURL. The owner links the workspace once the vendor approves.
	ApplicationURL string `json:"applicationURL,omitempty"`
	// ApplicationFields are asked of the owner and sent with the request.
	ApplicationFields []AuthField `json:"applicationFields,omitempty"`

	// MultipleAccounts lets a workspace link several of the vendor's accounts,
	// one for each of its brands. An app receives every account in
	// transport.RequestConnections and picks one with ConnectionForBrand.
	MultipleAccounts bool `json:"multipleAccounts,omitempty"`

	WorkspaceScopes []string `json:"workspaceScopes,omitempty"` // requested on every workspace link
	UserScopes      []string `json:"userScopes,omitempty"`      // requested on every user link

	// LegacyAuthentication names the auth data keys that linked a workspace
	// before this provider existed ("orgFid"). A workspace holding all of them
	// keeps working without the workspace link, and its owners are asked to
	// link again to replace them.
	LegacyAuthentication []string `json:"legacyAuthentication,omitempty"`
}

// AuthAddress is an http(s) address of the vendor account being linked, such
// as the subdomain an organisation signs in at.
type AuthAddress struct {
	Key     string           `json:"key"` // auth data key, and the {key} placeholder in AuthorizeURL
	Label   translation.Text `json:"label"`
	Help    translation.Text `json:"help,omitzero"`
	Example string           `json:"example,omitempty"`
}

// What an AuthField's Prefill may name: details the workspace keeps on the
// brand the account is requested for, or of the owner requesting it.
const (
	PrefillBrandCompany = "brand.company" // the company's legal name, else the brand's name
	PrefillBrandEmail   = "brand.email"   // its contact address
	PrefillBrandWebsite = "brand.website"
	PrefillBrandDomains = "brand.domains" // comma separated domains it takes payments on, else its website's
	PrefillUserEmail    = "user.email"    // the requesting owner's address
)

// AuthField is a detail the vendor asks of a workspace owner requesting an
// account, such as the company's legal name.
type AuthField struct {
	Key      string           `json:"key"` // name of the value in the request's details
	Label    translation.Text `json:"label"`
	Help     translation.Text `json:"help,omitzero"`
	Example  string           `json:"example,omitempty"`
	Required bool             `json:"required,omitempty"`
	// Prefill names the detail (Prefill...) the field starts with.
	Prefill string `json:"prefill,omitempty"`
}

// Connection opts an app into an AuthProvider belonging to its own vendor.
// Once the workspace (or user) is linked, Rubix sends the link to the app in
// transport.RequestConnections.
type Connection struct {
	// Provider is the ID of a provider the app declares itself, else of one
	// on the vendor's vendor app.
	Provider string          `json:"provider"`
	Scope    ActivationScope `json:"scope"`
	// Required blocks the app until the link exists. A required workspace
	// connection is linked while the app is installed.
	Required bool     `json:"required,omitempty"`
	Scopes   []string `json:"scopes,omitempty"` // OAuth scopes this app needs on the link
}

// ConnectionIdentity is the IdentityURL response: the vendor's workspace for a
// workspace link, the vendor's user for a user link.
type ConnectionIdentity struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

// IsVendorApp reports whether id is its vendor's vendor app.
func IsVendorApp(id GlobalAppID) bool {
	return id.VendorID != "" && id.VendorID == id.AppID
}

// VendorApp returns the vendor app ID for vendor.
func VendorApp(vendor string) GlobalAppID { return NewID(vendor, vendor) }

func (d *Definition) AuthProvider(id string) *AuthProvider {
	if d == nil {
		return nil
	}
	for i := range d.AuthProviders {
		if d.AuthProviders[i].ID == id {
			return &d.AuthProviders[i]
		}
	}
	return nil
}

// OAuthClientSecret is the client secret Rubix presents for clientID. It is
// derived from the signature key of the app declaring the provider, so
// rotating that key rotates the secret.
func OAuthClientSecret(signatureKey, clientID string) string {
	mac := hmac.New(sha256.New, []byte(signatureKey))
	mac.Write([]byte("rubix-oauth-client-v1\x00" + clientID))
	return hex.EncodeToString(mac.Sum(nil))
}
