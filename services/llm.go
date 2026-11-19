package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"unicode/utf8"
)

const (
	LLMType    = "llm"
	LLMVersion = 1

	LLMActionGenerate = "generate"
	LLMActionStream   = "stream"
	LLMActionModels   = "models"

	LLMPathGenerate = "/generate"
	LLMPathStream   = "/stream"
	LLMPathModels   = "/models"
)

// LLMDefinition returns the standard LLM v1 contract.
func LLMDefinition() Definition {
	return Definition{
		Type:        LLMType,
		Version:     LLMVersion,
		Description: "Generate text from a conversation with a large language model, with tools, image input, and structured output.",
		Endpoints: []Endpoint{
			{Action: LLMActionGenerate, Method: http.MethodPost, Path: LLMPathGenerate, Description: "Generate one complete response."},
			{Action: LLMActionStream, Method: http.MethodPost, Path: LLMPathStream, Description: "Generate one response as server-sent events."},
			{Action: LLMActionModels, Method: http.MethodPost, Path: LLMPathModels, Description: "List the models the selected instance can use."},
		},
	}
}

// LLM is the Go implementation interface for the LLM contract. HTTP handlers
// decode LLMRequest and return LLMResponse or LLMModelsResponse as JSON with
// status 200, or ErrorResponse with a non-2xx status.
type LLM interface {
	Generate(context.Context, LLMRequest) (*LLMResponse, error)
	// Stream calls emit with each increment of response text, in order, and
	// returns the complete response once generation ends. An emit error stops
	// generation and is returned.
	Stream(ctx context.Context, request LLMRequest, emit func(text string) error) (*LLMResponse, error)
	Models(context.Context) (*LLMModelsResponse, error)
}

// LLMRole identifies the author of a message.
type LLMRole string

const (
	LLMRoleUser      LLMRole = "user"
	LLMRoleAssistant LLMRole = "assistant"
)

// LLMContentType selects which LLMContent fields are set.
type LLMContentType string

const (
	LLMContentText       LLMContentType = "text"
	LLMContentImage      LLMContentType = "image"
	LLMContentToolUse    LLMContentType = "tool_use"
	LLMContentToolResult LLMContentType = "tool_result"
)

// LLMContent is one block of a message.
//
//   - text: Text. Either role.
//   - image: MediaType and base64 Data. User messages only.
//   - tool_use: ToolUseID, Name, and Input, a JSON object. Assistant messages only.
//   - tool_result: ToolUseID of the answered tool_use, Text as the tool's output,
//     and IsError when the tool failed. User messages only.
type LLMContent struct {
	Type      LLMContentType  `json:"type"`
	Text      string          `json:"text,omitempty"`
	MediaType string          `json:"mediaType,omitempty"`
	Data      string          `json:"data,omitempty"`
	ToolUseID string          `json:"toolUseId,omitempty"`
	Name      string          `json:"name,omitempty"`
	Input     json.RawMessage `json:"input,omitempty"`
	IsError   bool            `json:"isError,omitempty"`
	// Opaque is provider state attached to a block of a response. Callers return
	// the block unchanged, Opaque included, when they send it back in Messages;
	// a provider may be unable to continue a tool exchange without it.
	Opaque string `json:"opaque,omitempty"`
}

// LLMMessage is one conversation turn. Content must contain at least one block.
type LLMMessage struct {
	Role    LLMRole      `json:"role"`
	Content []LLMContent `json:"content"`
}

// LLMTool is a function the model may ask the caller to run. InputSchema is a
// JSON Schema object describing the tool_use Input.
type LLMTool struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

// LLMToolChoiceType controls whether the model calls tools.
type LLMToolChoiceType string

const (
	LLMToolChoiceAuto     LLMToolChoiceType = "auto"     // The model decides. Default when tools are supplied.
	LLMToolChoiceRequired LLMToolChoiceType = "required" // The model must call at least one tool.
	LLMToolChoiceNone     LLMToolChoiceType = "none"     // The model must answer without calling a tool.
	LLMToolChoiceTool     LLMToolChoiceType = "tool"     // The model must call the tool named by Name.
)

type LLMToolChoice struct {
	Type LLMToolChoiceType `json:"type"`
	Name string            `json:"name,omitempty"`
}

// LLMRequest is the JSON body of POST <provider path>/generate and /stream.
// Messages must be non-empty, begin with a user message, and end with a user
// message. Message content is data for the model: providers pass it through
// and never act on it themselves.
type LLMRequest struct {
	// Model is an ID from the models action. Empty selects the instance's default.
	Model string `json:"model,omitempty"`
	// System is the instruction that frames the whole conversation.
	System   string       `json:"system,omitempty"`
	Messages []LLMMessage `json:"messages"`
	// MaxOutputTokens caps generated tokens. Zero uses the provider's default;
	// negative is invalid.
	MaxOutputTokens int `json:"maxOutputTokens,omitempty"`
	// Temperature is a sampling hint from 0 (most deterministic) to 1 (most
	// varied), scaled by each provider to its vendor's range. Nil uses the
	// model's default. Providers ignore it for models with fixed sampling.
	Temperature *float64 `json:"temperature,omitempty"`
	// StopSequences end generation when the model produces one of them. The
	// matched sequence is excluded from the response.
	StopSequences []string       `json:"stopSequences,omitempty"`
	Tools         []LLMTool      `json:"tools,omitempty"`
	ToolChoice    *LLMToolChoice `json:"toolChoice,omitempty"`
	// OutputSchema is a JSON Schema object. When set, the response's text is a
	// single JSON document conforming to it.
	OutputSchema json.RawMessage `json:"outputSchema,omitempty"`
}

// LLMStopReason states why generation ended.
type LLMStopReason string

const (
	LLMStopEnd           LLMStopReason = "end"            // The model finished its answer.
	LLMStopMaxTokens     LLMStopReason = "max_tokens"     // The output token cap was reached; the content is truncated.
	LLMStopSequence      LLMStopReason = "stop_sequence"  // A requested stop sequence was produced.
	LLMStopToolUse       LLMStopReason = "tool_use"       // The model is waiting for results of the tool_use blocks in Content.
	LLMStopContentFilter LLMStopReason = "content_filter" // The vendor withheld or cut short the answer on policy grounds.
)

// LLMUsage counts tokens as billed by the vendor for this request.
type LLMUsage struct {
	InputTokens  int `json:"inputTokens"`
	OutputTokens int `json:"outputTokens"`
}

// LLMResponse is the assistant's turn. Content holds text and tool_use blocks
// in generation order; append it to Messages as an assistant message to
// continue the conversation. Model is the model that generated the response.
type LLMResponse struct {
	Model      string        `json:"model"`
	Content    []LLMContent  `json:"content"`
	StopReason LLMStopReason `json:"stopReason"`
	Usage      LLMUsage      `json:"usage"`
}

// Text returns the concatenated text blocks of the response.
func (r LLMResponse) Text() string {
	var b strings.Builder
	for _, c := range r.Content {
		if c.Type == LLMContentText {
			b.WriteString(c.Text)
		}
	}
	return b.String()
}

// LLMCapability is an optional request feature a model supports.
type LLMCapability string

const (
	LLMCapabilityTools            LLMCapability = "tools"
	LLMCapabilityImages           LLMCapability = "images"
	LLMCapabilityStructuredOutput LLMCapability = "structured_output"
)

// RequiredCapabilities returns the optional features the request uses, which
// its model must support.
func (r LLMRequest) RequiredCapabilities() []LLMCapability {
	tools, images := len(r.Tools) > 0, false
	for _, m := range r.Messages {
		for _, c := range m.Content {
			tools = tools || c.Type == LLMContentToolUse || c.Type == LLMContentToolResult
			images = images || c.Type == LLMContentImage
		}
	}
	var required []LLMCapability
	if tools {
		required = append(required, LLMCapabilityTools)
	}
	if images {
		required = append(required, LLMCapabilityImages)
	}
	if len(r.OutputSchema) > 0 {
		required = append(required, LLMCapabilityStructuredOutput)
	}
	return required
}

// LLMModel describes one model available to the selected instance. Token
// limits are zero when the vendor does not publish them.
type LLMModel struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
	// Default marks the model used when a request omits Model. At most one
	// model per response is the default.
	Default          bool            `json:"default,omitempty"`
	Capabilities     []LLMCapability `json:"capabilities,omitempty"`
	InputTokenLimit  int             `json:"inputTokenLimit,omitempty"`
	OutputTokenLimit int             `json:"outputTokenLimit,omitempty"`
}

// LLMModelsResponse is the JSON response of POST <provider path>/models, which
// takes an empty JSON object as its request. Models is always an array.
type LLMModelsResponse struct {
	Models []LLMModel `json:"models"`
}

// LLM request limits every provider accepts. Providers reject requests beyond
// them with 400 invalid_request.
const (
	LLMMaxRequestBytes = 16 << 20
	LLMMaxMessages     = 1024
	LLMMaxTools        = 128
	LLMMaxStopSequence = 16
	LLMMaxImageBytes   = 5 << 20 // Decoded size of one image.
)

// LLMImageMediaTypes are the image formats every image-capable provider accepts.
var LLMImageMediaTypes = []string{"image/png", "image/jpeg", "image/gif", "image/webp"}

func llmInvalid(message string) error {
	return Error{Code: ErrorCodeInvalidRequest, Message: message}
}

// Validate checks the request against the contract's structural rules and
// returns an Error with code invalid_request for the first one it breaks.
// Model availability and per-model capabilities are the provider's to check.
func (r LLMRequest) Validate() error {
	if !utf8.ValidString(r.System) {
		return llmInvalid("system must be valid UTF-8")
	}
	if len(r.Messages) == 0 || len(r.Messages) > LLMMaxMessages {
		return llmInvalid("messages must contain between 1 and 1024 messages")
	}
	if r.Messages[0].Role != LLMRoleUser || r.Messages[len(r.Messages)-1].Role != LLMRoleUser {
		return llmInvalid("messages must begin and end with a user message")
	}
	if r.MaxOutputTokens < 0 {
		return llmInvalid("maxOutputTokens cannot be negative")
	}
	if r.Temperature != nil && !(*r.Temperature >= 0 && *r.Temperature <= 1) {
		return llmInvalid("temperature must be between 0 and 1")
	}
	if len(r.StopSequences) > LLMMaxStopSequence {
		return llmInvalid("too many stopSequences")
	}
	for _, s := range r.StopSequences {
		if s == "" {
			return llmInvalid("stopSequences cannot contain an empty string")
		}
	}
	tools := map[string]bool{}
	if len(r.Tools) > LLMMaxTools {
		return llmInvalid("too many tools")
	}
	for _, t := range r.Tools {
		if !validLLMToolName(t.Name) || tools[t.Name] {
			return llmInvalid("tool names must be unique and use 1-64 letters, digits, underscores, or hyphens")
		}
		if !jsonObject(t.InputSchema) {
			return llmInvalid("tool inputSchema must be a JSON Schema object")
		}
		tools[t.Name] = true
	}
	if c := r.ToolChoice; c != nil {
		switch c.Type {
		case LLMToolChoiceAuto, LLMToolChoiceNone:
		case LLMToolChoiceRequired:
			if len(tools) == 0 {
				return llmInvalid("toolChoice requires tools")
			}
		case LLMToolChoiceTool:
			if !tools[c.Name] {
				return llmInvalid("toolChoice names an undefined tool")
			}
		default:
			return llmInvalid("invalid toolChoice type")
		}
		if c.Type != LLMToolChoiceTool && c.Name != "" {
			return llmInvalid("toolChoice name applies only to type tool")
		}
	}
	if len(r.OutputSchema) > 0 && !jsonObject(r.OutputSchema) {
		return llmInvalid("outputSchema must be a JSON Schema object")
	}
	return r.validateMessages()
}

// validateMessages checks each block and that tool results answer exactly the
// tool calls of the assistant message before them.
func (r LLMRequest) validateMessages() error {
	var pending map[string]bool // tool_use IDs of the previous assistant message
	for i, m := range r.Messages {
		if len(m.Content) == 0 {
			return llmInvalid("message content cannot be empty")
		}
		if i > 0 && m.Role == r.Messages[i-1].Role {
			return llmInvalid("messages must alternate between user and assistant")
		}
		calls, results := map[string]bool{}, map[string]bool{}
		for _, c := range m.Content {
			if !utf8.ValidString(c.Text) {
				return llmInvalid("content text must be valid UTF-8")
			}
			switch c.Type {
			case LLMContentText:
				if c.Text == "" {
					return llmInvalid("text content cannot be empty")
				}
			case LLMContentImage:
				if m.Role != LLMRoleUser {
					return llmInvalid("image content belongs in user messages")
				}
				if err := validLLMImage(c); err != nil {
					return err
				}
			case LLMContentToolUse:
				if m.Role != LLMRoleAssistant {
					return llmInvalid("tool_use content belongs in assistant messages")
				}
				if c.ToolUseID == "" || c.Name == "" || calls[c.ToolUseID] || !jsonObject(c.Input) {
					return llmInvalid("tool_use content needs a unique toolUseId, a name, and a JSON object input")
				}
				calls[c.ToolUseID] = true
			case LLMContentToolResult:
				if m.Role != LLMRoleUser {
					return llmInvalid("tool_result content belongs in user messages")
				}
				if !pending[c.ToolUseID] || results[c.ToolUseID] {
					return llmInvalid("tool_result must answer one tool_use of the preceding assistant message")
				}
				results[c.ToolUseID] = true
			default:
				return llmInvalid("invalid content type")
			}
		}
		switch m.Role {
		case LLMRoleUser:
			if len(results) != len(pending) {
				return llmInvalid("every tool_use needs a tool_result in the next user message")
			}
			pending = nil
		case LLMRoleAssistant:
			pending = calls
		default:
			return llmInvalid("invalid message role")
		}
	}
	return nil
}

func validLLMImage(c LLMContent) error {
	supported := false
	for _, t := range LLMImageMediaTypes {
		supported = supported || c.MediaType == t
	}
	if !supported {
		return llmInvalid("image mediaType must be image/png, image/jpeg, image/gif, or image/webp")
	}
	if c.Data == "" || base64.StdEncoding.DecodedLen(len(c.Data)) > LLMMaxImageBytes+2 {
		return llmInvalid("image data must be non-empty and at most 5 MiB decoded")
	}
	if _, err := base64.StdEncoding.DecodeString(c.Data); err != nil {
		return llmInvalid("image data must be standard padded base64")
	}
	return nil
}

func validLLMToolName(name string) bool {
	if name == "" || len(name) > 64 {
		return false
	}
	for _, c := range name {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' || c == '-') {
			return false
		}
	}
	return true
}

func jsonObject(raw json.RawMessage) bool {
	var object map[string]json.RawMessage
	return json.Unmarshal(raw, &object) == nil && object != nil
}
