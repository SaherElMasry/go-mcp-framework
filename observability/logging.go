package observability

import (
	"context"
	"log/slog"
	"os"
)

// LoggingConfig configures logging
type LoggingConfig struct {
	Level     string // ✅ Changed from LogLevel to string
	Format    string
	AddSource bool
}

// NewLogger creates a new structured logger
func NewLogger(config LoggingConfig) *slog.Logger {
	var level slog.Level
	switch config.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: config.AddSource,
	}

	var handler slog.Handler
	if config.Format == "json" {
		handler = slog.NewJSONHandler(os.Stderr, opts)
	} else {
		handler = slog.NewTextHandler(os.Stderr, opts)
	}

	return slog.New(handler)
}

// LoggerFromContext extracts logger from context or returns default
func LoggerFromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}

// ContextWithLogger adds logger to context
func ContextWithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

type loggerKey struct{}

// package observability

// import (
// 	"context"
// 	"log/slog"
// 	"os"
// )

// // LogLevel represents log severity levels
// type LogLevel string

// const (
// 	LogLevelDebug LogLevel = "debug"
// 	LogLevelInfo  LogLevel = "info"
// 	LogLevelWarn  LogLevel = "warn"
// 	LogLevelError LogLevel = "error"
// )

// // LoggingConfig configures logging
// type LoggingConfig struct {
// 	Level     string
// 	Format    string
// 	AddSource bool
// }

// // NewLogger creates a new structured logger
// func NewLogger(config LoggingConfig) *slog.Logger {
// 	var level slog.Level
// 	switch LogLevel(config.Level) {
// 	case LogLevelDebug:
// 		level = slog.LevelDebug
// 	case LogLevelInfo:
// 		level = slog.LevelInfo
// 	case LogLevelWarn:
// 		level = slog.LevelWarn
// 	case LogLevelError:
// 		level = slog.LevelError
// 	default:
// 		level = slog.LevelInfo
// 	}

// 	opts := &slog.HandlerOptions{
// 		Level:     level,
// 		AddSource: config.AddSource,
// 	}

// 	var handler slog.Handler
// 	if config.Format == "json" {
// 		handler = slog.NewJSONHandler(os.Stderr, opts)
// 	} else {
// 		handler = slog.NewTextHandler(os.Stderr, opts)
// 	}

// 	return slog.New(handler)
// }

// // LoggerFromContext extracts logger from context or returns default
// func LoggerFromContext(ctx context.Context) *slog.Logger {
// 	if logger, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok {
// 		return logger
// 	}
// 	return slog.Default()
// }

// // ContextWithLogger adds logger to context
// func ContextWithLogger(ctx context.Context, logger *slog.Logger) context.Context {
// 	return context.WithValue(ctx, loggerKey{}, logger)
// }

// type loggerKey struct{}
