package transport

import (
	"bufio"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/textproto"
	"strconv"
	"strings"
	"time"

	"github.com/openbyte-os/sdk-go/app"
)

type RequestContext struct {
	WorkspaceID    string
	UserID         string
	signature      string
	TraceID        string
	UserIP         string
	UserAgent      string
	Authorization  []app.PermissionStatement
	Authentication json.RawMessage
	OwnAppID       *app.GlobalAppID
}

func NewContext(headers map[string][]string, ownAppID *app.GlobalAppID) *RequestContext {
	c := &RequestContext{
		OwnAppID: ownAppID,
	}
	c.ApplyHeaders(headers)
	return c
}

func NewContextFromRaw(rawHeaders io.Reader, ownAppID *app.GlobalAppID) (*RequestContext, error) {
	reader := textproto.NewReader(bufio.NewReader(rawHeaders))
	headers, err := reader.ReadMIMEHeader()
	if headers != nil && io.EOF != err && err != nil {
		return nil, err
	}
	return NewContext(headers, ownAppID), nil
}

func (r *RequestContext) ApplyHeaders(headers map[string][]string) {
	for k, vs := range headers {
		if len(vs) == 0 {
			continue
		}
		switch strings.ToLower(k) {
		case RequestWorkspaceID:
			r.WorkspaceID = vs[0]
		case RequestUserID:
			r.UserID = vs[0]
		case RequestSignature:
			r.signature = vs[0]
		case RequestTraceID:
			r.TraceID = vs[0]
		case RequestUserIP:
			r.UserIP = vs[0]
		case RequestUserAgent:
			r.UserAgent = vs[0]
		case RequestAuthentication:
			r.Authentication = json.RawMessage(vs[0])
		case RequestAuthorization:
			_ = json.Unmarshal([]byte(vs[0]), &r.Authorization)
		}
	}
}

func (r RequestContext) Verify(signatureKey string, maxTimeDiff int64) error {
	if r.signature == "" || !strings.Contains(r.signature, "/") {
		return errors.New("invalid or missing signature")
	}
	splits := strings.SplitN(r.signature, "/", 2)

	timestamp, _ := strconv.ParseInt(splits[1], 10, 64)

	now := time.Now().Unix()
	if timestamp > (now+maxTimeDiff) || timestamp < (now-maxTimeDiff) {
		return errors.New("signature outside of available time window")
	}

	verifyString := ""
	verifyString += r.WorkspaceID
	verifyString += r.UserID
	verifyString += signatureKey
	verifyString += r.TraceID
	verifyString += r.UserIP
	verifyString += r.UserAgent
	verifyString += splits[1]

	signature := sha256.Sum256([]byte(verifyString))

	if subtle.ConstantTimeCompare([]byte(splits[0]), []byte(fmt.Sprintf("%x", signature))) != 1 {
		return errors.New("unable to verify signature")
	}

	return nil
}

func (r RequestContext) HasPermission(perm app.ScopedKey) bool {
	allowed := false
	if perm.GlobalAppID.VendorID == "" && perm.GlobalAppID.AppID == "" && r.OwnAppID != nil {
		perm.GlobalAppID = *r.OwnAppID
	}

	for _, statement := range r.Authorization {
		if perm.GlobalAppID.Matches(statement.Permission.GlobalAppID, false) && perm.Key == statement.Permission.Key {
			if statement.Effect == app.PermissionEffectDeny {
				return false
			}
			if statement.Effect == app.PermissionEffectAllow {
				allowed = true
			}
		}
	}
	return allowed
}

func (r RequestContext) HasResourcePermission(perm app.ScopedKey, resource string) bool {
	allowed := false
	for _, statement := range r.Authorization {
		if perm.GlobalAppID.Matches(statement.Permission.GlobalAppID, false) && perm.Key == statement.Permission.Key {
			matchesResource := statement.Resource == app.PermissionResourceAll || resource == statement.Resource ||
				(strings.HasSuffix(statement.Resource, app.PermissionResourceAll) &&
					strings.HasPrefix(resource, statement.Resource[:len(statement.Resource)-1]))

			if matchesResource {
				if statement.Effect == app.PermissionEffectDeny {
					return false
				}
				if statement.Effect == app.PermissionEffectAllow {
					allowed = true
				}
			}
		}
	}
	return allowed
}

func (r RequestContext) PermittedResources(perm app.ScopedKey) []string {
	var resources []string
	for _, statement := range r.Authorization {
		if perm.GlobalAppID.Matches(statement.Permission.GlobalAppID, false) && perm.Key == statement.Permission.Key && statement.Effect == app.PermissionEffectAllow {
			resources = append(resources, statement.Resource)
		}
	}
	return resources
}
func (r RequestContext) DeniedResources(perm app.ScopedKey) []string {
	var resources []string
	for _, statement := range r.Authorization {
		if perm.GlobalAppID.Matches(statement.Permission.GlobalAppID, false) && perm.Key == statement.Permission.Key && statement.Effect == app.PermissionEffectDeny {
			resources = append(resources, statement.Resource)
		}
	}
	return resources
}
