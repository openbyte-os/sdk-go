package gateway_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/openbyte-os/sdk-go/app"
	"github.com/openbyte-os/sdk-go/gateway"
)

func envelope(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]any{"meta": map[string]string{"traceID": "test-trace"}, "status": map[string]int{"code": code}, "data": data})
}

func client(t *testing.T, url string, options ...gateway.Option) *gateway.Client {
	t.Helper()
	c, err := gateway.NewClient(url, "old-token", options...)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestWorkspaceCallsAndCredentialLifecycle(t *testing.T) {
	var rotated, revoked atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		want := "Bearer old-token"
		if rotated.Load() {
			want = "Bearer new-token"
		}
		if r.Header.Get("Authorization") != want || revoked.Load() {
			http.Error(w, "Unauthorized", 401)
			return
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Error("missing JSON accept header")
		}
		switch r.URL.Path {
		case "/prefix/v1/workspaces/ws-a/identity":
			if r.Method != "GET" {
				t.Error("incorrect identity method")
			}
			envelope(w, 200, map[string]string{"app": "vendor/caller", "workspace": "ws-a"})
		case "/prefix/v1/workspaces/ws-a/services":
			envelope(w, 200, []gateway.Service{{Type: "echo", Version: 1, Actions: []string{"echo"}}})
		case "/prefix/v1/workspaces/ws-a/services/echo/1/echo", "/prefix/v1/workspaces/ws-a/apps/vendor/target/operations/ping":
			if r.Method != "POST" || r.Header.Get("Content-Type") != "application/json" {
				t.Error("incorrect JSON request")
			}
			body, _ := io.ReadAll(r.Body)
			if string(body) != `{"text":"hello"}` {
				t.Errorf("incorrect request body %s", body)
			}
			if strings.Contains(r.URL.Path, "/services/") {
				w.Header().Set("X-Kx-Trace-Id", "provider-trace")
				w.WriteHeader(202)
				_, _ = io.WriteString(w, `{"text":"echo"}`)
			} else {
				envelope(w, 201, map[string]string{"text": "pong"})
			}
		case "/prefix/v1/workspaces/ws-a/token/rotate":
			if r.Method != "POST" {
				t.Error("incorrect rotation method")
			}
			rotated.Store(true)
			envelope(w, 200, map[string]string{"token": "new-token", "tokenType": "Bearer"})
		case "/prefix/v1/workspaces/ws-a/token":
			if r.Method != "DELETE" {
				t.Error("incorrect revocation method")
			}
			revoked.Store(true)
			w.WriteHeader(204)
		default:
			t.Errorf("unexpected route: %s", r.URL)
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	c := client(t, server.URL+"/prefix/")
	ws := c.Workspace("ws-a")
	ctx := context.Background()
	identity, err := ws.Identity(ctx)
	if err != nil || identity.Workspace != "ws-a" || identity.App != "vendor/caller" {
		t.Fatalf("identity: %+v %v", identity, err)
	}
	services, err := ws.Services(ctx)
	if err != nil || len(services) != 1 || services[0].Type != "echo" {
		t.Fatalf("services: %+v %v", services, err)
	}
	var echo struct {
		Text string `json:"text"`
	}
	result, err := ws.CallService(ctx, "echo", 1, "echo", map[string]string{"text": "hello"}, &echo)
	if err != nil || echo.Text != "echo" || result.StatusCode != 202 || result.TraceID != "provider-trace" {
		t.Fatalf("service response: %+v %+v %v", echo, result, err)
	}
	result, err = ws.CallOperation(ctx, app.NewID("vendor", "target"), "ping", map[string]string{"text": "hello"}, &echo)
	if err != nil || echo.Text != "pong" || result.StatusCode != 201 || result.TraceID != "test-trace" {
		t.Fatalf("operation response: %+v %+v %v", echo, result, err)
	}
	issued, err := ws.RotateToken(ctx)
	if err != nil || issued.Token != "new-token" {
		t.Fatal("rotation failed", err)
	}
	_, err = ws.Identity(ctx)
	var apiError *gateway.Error
	if !errors.As(err, &apiError) || apiError.StatusCode != 401 {
		t.Fatal("old client did not retain its old token")
	}
	replacement, err := c.WithToken(issued.Token)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := replacement.Workspace("ws-a").Identity(ctx); err != nil {
		t.Fatal(err)
	}
	if err := replacement.Workspace("ws-a").RevokeToken(ctx); err != nil {
		t.Fatal(err)
	}
	_, err = replacement.Workspace("ws-a").Identity(ctx)
	if !errors.As(err, &apiError) || apiError.StatusCode != 401 {
		t.Fatal("revocation not reflected")
	}
}

func TestErrorsPreserveSourceDiagnosticsAndTrace(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
		_, _ = io.WriteString(w, `{"meta":{"traceID":"denied-trace"},"status":{"code":403},"data":null,"error":{"type":"source_ip_forbidden","detail":"Source IP is not permitted","sourceIP":"203.0.113.1","peerIP":"10.2.32.1","reason":"no allowlist match"}}`)
	}))
	defer server.Close()
	_, err := client(t, server.URL).Workspace("ws").Identity(context.Background())
	var failure *gateway.Error
	if !errors.As(err, &failure) || failure.StatusCode != 403 || failure.Type != "source_ip_forbidden" || failure.TraceID != "denied-trace" || failure.SourceIP != "203.0.113.1" || failure.PeerIP != "10.2.32.1" || failure.Reason != "no allowlist match" {
		t.Fatalf("lost diagnostics: %v", err)
	}
}

func TestRedirectsAreNeverFollowedWithCustomHTTPClient(t *testing.T) {
	var leaked atomic.Int32
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { leaked.Add(1) }))
	defer target.Close()
	redirect := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, target.URL, 307) }))
	defer redirect.Close()
	custom := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return nil }}
	_, err := client(t, redirect.URL, gateway.WithHTTPClient(custom)).Workspace("ws").Identity(context.Background())
	var failure *gateway.Error
	if !errors.As(err, &failure) || failure.StatusCode != 307 || leaked.Load() != 0 {
		t.Fatal("redirect was followed", err)
	}
	if custom.CheckRedirect == nil {
		t.Fatal("caller HTTP client was mutated")
	}
}

func TestValidationCancellationAndResponseLimits(t *testing.T) {
	for _, base := range []string{"", "ftp://example.test", "https://user:secret@example.test", "https://example.test?token=secret", "https://example.test?", "https://example.test#fragment"} {
		if _, err := gateway.NewClient(base, "token"); err == nil {
			t.Fatal("invalid base URL accepted")
		}
	}
	for _, token := range []string{"", "secret\r\nInjected: header", " token", "two tokens"} {
		if _, err := gateway.NewClient("https://example.test", token); err == nil {
			t.Fatal("invalid token accepted")
		}
	}
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		_, _ = io.WriteString(w, strings.Repeat("x", 100))
	}))
	defer server.Close()
	c := client(t, server.URL)
	for _, id := range []string{"", "..", "other/workspace", "a?query", "%2f"} {
		if _, err := c.Workspace(id).Identity(context.Background()); err == nil {
			t.Fatal("invalid workspace accepted")
		}
	}
	if _, err := c.Workspace("ws").CallService(context.Background(), "echo", 1, "../identity", nil, nil); err == nil {
		t.Fatal("path traversal accepted")
	}
	if _, err := c.Workspace("ws").CallService(context.Background(), "echo", 0, "echo", nil, nil); err == nil {
		t.Fatal("invalid version accepted")
	}
	if _, err := c.Workspace("ws").CallService(context.Background(), "echo", 1, "echo", json.RawMessage(`invalid`), nil); err == nil {
		t.Fatal("invalid JSON accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := c.Workspace("ws").Identity(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("context cancellation lost: %v", err)
	}
	if calls.Load() != 0 {
		t.Fatal("invalid request was sent")
	}
	if _, err := client(t, server.URL, gateway.WithMaxResponseBytes(10)).Workspace("ws").Identity(context.Background()); err == nil || !strings.Contains(err.Error(), "size limit") {
		t.Fatal("oversized response accepted", err)
	}
	if calls.Load() != 1 {
		t.Fatal("unexpected automatic retry")
	}
}

func TestConcurrentWorkspaceScoping(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		envelope(w, 200, gateway.Identity{App: "vendor/caller", Workspace: parts[3]})
	}))
	defer server.Close()
	c := client(t, server.URL)
	var wg sync.WaitGroup
	for _, id := range []string{"first", "second", "third"} {
		for range 5 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				identity, err := c.Workspace(id).Identity(context.Background())
				if err != nil || identity.Workspace != id {
					t.Errorf("workspace leaked: %+v %v", identity, err)
				}
			}()
		}
	}
	wg.Wait()
}

func TestCredentialFormattingIsRedacted(t *testing.T) {
	c := client(t, "https://example.test")
	for _, value := range []any{c, c.Workspace("ws"), gateway.Token{Token: "old-token", TokenType: "Bearer"}} {
		for _, format := range []string{"%v", "%+v", "%#v"} {
			if strings.Contains(fmt.Sprintf(format, value), "old-token") {
				t.Fatal("token appeared in formatted client or rotation result")
			}
		}
	}
}
