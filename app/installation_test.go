package app

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

func TestCreateSignatureKey(t *testing.T) {
	tests := []struct {
		name        string
		workspaceID string
		appSecret   string
	}{
		{name: "basic inputs", workspaceID: "workspace-123", appSecret: "secret-abc"},
		{name: "empty workspace", workspaceID: "", appSecret: "secret"},
		{name: "empty secret", workspaceID: "workspace", appSecret: ""},
		{name: "both empty", workspaceID: "", appSecret: ""},
		{name: "special characters", workspaceID: "ws/special@chars", appSecret: "sec!ret#$%"},
		{name: "unicode inputs", workspaceID: "workspace-日本語", appSecret: "秘密"},
		{name: "long inputs", workspaceID: "a]very-long-workspace-identifier-that-exceeds-normal-length", appSecret: "an-equally-long-application-secret-value-for-testing"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CreateSignatureKey(tt.workspaceID, tt.appSecret)

			// Verify against independently computed SHA256
			expected := fmt.Sprintf("%x", sha256.Sum256([]byte(tt.workspaceID+"\x00"+tt.appSecret)))
			if result != expected {
				t.Errorf("CreateSignatureKey(%q, %q) = %q, want %q", tt.workspaceID, tt.appSecret, result, expected)
			}

			// SHA256 hex output is always 64 characters
			if len(result) != 64 {
				t.Errorf("expected 64 character hex string, got %d characters", len(result))
			}
		})
	}
}

func TestCreateSignatureKey_Deterministic(t *testing.T) {
	a := CreateSignatureKey("ws", "secret")
	b := CreateSignatureKey("ws", "secret")
	if a != b {
		t.Error("same inputs should produce identical output")
	}
}

func TestCreateSignatureKey_DifferentInputsDifferentOutputs(t *testing.T) {
	a := CreateSignatureKey("workspace-1", "secret")
	b := CreateSignatureKey("workspace-2", "secret")
	if a == b {
		t.Error("different workspace IDs should produce different keys")
	}

	c := CreateSignatureKey("workspace", "secret-1")
	d := CreateSignatureKey("workspace", "secret-2")
	if c == d {
		t.Error("different secrets should produce different keys")
	}
}

func TestCreateSignatureKey_NullSeparator(t *testing.T) {
	// The null byte separator ensures "ab" + "c" differs from "a" + "bc"
	key1 := CreateSignatureKey("ab", "c")
	key2 := CreateSignatureKey("a", "bc")
	if key1 == key2 {
		t.Error("null separator should prevent collisions between shifted inputs")
	}
}
