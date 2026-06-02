package app

import (
	"encoding/json"
	"testing"

	"github.com/openbyte-os/sdk-go/translation"
)

func TestFromJson(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		wantErr  bool
		validate func(t *testing.T, d *Definition)
	}{
		{
			name:    "invalid JSON returns error",
			input:   []byte(`{not valid json`),
			wantErr: true,
		},
		{
			name:    "empty bytes returns error",
			input:   []byte(``),
			wantErr: true,
		},
		{
			name:    "null JSON returns empty definition",
			input:   []byte(`null`),
			wantErr: false,
			validate: func(t *testing.T, d *Definition) {
				// json.Unmarshal into a pre-allocated pointer with null input leaves it zeroed
				if d == nil {
					t.Fatal("expected non-nil definition")
				}
				if d.ID.VendorID != "" || d.ID.AppID != "" {
					t.Error("expected empty ID fields for null JSON")
				}
			},
		},
		{
			name:    "empty object returns empty definition",
			input:   []byte(`{}`),
			wantErr: false,
			validate: func(t *testing.T, d *Definition) {
				if d == nil {
					t.Fatal("expected non-nil definition")
				}
				if d.ID.VendorID != "" || d.ID.AppID != "" {
					t.Error("expected empty ID fields")
				}
			},
		},
		{
			name:  "minimal definition with ID and name",
			input: []byte(`{"id":{"vendorID":"test-vendor","appID":"test-app"},"name":{"fallback":"Test App"}}`),
			validate: func(t *testing.T, d *Definition) {
				if d.ID.VendorID != "test-vendor" {
					t.Errorf("vendorID = %q, want %q", d.ID.VendorID, "test-vendor")
				}
				if d.ID.AppID != "test-app" {
					t.Errorf("appID = %q, want %q", d.ID.AppID, "test-app")
				}
				if d.Name.Fallback != "Test App" {
					t.Errorf("name fallback = %q, want %q", d.Name.Fallback, "Test App")
				}
			},
		},
		{
			name: "full definition with nested types",
			input: func() []byte {
				d := Definition{
					ID:          NewID("acme", "billing"),
					Endpoint:    "https://billing.example.com",
					MCPEndpoint: "https://mcp.example.com",
					APIEndpoint: "https://api.example.com",
					DefaultPath: "/dashboard",
					Name:        translation.String("Billing App"),
					Description: translation.String("Manages billing"),
					UIMode:      UIModeFull,
					Category:    CategoryBilling,
					Icon:        "payments",
					Permissions: []Permission{
						{Key: "manage", Name: translation.String("Manage")},
					},
					Paths: []Path{
						{ID: "home", Path: "/"},
						{ID: "invoice", Path: "/invoices/{id}"},
					},
					PrefixRedirect: map[string]string{"INV": "invoices/$1"},
					QuickCodes:     map[string]string{"INV": "invoices/$1"},
					SystemApp:      true,
					Homepage:       "https://example.com",
					SupportEmail:   "support@example.com",
				}
				b, _ := json.Marshal(d)
				return b
			}(),
			validate: func(t *testing.T, d *Definition) {
				if d.ID.VendorID != "acme" || d.ID.AppID != "billing" {
					t.Errorf("ID = %v, want acme/billing", d.ID)
				}
				if d.Endpoint != "https://billing.example.com" {
					t.Errorf("Endpoint = %q", d.Endpoint)
				}
				if d.MCPEndpoint != "https://mcp.example.com" {
					t.Errorf("MCPEndpoint = %q", d.MCPEndpoint)
				}
				if d.APIEndpoint != "https://api.example.com" {
					t.Errorf("APIEndpoint = %q", d.APIEndpoint)
				}
				if d.DefaultPath != "/dashboard" {
					t.Errorf("DefaultPath = %q", d.DefaultPath)
				}
				if d.UIMode != UIModeFull {
					t.Errorf("UIMode = %q, want %q", d.UIMode, UIModeFull)
				}
				if d.Category != CategoryBilling {
					t.Errorf("Category = %q, want %q", d.Category, CategoryBilling)
				}
				if len(d.Permissions) != 1 || d.Permissions[0].Key != "manage" {
					t.Errorf("Permissions = %v", d.Permissions)
				}
				if len(d.Paths) != 2 {
					t.Errorf("Paths count = %d, want 2", len(d.Paths))
				}
				if d.PrefixRedirect["INV"] != "invoices/$1" {
					t.Errorf("PrefixRedirect[INV] = %q", d.PrefixRedirect["INV"])
				}
				if !d.SystemApp {
					t.Error("expected SystemApp to be true")
				}
				if d.Homepage != "https://example.com" {
					t.Errorf("Homepage = %q", d.Homepage)
				}
				if d.SupportEmail != "support@example.com" {
					t.Errorf("SupportEmail = %q", d.SupportEmail)
				}
			},
		},
		{
			name:  "unknown fields are silently ignored",
			input: []byte(`{"id":{"vendorID":"v","appID":"a"},"unknownField":"value"}`),
			validate: func(t *testing.T, d *Definition) {
				if d.ID.VendorID != "v" {
					t.Errorf("vendorID = %q, want %q", d.ID.VendorID, "v")
				}
			},
		},
		{
			name:  "boolean fields default to false",
			input: []byte(`{"id":{"vendorID":"v","appID":"a"}}`),
			validate: func(t *testing.T, d *Definition) {
				if d.SystemApp {
					t.Error("expected SystemApp to default to false")
				}
				if d.SkipCSRFValidation {
					t.Error("expected SkipCSRFValidation to default to false")
				}
				if d.ProvideBlueprints {
					t.Error("expected ProvideBlueprints to default to false")
				}
				if d.ProvideHelp {
					t.Error("expected ProvideHelp to default to false")
				}
			},
		},
		{
			name:  "definition with translations",
			input: []byte(`{"name":{"fallback":"Hello","translations":{"fr":"Bonjour","de":"Hallo"}}}`),
			validate: func(t *testing.T, d *Definition) {
				if d.Name.Fallback != "Hello" {
					t.Errorf("name fallback = %q, want %q", d.Name.Fallback, "Hello")
				}
				if d.Name.Get("fr") != "Bonjour" {
					t.Errorf("name fr = %q, want %q", d.Name.Get("fr"), "Bonjour")
				}
				if d.Name.Get("de") != "Hallo" {
					t.Errorf("name de = %q, want %q", d.Name.Get("de"), "Hallo")
				}
				// Missing language falls back
				if d.Name.Get("es") != "Hello" {
					t.Errorf("name es fallback = %q, want %q", d.Name.Get("es"), "Hello")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def, err := FromJson(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.validate != nil {
				tt.validate(t, def)
			}
		})
	}
}

func TestGetHash(t *testing.T) {
	tests := []struct {
		name          string
		def           *Definition
		updateIfEmpty bool
		wantEmpty     bool
		wantStored    bool // whether the hash should be stored on the definition after the call
	}{
		{
			name:      "nil receiver returns empty string",
			def:       nil,
			wantEmpty: true,
		},
		{
			name:          "empty hash computes and stores when updateIfEmpty is true",
			def:           &Definition{Name: translation.String("Test")},
			updateIfEmpty: true,
			wantStored:    true,
		},
		{
			name:          "empty hash computes but does not store when updateIfEmpty is false",
			def:           &Definition{Name: translation.String("Test")},
			updateIfEmpty: false,
			wantStored:    false,
		},
		{
			name:          "pre-set hash returns as-is without recomputing",
			def:           &Definition{Hash: "preset-hash-value"},
			updateIfEmpty: true,
			wantEmpty:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.def.GetHash(tt.updateIfEmpty)

			if tt.wantEmpty {
				if result != "" {
					t.Errorf("GetHash() = %q, want empty string", result)
				}
				return
			}

			if tt.def != nil && tt.def.Hash == "preset-hash-value" {
				if result != "preset-hash-value" {
					t.Errorf("GetHash() = %q, want %q", result, "preset-hash-value")
				}
				return
			}

			if result == "" {
				t.Error("GetHash() returned empty string, expected computed hash")
			}

			if tt.wantStored {
				if tt.def.Hash == "" {
					t.Error("expected hash to be stored on definition")
				}
				if tt.def.Hash != result {
					t.Errorf("stored hash %q does not match returned hash %q", tt.def.Hash, result)
				}
			} else {
				if tt.def.Hash != "" {
					t.Errorf("expected hash to remain empty, got %q", tt.def.Hash)
				}
			}
		})
	}

	t.Run("same definition produces same hash", func(t *testing.T) {
		d1 := &Definition{Name: translation.String("Same")}
		d2 := &Definition{Name: translation.String("Same")}
		h1 := d1.GetHash(false)
		h2 := d2.GetHash(false)
		if h1 != h2 {
			t.Errorf("identical definitions produced different hashes: %q vs %q", h1, h2)
		}
	})

	t.Run("different definitions produce different hashes", func(t *testing.T) {
		d1 := &Definition{Name: translation.String("App A")}
		d2 := &Definition{Name: translation.String("App B")}
		h1 := d1.GetHash(false)
		h2 := d2.GetHash(false)
		if h1 == h2 {
			t.Errorf("different definitions produced same hash: %q", h1)
		}
	})

	t.Run("hash is 32 character hex string", func(t *testing.T) {
		d := &Definition{Name: translation.String("Test")}
		h := d.GetHash(false)
		if len(h) != 32 {
			t.Errorf("hash length = %d, want 32", len(h))
		}
		for _, c := range h {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
				t.Errorf("hash contains non-hex character: %c", c)
			}
		}
	})

	t.Run("computing hash with updateIfEmpty true is idempotent", func(t *testing.T) {
		d := &Definition{Name: translation.String("Stable")}
		h1 := d.GetHash(true)
		h2 := d.GetHash(true) // now Hash is set, should return stored value
		if h1 != h2 {
			t.Errorf("second call returned different hash: %q vs %q", h1, h2)
		}
	})
}

func TestWithPath(t *testing.T) {
	tests := []struct {
		name       string
		initial    []Path
		add        []Path
		wantCount  int
		wantLastID string
	}{
		{
			name:       "append single path to empty definition",
			initial:    nil,
			add:        []Path{{ID: "home", Path: "/"}},
			wantCount:  1,
			wantLastID: "home",
		},
		{
			name:       "append multiple paths at once",
			initial:    nil,
			add:        []Path{{ID: "a", Path: "/a"}, {ID: "b", Path: "/b"}, {ID: "c", Path: "/c"}},
			wantCount:  3,
			wantLastID: "c",
		},
		{
			name:       "append to existing paths",
			initial:    []Path{{ID: "existing", Path: "/existing"}},
			add:        []Path{{ID: "new", Path: "/new"}},
			wantCount:  2,
			wantLastID: "new",
		},
		{
			name:       "append no paths is a no-op",
			initial:    []Path{{ID: "only", Path: "/only"}},
			add:        []Path{},
			wantCount:  1,
			wantLastID: "only",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Definition{Paths: tt.initial}
			result := d.WithPath(tt.add...)

			if result != d {
				t.Error("WithPath should return the same definition pointer for chaining")
			}
			if len(d.Paths) != tt.wantCount {
				t.Errorf("path count = %d, want %d", len(d.Paths), tt.wantCount)
			}
			if tt.wantCount > 0 {
				lastID := d.Paths[len(d.Paths)-1].ID
				if lastID != tt.wantLastID {
					t.Errorf("last path ID = %q, want %q", lastID, tt.wantLastID)
				}
			}
		})
	}

	t.Run("chaining multiple WithPath calls", func(t *testing.T) {
		d := &Definition{}
		d.WithPath(Path{ID: "a", Path: "/a"}).
			WithPath(Path{ID: "b", Path: "/b"}).
			WithPath(Path{ID: "c", Path: "/c"})

		if len(d.Paths) != 3 {
			t.Errorf("path count = %d, want 3", len(d.Paths))
		}
		ids := []string{d.Paths[0].ID, d.Paths[1].ID, d.Paths[2].ID}
		want := []string{"a", "b", "c"}
		for i, id := range ids {
			if id != want[i] {
				t.Errorf("path[%d].ID = %q, want %q", i, id, want[i])
			}
		}
	})
}

func TestVendorFromJson(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		wantErr  bool
		validate func(t *testing.T, v *Vendor)
	}{
		{
			name:    "invalid JSON returns error",
			input:   []byte(`{invalid`),
			wantErr: true,
		},
		{
			name:    "empty bytes returns error",
			input:   []byte(``),
			wantErr: true,
		},
		{
			name:    "empty object returns empty vendor",
			input:   []byte(`{}`),
			wantErr: false,
			validate: func(t *testing.T, v *Vendor) {
				if v.ID != "" {
					t.Errorf("ID = %q, want empty", v.ID)
				}
			},
		},
		{
			name:  "valid vendor with all fields",
			input: []byte(`{"id":"acme","name":{"fallback":"Acme Corp"},"description":{"fallback":"Makes everything"}}`),
			validate: func(t *testing.T, v *Vendor) {
				if v.ID != "acme" {
					t.Errorf("ID = %q, want %q", v.ID, "acme")
				}
				if v.Name.Fallback != "Acme Corp" {
					t.Errorf("Name = %q, want %q", v.Name.Fallback, "Acme Corp")
				}
				if v.Description.Fallback != "Makes everything" {
					t.Errorf("Description = %q, want %q", v.Description.Fallback, "Makes everything")
				}
			},
		},
		{
			name:  "vendor with translations",
			input: []byte(`{"id":"test","name":{"fallback":"Test","translations":{"fr":"Tester"}},"description":{"fallback":"Desc"}}`),
			validate: func(t *testing.T, v *Vendor) {
				if v.Name.Get("fr") != "Tester" {
					t.Errorf("Name fr = %q, want %q", v.Name.Get("fr"), "Tester")
				}
				if v.Name.Get("en") != "Test" {
					t.Errorf("Name en fallback = %q, want %q", v.Name.Get("en"), "Test")
				}
			},
		},
		{
			name:    "array JSON returns error",
			input:   []byte(`[1,2,3]`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := VendorFromJson(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.validate != nil {
				tt.validate(t, v)
			}
		})
	}
}
