package gateway

import (
	"context"
	"net/http"
	"net/url"
)

// Declare a dependency on MembersType version MembersVersion with MembersAction to call Members.
const (
	MembersType    = "workspace-members"
	MembersVersion = 1
	MembersAction  = "list"
)

type Member struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Type   string `json:"type"`
	BPO    string `json:"bpo,omitempty"`
	Source string `json:"source"`
}

// Members lists the workspace's active members. A non-empty user who is not an owner narrows the list to the BPOs
// they belong to or manage; a user in no BPO sees everyone. The filter is not authentication.
func (w *WorkspaceClient) Members(ctx context.Context, user string) ([]Member, error) {
	req, err := w.newRequest(ctx, http.MethodGet, []string{"members"}, nil)
	if err != nil {
		return nil, err
	}
	if user != "" {
		req.URL.RawQuery = url.Values{"user": {user}}.Encode()
	}
	response, err := w.send(req)
	if err != nil {
		return nil, err
	}
	var members []Member
	if err := decodeEnvelope(response, &members); err != nil {
		return nil, err
	}
	return members, nil
}
