package gateway

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Identity struct {
	App       string `json:"app"`
	Workspace string `json:"workspace"`
}

type Service struct {
	Type    string   `json:"type"`
	Version int      `json:"version"`
	Actions []string `json:"actions"`
}

// Token contains a secret returned once by RotateToken. Store it privately.
type Token struct {
	Token     string `json:"token"`
	TokenType string `json:"tokenType"`
}

func (t Token) String() string {
	return fmt.Sprintf("gateway.Token{Token:[REDACTED], TokenType:%q}", t.TokenType)
}

func (t Token) GoString() string { return t.String() }

type Response struct {
	StatusCode int
	TraceID    string
	Header     http.Header
	Body       []byte
}

// Error is a non-success HTTP response. Use errors.As to inspect it. Non-Gateway
// provider/proxy errors have StatusCode and a generic Detail without a Type.
type Error struct {
	StatusCode int    `json:"-"`
	TraceID    string `json:"-"`
	Type       string `json:"type"`
	Detail     string `json:"detail"`
	SourceIP   string `json:"sourceIP,omitempty"`
	PeerIP     string `json:"peerIP,omitempty"`
	Reason     string `json:"reason,omitempty"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("gateway: HTTP %d: %s", e.StatusCode, e.Detail)
}

type wireResponse struct {
	Meta struct {
		TraceID string `json:"traceID"`
	} `json:"meta"`
	Status struct {
		Code int `json:"code"`
	} `json:"status"`
	Data  json.RawMessage `json:"data"`
	Error *Error          `json:"error"`
}

func responseError(response *Response) *Error {
	result := &Error{StatusCode: response.StatusCode, TraceID: response.TraceID, Detail: http.StatusText(response.StatusCode)}
	var envelope wireResponse
	if json.Unmarshal(response.Body, &envelope) == nil && envelope.Status.Code == response.StatusCode && envelope.Error != nil {
		result = envelope.Error
		result.StatusCode = response.StatusCode
		result.TraceID = envelope.Meta.TraceID
		if result.Detail == "" {
			result.Detail = http.StatusText(response.StatusCode)
		}
	}
	return result
}
