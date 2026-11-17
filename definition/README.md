# Definition helpers

Build app metadata in Go and serve it from the standard `/_kubex/` endpoints.
The package uses the SDK's `app` types throughout, including translations.

```go
metadata, err := definition.New(app.Definition{
    ID:   app.NewID("acme", "tasks"),
    Name: translation.String("Tasks"),
}).WithListing(
    definition.NewListing("Keep your team on track").
        WithOverview("Plan and manage your team's work.").
        WithHighlights(app.ListingHighlight{
            Icon: "check@lu", Title: "Track progress",
        }).Build(),
).WithHelp(
    definition.NewHelpArticle("start", "Getting started", "Create your first task.").
        WithTags("tasks").Build(),
).WithBlueprints(
    definition.NewBlueprint("acme/team", "Team workspace", "1.0.0").
        WithApps(app.BlueprintApp{
            VendorID: "acme", AppID: "tasks", Required: true,
        }).Build(),
).WithVersion(app.Version{Version: "1.0.0"}).Build()
if err != nil {
    return err
}

mux := http.NewServeMux()
mux.Handle(definition.Prefix, metadata) // /_kubex/; do not use StripPrefix
```

Import `github.com/openbyte-os/sdk-go/definition`, `github.com/openbyte-os/sdk-go/app`,
`github.com/openbyte-os/sdk-go/translation`, and `net/http`.
See [example_test.go](example_test.go) for a complete executable example.

The built handler supports three entry points:

- `mux.Handle(definition.Prefix, metadata)` mounts all metadata under `/_kubex/`.
- `metadata.Register(mux)` registers only configured endpoint paths, leaving other
  `/_kubex/` paths available to the application.
- `response := metadata.Response(request)` returns a fresh `*http.Response` for
  adapters. The caller must close `response.Body`.

The handler also implements `http.Handler` directly and can be passed to
`http.ListenAndServe`. Requests must retain their full `/_kubex/...` path.

| Endpoint | Enabled by |
| --- | --- |
| `app.PathDefinition` | Always |
| `app.PathListing` | `WithListing` |
| `app.PathHelp` | `WithHelp` |
| `app.PathBlueprints` | `WithBlueprints` |
| `app.PathVersion` | `WithVersion`, `WithVersionFile`, or a version provider |
| `app.PathHealthCheck` | `WithHealthCheck` or a health provider |

GET returns JSON; HEAD returns the same status and headers without a body.
Other methods on configured endpoints return 405 with `Allow: GET, HEAD`.
Unknown or unconfigured paths return 404. Listing image paths are not served by
this handler; register asset handlers separately where needed.

`WithHelp` and `WithBlueprints` append values and automatically advertise the
corresponding definition capability. Calling them without arguments enables an
empty collection response (`[]`). Existing capability flags are preserved, so
an application can still serve those endpoints itself. `WithListing` and
`WithVersion` replace their previous values.

Definition collection helpers such as `WithPaths`, `WithPermissions`,
`WithOperations`, `WithConnections`, and `WithNavigation` append SDK values.
Use `builder.Configure(func(d *app.Definition) { ... })` or `builder.Definition`
for any other field. Listing, help, and blueprint builders also expose their
underlying SDK values, so all fields remain available, including translations.
Their `Build` methods return shallow values for use in `WithListing`, `WithHelp`,
and `WithBlueprints`; plain SDK struct literals work equally well.

Configure builders before concurrent use. The top-level `Build` serializes a
snapshot of static metadata unaffected by subsequent edits to inputs. Live
providers must support concurrent calls. It recomputes the definition hash and reports JSON encoding errors;
it does not validate platform business rules. Extended metadata is not part of
the SDK's definition hash. Live health callbacks receive the definition hash automatically. Activation,
configuration, and about handlers remain application-owned.

Build a new handler to publish changed metadata. `Register` follows
`http.ServeMux.Handle` semantics and panics if routes conflict; do not register
multiple snapshots on the same mux.

## K4-style services and live providers

Existing K4 blueprint and help providers work without an adapter or an SDK
dependency on K4:

```go
metadata, err := definition.New(appDefinition).
    WithMetadataProviders(application).
    WithVersionFile("version.json", app.Version{Environment: "production"}).
    WithHealthCheck(func() app.HealthCheckResponse {
        // Read synchronized lifecycle/dependency state owned by your service.
        return app.HealthCheckResponse{
            Status: "OK", IsHealthy: serviceReady(), NextRetry: 5,
        }
    }).Build()
if err != nil {
    return err
}
webHandler.HandleFunc(definition.Prefix, metadata.ServeHTTP)
// Or metadata.Register(webHandler) to register individual endpoints.
```

`WithMetadataProviders` discovers these interfaces independently:

| Interface | Method |
| --- | --- |
| `BlueprintProvider` | `AvailableBlueprints() []*app.BlueprintDefinition` |
| `HelpArticleProvider` | `HelpArticles() app.HelpArticlesResponse` |
| `ListingProvider` | `Listing() app.Listing` |
| `VersionProvider` | `Version() app.Version` |
| `HealthCheckProvider` | `HealthCheck() app.HealthCheckResponse` |

Providers run on each GET or HEAD and must support concurrent calls. They override
static data for the same endpoint regardless of call order; the last provider
wins. Unsupported interfaces are ignored. Do not pass typed nil providers.
Blueprint and help providers automatically enable their definition flags. Nil
blueprint pointers are skipped, and empty blueprint/help collections encode as
`[]`. Blueprint output uses the SDK envelope `{"blueprints":[...]}`, whereas K4's
older listener emits a bare array.

Health checks return 200 when healthy and 503 otherwise, preserving status and
retry information and replacing the hash with the served definition's hash.
Live responses send `Cache-Control: no-store`. The SDK does not start lifecycle
listeners or infer service health; the callback controls readiness and shutdown.

Version files are read once at Build. JSON fields override supplied defaults;
`buildID` takes precedence over K4's legacy `build` key. No build timestamp is
invented. Missing, malformed, or non-object files fail Build. To make a file
optional in development, call `LoadVersionFile` yourself, handle `os.ErrNotExist`,
and only call `WithVersion` when it loads successfully. Alternatively pass an
existing `k4.GetVersionData().Version` to `WithVersion` after checking `LoadedFile`.

K4's current `Service.WithHttp` registers its own exact metadata routes. Those
routes take precedence over a prefix mount, and registering duplicate exact
routes panics. To adopt this package in K4, replace those built-in metadata
registrations with this handler; keep service startup and custom about handling
in K4. K4's route-group `Mount` strips prefixes, so use `Handle`/`HandleFunc` for
this handler to preserve the full `/_kubex/...` path.
