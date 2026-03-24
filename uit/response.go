package uit

import (
	"net/http"
	"strings"

	"github.com/openbyte-os/sdk-go/transport"
)

type Response struct {
	w http.ResponseWriter
}

func NewResponse(w http.ResponseWriter) *Response {
	return &Response{w: w}
}

func (r *Response) NoContent() { r.w.WriteHeader(http.StatusNoContent) }

func Refresh(w http.ResponseWriter) { w.Header().Set(transport.ResponseRefresh, "self") }
func RefreshFragment(w http.ResponseWriter, fragments ...string) {
	w.Header().Set(transport.ResponseRefresh, strings.Join(fragments, ","))
}
func RefreshSpace(w http.ResponseWriter, uri ...string) {
	//app-space[uri="/${uri}"]
	// Provide the URI of the spaces to refresh
	w.Header().Set(transport.ResponseRefresh, strings.Join(uri, ","))
}
func RefreshReferer(w http.ResponseWriter) { w.Header().Set(transport.ResponseRefresh, "referer") }

func CloseModal(w http.ResponseWriter) { w.Header().Set(transport.ResponseCloseModal, "1") }
