package uit

import (
	"net/http"

	"github.com/openbyte-os/sdk-go/transport"
)

func Brands(r *http.Request) []transport.Resource {
	return transport.GetResources(r.Header.Get(transport.RequestBrands))
}

func Departments(r *http.Request) []transport.Resource {
	return transport.GetResources(r.Header.Get(transport.RequestDepartments))
}

func Channels(r *http.Request) []transport.Resource {
	return transport.GetResources(r.Header.Get(transport.RequestChannels))
}
