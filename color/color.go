// color/color.go
package color

import (
	"fmt"
	"io"
	"os"
	"sync"
)

// Color represents an ANSI color code
type Color int

const (
	// Text colors
	ColorBlack Color = iota + 30
	ColorRed
	ColorGreen
	ColorYellow
	ColorBlue
	ColorMagenta
	ColorCyan
	ColorWhite
	ColorDefault Color = 39

	// Bright colors
	ColorBrightBlack   Color = 90
	ColorBrightRed     Color = 91
	ColorBrightGreen   Color = 92
	ColorBrightYellow  Color = 93
	ColorBrightBlue    Color = 94
	ColorBrightMagenta Color = 95
	ColorBrightCyan    Color = 96
	ColorBrightWhite   Color = 97
)

// Attribute represents text attributes
type Attribute int

const (
	Reset Attribute = iota
	Bold
	Dim
	Italic
	Underline
	Blink
	Reverse
	Hidden
	Strikethrough
)

var (
	// NoColor disables all colors globally
	NoColor = false

	// Output is the default output writer (os.Stdout)
	Output io.Writer = os.Stdout

	// ForceColor forces color output even when not in terminal
	ForceColor = false

	mu sync.RWMutex
)

// Colorize wraps text with color codes
func Colorize(text string, c Color, attrs ...Attribute) string {
	if NoColor && !ForceColor {
		return text
	}

	var codes []int
	for _, attr := range attrs {
		codes = append(codes, int(attr))
	}
	codes = append(codes, int(c))

	var format string
	if len(codes) > 0 {
		format = "\x1b["
		for i, code := range codes {
			if i > 0 {
				format += ";"
			}
			format += fmt.Sprintf("%d", code)
		}
		format += "m%s\x1b[0m"
		return fmt.Sprintf(format, text)
	}

	return text
}

// Print formats and prints colored text
func Print(c Color, format string, args ...interface{}) {
	mu.RLock()
	defer mu.RUnlock()
	fmt.Fprint(Output, Colorize(fmt.Sprintf(format, args...), c))
}

// Println formats and prints colored text with newline
func Println(c Color, format string, args ...interface{}) {
	mu.RLock()
	defer mu.RUnlock()
	fmt.Fprintln(Output, Colorize(fmt.Sprintf(format, args...), c))
}

// Sprint returns colored string
func Sprint(c Color, format string, args ...interface{}) string {
	return Colorize(fmt.Sprintf(format, args...), c)
}

// Sprintf returns colored formatted string
func Sprintf(c Color, format string, args ...interface{}) string {
	return Colorize(fmt.Sprintf(format, args...), c)
}

// Helper functions for common colors

// Red returns text in red
func Red(format string, args ...interface{}) string {
	return Colorize(fmt.Sprintf(format, args...), ColorRed)
}

// Green returns text in green
func Green(format string, args ...interface{}) string {
	return Colorize(fmt.Sprintf(format, args...), ColorGreen)
}

// Yellow returns text in yellow
func Yellow(format string, args ...interface{}) string {
	return Colorize(fmt.Sprintf(format, args...), ColorYellow)
}

// Blue returns text in blue
func Blue(format string, args ...interface{}) string {
	return Colorize(fmt.Sprintf(format, args...), ColorBlue)
}

// Magenta returns text in magenta
func Magenta(format string, args ...interface{}) string {
	return Colorize(fmt.Sprintf(format, args...), ColorMagenta)
}

// Cyan returns text in cyan
func Cyan(format string, args ...interface{}) string {
	return Colorize(fmt.Sprintf(format, args...), ColorCyan)
}

// White returns text in white
func White(format string, args ...interface{}) string {
	return Colorize(fmt.Sprintf(format, args...), ColorWhite)
}

// Gray returns text in gray (bright black)
func Gray(format string, args ...interface{}) string {
	return Colorize(fmt.Sprintf(format, args...), ColorBrightBlack)
}

// MakeBold makes text bold
func MakeBold(text string) string {
	return Colorize(text, ColorDefault, Bold)
}

// MakeDim makes text dim
func MakeDim(text string) string {
	return Colorize(text, ColorDefault, Dim)
}

// MakeItalic makes text italic
func MakeItalic(text string) string {
	return Colorize(text, ColorDefault, Italic)
}

// MakeUnderline underlines text
func MakeUnderline(text string) string {
	return Colorize(text, ColorDefault, Underline)
}

// MakeStrikethrough strikes through text
func MakeStrikethrough(text string) string {
	return Colorize(text, ColorDefault, Strikethrough)
}

// Bright color helpers
func BrightRed(format string, args ...interface{}) string {
	return Colorize(fmt.Sprintf(format, args...), ColorBrightRed)
}

func BrightGreen(format string, args ...interface{}) string {
	return Colorize(fmt.Sprintf(format, args...), ColorBrightGreen)
}

func BrightYellow(format string, args ...interface{}) string {
	return Colorize(fmt.Sprintf(format, args...), ColorBrightYellow)
}

func BrightBlue(format string, args ...interface{}) string {
	return Colorize(fmt.Sprintf(format, args...), ColorBrightBlue)
}

func BrightMagenta(format string, args ...interface{}) string {
	return Colorize(fmt.Sprintf(format, args...), ColorBrightMagenta)
}

func BrightCyan(format string, args ...interface{}) string {
	return Colorize(fmt.Sprintf(format, args...), ColorBrightCyan)
}

// Semantic color helpers for common use cases

// Success prints success message in green
func Success(format string, args ...interface{}) string {
	return Colorize("✓ "+fmt.Sprintf(format, args...), ColorGreen, Bold)
}

// Error prints error message in red
func Error(format string, args ...interface{}) string {
	return Colorize("✗ "+fmt.Sprintf(format, args...), ColorRed, Bold)
}

// Warning prints warning message in yellow
func Warning(format string, args ...interface{}) string {
	return Colorize("⚠ "+fmt.Sprintf(format, args...), ColorYellow, Bold)
}

// Info prints info message in cyan
func Info(format string, args ...interface{}) string {
	return Colorize("ℹ "+fmt.Sprintf(format, args...), ColorCyan)
}

// Debug prints debug message in gray
func Debug(format string, args ...interface{}) string {
	return Colorize("🐛 "+fmt.Sprintf(format, args...), ColorBrightBlack)
}

// Highlight prints highlighted text
func Highlight(format string, args ...interface{}) string {
	return Colorize(fmt.Sprintf(format, args...), ColorBrightYellow, Bold)
}

// Subtle prints subtle text
func Subtle(format string, args ...interface{}) string {
	return Colorize(fmt.Sprintf(format, args...), ColorBrightBlack)
}

// Enable enables color output
func Enable() {
	mu.Lock()
	defer mu.Unlock()
	NoColor = false
}

// Disable disables color output
func Disable() {
	mu.Lock()
	defer mu.Unlock()
	NoColor = true
}

// IsEnabled returns whether colors are enabled
func IsEnabled() bool {
	mu.RLock()
	defer mu.RUnlock()
	return !NoColor || ForceColor
}
