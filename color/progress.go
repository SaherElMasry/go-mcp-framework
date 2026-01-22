// color/progress.go
package color

import (
	"fmt"
	"strings"
	"time"
)

// ProgressBar represents a colored progress bar
type ProgressBar struct {
	Total   int64
	Current int64
	Width   int
	Prefix  string
}

// NewProgressBar creates a new progress bar
func NewProgressBar(total int64) *ProgressBar {
	return &ProgressBar{
		Total: total,
		Width: 40,
	}
}

// Update updates the progress bar
func (pb *ProgressBar) Update(current int64) string {
	pb.Current = current
	return pb.String()
}

// String returns the colored progress bar string
func (pb *ProgressBar) String() string {
	if NoColor {
		return pb.plainString()
	}

	percentage := float64(pb.Current) / float64(pb.Total) * 100
	filled := int(float64(pb.Width) * float64(pb.Current) / float64(pb.Total))
	empty := pb.Width - filled

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)

	var percentColor Color
	switch {
	case percentage >= 100:
		percentColor = ColorBrightGreen
	case percentage >= 75:
		percentColor = ColorGreen
	case percentage >= 50:
		percentColor = ColorYellow
	case percentage >= 25:
		percentColor = ColorBrightYellow
	default:
		percentColor = ColorRed
	}

	prefix := ""
	if pb.Prefix != "" {
		// Fix: Use %s format with pb.Prefix
		prefix = Cyan("%s", pb.Prefix) + " "
	}

	return fmt.Sprintf("%s[%s] %s %d/%d",
		prefix,
		Colorize(bar, percentColor),
		Colorize(fmt.Sprintf("%.1f%%", percentage), percentColor, Bold),
		pb.Current,
		pb.Total,
	)
}

func (pb *ProgressBar) plainString() string {
	percentage := float64(pb.Current) / float64(pb.Total) * 100
	filled := int(float64(pb.Width) * float64(pb.Current) / float64(pb.Total))
	empty := pb.Width - filled

	bar := strings.Repeat("=", filled) + strings.Repeat("-", empty)

	prefix := ""
	if pb.Prefix != "" {
		prefix = pb.Prefix + " "
	}

	return fmt.Sprintf("%s[%s] %.1f%% %d/%d",
		prefix, bar, percentage, pb.Current, pb.Total,
	)
}

// Spinner represents a loading spinner
type Spinner struct {
	frames []string
	index  int
	text   string
}

// NewSpinner creates a new spinner
func NewSpinner() *Spinner {
	return &Spinner{
		frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		index:  0,
	}
}

// Next returns the next spinner frame
func (s *Spinner) Next() string {
	frame := s.frames[s.index]
	s.index = (s.index + 1) % len(s.frames)

	if NoColor {
		return fmt.Sprintf("%s %s", frame, s.text)
	}

	return fmt.Sprintf("%s %s",
		Colorize(frame, ColorCyan, Bold),
		s.text,
	)
}

// SetText sets the spinner text
func (s *Spinner) SetText(text string) {
	s.text = text
}

// Table represents a colored table
type Table struct {
	Headers []string
	Rows    [][]string
}

// NewTable creates a new table
func NewTable(headers ...string) *Table {
	return &Table{
		Headers: headers,
	}
}

// AddRow adds a row to the table
func (t *Table) AddRow(cols ...string) {
	t.Rows = append(t.Rows, cols)
}

// String returns the colored table string
func (t *Table) String() string {
	if len(t.Headers) == 0 {
		return ""
	}

	// Calculate column widths
	widths := make([]int, len(t.Headers))
	for i, h := range t.Headers {
		widths[i] = len(h)
	}

	for _, row := range t.Rows {
		for i, col := range row {
			if i < len(widths) && len(col) > widths[i] {
				widths[i] = len(col)
			}
		}
	}

	var buf strings.Builder

	// Header
	buf.WriteString(t.renderRow(t.Headers, widths, true))
	buf.WriteString("\n")

	// Separator
	for i, w := range widths {
		if i > 0 {
			buf.WriteString("─┼─")
		}
		buf.WriteString(strings.Repeat("─", w))
	}
	buf.WriteString("\n")

	// Rows
	for _, row := range t.Rows {
		buf.WriteString(t.renderRow(row, widths, false))
		buf.WriteString("\n")
	}

	return buf.String()
}

func (t *Table) renderRow(cols []string, widths []int, isHeader bool) string {
	var parts []string
	for i, col := range cols {
		width := widths[i]
		padded := col + strings.Repeat(" ", width-len(col))

		if NoColor {
			parts = append(parts, padded)
		} else if isHeader {
			parts = append(parts, Colorize(padded, ColorCyan, Bold))
		} else {
			parts = append(parts, padded)
		}
	}
	return strings.Join(parts, " │ ")
}

// Box draws a box around text
func Box(text string, width int) string {
	lines := strings.Split(text, "\n")
	maxLen := 0
	for _, line := range lines {
		if len(line) > maxLen {
			maxLen = len(line)
		}
	}

	if width > 0 && width > maxLen {
		maxLen = width
	}

	var buf strings.Builder

	// Top border
	topBorder := "┌" + strings.Repeat("─", maxLen+2) + "┐\n"
	if NoColor {
		buf.WriteString(topBorder)
	} else {
		// Fix: Use %s format
		buf.WriteString(Cyan("%s", topBorder))
	}

	// Content
	for _, line := range lines {
		padding := strings.Repeat(" ", maxLen-len(line))
		if NoColor {
			buf.WriteString(fmt.Sprintf("│ %s%s │\n", line, padding))
		} else {
			// Fix: Don't use "│" as format string
			leftBorder := Cyan("%s", "│")
			rightBorder := Cyan("%s", "│")
			buf.WriteString(fmt.Sprintf("%s %s%s %s\n",
				leftBorder, line, padding, rightBorder))
		}
	}

	// Bottom border
	bottomBorder := "└" + strings.Repeat("─", maxLen+2) + "┘"
	if NoColor {
		buf.WriteString(bottomBorder)
	} else {
		// Fix: Use %s format
		buf.WriteString(Cyan("%s", bottomBorder))
	}

	return buf.String()
}

// Banner creates a large banner with text
func Banner(title, subtitle string) string {
	var buf strings.Builder

	width := 60

	if NoColor {
		buf.WriteString("╔" + strings.Repeat("═", width) + "╗\n")
		buf.WriteString("║" + centerText(title, width) + "║\n")
		if subtitle != "" {
			buf.WriteString("║" + centerText(subtitle, width) + "║\n")
		}
		buf.WriteString("╚" + strings.Repeat("═", width) + "╝\n")
	} else {
		// Fix: Don't use strings as format
		topBorder := "╔" + strings.Repeat("═", width) + "╗\n"
		buf.WriteString(Cyan("%s", topBorder))
		
		leftBorder := Cyan("%s", "║")
		rightBorder := Cyan("%s", "║\n")
		buf.WriteString(leftBorder + Colorize(centerText(title, width), ColorBrightWhite, Bold) + rightBorder)
		
		if subtitle != "" {
			buf.WriteString(leftBorder + Colorize(centerText(subtitle, width), ColorBrightBlack) + rightBorder)
		}
		
		bottomBorder := "╚" + strings.Repeat("═", width) + "╝\n"
		buf.WriteString(Cyan("%s", bottomBorder))
	}

	return buf.String()
}

func centerText(text string, width int) string {
	if len(text) >= width {
		return text[:width]
	}
	padding := (width - len(text)) / 2
	return strings.Repeat(" ", padding) + text + strings.Repeat(" ", width-len(text)-padding)
}

// Duration formats a duration with color
func Duration(d time.Duration) string {
	var c Color
	switch {
	case d < 100*time.Millisecond:
		c = ColorBrightGreen
	case d < 500*time.Millisecond:
		c = ColorGreen
	case d < 1*time.Second:
		c = ColorYellow
	case d < 5*time.Second:
		c = ColorBrightYellow
	default:
		c = ColorRed
	}

	return Colorize(d.String(), c)
}

// Size formats a byte size with color
func Size(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return Colorize(fmt.Sprintf("%d B", bytes), ColorBrightGreen)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB", "PB"}
	value := float64(bytes) / float64(div)

	var c Color
	switch exp {
	case 0: // KB
		c = ColorBrightGreen
	case 1: // MB
		c = ColorGreen
	case 2: // GB
		c = ColorYellow
	default: // TB+
		c = ColorRed
	}

	return Colorize(fmt.Sprintf("%.1f %s", value, units[exp]), c)
}
