package transport

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/openbyte-os/sdk-go/app"
)

func TestNewContextFromRaw(t *testing.T) {
	t.Run("parses all known headers", func(t *testing.T) {
		raw := bytes.NewReader([]byte("X-Kx-Workspace-Id: ws-123\nX-Kx-User-Id: user-456\nX-Kx-Trace-Id: trace-789\nX-Kx-User-Ip: 10.0.0.1\nX-Kx-User-Agent: TestAgent/1.0\nX-Kx-Signature: sig/123\n\n"))
		r, err := NewContextFromRaw(raw)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if r.WorkspaceID != "ws-123" {
			t.Errorf("WorkspaceID = %q, want %q", r.WorkspaceID, "ws-123")
		}
		if r.UserID != "user-456" {
			t.Errorf("UserID = %q, want %q", r.UserID, "user-456")
		}
		if r.TraceID != "trace-789" {
			t.Errorf("TraceID = %q, want %q", r.TraceID, "trace-789")
		}
		if r.UserIP != "10.0.0.1" {
			t.Errorf("UserIP = %q, want %q", r.UserIP, "10.0.0.1")
		}
		if r.UserAgent != "TestAgent/1.0" {
			t.Errorf("UserAgent = %q, want %q", r.UserAgent, "TestAgent/1.0")
		}
	})

	t.Run("handles empty reader", func(t *testing.T) {
		raw := bytes.NewReader([]byte("\n"))
		r, err := NewContextFromRaw(raw)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if r.WorkspaceID != "" {
			t.Errorf("expected empty WorkspaceID, got %q", r.WorkspaceID)
		}
	})

	t.Run("original test - full header set", func(t *testing.T) {
		raw := bytes.NewReader([]byte("Host: connectors.chargehive.cubex-local.com:8873\nUser-Agent: Kubex Rubix/1.0\nContent-Type: \nX-Kx-Authentication: {\"chive-access-token\":\"dGVzdC1wcm9qZWN0OmQzNDdmZDI3LTFiMTMtNGE0Mi1hNDZiLTBlYTA1MzUzZjFiOTpDN0RCQ1lVVzdLMFdeNzkxYWVmYjc0YmJhNGU2MjllNzg=\",\"chive-project-id\":\"test-project\"}\nX-Kx-Authorization: [{\"e\":\"Allow\",\"p\":{\"vendorID\":\"test-vendor\",\"appID\":\"test-app\",\"Key\":\"view-configuration\"},\"r\":\"*\"}]\nX-Kx-Signature: bb557ad1806c359a6ac9d52046c7fd6dcb2e216d60312f516182e8482d03cf75/1632220779\nX-Kx-Trace-Id: 32eb2a12-9ab4-4dd6-8143-4388bb982ee3\nX-Kx-User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/93.0.4577.82 Safari/537.36\nX-Kx-User-Id: EFIDFIID-ZFA8TK5L6-MISCR-JY6QKDP\nX-Kx-User-Ip: 127.0.0.1\nX-Kx-Workspace-Id: 6daeb9c0-898a-4ee8-bc1e-96ac82e9fe6b\nX-Requested-With: XMLHttpRequest\nAccept-Encoding: gzip\n\n"))
		r, err := NewContextFromRaw(raw)
		if err != nil {
			t.Fatalf("%s", err.Error())
		}
		if r.WorkspaceID == "" {
			t.Fatal("Unable to decode headers")
		}
	})
}

func TestNewContext(t *testing.T) {
	t.Run("populates all fields from headers", func(t *testing.T) {
		headers := map[string][]string{
			RequestWorkspaceID: {"workspace-1"},
			RequestUserID:      {"user-1"},
			RequestTraceID:     {"trace-1"},
			RequestUserIP:      {"192.168.1.1"},
			RequestUserAgent:   {"MyAgent"},
			RequestSignature:   {"abc/123"},
		}
		ctx := NewContext(headers)
		if ctx.WorkspaceID != "workspace-1" {
			t.Errorf("WorkspaceID = %q, want %q", ctx.WorkspaceID, "workspace-1")
		}
		if ctx.UserID != "user-1" {
			t.Errorf("UserID = %q, want %q", ctx.UserID, "user-1")
		}
		if ctx.TraceID != "trace-1" {
			t.Errorf("TraceID = %q, want %q", ctx.TraceID, "trace-1")
		}
		if ctx.UserIP != "192.168.1.1" {
			t.Errorf("UserIP = %q, want %q", ctx.UserIP, "192.168.1.1")
		}
		if ctx.UserAgent != "MyAgent" {
			t.Errorf("UserAgent = %q, want %q", ctx.UserAgent, "MyAgent")
		}
	})

	t.Run("empty headers produces empty context", func(t *testing.T) {
		ctx := NewContext(map[string][]string{})
		if ctx.WorkspaceID != "" || ctx.UserID != "" {
			t.Error("expected empty context fields")
		}
	})

	t.Run("nil headers produces empty context", func(t *testing.T) {
		ctx := NewContext(nil)
		if ctx.WorkspaceID != "" {
			t.Error("expected empty WorkspaceID")
		}
	})

	t.Run("ignores empty value slices", func(t *testing.T) {
		headers := map[string][]string{
			RequestWorkspaceID: {},
			RequestUserID:      {"user-1"},
		}
		ctx := NewContext(headers)
		if ctx.WorkspaceID != "" {
			t.Errorf("expected empty WorkspaceID when value slice is empty, got %q", ctx.WorkspaceID)
		}
		if ctx.UserID != "user-1" {
			t.Errorf("UserID = %q, want %q", ctx.UserID, "user-1")
		}
	})

	t.Run("uses first value from multi-value header", func(t *testing.T) {
		headers := map[string][]string{
			RequestWorkspaceID: {"first", "second"},
		}
		ctx := NewContext(headers)
		if ctx.WorkspaceID != "first" {
			t.Errorf("WorkspaceID = %q, want %q", ctx.WorkspaceID, "first")
		}
	})
}

func TestApplyHeaders_Authorization(t *testing.T) {
	gaid := app.NewID("test-vendor", "test-app")
	statements := []app.PermissionStatement{
		{Effect: app.PermissionEffectAllow, Permission: app.ScopedKey{GlobalAppID: gaid, Key: "read"}, Resource: "*"},
		{Effect: app.PermissionEffectDeny, Permission: app.ScopedKey{GlobalAppID: gaid, Key: "write"}, Resource: "secret"},
	}
	authJSON, _ := json.Marshal(statements)

	headers := map[string][]string{
		RequestAuthorization: {string(authJSON)},
	}
	ctx := NewContext(headers)

	if len(ctx.Authorization) != 2 {
		t.Fatalf("expected 2 authorization statements, got %d", len(ctx.Authorization))
	}
	if ctx.Authorization[0].Effect != app.PermissionEffectAllow {
		t.Errorf("first statement effect = %q, want %q", ctx.Authorization[0].Effect, app.PermissionEffectAllow)
	}
	if ctx.Authorization[1].Resource != "secret" {
		t.Errorf("second statement resource = %q, want %q", ctx.Authorization[1].Resource, "secret")
	}
}

func TestApplyHeaders_Authentication(t *testing.T) {
	authData := `{"token":"abc123","project":"proj-1"}`
	headers := map[string][]string{
		RequestAuthentication: {authData},
	}
	ctx := NewContext(headers)

	if string(ctx.Authentication) != authData {
		t.Errorf("Authentication = %q, want %q", string(ctx.Authentication), authData)
	}
}

func TestApplyHeaders_InvalidAuthorizationJSON(t *testing.T) {
	headers := map[string][]string{
		RequestAuthorization: {"not-valid-json"},
	}
	ctx := NewContext(headers)
	if len(ctx.Authorization) != 0 {
		t.Errorf("expected empty Authorization for invalid JSON, got %d entries", len(ctx.Authorization))
	}
}

// helper to generate a valid signature
func makeSignature(workspaceID, userID, signatureKey, traceID, userIP, userAgent string, ts int64) string {
	verifyString := workspaceID + userID + signatureKey + traceID + userIP + userAgent + strconv.FormatInt(ts, 10)
	sig := sha256.Sum256([]byte(verifyString))
	return fmt.Sprintf("%x/%d", sig, ts)
}

func TestVerify(t *testing.T) {
	signatureKey := "my-secret-key"
	now := time.Now().Unix()

	newSignedContext := func(ts int64) *RequestContext {
		ctx := &RequestContext{
			WorkspaceID: "ws-1",
			UserID:      "user-1",
			TraceID:     "trace-1",
			UserIP:      "10.0.0.1",
			UserAgent:   "Agent/1",
		}
		ctx.signature = makeSignature(ctx.WorkspaceID, ctx.UserID, signatureKey, ctx.TraceID, ctx.UserIP, ctx.UserAgent, ts)
		return ctx
	}

	t.Run("valid signature", func(t *testing.T) {
		ctx := newSignedContext(now)
		if err := ctx.Verify(signatureKey, 60); err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	})

	t.Run("empty signature", func(t *testing.T) {
		ctx := &RequestContext{}
		err := ctx.Verify(signatureKey, 60)
		if err == nil {
			t.Error("expected error for empty signature")
		}
		if err.Error() != "invalid or missing signature" {
			t.Errorf("unexpected error message: %s", err.Error())
		}
	})

	t.Run("signature without slash", func(t *testing.T) {
		ctx := &RequestContext{signature: "noseparator"}
		err := ctx.Verify(signatureKey, 60)
		if err == nil {
			t.Error("expected error for signature without slash")
		}
	})

	t.Run("expired timestamp", func(t *testing.T) {
		ctx := newSignedContext(now - 120)
		err := ctx.Verify(signatureKey, 60)
		if err == nil {
			t.Error("expected error for expired timestamp")
		}
		if err.Error() != "signature outside of available time window" {
			t.Errorf("unexpected error: %s", err.Error())
		}
	})

	t.Run("future timestamp", func(t *testing.T) {
		ctx := newSignedContext(now + 120)
		err := ctx.Verify(signatureKey, 60)
		if err == nil {
			t.Error("expected error for future timestamp")
		}
		if err.Error() != "signature outside of available time window" {
			t.Errorf("unexpected error: %s", err.Error())
		}
	})

	t.Run("wrong signature key", func(t *testing.T) {
		ctx := newSignedContext(now)
		err := ctx.Verify("wrong-key", 60)
		if err == nil {
			t.Error("expected error for wrong key")
		}
		if err.Error() != "unable to verify signature" {
			t.Errorf("unexpected error: %s", err.Error())
		}
	})

	t.Run("tampered workspace ID", func(t *testing.T) {
		ctx := newSignedContext(now)
		ctx.WorkspaceID = "tampered"
		err := ctx.Verify(signatureKey, 60)
		if err == nil {
			t.Error("expected error for tampered data")
		}
	})

	t.Run("timestamp at boundary is valid", func(t *testing.T) {
		ctx := newSignedContext(now + 59)
		if err := ctx.Verify(signatureKey, 60); err != nil {
			t.Errorf("expected valid at boundary, got %v", err)
		}
	})
}

func makeTestStatements() (app.GlobalAppID, []app.PermissionStatement) {
	gaid := app.NewID("test-vendor", "test-app")
	return gaid, []app.PermissionStatement{
		{Effect: app.PermissionEffectAllow, Permission: app.ScopedKey{GlobalAppID: gaid, Key: "read"}, Resource: "*"},
		{Effect: app.PermissionEffectAllow, Permission: app.ScopedKey{GlobalAppID: gaid, Key: "write"}, Resource: "docs/*"},
		{Effect: app.PermissionEffectDeny, Permission: app.ScopedKey{GlobalAppID: gaid, Key: "write"}, Resource: "docs/secret"},
		{Effect: app.PermissionEffectAllow, Permission: app.ScopedKey{GlobalAppID: gaid, Key: "admin"}, Resource: "panel"},
		{Effect: app.PermissionEffectDeny, Permission: app.ScopedKey{GlobalAppID: gaid, Key: "admin"}, Resource: "panel"},
	}
}

func TestHasPermission(t *testing.T) {
	gaid, stmts := makeTestStatements()
	ctx := &RequestContext{Authorization: stmts}

	tests := []struct {
		name string
		perm app.ScopedKey
		want bool
	}{
		{"allowed permission", app.ScopedKey{GlobalAppID: gaid, Key: "read"}, true},
		{"deny for write overrides allow (deny seen during iteration)", app.ScopedKey{GlobalAppID: gaid, Key: "write"}, false},
		{"deny takes precedence - admin has both allow and deny", app.ScopedKey{GlobalAppID: gaid, Key: "admin"}, false},
		{"no matching permission", app.ScopedKey{GlobalAppID: gaid, Key: "delete"}, false},
		{"wrong vendor", app.ScopedKey{GlobalAppID: app.NewID("other-vendor", "test-app"), Key: "read"}, false},
		{"wrong app", app.ScopedKey{GlobalAppID: app.NewID("test-vendor", "other-app"), Key: "read"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ctx.HasPermission(tt.perm)
			if got != tt.want {
				t.Errorf("HasPermission(%v) = %v, want %v", tt.perm, got, tt.want)
			}
		})
	}

	t.Run("empty authorization", func(t *testing.T) {
		empty := &RequestContext{}
		if empty.HasPermission(app.ScopedKey{GlobalAppID: gaid, Key: "read"}) {
			t.Error("expected false for empty authorization")
		}
	})
}

func TestHasResourcePermission(t *testing.T) {
	gaid, stmts := makeTestStatements()
	ctx := &RequestContext{Authorization: stmts}

	tests := []struct {
		name     string
		perm     app.ScopedKey
		resource string
		want     bool
	}{
		{"wildcard resource allows any", app.ScopedKey{GlobalAppID: gaid, Key: "read"}, "anything", true},
		{"prefix match docs/*", app.ScopedKey{GlobalAppID: gaid, Key: "write"}, "docs/readme", true},
		{"prefix match docs/ root", app.ScopedKey{GlobalAppID: gaid, Key: "write"}, "docs/", true},
		{"deny exact match overrides prefix allow", app.ScopedKey{GlobalAppID: gaid, Key: "write"}, "docs/secret", false},
		{"exact match resource", app.ScopedKey{GlobalAppID: gaid, Key: "admin"}, "panel", false}, // deny takes precedence
		{"no matching resource", app.ScopedKey{GlobalAppID: gaid, Key: "write"}, "images/photo", false},
		{"no matching permission key", app.ScopedKey{GlobalAppID: gaid, Key: "delete"}, "docs/readme", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ctx.HasResourcePermission(tt.perm, tt.resource)
			if got != tt.want {
				t.Errorf("HasResourcePermission(%v, %q) = %v, want %v", tt.perm, tt.resource, got, tt.want)
			}
		})
	}

	t.Run("empty authorization", func(t *testing.T) {
		empty := &RequestContext{}
		if empty.HasResourcePermission(app.ScopedKey{GlobalAppID: gaid, Key: "read"}, "anything") {
			t.Error("expected false for empty authorization")
		}
	})
}

func TestPermittedResources(t *testing.T) {
	gaid, stmts := makeTestStatements()
	ctx := &RequestContext{Authorization: stmts}

	t.Run("returns allowed resources for write", func(t *testing.T) {
		got := ctx.PermittedResources(app.ScopedKey{GlobalAppID: gaid, Key: "write"})
		if len(got) != 1 || got[0] != "docs/*" {
			t.Errorf("PermittedResources(write) = %v, want [docs/*]", got)
		}
	})

	t.Run("returns allowed resources for read", func(t *testing.T) {
		got := ctx.PermittedResources(app.ScopedKey{GlobalAppID: gaid, Key: "read"})
		if len(got) != 1 || got[0] != "*" {
			t.Errorf("PermittedResources(read) = %v, want [*]", got)
		}
	})

	t.Run("returns allowed resources for admin (includes allow even if deny exists)", func(t *testing.T) {
		got := ctx.PermittedResources(app.ScopedKey{GlobalAppID: gaid, Key: "admin"})
		if len(got) != 1 || got[0] != "panel" {
			t.Errorf("PermittedResources(admin) = %v, want [panel]", got)
		}
	})

	t.Run("no matching key returns nil", func(t *testing.T) {
		got := ctx.PermittedResources(app.ScopedKey{GlobalAppID: gaid, Key: "delete"})
		if got != nil {
			t.Errorf("PermittedResources(delete) = %v, want nil", got)
		}
	})

	t.Run("empty authorization returns nil", func(t *testing.T) {
		empty := &RequestContext{}
		got := empty.PermittedResources(app.ScopedKey{GlobalAppID: gaid, Key: "read"})
		if got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})
}

func TestDeniedResources(t *testing.T) {
	gaid, stmts := makeTestStatements()
	ctx := &RequestContext{Authorization: stmts}

	t.Run("returns denied resources for write", func(t *testing.T) {
		got := ctx.DeniedResources(app.ScopedKey{GlobalAppID: gaid, Key: "write"})
		if len(got) != 1 || got[0] != "docs/secret" {
			t.Errorf("DeniedResources(write) = %v, want [docs/secret]", got)
		}
	})

	t.Run("returns denied resources for admin", func(t *testing.T) {
		got := ctx.DeniedResources(app.ScopedKey{GlobalAppID: gaid, Key: "admin"})
		if len(got) != 1 || got[0] != "panel" {
			t.Errorf("DeniedResources(admin) = %v, want [panel]", got)
		}
	})

	t.Run("no denied resources for read", func(t *testing.T) {
		got := ctx.DeniedResources(app.ScopedKey{GlobalAppID: gaid, Key: "read"})
		if got != nil {
			t.Errorf("DeniedResources(read) = %v, want nil", got)
		}
	})

	t.Run("no matching key returns nil", func(t *testing.T) {
		got := ctx.DeniedResources(app.ScopedKey{GlobalAppID: gaid, Key: "delete"})
		if got != nil {
			t.Errorf("DeniedResources(delete) = %v, want nil", got)
		}
	})
}
