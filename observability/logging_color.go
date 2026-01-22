// observability/logging_color.go
package observability

import (
	"io"
	"log/slog"
	"os"

	"github.com/SaherElMasry/go-mcp-framework/color"
)

// SetupColoredLogger configures a colored logger for the application
func SetupColoredLogger(level, format string, output io.Writer) *slog.Logger {
	var handler slog.Handler

	// Auto-detect color support
	color.AutoDetect()

	// Parse log level
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	// Create handler based on format
	if format == "text" || format == "colored" {
		handler = color.NewColoredHandler(output, &color.ColoredHandlerOptions{
			Level:      logLevel,
			TimeFormat: "15:04:05.000",
			Writer:     output,
			NoColor:    !color.IsEnabled(),
		})
	} else {
		// Fallback to JSON
		handler = slog.NewJSONHandler(output, &slog.HandlerOptions{
			Level: logLevel,
		})
	}

	return slog.New(handler)
}

// SetupDefaultColoredLogger creates a colored logger with sensible defaults
func SetupDefaultColoredLogger() *slog.Logger {
	return SetupColoredLogger("info", "colored", os.Stdout)
}
