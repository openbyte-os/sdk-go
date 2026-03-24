package app

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/openbyte-os/sdk-go/translation"
)

const VendorAppID = "kx" // setting your application ID to this value should be done for vendor shared

type Definition struct {
	ID GlobalAppID `json:"id"`

	Endpoint    string `json:"endpoint,omitempty"`
	MCPEndpoint string `json:"mcpEndpoint,omitempty"`
	APIEndpoint string `json:"apiEndpoint,omitempty"`

	DefaultPath string `json:"defaultPath,omitempty"` // Default path to use when opening the app

	Name        translation.Text `json:"name"`
	Description translation.Text `json:"description,omitempty"`

	UIMode             UIMode   `json:"UIMode,omitempty"`
	Category           Category `json:"category,omitempty"`
	Icon               string   `json:"icon,omitempty"` // Default icon to use for this application
	SkipCSRFValidation bool     `json:"skipCSRFValidation,omitempty"`

	Permissions []Permission `json:"permissions,omitempty"` // Permissions made available by this application

	MCPCapabilities      []MCPCapability        // Capabilities this app has when running within the MCP, e.g. to receive events or link with other apps
	Operations           []Operation            `json:"operations,omitempty"`           // APIs available for app-to-app calls
	Dependencies         []Dependency           `json:"dependencies,omitempty"`         // APIs this app needs from other apps
	Providers            []ProviderRegistration `json:"providers,omitempty"`            // Services this app provides
	ProviderDependencies []ProviderDependency   `json:"providerDependencies,omitempty"` // Provider types this app needs
	Roles                []PermissionPolicy     `json:"roles,omitempty"`                // Roles made available by this application
	Paths                []Path                 `json:"paths,omitempty"`
	Unify                []IntegrationPoint     `json:"unify,omitempty"` // How to link with other applications
	ActivationSteps      []ActivationStep       `json:"activationSteps,omitempty"`
	ListenToEvents       []ScopedKey            `json:"listenToEvents,omitempty"`
	Navigation           []Navigation           `json:"navigation,omitempty"` // App navigation
	NavigationUri        string                 `json:"navigationUri,omitempty"`

	Configuration     []SettingsPage
	ConfigurationNav  []EntryPoint `json:"configurationNav,omitempty"` // Configuration nav items are locked to the bottom of the app nav
	ConfigurationPath string       // the path to use when settings are managed by the app

	Homepage       string `json:"homepage,omitempty"`       // https:// url
	TermsOfService string `json:"termsOfService,omitempty"` // https:// url
	PrivacyPolicy  string `json:"privacyPolicy,omitempty"`  // https:// url

	SupportEmail string `json:"supportEmail,omitempty"`

	PrefixRedirect map[string]string `json:"prefixRedirect,omitempty"` // Matching prefixes to redirect  e.g. [CST:CST => 'view/$1'] $1 includes the prefix
	QuickCodes     map[string]string `json:"quickCodes,omitempty"`     // Matching Codes to redirect  e.g. [CST => 'view/$1'] $1 is replaced by everything after the code
	SearchPatterns []SearchPattern   `json:"searchPatterns,omitempty"` // Search patterns to use for this app, used in the global search

	QuickActions []EntryPoint `json:"quickActions,omitempty"` // Quick actions made available in the outer shell

	SearchPanelPath string       `json:"searchPanelPath,omitempty"` // Path to post search queries to
	SearchResults   []EntryPoint `json:"searchResults"`             // To surface pages in the global search results, e.g. Linking to sub pages

	PermittedProxyPaths []string `json:"permittedProxyPaths,omitempty"` // Paths that can be proxied by the platform, without auth / modification

	SystemApp         bool     `json:"systemApp,omitempty"`         // Only system users can see/access this app
	AllowedUsers      []string `json:"allowedUsers,omitempty"`      // Specific user IDs with access; empty = no restriction
	ProvideBlueprints bool     `json:"provideBlueprints,omitempty"` // App serves blueprint definitions via /_kubex/blueprints

	Hash string `json:"hash,omitempty"` // Hash of the definition for change detection, the latest hash can be returned in HealthResponse
}

func (d *Definition) GetHash(updateIfEmpty bool) string {
	if d == nil {
		return ""
	}
	if d.Hash == "" {
		cp := *d
		cp.Hash = ""
		jsonBytes, _ := json.Marshal(cp)
		result := fmt.Sprintf("%x", md5.Sum(jsonBytes))
		if updateIfEmpty {
			d.Hash = result
		}
		return result
	}
	return d.Hash
}

func (d *Definition) WithPath(path ...Path) *Definition {
	d.Paths = append(d.Paths, path...)
	return d
}

func FromJson(jsonBytes []byte) (*Definition, error) {
	def := &Definition{}
	if err := json.Unmarshal(jsonBytes, def); err != nil {
		return nil, errors.New("unable to decode app definition json")
	}
	return def, nil
}
