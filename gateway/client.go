// Package gateway provides an HTTP client for Rubix Gateway's workspace-scoped API.
package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/openbyte-os/sdk-go/app"
)

const DefaultMaxResponseBytes int64 = 8 << 20

// Client can be shared by goroutines. Credentials and configuration are immutable.
type Client struct {
	baseURL          string
	token            string
	httpClient       *http.Client
	maxResponseBytes int64
}

func (c *Client) String() string {
	return fmt.Sprintf("gateway.Client{BaseURL:%q, Token:[REDACTED]}", c.baseURL)
}

func (c *Client) GoString() string { return c.String() }

type Option func(*Client) error

// WithHTTPClient copies client settings. Redirects are always disabled to keep
// the Bearer token on the configured origin. The supplied client is not modified.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) error {
		if client == nil {
			return errors.New("gateway: HTTP client is required")
		}
		clone := *client
		clone.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
		c.httpClient = &clone
		return nil
	}
}

func WithMaxResponseBytes(limit int64) Option {
	return func(c *Client) error {
		if limit <= 0 || limit == int64(^uint64(0)>>1) {
			return errors.New("gateway: response size limit must be positive and bounded")
		}
		c.maxResponseBytes = limit
		return nil
	}
}

// NewClient takes an origin (optionally with a path prefix), not a workspace URL.
// HTTP supports direct in-cluster calls; use HTTPS for public endpoints.
func NewClient(baseURL, token string, options ...Option) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil || u.Hostname() == "" || (u.Scheme != "https" && u.Scheme != "http") || u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || u.Opaque != "" {
		return nil, errors.New("gateway: base URL must be an HTTP(S) origin without credentials, query or fragment")
	}
	if err := validToken(token); err != nil {
		return nil, err
	}
	c := &Client{baseURL: strings.TrimRight(u.String(), "/"), token: token, maxResponseBytes: DefaultMaxResponseBytes}
	_ = WithHTTPClient(&http.Client{Timeout: 30 * time.Second})(c)
	for _, option := range options {
		if option == nil {
			return nil, errors.New("gateway: nil client option")
		}
		if err := option(c); err != nil {
			return nil, err
		}
	}
	return c, nil
}

func validToken(token string) error {
	if token == "" || strings.ContainsFunc(token, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) }) {
		return errors.New("gateway: a nonempty Bearer token without whitespace is required")
	}
	return nil
}

// WithToken returns a new client. Persist a rotated token before adopting it.
func (c *Client) WithToken(token string) (*Client, error) {
	if err := validToken(token); err != nil {
		return nil, err
	}
	clone := *c
	clone.token = token
	return &clone, nil
}

type WorkspaceClient struct {
	client    *Client
	workspace string
}

func (w *WorkspaceClient) String() string {
	return fmt.Sprintf("gateway.WorkspaceClient{Workspace:%q, Client:%s}", w.workspace, w.client)
}

func (w *WorkspaceClient) GoString() string { return w.String() }

// Workspace scopes subsequent requests without changing the shared client.
func (c *Client) Workspace(id string) *WorkspaceClient {
	return &WorkspaceClient{client: c, workspace: id}
}

func validSegment(value string) bool {
	return value != "" && value != "." && value != ".." && strings.TrimSpace(value) == value && !strings.ContainsAny(value, "/\\?#%") && !strings.ContainsFunc(value, unicode.IsControl)
}

func (w *WorkspaceClient) request(ctx context.Context, method string, segments []string, input any) (*Response, error) {
	if !validSegment(w.workspace) {
		return nil, errors.New("gateway: invalid workspace ID")
	}
	parts := []string{"v1", "workspaces", url.PathEscape(w.workspace)}
	for _, segment := range segments {
		if !validSegment(segment) {
			return nil, errors.New("gateway: invalid route segment")
		}
		parts = append(parts, url.PathEscape(segment))
	}
	var body io.Reader
	if method == http.MethodPost {
		encoded, err := json.Marshal(input)
		if err != nil {
			return nil, fmt.Errorf("gateway: encode request: %w", err)
		}
		body = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, w.client.baseURL+"/"+strings.Join(parts, "/"), body)
	if err != nil {
		return nil, fmt.Errorf("gateway: create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+w.client.token)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := w.client.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gateway: send request: %w", err)
	}
	defer resp.Body.Close()
	payload, err := io.ReadAll(io.LimitReader(resp.Body, w.client.maxResponseBytes+1))
	if err != nil {
		return nil, fmt.Errorf("gateway: read response: %w", err)
	}
	if int64(len(payload)) > w.client.maxResponseBytes {
		return nil, errors.New("gateway: response exceeds size limit")
	}
	response := &Response{StatusCode: resp.StatusCode, Header: resp.Header.Clone(), Body: payload, TraceID: resp.Header.Get("X-Kx-Trace-Id")}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return response, responseError(response)
	}
	return response, nil
}

func decodeEnvelope(response *Response, output any) error {
	var envelope wireResponse
	if json.Unmarshal(response.Body, &envelope) != nil || envelope.Status.Code != response.StatusCode || len(envelope.Data) == 0 || envelope.Error != nil {
		return errors.New("gateway: invalid response envelope")
	}
	response.TraceID = envelope.Meta.TraceID
	if output != nil {
		if err := json.Unmarshal(envelope.Data, output); err != nil {
			return fmt.Errorf("gateway: decode response data: %w", err)
		}
	}
	return nil
}

func (w *WorkspaceClient) Identity(ctx context.Context) (Identity, error) {
	var identity Identity
	response, err := w.request(ctx, http.MethodGet, []string{"identity"}, nil)
	if err != nil {
		return identity, err
	}
	err = decodeEnvelope(response, &identity)
	return identity, err
}

func (w *WorkspaceClient) Services(ctx context.Context) ([]Service, error) {
	response, err := w.request(ctx, http.MethodGet, []string{"services"}, nil)
	if err != nil {
		return nil, err
	}
	var services []Service
	if err := decodeEnvelope(response, &services); err != nil {
		return nil, err
	}
	return services, nil
}

// CallService decodes the provider's raw JSON response, which is not wrapped in
// a Gateway envelope. Pass nil as output to consume Response.Body yourself.
func (w *WorkspaceClient) CallService(ctx context.Context, kind string, version int, action string, input, output any) (*Response, error) {
	if version <= 0 {
		return nil, errors.New("gateway: service version must be positive")
	}
	response, err := w.request(ctx, http.MethodPost, []string{"services", kind, strconv.Itoa(version), action}, input)
	if err != nil {
		return response, err
	}
	if output != nil && response.StatusCode != http.StatusNoContent {
		if err := json.Unmarshal(response.Body, output); err != nil {
			return response, fmt.Errorf("gateway: decode service response: %w", err)
		}
	}
	return response, nil
}

// CallOperation decodes data from the Gateway operation response envelope.
func (w *WorkspaceClient) CallOperation(ctx context.Context, target app.GlobalAppID, operation string, input, output any) (*Response, error) {
	response, err := w.request(ctx, http.MethodPost, []string{"apps", target.VendorID, target.AppID, "operations", operation}, input)
	if err != nil {
		return response, err
	}
	if response.StatusCode == http.StatusNoContent && len(response.Body) == 0 {
		return response, nil
	}
	return response, decodeEnvelope(response, output)
}

// RotateToken retires the current token immediately. It does not persist or
// install the replacement. Do not retry if the response is lost: rotation may
// already have succeeded. Recovery requires a new token from the developer portal.
func (w *WorkspaceClient) RotateToken(ctx context.Context) (Token, error) {
	var token Token
	response, err := w.request(ctx, http.MethodPost, []string{"token", "rotate"}, nil)
	if err != nil {
		return token, err
	}
	if err := decodeEnvelope(response, &token); err != nil {
		return Token{}, err
	}
	if validToken(token.Token) != nil || token.TokenType != "Bearer" {
		return Token{}, errors.New("gateway: invalid token rotation response")
	}
	return token, nil
}

func (w *WorkspaceClient) RevokeToken(ctx context.Context) error {
	_, err := w.request(ctx, http.MethodDelete, []string{"token"}, nil)
	return err
}
