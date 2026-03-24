package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/openbyte-os/sdk-go/transport"
)

func TestNewAlert(t *testing.T) {
	a := NewAlert("Title", "Description", "icon-name")
	if a.Title != "Title" {
		t.Errorf("Title = %q, want %q", a.Title, "Title")
	}
	if a.Description != "Description" {
		t.Errorf("Description = %q, want %q", a.Description, "Description")
	}
	if a.Icon != "icon-name" {
		t.Errorf("Icon = %q, want %q", a.Icon, "icon-name")
	}
	if a.Closer {
		t.Error("Closer should be false by default")
	}
	if a.Style != "" {
		t.Errorf("Style = %q, want empty", a.Style)
	}
	if len(a.Actions) != 0 {
		t.Errorf("Actions should be empty, got %d", len(a.Actions))
	}
}

func TestAlertStyleMethods(t *testing.T) {
	tests := []struct {
		name   string
		method func(*Alert) *Alert
		want   AlertStyle
	}{
		{"Info", (*Alert).Info, AlertStyleInfo},
		{"Warning", (*Alert).Warning, AlertStyleWarning},
		{"Success", (*Alert).Success, AlertStyleSuccess},
		{"Danger", (*Alert).Danger, AlertStyleDanger},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewAlert("t", "d", "i")
			result := tt.method(a)
			if result != a {
				t.Error("expected method to return same pointer for chaining")
			}
			if a.Style != tt.want {
				t.Errorf("Style = %q, want %q", a.Style, tt.want)
			}
		})
	}
}

func TestAlertWithClose(t *testing.T) {
	a := NewAlert("t", "d", "i")
	result := a.WithClose()
	if result != a {
		t.Error("expected WithClose to return same pointer")
	}
	if !a.Closer {
		t.Error("Closer should be true after WithClose()")
	}
}

func TestAlertWithAction(t *testing.T) {
	a := NewAlert("t", "d", "i")

	t.Run("add single action", func(t *testing.T) {
		btn := AlertButton{Type: "link", Content: "Click", Link: "/go", Uri: "/uri", Target: "_blank"}
		result := a.WithAction(btn)
		if result != a {
			t.Error("expected WithAction to return same pointer")
		}
		if len(a.Actions) != 1 {
			t.Fatalf("expected 1 action, got %d", len(a.Actions))
		}
		if a.Actions[0].Content != "Click" {
			t.Errorf("action Content = %q, want %q", a.Actions[0].Content, "Click")
		}
	})

	t.Run("add multiple actions at once", func(t *testing.T) {
		a2 := NewAlert("t", "d", "i")
		a2.WithAction(
			AlertButton{Content: "A"},
			AlertButton{Content: "B"},
		)
		if len(a2.Actions) != 2 {
			t.Fatalf("expected 2 actions, got %d", len(a2.Actions))
		}
	})

	t.Run("actions accumulate", func(t *testing.T) {
		a3 := NewAlert("t", "d", "i")
		a3.WithAction(AlertButton{Content: "A"})
		a3.WithAction(AlertButton{Content: "B"})
		if len(a3.Actions) != 2 {
			t.Fatalf("expected 2 actions, got %d", len(a3.Actions))
		}
	})
}

func TestAlertChaining(t *testing.T) {
	a := NewAlert("Title", "Desc", "icon").
		Info().
		WithClose().
		WithAction(AlertButton{Content: "OK"})

	if a.Style != AlertStyleInfo {
		t.Errorf("Style = %q, want %q", a.Style, AlertStyleInfo)
	}
	if !a.Closer {
		t.Error("expected Closer to be true")
	}
	if len(a.Actions) != 1 {
		t.Errorf("expected 1 action, got %d", len(a.Actions))
	}
}

func TestAlertWriteHeader(t *testing.T) {
	a := NewAlert("Test", "Desc", "icon").Success().WithClose()
	w := httptest.NewRecorder()
	a.WriteHeader(w)

	headerVal := w.Header().Get(transport.ResponseAlert)
	if headerVal == "" {
		t.Fatal("expected ResponseAlert header to be set")
	}

	var decoded Alert
	if err := json.Unmarshal([]byte(headerVal), &decoded); err != nil {
		t.Fatalf("failed to unmarshal header JSON: %v", err)
	}
	if decoded.Title != "Test" {
		t.Errorf("decoded Title = %q, want %q", decoded.Title, "Test")
	}
	if decoded.Style != AlertStyleSuccess {
		t.Errorf("decoded Style = %q, want %q", decoded.Style, AlertStyleSuccess)
	}
	if !decoded.Closer {
		t.Error("decoded Closer should be true")
	}
}

func TestAlertText(t *testing.T) {
	w := httptest.NewRecorder()
	AlertText(w, "hello")
	if got := w.Header().Get(transport.ResponseAlert); got != "hello" {
		t.Errorf("ResponseAlert = %q, want %q", got, "hello")
	}
}

func TestAlertInfo(t *testing.T) {
	w := httptest.NewRecorder()
	AlertInfo(w, "info msg")
	if got := w.Header().Get(transport.ResponseAlertInfo); got != "info msg" {
		t.Errorf("ResponseAlertInfo = %q, want %q", got, "info msg")
	}
}

func TestAlertSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	AlertSuccess(w, "success msg")
	if got := w.Header().Get(transport.ResponseAlertSuccess); got != "success msg" {
		t.Errorf("ResponseAlertSuccess = %q, want %q", got, "success msg")
	}
}

func TestAlertWarning(t *testing.T) {
	w := httptest.NewRecorder()
	AlertWarning(w, "warning msg")
	if got := w.Header().Get(transport.ResponseAlertWarning); got != "warning msg" {
		t.Errorf("ResponseAlertWarning = %q, want %q", got, "warning msg")
	}
}

func TestAlertDanger(t *testing.T) {
	w := httptest.NewRecorder()
	AlertDanger(w, "danger msg")
	if got := w.Header().Get(transport.ResponseAlertDanger); got != "danger msg" {
		t.Errorf("ResponseAlertDanger = %q, want %q", got, "danger msg")
	}
}

func TestAlertHelperFunctions_EmptyString(t *testing.T) {
	tests := []struct {
		name   string
		fn     func(http.ResponseWriter, string)
		header string
	}{
		{"AlertText empty", AlertText, transport.ResponseAlert},
		{"AlertInfo empty", AlertInfo, transport.ResponseAlertInfo},
		{"AlertSuccess empty", AlertSuccess, transport.ResponseAlertSuccess},
		{"AlertWarning empty", AlertWarning, transport.ResponseAlertWarning},
		{"AlertDanger empty", AlertDanger, transport.ResponseAlertDanger},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			tt.fn(w, "")
			if got := w.Header().Get(tt.header); got != "" {
				t.Errorf("expected empty header value, got %q", got)
			}
		})
	}
}
