// color/terminal.go
package color

import (
	"io"
	"os"
	"runtime"
	"strings"

	"golang.org/x/sys/unix"
)

// IsTerminal returns true if the writer is a terminal
func IsTerminal(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		fd := int(f.Fd())
		_, err := unix.IoctlGetTermios(fd, unix.TCGETS)
		return err == nil
	}
	return false
}

// IsTTY returns true if stdout is a TTY
func IsTTY() bool {
	return IsTerminal(os.Stdout)
}

// SupportsColor returns true if the terminal supports color
func SupportsColor() bool {
	// Check NO_COLOR environment variable
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	// Check TERM environment variable
	term := os.Getenv("TERM")
	if term == "dumb" {
		return false
	}

	// Check if output is a terminal
	if !IsTTY() {
		return false
	}

	// Check color support
	if strings.Contains(term, "color") || strings.Contains(term, "xterm") ||
		strings.Contains(term, "screen") || term == "cygwin" ||
		term == "linux" || term == "ansi" {
		return true
	}

	// Windows check
	if runtime.GOOS == "windows" {
		// Windows 10+ supports ANSI colors
		return true
	}

	return true
}

// AutoDetect automatically enables or disables colors based on terminal support
func AutoDetect() {
	if SupportsColor() {
		Enable()
	} else {
		Disable()
	}
}

// GetTerminalSize returns terminal width and height
func GetTerminalSize() (width, height int, err error) {
	ws, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	if err != nil {
		return 80, 24, err
	}
	return int(ws.Col), int(ws.Row), nil
}
