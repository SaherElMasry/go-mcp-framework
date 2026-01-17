package observability

import (
	"log/slog"
	"os"
)

// LoggingConfig is imported from framework package when needed
// It's defined in framework/config.go

// SetupLogging configures structured logging based on config
func SetupLogging(config interface{}) *slog.Logger {
	// Use type assertion to extract fields
	var level slog.Level
	var format string
	var addSource bool

	// Handle the config as a generic interface
	// This works with any struct that has Level, Format, AddSource fields
	if cfg, ok := config.(struct {
		Level     string
		Format    string
		AddSource bool
	}); ok {
		level = parseLevel(cfg.Level)
		format = cfg.Format
		addSource = cfg.AddSource
	} else {
		// Default values if type assertion fails
		level = slog.LevelInfo
		format = "json"
		addSource = false
	}

	// Create handler options
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: addSource,
	}

	var handler slog.Handler

	// Choose format
	switch format {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	case "text":
		handler = slog.NewTextHandler(os.Stdout, opts)
	default:
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}

// parseLevel converts string to slog.Level
func parseLevel(levelStr string) slog.Level {
	switch levelStr {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
