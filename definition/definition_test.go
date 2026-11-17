package definition_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"

	"github.com/openbyte-os/sdk-go/app"
	"github.com/openbyte-os/sdk-go/definition"
	"github.com/openbyte-os/sdk-go/translation"
)

func TestSnapshotAndContracts(t *testing.T) {
	article := definition.NewHelpArticle("start", "Getting started", "Welcome").WithTags("start").Build()
	listing := definition.NewListing("Get things done").WithHighlights(app.ListingHighlight{Title: "Fast"}).Build()
	blueprint := definition.NewBlueprint("acme/start", "Start", "1.0.0").WithApps(app.BlueprintApp{VendorID: "acme", AppID: "tasks", Required: true}).Build()
	builder := definition.New(app.Definition{ID: app.NewID("acme", "tasks"), Name: translation.String("Tasks"), Hash: "stale"}).
		WithListing(listing).WithHelp(article).WithBlueprints(blueprint).WithVersion(app.Version{Version: "1.0.0"})
	handler, err := builder.Build()
	if err != nil {
		t.Fatal(err)
	}
	// Changes to nested inputs and the builder cannot alter a published snapshot.
	article.Tags[0] = "changed"
	listing.Highlights[0].Title = "changed"
	blueprint.Apps[0].AppID = "changed"
	builder.Definition.Name.Fallback = "changed"
	builder.WithHelp(app.HelpArticle{ID: "later"})
	get := func(path string, value any) {
		t.Helper()
		response := handler.Response(httptest.NewRequest(http.MethodGet, path, nil))
		defer response.Body.Close()
		if response.StatusCode != 200 {
			t.Fatalf("%s: %d", path, response.StatusCode)
		}
		if err := json.NewDecoder(response.Body).Decode(value); err != nil {
			t.Fatal(err)
		}
	}
	var def app.Definition
	get(app.PathDefinition, &def)
	if !def.ProvideHelp || !def.ProvideBlueprints || def.Name.Fallback != "Tasks" {
		t.Fatalf("definition: %+v", def)
	}
	hash := def.Hash
	def.Hash = ""
	if hash == "stale" || hash != def.GetHash(false) || hash != handler.DefinitionHash() {
		t.Fatal("incorrect definition hash")
	}
	var help app.HelpArticlesResponse
	get(app.PathHelp, &help)
	if len(help.Articles) != 1 || help.Articles[0].Tags[0] != "start" {
		t.Fatalf("help: %+v", help)
	}
	var gotListing app.Listing
	get(app.PathListing, &gotListing)
	if gotListing.Highlights[0].Title != "Fast" {
		t.Fatal("listing mutated")
	}
	var blueprints app.BlueprintsResponse
	get(app.PathBlueprints, &blueprints)
	if blueprints.Blueprints[0].Apps[0].AppID != "tasks" {
		t.Fatal("blueprint mutated")
	}
	var version app.Version
	get(app.PathVersion, &version)
	if version.Version != "1.0.0" {
		t.Fatal("wrong version")
	}
}

func TestHTTPEntryPoints(t *testing.T) {
	h, err := definition.New(app.Definition{}).WithHelp().WithBlueprints().Build()
	if err != nil {
		t.Fatal(err)
	}
	individual := http.NewServeMux()
	h.Register(individual)
	prefix := http.NewServeMux()
	prefix.Handle(definition.Prefix, h)
	prefix.HandleFunc("/app", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) })
	for _, tc := range []struct {
		method, path string
		status       int
		body         string
	}{
		{"GET", app.PathHelp, 200, `{"articles":[]}`},
		{"GET", app.PathBlueprints, 200, `{"blueprints":[]}`},
		{"HEAD", app.PathHelp, 200, ""},
		{"POST", app.PathHelp, 405, `{"error":"method not allowed"}`},
		{"GET", app.PathListing, 404, `{"error":"not found"}`},
		{"GET", app.PathHelp + "/extra", 404, `{"error":"not found"}`},
		{"GET", "/_kubex/unknown", 404, `{"error":"not found"}`},
		{"HEAD", "/_kubex/unknown", 404, ""},
		{"GET", app.PathHelp + "?language=en", 200, `{"articles":[]}`},
	} {
		t.Run(tc.method+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			response := h.Response(req)
			data, err := io.ReadAll(response.Body)
			response.Body.Close()
			if err != nil {
				t.Fatal(err)
			}
			if response.StatusCode != tc.status || string(data) != tc.body {
				t.Fatalf("response: %d %s", response.StatusCode, data)
			}
			for _, target := range []http.Handler{h, prefix} {
				recorder := httptest.NewRecorder()
				target.ServeHTTP(recorder, req)
				if recorder.Code != tc.status || recorder.Body.String() != tc.body {
					t.Fatalf("handler: %d %s", recorder.Code, recorder.Body)
				}
				if !reflect.DeepEqual(recorder.Header(), response.Header) {
					t.Fatal("response/handler headers differ")
				}
			}
			recorder := httptest.NewRecorder()
			individual.ServeHTTP(recorder, req)
			if recorder.Code != tc.status {
				t.Fatalf("registered status: %d", recorder.Code)
			}
			if tc.status == 405 && response.Header.Get("Allow") != "GET, HEAD" {
				t.Fatal("missing Allow")
			}
			if tc.method != "HEAD" && response.ContentLength != int64(len(data)) {
				t.Fatal("wrong length")
			}
			if tc.method == "HEAD" && response.Header.Get("Content-Length") == "0" {
				t.Fatal("HEAD should retain GET length")
			}
			if response.Header.Get("Content-Length") != strconv.FormatInt(response.ContentLength, 10) {
				t.Fatal("inconsistent length")
			}
		})
	}
	recorder := httptest.NewRecorder()
	prefix.ServeHTTP(recorder, httptest.NewRequest("GET", "/app", nil))
	if recorder.Code != 204 {
		t.Fatal("prefix interfered with app route")
	}
}

func TestResponseIsolation(t *testing.T) {
	h, err := definition.New(app.Definition{}).WithHelp().Build()
	if err != nil {
		t.Fatal(err)
	}
	first := h.Response(httptest.NewRequest("GET", app.PathHelp, nil))
	first.Header.Set("Content-Type", "changed")
	io.ReadAll(first.Body)
	first.Body.Close()
	second := h.Response(httptest.NewRequest("GET", app.PathHelp, nil))
	defer second.Body.Close()
	body, _ := io.ReadAll(second.Body)
	if string(body) != `{"articles":[]}` || second.Header.Get("Content-Type") != "application/json" {
		t.Fatal("responses share mutable state")
	}
}
