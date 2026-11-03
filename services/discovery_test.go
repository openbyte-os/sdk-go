package services_test

import (
	"encoding/json"
	"testing"

	"github.com/openbyte-os/sdk-go/services"
	"github.com/openbyte-os/sdk-go/translation"
)

func TestDiscoveryResponse(t *testing.T) {
	want := services.DiscoveryResponse{Services: []services.Instance{
		{Key: "production", Name: translation.String("Production Google account")},
		{Key: "backup", Name: translation.FromMap(map[string]string{"en": "Backup", "fr": "Secours"}, "en")},
	}}
	encoded, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	var got services.DiscoveryResponse
	if err := json.Unmarshal(encoded, &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Services) != 2 || got.Services[0].Key != "production" || got.Services[1].Name.Get("fr") != "Secours" {
		t.Fatalf("round trip: %+v", got)
	}
	encoded, err = json.Marshal(services.DiscoveryResponse{Services: []services.Instance{}})
	if err != nil || string(encoded) != `{"services":[]}` {
		t.Fatalf("empty discovery response: %s %v", encoded, err)
	}
	var legacy services.DiscoveryResponse
	if err := json.Unmarshal([]byte(`{"services":[{"key":"default"}],"future":true}`), &legacy); err != nil || legacy.Services[0].Name.Get("en") != "" {
		t.Fatalf("optional name: %+v %v", legacy, err)
	}
}
