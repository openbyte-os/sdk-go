package uit

import (
	"net/http"
	"strings"

	"github.com/openbyte-os/sdk-go/transport"
)

type ResponseWriter struct {
	headers map[string]string
	code    int
}

func Response() *ResponseWriter {
	return &ResponseWriter{headers: make(map[string]string)}
}

func (r *ResponseWriter) Refresh() *ResponseWriter {
	r.headers[transport.ResponseRefresh] = "self"
	return r
}

func (r *ResponseWriter) NoContent() *ResponseWriter {
	r.code = http.StatusNoContent
	return r
}

func (r *ResponseWriter) RefreshFragment(fragments ...string) *ResponseWriter {
	r.headers[transport.ResponseRefresh] = strings.Join(fragments, ",")
	return r
}

func (r *ResponseWriter) RefreshSpace(uri ...string) *ResponseWriter {
	//app-space[uri="/${uri}"] Provide the URI of the spaces to refresh
	r.headers[transport.ResponseRefresh] = strings.Join(uri, ",")
	return r
}

func (r *ResponseWriter) RefreshReferer() *ResponseWriter {
	r.headers[transport.ResponseRefresh] = "referer"
	return r
}

func (r *ResponseWriter) CloseModal() *ResponseWriter {
	r.headers[transport.ResponseCloseModal] = "1"
	return r
}

func (r *ResponseWriter) Write(w http.ResponseWriter) {
	if r.headers != nil {
		for k, v := range r.headers {
			w.Header().Set(k, v)
		}
	}

	if r.code != 0 {
		w.WriteHeader(r.code)
	}
}
