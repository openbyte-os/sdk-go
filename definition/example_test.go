package definition_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/openbyte-os/sdk-go/app"
	"github.com/openbyte-os/sdk-go/definition"
	"github.com/openbyte-os/sdk-go/translation"
)

func Example() {
	metadata, err := definition.New(app.Definition{
		ID:   app.NewID("acme", "tasks"),
		Name: translation.String("Tasks"),
	}).WithListing(
		definition.NewListing("Keep your team on track").
			WithOverview("Plan and manage your team's work.").
			WithHighlights(app.ListingHighlight{Icon: "check@lu", Title: "Track progress"}).Build(),
	).WithHelp(
		definition.NewHelpArticle("getting-started", "Getting started", "Create your first task.").
			WithTags("tasks").Build(),
	).WithBlueprints(
		definition.NewBlueprint("acme/team", "Team workspace", "1.0.0").
			WithApps(app.BlueprintApp{VendorID: "acme", AppID: "tasks", Required: true}).Build(),
	).WithVersion(app.Version{Version: "1.0.0"}).Build()
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.Handle(definition.Prefix, metadata)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Your application")
	})

	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, httptest.NewRequest("GET", app.PathHelp, nil))
	fmt.Println(recorder.Code)
	fmt.Println(recorder.Header().Get("Content-Type"))
	// Output:
	// 200
	// application/json
}
