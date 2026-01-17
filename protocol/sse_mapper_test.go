package protocol

import (
	"testing"
	"time"

	"github.com/SaherElMasry/go-mcp-framework/engine"
)

func TestSSEMessage_Format(t *testing.T) {
	tests := []struct {
		name     string
		msg      sseMessage
		expected string
	}{
		{
			name: "simple message",
			msg: sseMessage{
				Event: "test",
				ID:    "123",
				Data:  "hello",
			},
			expected: "event: test\nid: 123\ndata: hello\n\n",
		},
		{
			name: "message without event",
			msg: sseMessage{
				Data: "data only",
			},
			expected: "data: data only\n\n",
		},
		{
			name: "multi-line data",
			msg: sseMessage{
				Event: "message",
				Data:  "line1\nline2\nline3",
			},
			expected: "event: message\ndata: line1\ndata: line2\ndata: line3\n\n",
		},
		{
			name: "empty data",
			msg: sseMessage{
				Event: "ping",
				Data:  "",
			},
			expected: "event: ping\ndata: \n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.msg.format()
			if result != tt.expected {
				t.Errorf("format() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestEventToSSE(t *testing.T) {
	tests := []struct {
		name      string
		event     engine.Event
		requestID string
		wantEvent string
		wantData  bool // Check if data is non-empty
	}{
		{
			name: "start event",
			event: engine.NewStartEvent(
				"test_tool",
				"req-123",
				map[string]interface{}{"arg": "value"},
			),
			requestID: "req-123",
			wantEvent: "start",
			wantData:  true,
		},
		{
			name: "data event",
			event: engine.NewDataEvent(
				map[string]string{"result": "test"},
				1,
			),
			requestID: "req-123",
			wantEvent: "data",
			wantData:  true,
		},
		{
			name: "progress event",
			event: engine.NewProgressEvent(
				50,
				100,
				"Processing...",
			),
			requestID: "req-123",
			wantEvent: "progress",
			wantData:  true,
		},
		{
			name: "end event",
			event: engine.NewEndEvent(
				time.Second,
				42,
				"Completed successfully",
			),
			requestID: "req-123",
			wantEvent: "end",
			wantData:  true,
		},
		{
			name: "error event",
			event: engine.NewErrorEvent(
				nil,
				"test error",
				false,
			),
			requestID: "req-123",
			wantEvent: "error",
			wantData:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := eventToSSE(tt.event, tt.requestID)

			if result.Event != tt.wantEvent {
				t.Errorf("eventToSSE() event = %q, want %q", result.Event, tt.wantEvent)
			}

			if result.ID != tt.requestID {
				t.Errorf("eventToSSE() id = %q, want %q", result.ID, tt.requestID)
			}

			if tt.wantData && result.Data == "" {
				t.Errorf("eventToSSE() data is empty, want non-empty")
			}

			// Verify data is valid JSON
			if result.Data != "" {
				// Simple check - data should start with { or [
				if result.Data[0] != '{' && result.Data[0] != '[' {
					t.Errorf("eventToSSE() data is not valid JSON: %s", result.Data)
				}
			}
		})
	}
}

func TestStreamEventsToSSE(t *testing.T) {
	// Create event channel
	events := make(chan engine.Event, 5)
	requestID := "req-test"

	// Send test events
	go func() {
		events <- engine.NewStartEvent("test_tool", requestID, nil)
		events <- engine.NewDataEvent(map[string]string{"key": "value"}, 1)
		events <- engine.NewProgressEvent(1, 2, "Half done")
		events <- engine.NewEndEvent(time.Millisecond*100, 3, "Done")
		close(events)
	}()

	// Convert to SSE
	sseMessages := streamEventsToSSE(events, requestID)

	// Collect all messages
	var messages []sseMessage
	for msg := range sseMessages {
		messages = append(messages, msg)
	}

	// Verify count
	expectedCount := 4 // start, data, progress, end
	if len(messages) != expectedCount {
		t.Errorf("streamEventsToSSE() got %d messages, want %d", len(messages), expectedCount)
	}

	// Verify order
	expectedEvents := []string{"start", "data", "progress", "end"}
	for i, msg := range messages {
		if msg.Event != expectedEvents[i] {
			t.Errorf("message[%d] event = %q, want %q", i, msg.Event, expectedEvents[i])
		}
		if msg.ID != requestID {
			t.Errorf("message[%d] id = %q, want %q", i, msg.ID, requestID)
		}
	}
}

func TestEventToSSE_EdgeCases(t *testing.T) {
	t.Run("nil data in event", func(t *testing.T) {
		event := engine.Event{
			Type:      engine.EventData,
			Timestamp: time.Now(),
			Data:      nil,
		}

		result := eventToSSE(event, "req-123")

		if result.Event != "data" {
			t.Errorf("event type = %q, want %q", result.Event, "data")
		}

		// Should handle nil data gracefully
		if result.Data == "" {
			t.Error("data should not be empty even with nil input")
		}
	})

	t.Run("empty request ID", func(t *testing.T) {
		event := engine.NewDataEvent("test", 1)
		result := eventToSSE(event, "")

		if result.ID != "" {
			t.Errorf("expected empty ID, got %q", result.ID)
		}
	})
}

func TestSSEMessage_Format_EdgeCases(t *testing.T) {
	t.Run("data with carriage returns", func(t *testing.T) {
		msg := sseMessage{
			Event: "test",
			Data:  "line1\r\nline2\r\nline3",
		}

		result := msg.format()

		// Should handle \r\n properly
		if !contains(result, "data: line1") {
			t.Error("should contain first line")
		}
		if !contains(result, "data: line2") {
			t.Error("should contain second line")
		}
		if !contains(result, "data: line3") {
			t.Error("should contain third line")
		}
	})

	t.Run("very long data", func(t *testing.T) {
		longData := make([]byte, 10000)
		for i := range longData {
			longData[i] = 'a'
		}

		msg := sseMessage{
			Event: "test",
			Data:  string(longData),
		}

		result := msg.format()

		// Should handle long data
		if len(result) < len(longData) {
			t.Error("formatted message should include all data")
		}
	})
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
