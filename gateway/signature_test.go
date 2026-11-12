package gateway

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/openbyte-os/sdk-go/transport"
)

const testKey = "a-test-only-signing-key-at-least-32-bytes"

const testDefinition = `{"id":{"vendorID":"acme","appID":"billing"},"name":{"fallback":"Billing"}}`

func signedRequest(body string, now time.Time) *http.Request {
	return signedRequestWithDefinition(body, "", now)
}

func signedRequestWithDefinition(body, definition string, now time.Time) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/_kxop/get?x=1", strings.NewReader(body))
	r.Header.Set(transport.RequestWorkspaceID, "ws-1")
	r.Header.Set(transport.RequestCallerApp, "acme/billing")
	r.Header.Set(transport.RequestTraceID, "trace-1")
	if definition != "" {
		r.Header.Set(transport.RequestAppDefinition, definition)
	}
	SignRequest(r, []byte(body), testKey, now)
	return r
}

func TestVerifyRequestAcceptsGatewaySignature(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	for _, keys := range []string{testKey, "an-older-key-still-being-rotated-out, " + testKey} {
		caller, err := VerifyRequest(signedRequest(`{"id":"1"}`, now), []byte(`{"id":"1"}`), keys, now.Add(10*time.Second))
		if err != nil || caller.Workspace != "ws-1" || caller.App.String() != "acme/billing" || caller.Definition != nil {
			t.Fatalf("keys %q: caller %+v, err %v", keys, caller, err)
		}
	}
}

func TestVerifyRequestReturnsTheSignedCallerDefinition(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	body := `{"id":"1"}`
	caller, err := VerifyRequest(signedRequestWithDefinition(body, testDefinition, now), []byte(body), testKey, now)
	if err != nil || caller.Definition == nil || caller.Definition.ID.String() != "acme/billing" {
		t.Fatalf("caller %+v, err %v", caller, err)
	}
	other := `{"id":{"vendorID":"acme","appID":"other"},"name":{"fallback":"Other"}}`
	caller, err = VerifyRequest(signedRequestWithDefinition(body, other, now), []byte(body), testKey, now)
	if err != nil || caller.Definition != nil {
		t.Fatalf("another app's definition was attributed to the caller: %+v, %v", caller, err)
	}
	for name, change := range map[string]func(*http.Request){
		"changed definition":  func(r *http.Request) { r.Header.Set(transport.RequestAppDefinition, other) },
		"removed definition":  func(r *http.Request) { r.Header.Del(transport.RequestAppDefinition) },
		"repeated definition": func(r *http.Request) { r.Header.Add(transport.RequestAppDefinition, testDefinition) },
	} {
		t.Run(name, func(t *testing.T) {
			r := signedRequestWithDefinition(body, testDefinition, now)
			change(r)
			if _, err := VerifyRequest(r, []byte(body), testKey, now); !errors.Is(err, ErrUnsigned) {
				t.Fatalf("got %v, want ErrUnsigned", err)
			}
		})
	}
}

func TestVerifyRequestRejectsAnythingElse(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	body := `{"id":"1"}`
	for name, tc := range map[string]struct {
		change func(*http.Request)
		body   string
		keys   string
		at     time.Time
	}{
		"different body":      {func(*http.Request) {}, `{"id":"2"}`, testKey, now},
		"wrong key":           {func(*http.Request) {}, body, "some-other-key", now},
		"no keys":             {func(*http.Request) {}, body, "", now},
		"expired":             {func(*http.Request) {}, body, testKey, now.Add(2 * time.Minute)},
		"from the future":     {func(*http.Request) {}, body, testKey, now.Add(-time.Minute)},
		"changed workspace":   {func(r *http.Request) { r.Header.Set(transport.RequestWorkspaceID, "ws-2") }, body, testKey, now},
		"changed caller":      {func(r *http.Request) { r.Header.Set(transport.RequestCallerApp, "acme/other") }, body, testKey, now},
		"changed path":        {func(r *http.Request) { r.URL.Path = "/_kxop/update" }, body, testKey, now},
		"user identity":       {func(r *http.Request) { r.Header.Set(transport.RequestUserID, "user-1") }, body, testKey, now},
		"bearer token":        {func(r *http.Request) { r.Header.Set("Authorization", "Bearer x") }, body, testKey, now},
		"missing signature":   {func(r *http.Request) { r.Header.Del(SignatureHeader) }, body, testKey, now},
		"repeated workspace":  {func(r *http.Request) { r.Header.Add(transport.RequestWorkspaceID, "ws-1") }, body, testKey, now},
		"malformed signature": {func(r *http.Request) { r.Header.Set(SignatureHeader, "v2/1/abcd") }, body, testKey, now},
		"added definition":    {func(r *http.Request) { r.Header.Set(transport.RequestAppDefinition, testDefinition) }, body, testKey, now},
	} {
		t.Run(name, func(t *testing.T) {
			r := signedRequest(body, now)
			tc.change(r)
			if _, err := VerifyRequest(r, []byte(tc.body), tc.keys, tc.at); !errors.Is(err, ErrUnsigned) {
				t.Fatalf("got %v, want ErrUnsigned", err)
			}
		})
	}
}

// Keep in sync with Rubix's TestGatewayRequestSignatureVector and TestGatewayRequestSignatureVectorWithDefinition.
func TestSignRequestMatchesRubixVectors(t *testing.T) {
	for definition, want := range map[string]string{
		"":                                    "v1/1800000000/11ae560accffcdc61c382c9cea6476462d9cfd7162ae3f0fa912e7e8e8c7994b",
		`{"id":{"vendorID":"v","appID":"a"}}`: "v1/1800000000/8e3fb43868507cfbe9e9aad251929181931468b6b4a8bff6ba693c790bac16b0",
	} {
		r := httptest.NewRequest(http.MethodPost, "/_kxop/publish", nil)
		r.Header.Set(transport.RequestWorkspaceID, "ws")
		r.Header.Set(transport.RequestCallerApp, "v/a")
		r.Header.Set(transport.RequestTraceID, "trace")
		if definition != "" {
			r.Header.Set(transport.RequestAppDefinition, definition)
		}
		SignRequest(r, []byte(`{"title":"backup"}`), testKey, time.Unix(1800000000, 0))
		if got := r.Header.Get(SignatureHeader); got != want {
			t.Errorf("definition %q: signature %q, want %q", definition, got, want)
		}
	}
}
