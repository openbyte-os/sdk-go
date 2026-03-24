package app

import (
	"reflect"
	"testing"
)

func TestGlobalAppID_String(t *testing.T) {
	tests := []struct {
		name     string
		gaid     GlobalAppID
		expected string
	}{
		{"basic", GlobalAppID{VendorID: "vendor1", AppID: "app1"}, "vendor1/app1"},
		{"empty fields", GlobalAppID{}, "/"},
		{"empty vendor", GlobalAppID{AppID: "app1"}, "/app1"},
		{"empty app", GlobalAppID{VendorID: "vendor1"}, "vendor1/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.gaid.String()
			if got != tt.expected {
				t.Fatalf("String() = %q, expected %q", got, tt.expected)
			}
		})
	}
}

func TestGlobalAppID_Tertiary(t *testing.T) {
	tests := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{"set and get tertiary", func(t *testing.T) {
			g := NewID("v", "a")
			g.SetTertiary("extra")
			if g.Tertiary() != "extra" {
				t.Fatalf("Tertiary() = %q, expected %q", g.Tertiary(), "extra")
			}
		}},
		{"clear tertiary returns old value", func(t *testing.T) {
			g := NewID("v", "a")
			g.SetTertiary("old")
			cleared := g.ClearTertiary()
			if cleared != "old" {
				t.Fatalf("ClearTertiary() = %q, expected %q", cleared, "old")
			}
			if g.Tertiary() != "" {
				t.Fatalf("Tertiary() after clear = %q, expected empty", g.Tertiary())
			}
		}},
		{"clear empty tertiary", func(t *testing.T) {
			g := NewID("v", "a")
			cleared := g.ClearTertiary()
			if cleared != "" {
				t.Fatalf("ClearTertiary() = %q, expected empty", cleared)
			}
		}},
		{"overwrite tertiary", func(t *testing.T) {
			g := NewID("v", "a")
			g.SetTertiary("first")
			g.SetTertiary("second")
			if g.Tertiary() != "second" {
				t.Fatalf("Tertiary() = %q, expected %q", g.Tertiary(), "second")
			}
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.fn)
	}
}

func TestGlobalAppID_AsPath(t *testing.T) {
	tests := []struct {
		name     string
		gaid     GlobalAppID
		tertiary string
		expected string
	}{
		{"with tertiary", GlobalAppID{VendorID: "v", AppID: "a"}, "extra", "v/a/extra"},
		{"without tertiary", GlobalAppID{VendorID: "v", AppID: "a"}, "", "v/a/"},
		{"multi-segment tertiary", GlobalAppID{VendorID: "v", AppID: "a"}, "x/y/z", "v/a/x/y/z"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.gaid.SetTertiary(tt.tertiary)
			got := tt.gaid.AsPath()
			if got != tt.expected {
				t.Fatalf("AsPath() = %q, expected %q", got, tt.expected)
			}
		})
	}
}

func TestGlobalAppID_Validate(t *testing.T) {
	tests := []struct {
		name    string
		gaid    GlobalAppID
		strict  bool
		wantErr bool
	}{
		{"valid non-strict", GlobalAppID{VendorID: "my-vendor", AppID: "my-app"}, false, false},
		{"valid strict", GlobalAppID{VendorID: "my-vendor", AppID: "my-app"}, true, false},
		{"valid with tertiary non-strict", GlobalAppID{VendorID: "my-vendor", AppID: "my-app", tertiary: "extra"}, false, false},
		{"tertiary strict fails", GlobalAppID{VendorID: "my-vendor", AppID: "my-app", tertiary: "extra"}, true, true},
		{"invalid vendor", GlobalAppID{VendorID: "A", AppID: "my-app"}, false, true},
		{"invalid app", GlobalAppID{VendorID: "my-vendor", AppID: "B"}, false, true},
		{"empty vendor", GlobalAppID{VendorID: "", AppID: "my-app"}, false, true},
		{"empty app", GlobalAppID{VendorID: "my-vendor", AppID: ""}, false, true},
		{"both empty", GlobalAppID{}, false, true},
		{"vendor with uppercase", GlobalAppID{VendorID: "Vendor", AppID: "my-app"}, false, true},
		{"vendor with spaces", GlobalAppID{VendorID: "my vendor", AppID: "my-app"}, false, true},
		{"vendor starting with hyphen", GlobalAppID{VendorID: "-vendor", AppID: "my-app"}, false, true},
		{"vendor ending with hyphen", GlobalAppID{VendorID: "vendor-", AppID: "my-app"}, false, true},
		{"app starting with hyphen", GlobalAppID{VendorID: "my-vendor", AppID: "-app"}, false, true},
		{"numeric ids", GlobalAppID{VendorID: "123", AppID: "456"}, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.gaid.Validate(tt.strict)
			if tt.wantErr && err == nil {
				t.Fatalf("Validate(%v) expected error, got nil", tt.strict)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("Validate(%v) unexpected error: %v", tt.strict, err)
			}
		})
	}
}

func TestGlobalAppID_Matches(t *testing.T) {
	tests := []struct {
		name     string
		a        GlobalAppID
		b        GlobalAppID
		strict   bool
		expected bool
	}{
		{"same ids non-strict", NewID("v", "a"), NewID("v", "a"), false, true},
		{"same ids strict no tertiary", NewID("v", "a"), NewID("v", "a"), true, true},
		{"different vendor", NewID("v1", "a"), NewID("v2", "a"), false, false},
		{"different app", NewID("v", "a1"), NewID("v", "a2"), false, false},
		{"same ids different tertiary non-strict", func() GlobalAppID {
			g := NewID("v", "a")
			g.SetTertiary("x")
			return g
		}(), func() GlobalAppID {
			g := NewID("v", "a")
			g.SetTertiary("y")
			return g
		}(), false, true},
		{"same ids different tertiary strict", func() GlobalAppID {
			g := NewID("v", "a")
			g.SetTertiary("x")
			return g
		}(), func() GlobalAppID {
			g := NewID("v", "a")
			g.SetTertiary("y")
			return g
		}(), true, false},
		{"same ids same tertiary strict", func() GlobalAppID {
			g := NewID("v", "a")
			g.SetTertiary("x")
			return g
		}(), func() GlobalAppID {
			g := NewID("v", "a")
			g.SetTertiary("x")
			return g
		}(), true, true},
		{"both empty", GlobalAppID{}, GlobalAppID{}, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.a.Matches(tt.b, tt.strict)
			if got != tt.expected {
				t.Fatalf("Matches() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestNewID(t *testing.T) {
	g := NewID("my-vendor", "my-app")
	if g.VendorID != "my-vendor" {
		t.Fatalf("VendorID = %q, expected %q", g.VendorID, "my-vendor")
	}
	if g.AppID != "my-app" {
		t.Fatalf("AppID = %q, expected %q", g.AppID, "my-app")
	}
	if g.Tertiary() != "" {
		t.Fatalf("Tertiary() = %q, expected empty", g.Tertiary())
	}
}

func TestIDFromString(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantVendor   string
		wantApp      string
		wantTertiary string
	}{
		{"vendor and app", "vendor/app", "vendor", "app", ""},
		{"vendor app tertiary", "vendor/app/extra", "vendor", "app", "extra"},
		{"tertiary with slashes", "vendor/app/a/b/c", "vendor", "app", "a/b/c"},
		{"no slash", "onlyvendor", "", "", ""},
		{"empty string", "", "", "", ""},
		{"single slash", "/", "", "", ""},
		{"trailing slash", "vendor/app/", "vendor", "app", ""},
		{"empty vendor with app", "/app", "", "app", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IDFromString(tt.input)
			if got.VendorID != tt.wantVendor {
				t.Fatalf("VendorID = %q, expected %q", got.VendorID, tt.wantVendor)
			}
			if got.AppID != tt.wantApp {
				t.Fatalf("AppID = %q, expected %q", got.AppID, tt.wantApp)
			}
			if got.Tertiary() != tt.wantTertiary {
				t.Fatalf("Tertiary() = %q, expected %q", got.Tertiary(), tt.wantTertiary)
			}
		})
	}
}

func TestCreateID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"lowercase passthrough", "hello-world", "hello-world"},
		{"uppercase to lower", "Hello-World", "hello-world"},
		{"spaces to hyphens", "hello world", "hello-world"},
		{"special chars to hyphens", "hello@world!", "hello-world"},
		{"multiple special chars collapsed", "hello!!!world", "hello-world"},
		{"leading special chars trimmed", "---hello", "hello"},
		{"trailing special chars trimmed", "hello---", "hello"},
		{"leading and trailing trimmed", "!!hello!!", "hello"},
		{"all special", "!!!", ""},
		{"empty string", "", ""},
		{"numeric", "123", "123"},
		{"mixed alphanum and special", "My App v2.0", "my-app-v2-0"},
		{"underscores replaced", "my_app_id", "my-app-id"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CreateID(tt.input)
			if got != tt.expected {
				t.Fatalf("CreateID(%q) = %q, expected %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestValidateID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid simple", "abc", false},
		{"valid with hyphens", "my-app-id", false},
		{"valid numeric", "123", false},
		{"valid alphanumeric mix", "a1b2c3", false},
		{"valid start number end letter", "1abc", false},
		{"invalid uppercase", "ABC", true},
		{"invalid start hyphen", "-abc", true},
		{"invalid end hyphen", "abc-", true},
		{"invalid single char", "a", true},
		{"invalid two chars", "ab", true},
		{"invalid empty", "", true},
		{"invalid spaces", "a b c", true},
		{"invalid special chars", "a@b", true},
		{"invalid underscore", "a_b", true},
		{"valid min length three", "a-b", false},
		{"valid long id", "abcdefghijklmnop-1234567890", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateID(tt.input)
			if tt.wantErr && err == nil {
				t.Fatalf("ValidateID(%q) expected error, got nil", tt.input)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("ValidateID(%q) unexpected error: %v", tt.input, err)
			}
			if tt.wantErr && err != nil && err != ErrInvalidID {
				t.Fatalf("ValidateID(%q) error = %v, expected ErrInvalidID", tt.input, err)
			}
		})
	}
}

func TestScopedKey_String(t *testing.T) {
	tests := []struct {
		name     string
		sk       ScopedKey
		expected string
	}{
		{"basic", ScopedKey{GlobalAppID: NewID("v", "a"), Key: "k"}, "v/a/k"},
		{"empty key", ScopedKey{GlobalAppID: NewID("v", "a"), Key: ""}, "v/a/"},
		{"empty gaid", ScopedKey{Key: "k"}, "//k"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.sk.String()
			if got != tt.expected {
				t.Fatalf("String() = %q, expected %q", got, tt.expected)
			}
		})
	}
}

func TestScopedKeyFromString(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantVendor string
		wantApp    string
		wantKey    string
	}{
		{"full path", "vendor/app/mykey", "vendor", "app", "mykey"},
		{"no key", "vendor/app", "vendor", "app", ""},
		{"empty", "", "", "", ""},
		{"no slash", "onlyone", "", "", ""},
		{"key with slashes preserved as single key", "vendor/app/a/b", "vendor", "app", "a/b"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ScopedKeyFromString(tt.input)
			if got.VendorID != tt.wantVendor {
				t.Fatalf("VendorID = %q, expected %q", got.VendorID, tt.wantVendor)
			}
			if got.AppID != tt.wantApp {
				t.Fatalf("AppID = %q, expected %q", got.AppID, tt.wantApp)
			}
			if got.Key != tt.wantKey {
				t.Fatalf("Key = %q, expected %q", got.Key, tt.wantKey)
			}
			// tertiary should be cleared after extraction
			if got.Tertiary() != "" {
				t.Fatalf("Tertiary() = %q, expected empty after ScopedKeyFromString", got.Tertiary())
			}
		})
	}
}

func TestNewScopedKey(t *testing.T) {
	t.Run("with gaid", func(t *testing.T) {
		gaid := NewID("v", "a")
		sk := NewScopedKey("mykey", &gaid)
		expected := ScopedKey{GlobalAppID: NewID("v", "a"), Key: "mykey"}
		if !reflect.DeepEqual(sk, expected) {
			t.Fatalf("NewScopedKey() = %+v, expected %+v", sk, expected)
		}
	})

	t.Run("with nil gaid", func(t *testing.T) {
		sk := NewScopedKey("mykey", nil)
		if sk.Key != "mykey" {
			t.Fatalf("Key = %q, expected %q", sk.Key, "mykey")
		}
		if sk.VendorID != "" || sk.AppID != "" {
			t.Fatalf("expected empty GlobalAppID, got VendorID=%q AppID=%q", sk.VendorID, sk.AppID)
		}
	})

	t.Run("empty key with gaid", func(t *testing.T) {
		gaid := NewID("v", "a")
		sk := NewScopedKey("", &gaid)
		if sk.Key != "" {
			t.Fatalf("Key = %q, expected empty", sk.Key)
		}
		if sk.VendorID != "v" {
			t.Fatalf("VendorID = %q, expected %q", sk.VendorID, "v")
		}
	})
}

func TestGlobalAppID_RoundTrip(t *testing.T) {
	// Test that IDFromString -> String produces consistent results
	tests := []struct {
		name  string
		input string
		str   string
	}{
		{"basic round trip", "vendor/app", "vendor/app"},
		{"tertiary stripped from String", "vendor/app/extra", "vendor/app"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IDFromString(tt.input).String()
			if got != tt.str {
				t.Fatalf("IDFromString(%q).String() = %q, expected %q", tt.input, got, tt.str)
			}
		})
	}
}

func TestGlobalAppID_RoundTripAsPath(t *testing.T) {
	input := "vendor/app/extra"
	got := IDFromString(input).AsPath()
	if got != input {
		t.Fatalf("IDFromString(%q).AsPath() = %q, expected %q", input, got, input)
	}
}
