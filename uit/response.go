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

// RefreshDataTable reloads a data table's rows and leaves the page around it
// alone, so the reader keeps the search, filters and page they were on. With no
// selector it reloads the table the request came from.
func (r *ResponseWriter) RefreshDataTable(selectors ...string) *ResponseWriter {
	value := "self"
	if len(selectors) > 0 {
		value = strings.Join(selectors, ",")
	}
	r.headers[transport.ResponseRefreshDataTable] = value
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
