// observability/logging.go
package observability

import (
	"io"
	"log/slog"
	"os"

	"github.com/SaherElMasry/go-mcp-framework/color"
)

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level     string
	Format    string
	AddSource bool
	Output    io.Writer
}

// SetupLogging configures structured logging based on config
func SetupLogging(config interface{}) *slog.Logger {
	var handler slog.Handler

	// Auto-detect terminal
	color.AutoDetect()

	// Try to type assert to LoggingConfig
	var cfg LoggingConfig
	if logCfg, ok := config.(LoggingConfig); ok {
		cfg = logCfg
	} else if logCfgPtr, ok := config.(*LoggingConfig); ok {
		cfg = *logCfgPtr
	} else {
		// Try to extract fields from any struct
		if cfgStruct, ok := config.(struct {
			Level     string
			Format    string
			AddSource bool
			Output    io.Writer
		}); ok {
			cfg = LoggingConfig{
				Level:     cfgStruct.Level,
				Format:    cfgStruct.Format,
				AddSource: cfgStruct.AddSource,
				Output:    cfgStruct.Output,
			}
		} else {
			// Default values
			cfg = LoggingConfig{
				Level:     "info",
				Format:    "json",
				AddSource: false,
				Output:    os.Stdout,
			}
		}
	}

	// Ensure output is set
	if cfg.Output == nil {
		cfg.Output = os.Stdout
	}

	// Use colored handler for text format
	if cfg.Format == "text" && color.IsEnabled() {
		handler = color.NewColoredHandler(cfg.Output, &color.ColoredHandlerOptions{
			Level:      parseLevel(cfg.Level),
			TimeFormat: "15:04:05",
			Writer:     cfg.Output,
		})
	} else {
		// Create handler options
		opts := &slog.HandlerOptions{
			Level:     parseLevel(cfg.Level),
			AddSource: cfg.AddSource,
		}

		// Choose format
		switch cfg.Format {
		case "json":
			handler = slog.NewJSONHandler(cfg.Output, opts)
		case "text":
			handler = slog.NewTextHandler(cfg.Output, opts)
		default:
			handler = slog.NewJSONHandler(cfg.Output, opts)
		}
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
