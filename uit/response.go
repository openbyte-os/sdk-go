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

func (r *Response) Refresh() *Response {
	r.headers[transport.ResponseRefresh] = "self"
	return r
}

func (r *Response) NoContent() *Response {
	r.code = http.StatusNoContent
	return r
}

func (r *Response) RefreshFragment(fragments ...string) *Response {
	r.headers[transport.ResponseRefresh] = strings.Join(fragments, ",")
	return r
}

func (r *Response) RefreshSpace(uri ...string) *Response {
	//app-space[uri="/${uri}"] Provide the URI of the spaces to refresh
	r.headers[transport.ResponseRefresh] = strings.Join(uri, ",")
	return r
}

func (r *Response) RefreshReferer() *Response {
	r.headers[transport.ResponseRefresh] = "referer"
	return r
}

func (r *Response) CloseModal() *Response {
	r.headers[transport.ResponseCloseModal] = "1"
	return r
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
