package app

import (
	"encoding/json"
	"testing"
)

func TestBlueprintsResponseJsonRoundTrip(t *testing.T) {
	response := BlueprintsResponse{
		Blueprints: []BlueprintDefinition{{
			ID:          "acme/starter",
			Name:        "Starter",
			Description: "Starter blueprint",
			Icon:        "box",
			Version:     "1.0.0",
		}},
	}

	raw, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal blueprints response: %v", err)
	}

	decoded := BlueprintsResponse{}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal blueprints response: %v", err)
	}

	if len(decoded.Blueprints) != 1 {
		t.Fatalf("blueprints = %d, want 1", len(decoded.Blueprints))
	}
	if decoded.Blueprints[0].ID != "acme/starter" {
		t.Fatalf("blueprint id = %q, want %q", decoded.Blueprints[0].ID, "acme/starter")
	}
}
