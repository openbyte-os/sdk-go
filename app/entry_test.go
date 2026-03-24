package app

import (
	"testing"

	"github.com/openbyte-os/sdk-go/translation"
)

func TestNewEntryPoint(t *testing.T) {
	text := translation.String("Dashboard")
	ep := NewEntryPoint("/dashboard", text)

	if ep == nil {
		t.Fatal("NewEntryPoint returned nil")
	}
	if ep.DestinationPath != "/dashboard" {
		t.Errorf("DestinationPath = %q, want %q", ep.DestinationPath, "/dashboard")
	}
	if ep.Text.Fallback != "Dashboard" {
		t.Errorf("Text.Fallback = %q, want %q", ep.Text.Fallback, "Dashboard")
	}
}

func TestNewEntryPoint_EmptyValues(t *testing.T) {
	ep := NewEntryPoint("", translation.Text{})

	if ep == nil {
		t.Fatal("NewEntryPoint returned nil")
	}
	if ep.DestinationPath != "" {
		t.Errorf("DestinationPath = %q, want empty", ep.DestinationPath)
	}
}

func TestEntryPoint_WithLaunchMode(t *testing.T) {
	modes := []LaunchMode{
		LaunchModePage,
		LaunchModeModal,
		LaunchModeWindow,
		LaunchModeSlide,
		LaunchModeActionDrop,
		LaunchModeActionFill,
	}

	for _, mode := range modes {
		t.Run(string(mode), func(t *testing.T) {
			ep := NewEntryPoint("/test", translation.String("Test"))
			result := ep.WithLaunchMode(mode)

			if result != ep {
				t.Error("WithLaunchMode should return the same pointer for chaining")
			}
			if ep.LaunchMode != mode {
				t.Errorf("LaunchMode = %q, want %q", ep.LaunchMode, mode)
			}
		})
	}
}

func TestEntryPoint_WithIcon(t *testing.T) {
	ep := NewEntryPoint("/test", translation.String("Test"))
	result := ep.WithIcon("settings")

	if result != ep {
		t.Error("WithIcon should return the same pointer for chaining")
	}
	if ep.Icon != "settings" {
		t.Errorf("Icon = %q, want %q", ep.Icon, "settings")
	}
}

func TestEntryPoint_WithIcon_Empty(t *testing.T) {
	ep := NewEntryPoint("/test", translation.String("Test"))
	ep.WithIcon("")

	if ep.Icon != "" {
		t.Errorf("Icon = %q, want empty", ep.Icon)
	}
}

func TestEntryPoint_WithTitle(t *testing.T) {
	ep := NewEntryPoint("/test", translation.String("Test"))
	title := translation.String("My Title")
	result := ep.WithTitle(title)

	if result != ep {
		t.Error("WithTitle should return the same pointer for chaining")
	}
	if ep.Title.Fallback != "My Title" {
		t.Errorf("Title.Fallback = %q, want %q", ep.Title.Fallback, "My Title")
	}
}

func TestEntryPoint_Chaining(t *testing.T) {
	ep := NewEntryPoint("/home", translation.String("Home")).
		WithIcon("home").
		WithLaunchMode(LaunchModeModal).
		WithTitle(translation.String("Home Page"))

	if ep.DestinationPath != "/home" {
		t.Errorf("DestinationPath = %q, want %q", ep.DestinationPath, "/home")
	}
	if ep.Icon != "home" {
		t.Errorf("Icon = %q, want %q", ep.Icon, "home")
	}
	if ep.LaunchMode != LaunchModeModal {
		t.Errorf("LaunchMode = %q, want %q", ep.LaunchMode, LaunchModeModal)
	}
	if ep.Title.Fallback != "Home Page" {
		t.Errorf("Title.Fallback = %q, want %q", ep.Title.Fallback, "Home Page")
	}
}

func TestEntryPoint_OverwriteValues(t *testing.T) {
	ep := NewEntryPoint("/test", translation.String("Test")).
		WithIcon("icon1").
		WithIcon("icon2")

	if ep.Icon != "icon2" {
		t.Errorf("Icon = %q, want %q after overwrite", ep.Icon, "icon2")
	}
}
