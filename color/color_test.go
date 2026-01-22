// color/color_test.go
package color

import (
	"bytes"
	"testing"
)

func TestColorize(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		color    Color
		attrs    []Attribute
		noColor  bool
		expected string
	}{
		{
			name:     "red text",
			text:     "hello",
			color:    ColorRed,
			noColor:  false,
			expected: "\x1b[31mhello\x1b[0m",
		},
		{
			name:     "bold green",
			text:     "success",
			color:    ColorGreen,
			attrs:    []Attribute{Bold},
			noColor:  false,
			expected: "\x1b[1;32msuccess\x1b[0m",
		},
		{
			name:     "no color",
			text:     "plain",
			color:    ColorBlue,
			noColor:  true,
			expected: "plain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			NoColor = tt.noColor
			result := Colorize(tt.text, tt.color, tt.attrs...)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestHelperFunctions(t *testing.T) {
	NoColor = false

	tests := []struct {
		name     string
		fn       func(string, ...interface{}) string
		input    string
		contains string
	}{
		{"Red", Red, "error", "\x1b[31m"},
		{"Green", Green, "success", "\x1b[32m"},
		{"Yellow", Yellow, "warning", "\x1b[33m"},
		{"Blue", Blue, "info", "\x1b[34m"},
		{"Cyan", Cyan, "debug", "\x1b[36m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn(tt.input)
			if !bytes.Contains([]byte(result), []byte(tt.contains)) {
				t.Errorf("result %q doesn't contain %q", result, tt.contains)
			}
		})
	}
}

func TestSemanticColors(t *testing.T) {
	NoColor = false

	tests := []struct {
		name string
		fn   func(string, ...interface{}) string
		text string
	}{
		{"Success", Success, "done"},
		{"Error", Error, "failed"},
		{"Warning", Warning, "caution"},
		{"Info", Info, "notice"},
		{"Debug", Debug, "trace"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn(tt.text)
			if result == tt.text {
				t.Error("semantic color function didn't colorize text")
			}
		})
	}
}

func TestProgressBar(t *testing.T) {
	pb := NewProgressBar(100)
	pb.Prefix = "Loading"

	tests := []struct {
		current int64
		check   func(string) bool
	}{
		{0, func(s string) bool { return len(s) > 0 }},
		{50, func(s string) bool { return len(s) > 0 }},
		{100, func(s string) bool { return len(s) > 0 }},
	}

	for _, tt := range tests {
		result := pb.Update(tt.current)
		if !tt.check(result) {
			t.Errorf("progress bar check failed for current=%d", tt.current)
		}
	}
}

func TestSpinner(t *testing.T) {
	s := NewSpinner()
	s.SetText("Loading...")

	// Get a few frames
	frame1 := s.Next()
	frame2 := s.Next()

	if frame1 == frame2 {
		t.Error("spinner frames should be different")
	}

	if len(frame1) == 0 {
		t.Error("spinner frame should not be empty")
	}
}

func TestTable(t *testing.T) {
	table := NewTable("Name", "Age", "City")
	table.AddRow("Alice", "30", "NYC")
	table.AddRow("Bob", "25", "SF")

	result := table.String()
	if len(result) == 0 {
		t.Error("table should not be empty")
	}

	// Check that headers are present
	if !bytes.Contains([]byte(result), []byte("Name")) {
		t.Error("table should contain header 'Name'")
	}
}

func TestBox(t *testing.T) {
	text := "Hello\nWorld"
	result := Box(text, 0)

	if len(result) == 0 {
		t.Error("box should not be empty")
	}

	// Check for box characters
	if !bytes.Contains([]byte(result), []byte("┌")) {
		t.Error("box should contain top-left corner")
	}
}

func TestBanner(t *testing.T) {
	result := Banner("Test Server", "v1.0.0")

	if len(result) == 0 {
		t.Error("banner should not be empty")
	}

	if !bytes.Contains([]byte(result), []byte("Test Server")) {
		t.Error("banner should contain title")
	}
}

func TestEnableDisable(t *testing.T) {
	// Test enable
	Enable()
	if !IsEnabled() {
		t.Error("colors should be enabled")
	}

	// Test disable
	Disable()
	if IsEnabled() {
		t.Error("colors should be disabled")
	}
}
