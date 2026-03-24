package transport

// RequestWorkspaceID Workspace UUID
const RequestWorkspaceID = "x-kx-workspace-id"

// RequestUserID User UID
const RequestUserID = "x-kx-user-id"

// RequestSignature Kubex Signature
const RequestSignature = "x-kx-signature"

// RequestTraceID Kubex request Trace ID
const RequestTraceID = "x-kx-trace-id"

// RequestUserIP User IP
const RequestUserIP = "x-kx-user-ip"

// RequestUserAgent User Agent
const RequestUserAgent = "x-kx-user-agent"

// RequestAuthorization Json Authorizations, [{"k":"permission-key","e":"A","r":"*"},"k":"permission-key-2","e":"D","r":"uid1"]
const RequestAuthorization = "x-kx-authorization"

// RequestAuthentication JSON access credentials, provided by the app e.g. {"accessToken":"xx"}
const RequestAuthentication = "x-kx-authentication"

// RequestWorkspaceHost The full workspace host the user is currently in
const RequestWorkspaceHost = "x-kx-workspace-host" // Host of the workspace, used for routing

// RequestBrands CSV of permitted brands
const RequestBrands = "x-kx-brands"

// RequestDepartments CSV of permitted departments
const RequestDepartments = "x-kx-departments"

// RequestChannels CSV of permitted channels
const RequestChannels = "x-kx-channels"

// RequestConfiguration JSON configuration data for the app k:v pairs
const RequestConfiguration = "x-kx-configuration"

// RequestCallerApp identifies the calling app in an RPC request
const RequestCallerApp = "x-kx-caller-app"

const RequestEditorID = "x-kx-editor-id"
const RequestFetchStyle = "x-kx-fetch-style"
const RequestContextData = "x-kx-context-data"
const RequestActiveTab = "x-kx-active-tab"

// ResponseDebug Debug object for the browser
const ResponseDebug = "x-kubex-debug"

// ResponseAlert alert text
const ResponseAlert = "x-kubex-alert"

// ResponseAlertInfo alert text
const ResponseAlertInfo = "x-kubex-alert-info"

// ResponseAlertSuccess alert text
const ResponseAlertSuccess = "x-kubex-alert-success"

// ResponseAlertWarning alert text
const ResponseAlertWarning = "x-kubex-alert-warning"

// ResponseAlertDanger alert text
const ResponseAlertDanger = "x-kubex-alert-danger"

// ResponseAppendElement append element to the DOM
const ResponseAppendElement = "x-kubex-append-element"

// ResponseVanishElement vanish element from the DOM
const ResponseVanishElement = "x-kubex-vanish-element"

// ResponseCloseModal close modal
const ResponseCloseModal = "x-kubex-close-modal"

// ResponseBasisPx ideal viewport in pixels (not guaranteed)
const ResponseBasisPx = "x-kubex-basis-px"

// ResponseRemoveSelf remove self from the DOM
const ResponseRemoveSelf = "x-kubex-remove-self"

// ResponseRefresh refresh (self, referer or space-ID[#fragment])
const ResponseRefresh = "x-kubex-refresh"

// ResponseForwardUri forward the browser to a new URI
const ResponseForwardUri = "x-kubex-forward-uri"

// ResponseForwardGaid sets the gaid for the forward URI
const ResponseForwardGaid = "x-kubex-forward-gaid" // GAID of the forward URI

// ResponseCount item count
const ResponseCount = "x-kubex-count"

// ResponseColor item color
const ResponseColor = "x-kubex-color"

// ResponseClearValue clear the value of an input field
const ResponseClearValue = "x-kubex-clear-value"

const RequestPreferredLabels = "x-kx-preferred-labels"
const RequestConnectionID = "x-kx-connection-id"
const RequestConnectionToken = "x-kx-connection-token"
