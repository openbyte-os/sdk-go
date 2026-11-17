package definition

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/openbyte-os/sdk-go/app"
	"io"
	"net/http"
	"sort"
)

// Prefix is the route at which a Handler can be mounted without StripPrefix.
const Prefix = "/_kubex/"

// Handler serves static metadata and live provider callbacks captured by Build. It is safe
// for concurrent use when its providers are safe for concurrent calls. Mount it with mux.Handle(definition.Prefix, handler).
// Paths are exact: unknown paths and unconfigured endpoints return 404.
type Handler struct {
	documents map[string][]byte
	hash      string
	endpoints map[string]endpoint
}

var _ http.Handler = (*Handler)(nil)

// DefinitionHash returns the hash served in the definition. An application's
// live health check can use it in app.HealthCheckResponse.DefinitionHash.
func (h *Handler) DefinitionHash() string { return h.hash }

// Paths lists configured endpoint paths in sorted order.
func (h *Handler) Paths() []string {
	paths := make([]string, 0, len(h.documents))
	for path := range h.documents {
		paths = append(paths, path)
	}
	for path := range h.endpoints {
		if _, exists := h.documents[path]; !exists {
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)
	return paths
}

// Register attaches each configured endpoint to a mux or compatible registrar.
// Like http.ServeMux.HandleFunc, it panics on conflicting registrations. To own the
// whole prefix instead, call mux.Handle(Prefix, h).
func (h *Handler) Register(mux interface {
	HandleFunc(string, func(http.ResponseWriter, *http.Request))
}) {
	for _, path := range h.Paths() {
		mux.HandleFunc(path, h.ServeHTTP)
	}
}

// Response returns a fresh response for a request, including 404 and 405
// responses. The caller owns Body and should close it. GET and HEAD are supported.
// The request must have a non-nil URL and retain its full /_kubex path.
func (h *Handler) Response(r *http.Request) *http.Response {
	status := http.StatusOK
	headers := make(http.Header)
	headers.Set("Content-Type", "application/json")
	data, found := h.documents[r.URL.Path]
	provider := h.endpoints[r.URL.Path]
	if !found && provider == nil {
		status = http.StatusNotFound
		data = []byte(`{"error":"not found"}`)
	} else if r.Method != http.MethodGet && r.Method != http.MethodHead {
		status = http.StatusMethodNotAllowed
		headers.Set("Allow", "GET, HEAD")
		data = []byte(`{"error":"method not allowed"}`)
	}
	if provider != nil && status == http.StatusOK {
		value, code := provider(r)
		status = code
		if health, ok := value.(app.HealthCheckResponse); ok {
			health.DefinitionHash = h.hash
			value = health
		}
		var err error
		data, err = json.Marshal(value)
		if err != nil {
			status = http.StatusInternalServerError
			data = []byte(`{"error":"unable to encode response"}`)
		}
		headers.Set("Cache-Control", "no-store")
	}
	length := len(data)
	headers.Set("Content-Length", fmt.Sprint(length))
	if r.Method == http.MethodHead {
		data = nil
	}
	return &http.Response{
		Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
		StatusCode: status,
		Proto:      "HTTP/1.1", ProtoMajor: 1, ProtoMinor: 1,
		Header:        headers,
		Body:          io.NopCloser(bytes.NewReader(data)),
		ContentLength: int64(length),
		Request:       r,
	}
}

// ServeHTTP implements http.Handler using the same response as Response.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	response := h.Response(r)
	defer response.Body.Close()
	for key, values := range response.Header {
		w.Header()[key] = values
	}
	w.WriteHeader(response.StatusCode)
	_, _ = io.Copy(w, response.Body)
}
