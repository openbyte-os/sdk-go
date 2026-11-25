package gateway_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/openbyte-os/sdk-go/gateway"
)

func TestMembersSendsTheUserFilter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/workspaces/ws-a/members" {
			t.Errorf("request = %s %s", r.Method, r.URL.Path)
		}
		if got := r.URL.Query().Get("user"); got != "u 1" {
			t.Errorf("user = %q, want u 1", got)
		}
		envelope(w, 200, []gateway.Member{{ID: "u2", Name: "Bob", BPO: "acme"}})
	}))
	defer server.Close()

	members, err := client(t, server.URL).Workspace("ws-a").Members(context.Background(), "u 1")
	if err != nil {
		t.Fatal(err)
	}
	if len(members) != 1 || members[0].ID != "u2" || members[0].BPO != "acme" {
		t.Errorf("members = %+v", members)
	}
}
