package app

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/openbyte-os/sdk-go/translation"
)

func TestProviderRegistrationDisplayNameRoundTrip(t *testing.T) {
	want := ProviderRegistration{
		Type: "translation", Key: "basic-v2", Version: 1,
		Path: "/_services/translation/v1", Actions: []string{"translate"},
		Name: translation.FromMap(map[string]string{
			"en": "Google Cloud Translation Basic",
			"fr": "Traduction Google Cloud Basic",
		}, "en"),
	}
	encoded, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	var got ProviderRegistration
	if err := json.Unmarshal(encoded, &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round trip: got %+v, want %+v", got, want)
	}
	if got.Name.Get("fr") != "Traduction Google Cloud Basic" || got.Name.Get("de") != "Google Cloud Translation Basic" {
		t.Fatal("localized name or fallback lost")
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatal(err)
	}
	if _, ok := fields["name"]; !ok {
		t.Fatal("display name missing from JSON")
	}
	// Older consumers can still decode the advertised identity and endpoint.
	var oldConsumer struct {
		Type    string `json:"type"`
		Key     string `json:"key"`
		Version int    `json:"version"`
		Path    string `json:"path"`
	}
	if err := json.Unmarshal(encoded, &oldConsumer); err != nil {
		t.Fatal(err)
	}
	if oldConsumer.Key != want.Key || oldConsumer.Path != want.Path || oldConsumer.Type != want.Type || oldConsumer.Version != want.Version {
		t.Fatal("display name affected legacy identity or routing fields")
	}
}

func TestProviderRegistrationWithoutDisplayName(t *testing.T) {
	legacy := `{"type":"translation","key":"default","version":1,"path":"/_services/translation/v1","actions":["translate"]}`
	var got ProviderRegistration
	if err := json.Unmarshal([]byte(legacy), &got); err != nil {
		t.Fatal(err)
	}
	if got.Key != "default" || got.Name.Get("en") != "" {
		t.Fatalf("legacy registration changed: %+v", got)
	}
	encoded, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatal(err)
	}
	if _, exists := fields["name"]; exists {
		t.Fatal("zero-value name should be omitted from JSON")
	}
	var nullName ProviderRegistration
	if err := json.Unmarshal([]byte(`{"type":"translation","key":"default","name":null}`), &nullName); err != nil {
		t.Fatal(err)
	}
	if nullName.Name.Get("en") != "" {
		t.Fatal("null name should preserve key fallback behavior")
	}
}
