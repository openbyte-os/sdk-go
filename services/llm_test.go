package services_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/openbyte-os/sdk-go/app"
	"github.com/openbyte-os/sdk-go/services"
)

func userText(text string) services.LLMMessage {
	return services.LLMMessage{Role: services.LLMRoleUser, Content: []services.LLMContent{{Type: services.LLMContentText, Text: text}}}
}

func TestLLMWireContract(t *testing.T) {
	request := services.LLMRequest{Messages: []services.LLMMessage{userText("Hello")}}
	data, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"messages":[{"role":"user","content":[{"type":"text","text":"Hello"}]}]}` {
		t.Fatalf("request JSON: %s", data)
	}
	response := services.LLMResponse{
		Model:      "example-1",
		Content:    []services.LLMContent{{Type: services.LLMContentText, Text: "Hi"}, {Type: services.LLMContentText, Text: " there"}},
		StopReason: services.LLMStopEnd,
		Usage:      services.LLMUsage{InputTokens: 3, OutputTokens: 2},
	}
	data, err = json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"model":"example-1","content":[{"type":"text","text":"Hi"},{"type":"text","text":" there"}],"stopReason":"end","usage":{"inputTokens":3,"outputTokens":2}}` {
		t.Fatalf("response JSON: %s", data)
	}
	if response.Text() != "Hi there" {
		t.Fatalf("text: %q", response.Text())
	}
	data, err = json.Marshal(services.LLMModelsResponse{Models: []services.LLMModel{{ID: "example-1", Default: true, Capabilities: []services.LLMCapability{services.LLMCapabilityTools}}}})
	if err != nil || string(data) != `{"models":[{"id":"example-1","default":true,"capabilities":["tools"]}]}` {
		t.Fatalf("models JSON: %s %v", data, err)
	}
}

func TestLLMFullRequestRoundTrip(t *testing.T) {
	temperature := 0.2
	request := services.LLMRequest{
		Model: "example-1", System: "Be brief.", MaxOutputTokens: 256, Temperature: &temperature,
		StopSequences: []string{"END"},
		Messages: []services.LLMMessage{
			{Role: services.LLMRoleUser, Content: []services.LLMContent{
				{Type: services.LLMContentText, Text: "What is the weather in this photo's city?"},
				{Type: services.LLMContentImage, MediaType: "image/png", Data: "aGVsbG8="},
			}},
			{Role: services.LLMRoleAssistant, Content: []services.LLMContent{
				{Type: services.LLMContentToolUse, ToolUseID: "call-1", Name: "get_weather", Input: json.RawMessage(`{"city":"Paris"}`), Opaque: "state"},
			}},
			{Role: services.LLMRoleUser, Content: []services.LLMContent{
				{Type: services.LLMContentToolResult, ToolUseID: "call-1", Text: "18C, cloudy"},
			}},
		},
		Tools:        []services.LLMTool{{Name: "get_weather", Description: "Current weather for a city.", InputSchema: json.RawMessage(`{"type":"object"}`)}},
		ToolChoice:   &services.LLMToolChoice{Type: services.LLMToolChoiceAuto},
		OutputSchema: json.RawMessage(`{"type":"object"}`),
	}
	if err := request.Validate(); err != nil {
		t.Fatal(err)
	}
	if got := request.RequiredCapabilities(); !reflect.DeepEqual(got, []services.LLMCapability{services.LLMCapabilityTools, services.LLMCapabilityImages, services.LLMCapabilityStructuredOutput}) {
		t.Fatalf("required capabilities: %v", got)
	}
	if got := (services.LLMRequest{Messages: []services.LLMMessage{userText("Hi")}}).RequiredCapabilities(); len(got) != 0 {
		t.Fatalf("plain text requires no capabilities: %v", got)
	}
	if (services.Error{Code: services.ErrorCodeUnsupported}).HTTPStatus() != 422 || (services.Error{Code: "future"}).HTTPStatus() != 500 {
		t.Fatal("error status mapping")
	}
	data, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"model":"example-1","system":"Be brief.","messages":[{"role":"user","content":[{"type":"text","text":"What is the weather in this photo's city?"},{"type":"image","mediaType":"image/png","data":"aGVsbG8="}]},{"role":"assistant","content":[{"type":"tool_use","toolUseId":"call-1","name":"get_weather","input":{"city":"Paris"},"opaque":"state"}]},{"role":"user","content":[{"type":"tool_result","text":"18C, cloudy","toolUseId":"call-1"}]}],"maxOutputTokens":256,"temperature":0.2,"stopSequences":["END"],"tools":[{"name":"get_weather","description":"Current weather for a city.","inputSchema":{"type":"object"}}],"toolChoice":{"type":"auto"},"outputSchema":{"type":"object"}}`
	if string(data) != want {
		t.Fatalf("request JSON: %s", data)
	}
	var decoded services.LLMRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded, request) {
		t.Fatalf("request did not round trip: %+v", decoded)
	}
}

func TestLLMValidate(t *testing.T) {
	object := json.RawMessage(`{"type":"object"}`)
	tool := []services.LLMTool{{Name: "lookup", InputSchema: object}}
	call := services.LLMMessage{Role: services.LLMRoleAssistant, Content: []services.LLMContent{{Type: services.LLMContentToolUse, ToolUseID: "a", Name: "lookup", Input: json.RawMessage(`{}`)}}}
	result := func(id string) services.LLMMessage {
		return services.LLMMessage{Role: services.LLMRoleUser, Content: []services.LLMContent{{Type: services.LLMContentToolResult, ToolUseID: id}}}
	}
	assistant := services.LLMMessage{Role: services.LLMRoleAssistant, Content: []services.LLMContent{{Type: services.LLMContentText, Text: "Hi"}}}
	hot, negative := 1.5, -1.0
	for _, tc := range []struct {
		name    string
		request services.LLMRequest
	}{
		{"no messages", services.LLMRequest{}},
		{"starts with assistant", services.LLMRequest{Messages: []services.LLMMessage{assistant, userText("Hi")}}},
		{"ends with assistant", services.LLMRequest{Messages: []services.LLMMessage{userText("Hi"), assistant}}},
		{"consecutive user", services.LLMRequest{Messages: []services.LLMMessage{userText("Hi"), userText("Again")}}},
		{"empty content", services.LLMRequest{Messages: []services.LLMMessage{{Role: services.LLMRoleUser}}}},
		{"empty text", services.LLMRequest{Messages: []services.LLMMessage{userText("")}}},
		{"unknown role", services.LLMRequest{Messages: []services.LLMMessage{userText("Hi"), {Role: "system", Content: assistant.Content}, userText("Hi")}}},
		{"unknown content", services.LLMRequest{Messages: []services.LLMMessage{{Role: services.LLMRoleUser, Content: []services.LLMContent{{Type: "audio"}}}}}},
		{"negative max tokens", services.LLMRequest{Messages: []services.LLMMessage{userText("Hi")}, MaxOutputTokens: -1}},
		{"hot temperature", services.LLMRequest{Messages: []services.LLMMessage{userText("Hi")}, Temperature: &hot}},
		{"negative temperature", services.LLMRequest{Messages: []services.LLMMessage{userText("Hi")}, Temperature: &negative}},
		{"empty stop sequence", services.LLMRequest{Messages: []services.LLMMessage{userText("Hi")}, StopSequences: []string{""}}},
		{"bad tool name", services.LLMRequest{Messages: []services.LLMMessage{userText("Hi")}, Tools: []services.LLMTool{{Name: "look up", InputSchema: object}}}},
		{"duplicate tool", services.LLMRequest{Messages: []services.LLMMessage{userText("Hi")}, Tools: append(tool, tool...)}},
		{"tool schema not object", services.LLMRequest{Messages: []services.LLMMessage{userText("Hi")}, Tools: []services.LLMTool{{Name: "lookup", InputSchema: json.RawMessage(`[]`)}}}},
		{"required without tools", services.LLMRequest{Messages: []services.LLMMessage{userText("Hi")}, ToolChoice: &services.LLMToolChoice{Type: services.LLMToolChoiceRequired}}},
		{"undefined tool choice", services.LLMRequest{Messages: []services.LLMMessage{userText("Hi")}, Tools: tool, ToolChoice: &services.LLMToolChoice{Type: services.LLMToolChoiceTool, Name: "other"}}},
		{"unknown tool choice", services.LLMRequest{Messages: []services.LLMMessage{userText("Hi")}, Tools: tool, ToolChoice: &services.LLMToolChoice{Type: "any"}}},
		{"output schema not object", services.LLMRequest{Messages: []services.LLMMessage{userText("Hi")}, OutputSchema: json.RawMessage(`"string"`)}},
		{"unanswered tool use", services.LLMRequest{Messages: []services.LLMMessage{userText("Hi"), call, userText("Never mind")}}},
		{"unknown tool result", services.LLMRequest{Messages: []services.LLMMessage{userText("Hi"), call, result("b")}}},
		{"orphan tool result", services.LLMRequest{Messages: []services.LLMMessage{result("a")}}},
		{"image in assistant", services.LLMRequest{Messages: []services.LLMMessage{userText("Hi"), {Role: services.LLMRoleAssistant, Content: []services.LLMContent{{Type: services.LLMContentImage, MediaType: "image/png", Data: "aGVsbG8="}}}, userText("Hi")}}},
		{"image media type", services.LLMRequest{Messages: []services.LLMMessage{{Role: services.LLMRoleUser, Content: []services.LLMContent{{Type: services.LLMContentImage, MediaType: "image/svg+xml", Data: "aGVsbG8="}}}}}},
		{"image base64", services.LLMRequest{Messages: []services.LLMMessage{{Role: services.LLMRoleUser, Content: []services.LLMContent{{Type: services.LLMContentImage, MediaType: "image/png", Data: "not base64!"}}}}}},
		{"image too large", services.LLMRequest{Messages: []services.LLMMessage{{Role: services.LLMRoleUser, Content: []services.LLMContent{{Type: services.LLMContentImage, MediaType: "image/png", Data: strings.Repeat("AAAA", services.LLMMaxImageBytes/3+2)}}}}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var e services.Error
			if err := tc.request.Validate(); !errors.As(err, &e) || e.Code != services.ErrorCodeInvalidRequest || e.Message == "" {
				t.Fatalf("expected invalid_request, got %v", err)
			}
		})
	}
	valid := services.LLMRequest{Messages: []services.LLMMessage{userText("Hi"), call, result("a")}, Tools: tool, ToolChoice: &services.LLMToolChoice{Type: services.LLMToolChoiceTool, Name: "lookup"}}
	if err := valid.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestLLMDefinition(t *testing.T) {
	def := services.LLMDefinition()
	if def.Type != "llm" || def.Version != 1 || len(def.Endpoints) != 3 {
		t.Fatalf("unexpected definition: %+v", def)
	}
	want := map[string][2]string{
		"generate": {http.MethodPost, "/generate"},
		"stream":   {http.MethodPost, "/stream"},
		"models":   {http.MethodPost, "/models"},
	}
	for _, endpoint := range def.Endpoints {
		if want[endpoint.Action] != [2]string{endpoint.Method, endpoint.Path} {
			t.Fatalf("unexpected endpoint: %+v", endpoint)
		}
	}
	def.Endpoints[0].Path = "/changed"
	if services.LLMDefinition().Endpoints[0].Path != "/generate" {
		t.Fatal("caller mutated the shared contract")
	}
}

func TestLLMDeclarationsRoundTrip(t *testing.T) {
	def := app.Definition{
		Providers: []app.ProviderRegistration{{
			Type: services.LLMType, Key: "default", Version: services.LLMVersion, Path: "/_services/llm/v1",
			Actions: []string{services.LLMActionGenerate, services.LLMActionStream, services.LLMActionModels},
		}},
		ProviderDependencies: []app.ProviderDependency{{
			Type: services.LLMType, Version: services.LLMVersion, Required: true,
			RequiredActions: []string{services.LLMActionGenerate},
		}},
	}
	data, err := json.Marshal(def)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := app.FromJson(data)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(def.Providers, decoded.Providers) || !reflect.DeepEqual(def.ProviderDependencies, decoded.ProviderDependencies) {
		t.Fatal("service declarations did not round trip")
	}
}

func TestLLMStream(t *testing.T) {
	final := &services.LLMResponse{
		Model: "example-1", StopReason: services.LLMStopEnd, Usage: services.LLMUsage{InputTokens: 1, OutputTokens: 2},
		Content: []services.LLMContent{{Type: services.LLMContentText, Text: "Hello\nworld"}},
	}
	var wire bytes.Buffer
	for _, e := range []services.LLMStreamEvent{
		{Type: services.LLMStreamText, Text: "Hello\n"},
		{Type: services.LLMStreamText, Text: "world"},
		{Type: services.LLMStreamDone, Response: final},
	} {
		if err := services.WriteLLMStreamEvent(&wire, e); err != nil {
			t.Fatal(err)
		}
	}
	if !strings.HasPrefix(wire.String(), "event: text\ndata: {\"text\":\"Hello\\n\"}\n\nevent: text\n") {
		t.Fatalf("stream framing: %q", wire.String())
	}
	// Comments and unknown events are forward-compatible padding.
	input := ": keep-alive\n\nevent: future\ndata: {}\n\n" + wire.String()
	var text strings.Builder
	got, err := services.ReadLLMStream(strings.NewReader(input), func(s string) error { text.WriteString(s); return nil })
	if err != nil {
		t.Fatal(err)
	}
	if text.String() != "Hello\nworld" || !reflect.DeepEqual(got, final) {
		t.Fatalf("stream result: %q %+v", text.String(), got)
	}

	wire.Reset()
	_ = services.WriteLLMStreamEvent(&wire, services.LLMStreamEvent{Type: services.LLMStreamText, Text: "partial"})
	_ = services.WriteLLMStreamEvent(&wire, services.LLMStreamEvent{Type: services.LLMStreamError, Error: &services.Error{Code: services.ErrorCodeUnavailable, Message: "Upstream closed"}})
	var e services.Error
	if _, err := services.ReadLLMStream(&wire, nil); !errors.As(err, &e) || e.Code != services.ErrorCodeUnavailable {
		t.Fatalf("expected the stream error, got %v", err)
	}
	if _, err := services.ReadLLMStream(strings.NewReader("event: text\ndata: {\"text\":\"cut\"}\n\n"), nil); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("expected a truncated stream error, got %v", err)
	}
	stop := errors.New("stop")
	if _, err := services.ReadLLMStream(strings.NewReader("event: text\ndata: {\"text\":\"cut\"}\n\n"), func(string) error { return stop }); !errors.Is(err, stop) {
		t.Fatalf("expected the emit error, got %v", err)
	}
}
