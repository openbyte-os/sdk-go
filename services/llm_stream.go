package services

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

// LLMStreamContentType is the Content-Type of a successful stream response.
const LLMStreamContentType = "text/event-stream"

// LLMStreamEventType names a server-sent event of the stream action.
type LLMStreamEventType string

const (
	LLMStreamText  LLMStreamEventType = "text"  // Text holds the next increment of response text.
	LLMStreamDone  LLMStreamEventType = "done"  // Response holds the complete response. Last event of a successful stream.
	LLMStreamError LLMStreamEventType = "error" // Error holds the failure. Last event of a failed stream.
)

// LLMStreamEvent is one server-sent event. Its Type is the SSE event name and
// the remaining fields are the event's JSON data. A stream ends with exactly
// one done or error event; one that closes without either has failed.
type LLMStreamEvent struct {
	Type     LLMStreamEventType `json:"-"`
	Text     string             `json:"text,omitempty"`
	Response *LLMResponse       `json:"response,omitempty"`
	Error    *Error             `json:"error,omitempty"`
}

// WriteLLMStreamEvent writes one event in server-sent event framing. Callers
// flush the underlying response after each event.
func WriteLLMStreamEvent(w io.Writer, e LLMStreamEvent) error {
	data, err := json.Marshal(e)
	if err != nil {
		return err
	}
	// JSON encoding escapes line breaks, so the data always fits one line.
	_, err = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", e.Type, data)
	return err
}

// ReadLLMStream reads a stream response body, calling emit for each text
// increment, and returns the complete response from the done event. An error
// event is returned as its Error. Events of unknown types are skipped.
func ReadLLMStream(r io.Reader, emit func(text string) error) (*LLMResponse, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64<<10), LLMMaxRequestBytes)
	var name string
	var data strings.Builder
	for scanner.Scan() {
		line := scanner.Text()
		if field, value, ok := strings.Cut(line, ":"); ok && line != "" {
			value = strings.TrimPrefix(value, " ")
			switch field {
			case "event":
				name = value
			case "data":
				if data.Len() > 0 {
					data.WriteByte('\n')
				}
				data.WriteString(value)
			}
			continue
		}
		if line != "" {
			continue
		}
		event := LLMStreamEvent{Type: LLMStreamEventType(name)}
		payload := data.String()
		name = ""
		data.Reset()
		if payload == "" {
			continue
		}
		switch event.Type {
		case LLMStreamText, LLMStreamDone, LLMStreamError:
		default:
			continue
		}
		if err := json.Unmarshal([]byte(payload), &event); err != nil {
			return nil, fmt.Errorf("invalid %s event: %w", event.Type, err)
		}
		switch event.Type {
		case LLMStreamText:
			if emit != nil {
				if err := emit(event.Text); err != nil {
					return nil, err
				}
			}
		case LLMStreamDone:
			if event.Response == nil {
				return nil, errors.New("done event without a response")
			}
			return event.Response, nil
		case LLMStreamError:
			if event.Error == nil {
				return nil, errors.New("error event without an error")
			}
			return nil, *event.Error
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return nil, io.ErrUnexpectedEOF
}
