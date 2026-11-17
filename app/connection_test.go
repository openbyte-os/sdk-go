package app

import "testing"

func TestVendorApp(t *testing.T) {
	if !IsVendorApp(NewID("fortifi", "fortifi")) || IsVendorApp(NewID("fortifi", "billing")) || IsVendorApp(GlobalAppID{}) {
		t.Fatal("IsVendorApp must match only vendor/vendor")
	}
	if VendorApp("ch") != NewID("ch", "ch") {
		t.Fatal("VendorApp(ch) must be ch/ch")
	}
}

func TestDefinitionConnectionsRoundTrip(t *testing.T) {
	def, err := FromJson([]byte(`{"id":{"vendorID":"fortifi","appID":"fortifi"},
		"authProviders":[{"id":"fortifi","clientID":"rubix","authorizeURL":"/oauth/authorize","tokenURL":"/oauth/token","identityURL":"/oauth/identity","workspaceScopes":["org"]}],
		"connections":[{"provider":"fortifi","scope":"workspace","required":true,"scopes":["billing.read"]}]}`))
	if err != nil {
		t.Fatal(err)
	}
	p := def.AuthProvider("fortifi")
	if p == nil || p.ClientID != "rubix" || p.WorkspaceScopes[0] != "org" {
		t.Fatalf("provider = %+v", p)
	}
	if def.AuthProvider("other") != nil {
		t.Fatal("unknown provider must be nil")
	}
	c := def.Connections[0]
	if c.Scope != ActivationScopeWorkspace || !c.Required || c.Scopes[0] != "billing.read" {
		t.Fatalf("connection = %+v", c)
	}
}

func TestOAuthClientSecret(t *testing.T) {
	a := OAuthClientSecret("key", "rubix")
	if len(a) != 64 || a != OAuthClientSecret("key", "rubix") {
		t.Fatal("secret must be a stable 64 char hex digest")
	}
	if a == OAuthClientSecret("key2", "rubix") || a == OAuthClientSecret("key", "other") {
		t.Fatal("secret must depend on both the key and the client")
	}
}
