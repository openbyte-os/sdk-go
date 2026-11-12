package gateway_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/openbyte-os/sdk-go/gateway"
)

func TestNotifyPublishesThroughNotificationsApp(t *testing.T) {
	var path string
	var got map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&got)
		envelope(w, http.StatusAccepted, map[string]bool{"changed": true})
	}))
	defer server.Close()

	err := client(t, server.URL).Workspace("ws").Notify(context.Background(), gateway.Notification{Category: "kudos", Title: "Nice work"})
	if err != nil {
		t.Fatal(err)
	}
	if path != "/v1/workspaces/ws/apps/kx/notifications/operations/publish" {
		t.Fatalf("path %q", path)
	}
	if id, _ := got["id"].(string); len(id) != 32 || got["revision"] != 1.0 || got["state"] != "posted" || got["category"] != "kudos" || got["title"] != "Nice work" {
		t.Fatalf("body %+v", got)
	}
}
