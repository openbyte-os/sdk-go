package app

import (
	"testing"

	"github.com/openbyte-os/sdk-go/translation"
)

func TestNewPath(t *testing.T) {
	p := NewPath("home", "/home")

	if p == nil {
		t.Fatal("NewPath returned nil")
	}
	if p.ID != "home" {
		t.Errorf("ID = %q, want %q", p.ID, "home")
	}
	if p.Path != "/home" {
		t.Errorf("Path = %q, want %q", p.Path, "/home")
	}
}

func TestNewPath_EmptyValues(t *testing.T) {
	p := NewPath("", "")
	if p.ID != "" {
		t.Errorf("ID = %q, want empty", p.ID)
	}
	if p.Path != "" {
		t.Errorf("Path = %q, want empty", p.Path)
	}
}

func TestPath_WithNavigation(t *testing.T) {
	p := NewPath("test", "/test")
	nav := EntryPoint{DestinationPath: "/sub", Text: translation.String("Sub")}
	result := p.WithNavigation(nav)

	if result != p {
		t.Error("WithNavigation should return the same pointer for chaining")
	}
	if len(p.Navigation) != 1 {
		t.Fatalf("Navigation length = %d, want 1", len(p.Navigation))
	}
	if p.Navigation[0].DestinationPath != "/sub" {
		t.Errorf("Navigation[0].DestinationPath = %q, want %q", p.Navigation[0].DestinationPath, "/sub")
	}
}

func TestPath_WithNavigation_Appends(t *testing.T) {
	p := NewPath("test", "/test")
	nav1 := EntryPoint{DestinationPath: "/a"}
	nav2 := EntryPoint{DestinationPath: "/b"}

	p.WithNavigation(nav1)
	p.WithNavigation(nav2)

	if len(p.Navigation) != 2 {
		t.Fatalf("Navigation length = %d, want 2", len(p.Navigation))
	}
	if p.Navigation[0].DestinationPath != "/a" {
		t.Errorf("Navigation[0] = %q, want /a", p.Navigation[0].DestinationPath)
	}
	if p.Navigation[1].DestinationPath != "/b" {
		t.Errorf("Navigation[1] = %q, want /b", p.Navigation[1].DestinationPath)
	}
}

func TestPath_WithNavigation_Variadic(t *testing.T) {
	p := NewPath("test", "/test")
	p.WithNavigation(
		EntryPoint{DestinationPath: "/a"},
		EntryPoint{DestinationPath: "/b"},
		EntryPoint{DestinationPath: "/c"},
	)
	if len(p.Navigation) != 3 {
		t.Fatalf("Navigation length = %d, want 3", len(p.Navigation))
	}
}

func TestPath_WithRequestPermissions(t *testing.T) {
	p := NewPath("test", "/test")
	perm := ScopedKey{Key: "read"}
	result := p.WithRequestPermissions(perm)

	if result != p {
		t.Error("WithRequestPermissions should return the same pointer")
	}
	if len(p.RequestPermissions) != 1 {
		t.Fatalf("RequestPermissions length = %d, want 1", len(p.RequestPermissions))
	}
	if p.RequestPermissions[0].Key != "read" {
		t.Errorf("RequestPermissions[0].Key = %q, want %q", p.RequestPermissions[0].Key, "read")
	}
}

func TestPath_WithRequiredPermissions(t *testing.T) {
	p := NewPath("test", "/test")
	perm := ScopedKey{Key: "admin"}
	result := p.WithRequiredPermissions(perm)

	if result != p {
		t.Error("WithRequiredPermissions should return the same pointer")
	}
	if len(p.RequiredPermissions) != 1 {
		t.Fatalf("RequiredPermissions length = %d, want 1", len(p.RequiredPermissions))
	}
	if p.RequiredPermissions[0].Key != "admin" {
		t.Errorf("RequiredPermissions[0].Key = %q, want %q", p.RequiredPermissions[0].Key, "admin")
	}
}

func TestPath_WithActions(t *testing.T) {
	p := NewPath("test", "/test")
	action := EntryPoint{DestinationPath: "/action", Icon: "save"}
	result := p.WithActions(action)

	if result != p {
		t.Error("WithActions should return the same pointer")
	}
	if len(p.Actions) != 1 {
		t.Fatalf("Actions length = %d, want 1", len(p.Actions))
	}
	if p.Actions[0].Icon != "save" {
		t.Errorf("Actions[0].Icon = %q, want %q", p.Actions[0].Icon, "save")
	}
}

func TestPath_WithActions_Appends(t *testing.T) {
	p := NewPath("test", "/test")
	p.WithActions(EntryPoint{Icon: "save"})
	p.WithActions(EntryPoint{Icon: "delete"})

	if len(p.Actions) != 2 {
		t.Fatalf("Actions length = %d, want 2", len(p.Actions))
	}
}

func TestPath_WithMethod(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", ""}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			p := NewPath("test", "/test")
			result := p.WithMethod(method)

			if result != p {
				t.Error("WithMethod should return the same pointer")
			}
			if p.Method != method {
				t.Errorf("Method = %q, want %q", p.Method, method)
			}
		})
	}
}

func TestPath_WithName(t *testing.T) {
	p := NewPath("test", "/test")
	name := translation.String("My Path")
	result := p.WithName(name)

	if result != p {
		t.Error("WithName should return the same pointer")
	}
	if p.Name.Fallback != "My Path" {
		t.Errorf("Name.Fallback = %q, want %q", p.Name.Fallback, "My Path")
	}
}

func TestPath_WithDescription(t *testing.T) {
	p := NewPath("test", "/test")
	desc := translation.String("A description")
	result := p.WithDescription(desc)

	if result != p {
		t.Error("WithDescription should return the same pointer")
	}
	if p.Description.Fallback != "A description" {
		t.Errorf("Description.Fallback = %q, want %q", p.Description.Fallback, "A description")
	}
}

func TestPath_WithoutBreadCrumbs(t *testing.T) {
	p := NewPath("test", "/test")
	if p.HideBreadcrumb {
		t.Error("HideBreadcrumb should default to false")
	}

	result := p.WithoutBreadCrumbs()
	if result != p {
		t.Error("WithoutBreadCrumbs should return the same pointer")
	}
	if !p.HideBreadcrumb {
		t.Error("HideBreadcrumb should be true after WithoutBreadCrumbs")
	}
}

func TestPath_WithoutHeader(t *testing.T) {
	p := NewPath("test", "/test")
	if p.HideHeader {
		t.Error("HideHeader should default to false")
	}

	result := p.WithoutHeader()
	if result != p {
		t.Error("WithoutHeader should return the same pointer")
	}
	if !p.HideHeader {
		t.Error("HideHeader should be true after WithoutHeader")
	}
}

func TestPath_WithKiosk(t *testing.T) {
	p := NewPath("test", "/test")
	if p.Kiosk {
		t.Error("Kiosk should default to false")
	}

	result := p.WithKiosk()
	if result != p {
		t.Error("WithKiosk should return the same pointer")
	}
	if !p.Kiosk {
		t.Error("Kiosk should be true after WithKiosk")
	}
}

func TestPath_FullChain(t *testing.T) {
	p := NewPath("dashboard", "/dashboard").
		WithMethod("GET").
		WithName(translation.String("Dashboard")).
		WithDescription(translation.String("Main dashboard")).
		WithNavigation(EntryPoint{DestinationPath: "/sub"}).
		WithActions(EntryPoint{Icon: "refresh"}).
		WithRequestPermissions(ScopedKey{Key: "read"}).
		WithRequiredPermissions(ScopedKey{Key: "admin"}).
		WithoutBreadCrumbs().
		WithoutHeader().
		WithKiosk()

	if p.ID != "dashboard" {
		t.Errorf("ID = %q, want dashboard", p.ID)
	}
	if p.Method != "GET" {
		t.Errorf("Method = %q, want GET", p.Method)
	}
	if p.Name.Fallback != "Dashboard" {
		t.Errorf("Name = %q, want Dashboard", p.Name.Fallback)
	}
	if p.Description.Fallback != "Main dashboard" {
		t.Errorf("Description = %q, want Main dashboard", p.Description.Fallback)
	}
	if len(p.Navigation) != 1 {
		t.Errorf("Navigation length = %d, want 1", len(p.Navigation))
	}
	if len(p.Actions) != 1 {
		t.Errorf("Actions length = %d, want 1", len(p.Actions))
	}
	if len(p.RequestPermissions) != 1 {
		t.Errorf("RequestPermissions length = %d, want 1", len(p.RequestPermissions))
	}
	if len(p.RequiredPermissions) != 1 {
		t.Errorf("RequiredPermissions length = %d, want 1", len(p.RequiredPermissions))
	}
	if !p.HideBreadcrumb {
		t.Error("HideBreadcrumb should be true")
	}
	if !p.HideHeader {
		t.Error("HideHeader should be true")
	}
	if !p.Kiosk {
		t.Error("Kiosk should be true")
	}
}
