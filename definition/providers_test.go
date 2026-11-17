package definition_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/openbyte-os/sdk-go/app"
	"github.com/openbyte-os/sdk-go/definition"
)

type serviceMetadata struct {
	unhealthy atomic.Bool
	calls     atomic.Int32
}

func (s *serviceMetadata) AvailableBlueprints() []*app.BlueprintDefinition {
	s.calls.Add(1)
	return []*app.BlueprintDefinition{nil, {ID: "acme/team"}}
}
func (s *serviceMetadata) HelpArticles() app.HelpArticlesResponse { return app.HelpArticlesResponse{} }
func (s *serviceMetadata) HealthCheck() app.HealthCheckResponse {
	return app.HealthCheckResponse{IsHealthy: !s.unhealthy.Load(), Status: "service", NextRetry: 5, DefinitionHash: "stale"}
}
func (s *serviceMetadata) Version() app.Version { return app.Version{Version: "2.0.0"} }
func (s *serviceMetadata) Listing() app.Listing { return app.Listing{Tagline: "Live"} }

// K4's WebHandler exposes HandleFunc, but not Handle.
type funcRegistrar struct{ mux *http.ServeMux }

func (r funcRegistrar) HandleFunc(path string, fn func(http.ResponseWriter, *http.Request)) {
	r.mux.HandleFunc(path, fn)
}

func TestLiveProviders(t *testing.T) {
	service := &serviceMetadata{}
	builder := definition.New(app.Definition{}).WithMetadataProviders(service).
		WithVersionFile("not-read-because-provider-wins", app.Version{})
	h, err := builder.Build()
	if err != nil {
		t.Fatal(err)
	}
	if service.calls.Load() != 0 {
		t.Fatal("provider called during build")
	}
	// Further configuration must not change the built handler's provider map.
	builder.WithHealthCheck(func() app.HealthCheckResponse { return app.HealthCheckResponse{} })
	mux := http.NewServeMux()
	h.Register(funcRegistrar{mux})
	get := func(method, path string) *httptest.ResponseRecorder {
		t.Helper()
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(method, path, nil))
		return w
	}
	for _, unhealthy := range []bool{false, true, false} {
		service.unhealthy.Store(unhealthy)
		response := get("GET", app.PathHealthCheck)
		want := 200
		if unhealthy {
			want = 503
		}
		if response.Code != want {
			t.Fatalf("status = %d, want %d", response.Code, want)
		}
		var health app.HealthCheckResponse
		if err := json.Unmarshal(response.Body.Bytes(), &health); err != nil {
			t.Fatal(err)
		}
		if health.DefinitionHash != h.DefinitionHash() || health.NextRetry != 5 || health.IsHealthy == unhealthy {
			t.Fatalf("health: %+v", health)
		}
		if response.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("health must not be cached")
		}
	}
	response := get("GET", app.PathBlueprints)
	var blueprints app.BlueprintsResponse
	if err := json.Unmarshal(response.Body.Bytes(), &blueprints); err != nil {
		t.Fatal(err)
	}
	if len(blueprints.Blueprints) != 1 || blueprints.Blueprints[0].ID != "acme/team" {
		t.Fatalf("blueprints: %+v", blueprints)
	}
	for i := 0; i < 2; i++ {
		get("GET", app.PathBlueprints)
	}
	if service.calls.Load() != 3 {
		t.Fatal("provider must be called on every GET")
	}
	if get("POST", app.PathBlueprints).Code != 405 || service.calls.Load() != 3 {
		t.Fatal("unsupported method invoked provider")
	}
	if r := get("HEAD", app.PathBlueprints); r.Code != 200 || r.Body.Len() != 0 || service.calls.Load() != 4 {
		t.Fatal("incorrect provider HEAD")
	}
	if r := get("GET", app.PathHelp); r.Body.String() != `{"articles":[]}` {
		t.Fatal(r.Body.String())
	}
	if r := get("GET", app.PathListing); r.Body.String() != `{"tagline":"Live"}` {
		t.Fatal(r.Body.String())
	}
	var def app.Definition
	if err := json.Unmarshal(get("GET", app.PathDefinition).Body.Bytes(), &def); err != nil {
		t.Fatal(err)
	}
	if !def.ProvideHelp || !def.ProvideBlueprints {
		t.Fatal("missing provider flags")
	}
	var version app.Version
	if err := json.Unmarshal(get("GET", app.PathVersion).Body.Bytes(), &version); err != nil {
		t.Fatal(err)
	}
	if version.Version != "2.0.0" {
		t.Fatal(version)
	}
}

func TestVersionFile(t *testing.T) {
	for _, tc := range []struct {
		name, data, build string
		fail              bool
	}{
		{"sdk", `{"version":"1.2.3","buildID":"sdk"}`, "sdk", false},
		{"legacy", `{"version":"1.2.3","build":"legacy"}`, "legacy", false},
		{"precedence", `{"buildID":"sdk","build":"legacy"}`, "sdk", false},
		{"invalid", `{`, "", true},
		{"null", `null`, "", true},
		{"array", `[]`, "", true},
		{"type", `{"version":12}`, "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "version.json")
			if err := os.WriteFile(path, []byte(tc.data), 0600); err != nil {
				t.Fatal(err)
			}
			h, err := definition.New(app.Definition{}).WithVersionFile(path, app.Version{Environment: "production"}).Build()
			if tc.fail {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			// Snapshot survives changes to the source file.
			if err := os.WriteFile(path, []byte(`{}`), 0600); err != nil {
				t.Fatal(err)
			}
			response := h.Response(httptest.NewRequest("GET", app.PathVersion, nil))
			defer response.Body.Close()
			var version app.Version
			if err := json.NewDecoder(response.Body).Decode(&version); err != nil {
				t.Fatal(err)
			}
			if version.BuildID != tc.build || version.Environment != "production" || version.BuildTime != "" {
				t.Fatalf("version: %+v", version)
			}
		})
	}
	b := definition.New(app.Definition{}).WithVersionFile(filepath.Join(t.TempDir(), "missing.json"), app.Version{})
	if _, err := b.Build(); err == nil {
		t.Fatal("missing file should fail")
	}
	if _, err := b.WithVersion(app.Version{}).Build(); err != nil {
		t.Fatal(err)
	}
}

func TestConcurrentProviderRequests(t *testing.T) {
	service := &serviceMetadata{}
	h, err := definition.New(app.Definition{}).WithMetadataProviders(service).Build()
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 16; i++ {
		t.Run("request", func(t *testing.T) {
			t.Parallel()
			for _, path := range []string{app.PathHealthCheck, app.PathBlueprints, app.PathHelp, app.PathVersion} {
				response := h.Response(httptest.NewRequest("GET", path, nil))
				response.Body.Close()
				if response.StatusCode != 200 {
					t.Fatalf("%s: %d", path, response.StatusCode)
				}
			}
		})
	}
}
