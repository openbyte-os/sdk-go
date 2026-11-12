# Rubix Gateway client

`github.com/openbyte-os/sdk-go/gateway` connects apps to Rubix Gateway over HTTP.
It uses an app's Gateway API token, independently of its inbound signing key.
The package uses the standard library and existing SDK app IDs; it has no new dependencies.

```go
client, err := gateway.NewClient(
    "https://gateway.chargehive.cloud",
    os.Getenv("RUBIX_API_TOKEN"),
)
if err != nil {
    return err
}

workspace := client.Workspace("your-workspace-id")
identity, err := workspace.Identity(ctx)
if err != nil {
    return err
}
fmt.Println(identity.App, identity.Workspace)

available, err := workspace.Services(ctx)
if err != nil {
    return err
}
fmt.Println(available)

var echo struct {
    Text string `json:"text"`
}
response, err := workspace.CallService(ctx, "echo", 1, "echo",
    map[string]string{"text": "Hello"}, &echo)
if err != nil {
    return err
}
fmt.Println(echo.Text, response.TraceID)
```

Import `context`, `fmt`, `os`, and `github.com/openbyte-os/sdk-go/gateway`
as needed; supply your own `ctx` and error handling. Use
`http://gateway.cloud.svc.cluster.local:6082` for direct in-cluster access, or
`https://gateway.chargehive.dev` for ChargeHive dev. A base URL may contain an
installation path prefix, but must not include `/v1/workspaces/...`, a query,
fragment, or credentials.

The app must be installed in the selected workspace. Declare service type,
version, and required actions in the app definition; a workspace administrator
must approve access and configure the workspace's provider. `Services` lists
only approved, enabled contracts. Public callers must also satisfy the configured
app, vendor, or global source-IP allowlist.

## App operations

```go
var pong struct {
    Message string `json:"message"`
}
response, err := workspace.CallOperation(ctx, app.NewID("vendor", "target"),
    "ping", map[string]any{}, &pong)
```

Import `github.com/openbyte-os/sdk-go/app`. The target must expose the operation
and the workspace must permit the inter-app call. No user session or signing key
is supplied by this client.

`CallService` decodes the provider's raw JSON response. `CallOperation` decodes
the `data` field of Gateway's response envelope. Both return HTTP status, trace
ID, headers, and the original body in `Response`. Use `json.RawMessage` for raw
JSON inputs, and nil output to inspect a response body yourself. Nil input is
sent as JSON `null`; use `map[string]any{}` when an action expects an empty object.

## Notifications

```go
err := workspace.Notify(ctx, gateway.Notification{
    Category: "assessment-completed",
    Title:    "Your interaction has been scored",
    Path:     "/assessments/123",
    Users:    []string{userID},
})
```

`Notify` publishes as the calling app through `kx/notifications`. `Category` must
be a notification type declared in the app definition. An empty `ID` publishes a
new notification each call; set `ID` and raise `Revision` to update one in place.
`State` defaults to `posted` and `Revision` to 1.

## Errors and transport

```go
var failure *gateway.Error
if errors.As(err, &failure) {
    fmt.Println(failure.StatusCode, failure.Type, failure.Detail, failure.TraceID)
    // SourceIP, PeerIP and Reason are available for source-IP denials.
}
```

Errors from non-Gateway providers or intermediaries retain the HTTP status and
a generic detail if they do not use Gateway's error envelope. Context cancellation
and transport errors support `errors.Is` / `errors.As` through wrapping.

The client has a 30-second default timeout, an 8 MiB response limit, adds no retries,
and never follows redirects. Clients and workspace handles can be shared
by goroutines. Configure transport or response limits with:

```go
client, err := gateway.NewClient(baseURL, token,
    gateway.WithHTTPClient(&http.Client{Timeout: 10 * time.Second}),
    gateway.WithMaxResponseBytes(16 << 20),
)
```

The supplied HTTP client is copied; redirects remain disabled to keep tokens on
the configured origin. Use HTTPS for public endpoints. Client and token formatting
redacts credentials, but the token's exported field and JSON form contain the raw
secret and must be stored privately.

## Rotation and revocation

```go
issued, err := workspace.RotateToken(ctx)
if err != nil {
    return err
}
if err := saveSecret(issued.Token); err != nil {
    return err // The old credential has already been retired.
}
client, err = client.WithToken(issued.Token)
if err != nil {
    return err
}
workspace = client.Workspace(workspaceID)
```

`saveSecret` is the app's own persistence function. Rotation returns the replacement
once and retires the old credential immediately. The SDK does not persist or adopt
it automatically. Existing clients and workspace handles retain their previous token.
Do not retry an uncertain rotation: if the response was lost, obtain a new token
through the developer portal. `workspace.RevokeToken(ctx)` revokes the current
credential for all workspaces.
