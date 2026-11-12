package gateway

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/openbyte-os/sdk-go/app"
	"github.com/openbyte-os/sdk-go/transport"
)

// SignatureHeader carries the gateway's signature on a request it forwards to
// an app operation: v1/<unix seconds>/<hex HMAC-SHA256>.
const SignatureHeader = "X-Kx-Gateway-Signature"

// ErrUnsigned means a request did not carry a valid gateway signature.
var ErrUnsigned = errors.New("gateway: request is not signed by the gateway")

// Caller is the app and workspace the gateway authenticated for a request.
type Caller struct {
	Workspace string
	App       app.GlobalAppID
	// Definition is the caller's public definition, which the gateway sends only
	// to system apps; nil when absent or not the caller's.
	Definition *app.PublicDefinition
}

// VerifyRequest authenticates a request the gateway forwarded to an app
// operation. keys is the app's signature key, or several comma-separated keys
// during rotation. body is the full request body. The gateway has already
// checked the caller's credential and its grant for the operation.
func VerifyRequest(r *http.Request, body []byte, keys string, now time.Time) (Caller, error) {
	for _, name := range []string{SignatureHeader, transport.RequestWorkspaceID, transport.RequestCallerApp, transport.RequestTraceID} {
		if len(r.Header.Values(name)) != 1 || strings.TrimSpace(r.Header.Get(name)) == "" {
			return Caller{}, ErrUnsigned
		}
	}
	// A gateway call carries the app's identity only.
	if len(r.Header.Values("Authorization")) > 0 || len(r.Header.Values(transport.RequestUserID)) > 0 || len(r.Header.Values(transport.RequestAppDefinition)) > 1 {
		return Caller{}, ErrUnsigned
	}
	parts := strings.Split(r.Header.Get(SignatureHeader), "/")
	if len(parts) != 3 || parts[0] != "v1" {
		return Caller{}, ErrUnsigned
	}
	timestamp, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || timestamp < now.Unix()-60 || timestamp > now.Unix()+5 {
		return Caller{}, ErrUnsigned
	}
	signature, err := hex.DecodeString(parts[2])
	if err != nil || len(signature) != sha256.Size {
		return Caller{}, ErrUnsigned
	}
	workspace := r.Header.Get(transport.RequestWorkspaceID)
	caller := app.IDFromString(r.Header.Get(transport.RequestCallerApp))
	if !validSegment(workspace) || app.ValidateVendorID(caller.VendorID) != nil || app.ValidateID(caller.AppID) != nil || caller.String() != r.Header.Get(transport.RequestCallerApp) {
		return Caller{}, ErrUnsigned
	}
	message := signedMessage(r, body, parts[1])
	valid := false
	for key := range strings.SplitSeq(keys, ",") {
		if key = strings.TrimSpace(key); key == "" {
			continue
		}
		mac := hmac.New(sha256.New, []byte(key))
		_, _ = mac.Write(message)
		valid = hmac.Equal(signature, mac.Sum(nil)) || valid
	}
	if !valid {
		return Caller{}, ErrUnsigned
	}
	verified := Caller{Workspace: workspace, App: caller}
	var def app.PublicDefinition
	if json.Unmarshal([]byte(r.Header.Get(transport.RequestAppDefinition)), &def) == nil && def.ID.String() == caller.String() {
		verified.Definition = &def
	}
	return verified, nil
}

// SignRequest signs a request the way the gateway does. The workspace, caller
// app, trace ID and any app definition headers must already be set.
func SignRequest(r *http.Request, body []byte, key string, now time.Time) {
	timestamp := strconv.FormatInt(now.Unix(), 10)
	mac := hmac.New(sha256.New, []byte(key))
	_, _ = mac.Write(signedMessage(r, body, timestamp))
	r.Header.Set(SignatureHeader, "v1/"+timestamp+"/"+hex.EncodeToString(mac.Sum(nil)))
}

func signedMessage(r *http.Request, body []byte, timestamp string) []byte {
	bodyHash := sha256.Sum256(body)
	fields := []string{
		"rubix-gateway-v1", r.Method, r.URL.RequestURI(),
		r.Header.Get(transport.RequestWorkspaceID), r.Header.Get(transport.RequestCallerApp), r.Header.Get(transport.RequestTraceID),
		timestamp, hex.EncodeToString(bodyHash[:]),
	}
	// Keep in sync with Rubix gateway/request_signature.go: bound only when sent.
	if def := r.Header.Get(transport.RequestAppDefinition); def != "" {
		defHash := sha256.Sum256([]byte(def))
		fields = append(fields, hex.EncodeToString(defHash[:]))
	}
	message, _ := json.Marshal(fields)
	return message
}
