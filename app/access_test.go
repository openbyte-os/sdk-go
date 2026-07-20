package app

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestAccessRequest_ReadOnly(t *testing.T) {
	ar := AccessRequest{ /*ReadOnly: true*/ }
	b, err := json.Marshal(ar)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(b), `"readOnly":true`) {
		t.Errorf("missing readOnly flag in %s", string(b))
	}

	ar2 := AccessRequest{}
	b2, _ := json.Marshal(ar2)
	if strings.Contains(string(b2), `"readOnly"`) {
		t.Errorf("readOnly should omitempty when false, got %s", string(b2))
	}
}
