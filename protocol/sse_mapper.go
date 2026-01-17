package protocol

import (
	"encoding/json"
	"strings"

	"github.com/SaherElMasry/go-mcp-framework/engine"
)

// sseMessage represents a Server-Sent Event message
type sseMessage struct {
	Event string
	ID    string
	Data  string
}

// format formats the SSE message according to SSE specification
func (m sseMessage) format() string {
	var sb strings.Builder

	if m.Event != "" {
		sb.WriteString("event: ")
		sb.WriteString(m.Event)
		sb.WriteString("\n")
	}

	if m.ID != "" {
		sb.WriteString("id: ")
		sb.WriteString(m.ID)
		sb.WriteString("\n")
	}

	// Handle multi-line data
	lines := strings.Split(m.Data, "\n")
	for _, line := range lines {
		sb.WriteString("data: ")
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	// Empty line to separate messages
	sb.WriteString("\n")

	return sb.String()
}

// eventToSSE converts an engine event to SSE message
func eventToSSE(event engine.Event, requestID string) sseMessage {
	msg := sseMessage{
		Event: event.Type.String(),
		ID:    requestID,
	}

	// Serialize event data to JSON
	data, err := json.Marshal(event.Data)
	if err != nil {
		// Fallback to error message
		msg.Data = `{"error":"failed to serialize event data"}`
	} else {
		msg.Data = string(data)
	}

	return msg
}

// streamEventsToSSE converts a channel of events to SSE messages
func streamEventsToSSE(events <-chan engine.Event, requestID string) <-chan sseMessage {
	messages := make(chan sseMessage)

	go func() {
		defer close(messages)

		for event := range events {
			msg := eventToSSE(event, requestID)
			messages <- msg
		}
	}()

	return messages
}

// FormatEventAsSSE is the public API for converting an event to SSE format
// This is used by the HTTP transport layer
func FormatEventAsSSE(event engine.Event, requestID string) string {
	msg := eventToSSE(event, requestID)
	return msg.format()
}
