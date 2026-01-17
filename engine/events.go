package engine

import "time"

// EventType represents the type of streaming event
type EventType int

const (
	// EventStart indicates execution has started
	EventStart EventType = iota

	// EventData indicates a data chunk
	EventData

	// EventProgress indicates progress update
	EventProgress

	// EventEnd indicates successful completion
	EventEnd

	// EventError indicates an error occurred
	EventError
)

// String returns the string representation of EventType
func (t EventType) String() string {
	switch t {
	case EventStart:
		return "start"
	case EventData:
		return "data"
	case EventProgress:
		return "progress"
	case EventEnd:
		return "end"
	case EventError:
		return "error"
	default:
		return "unknown"
	}
}

// Event represents a streaming event
type Event struct {
	Type      EventType
	Timestamp time.Time
	Data      interface{}
}

// StartPayload contains start event data
type StartPayload struct {
	ToolName  string                 `json:"tool_name"`
	RequestID string                 `json:"request_id"`
	Args      map[string]interface{} `json:"args,omitempty"`
}

// DataPayload contains data event payload
type DataPayload struct {
	Chunk    interface{} `json:"chunk"`
	Sequence int64       `json:"sequence"`
	Total    *int64      `json:"total,omitempty"`
}

// ProgressPayload contains progress event data
type ProgressPayload struct {
	Current    int64   `json:"current"`
	Total      int64   `json:"total"`
	Percentage float64 `json:"percentage"`
	Message    string  `json:"message,omitempty"`
}

// EndPayload contains completion event data
type EndPayload struct {
	Duration   time.Duration `json:"duration_ms"`
	EventCount int64         `json:"event_count,omitempty"`
	Summary    string        `json:"summary,omitempty"`
}

// ErrorPayload contains error event data
type ErrorPayload struct {
	Error     error  `json:"-"`
	Message   string `json:"message"`
	Retryable bool   `json:"retryable"`
}

// Event constructors

// NewStartEvent creates a start event
func NewStartEvent(toolName, requestID string, args map[string]interface{}) Event {
	return Event{
		Type:      EventStart,
		Timestamp: time.Now(),
		Data: StartPayload{
			ToolName:  toolName,
			RequestID: requestID,
			Args:      args,
		},
	}
}

// NewDataEvent creates a data event
func NewDataEvent(chunk interface{}, sequence int64) Event {
	return Event{
		Type:      EventData,
		Timestamp: time.Now(),
		Data: DataPayload{
			Chunk:    chunk,
			Sequence: sequence,
		},
	}
}

// NewProgressEvent creates a progress event
func NewProgressEvent(current, total int64, message string) Event {
	percentage := 0.0
	if total > 0 {
		percentage = (float64(current) / float64(total)) * 100.0
	}

	return Event{
		Type:      EventProgress,
		Timestamp: time.Now(),
		Data: ProgressPayload{
			Current:    current,
			Total:      total,
			Percentage: percentage,
			Message:    message,
		},
	}
}

// NewEndEvent creates an end event
func NewEndEvent(duration time.Duration, eventCount int64, summary string) Event {
	return Event{
		Type:      EventEnd,
		Timestamp: time.Now(),
		Data: EndPayload{
			Duration:   duration,
			EventCount: eventCount,
			Summary:    summary,
		},
	}
}

// NewErrorEvent creates an error event
func NewErrorEvent(err error, message string, retryable bool) Event {
	if message == "" && err != nil {
		message = err.Error()
	}

	return Event{
		Type:      EventError,
		Timestamp: time.Now(),
		Data: ErrorPayload{
			Error:     err,
			Message:   message,
			Retryable: retryable,
		},
	}
}
