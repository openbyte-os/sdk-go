package uit

import (
	"net/http"
	"strings"

	"github.com/openbyte-os/sdk-go/transport"
)

type Response struct {
	headers map[string]string
	code    int
}

func R() *Response {
	return &Response{headers: make(map[string]string)}
}

func (r *Response) Refresh() {
	r.headers[transport.ResponseRefresh] = "self"
}

func (r *Response) NoContent() {
	r.code = http.StatusNoContent
}

func (r *Response) RefreshFragment(fragments ...string) {
	r.headers[transport.ResponseRefresh] = strings.Join(fragments, ",")
}

func (r *Response) RefreshSpace(uri ...string) {
	//app-space[uri="/${uri}"] Provide the URI of the spaces to refresh
	r.headers[transport.ResponseRefresh] = strings.Join(uri, ",")
}

func (r *Response) RefreshReferer() {
	r.headers[transport.ResponseRefresh] = "referer"
}

func (r *Response) CloseModal() {
	r.headers[transport.ResponseCloseModal] = "1"
}

func (r *Response) Write(w http.ResponseWriter) {
	if r.headers != nil {
		for k, v := range r.headers {
			w.Header().Set(k, v)
		}
	}

	if r.code != 0 {
		w.WriteHeader(r.code)
	}
}
