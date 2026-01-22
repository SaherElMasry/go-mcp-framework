// color/logger.go
package color

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
)

// ColoredHandler is a slog.Handler that outputs colored logs
type ColoredHandler struct {
	opts  *ColoredHandlerOptions
	level slog.Level
	attrs []slog.Attr
	group string
}

// ColoredHandlerOptions configures the ColoredHandler
type ColoredHandlerOptions struct {
	Level      slog.Level
	AddSource  bool
	TimeFormat string
	Writer     io.Writer
	NoColor    bool
}

// NewColoredHandler creates a new colored log handler
func NewColoredHandler(w io.Writer, opts *ColoredHandlerOptions) *ColoredHandler {
	if opts == nil {
		opts = &ColoredHandlerOptions{
			Level:      slog.LevelInfo,
			TimeFormat: "15:04:05",
			Writer:     w,
		}
	}
	if opts.Writer == nil {
		opts.Writer = w
	}
	if opts.TimeFormat == "" {
		opts.TimeFormat = "15:04:05"
	}

	return &ColoredHandler{
		opts:  opts,
		level: opts.Level,
	}
}

// Enabled implements slog.Handler
func (h *ColoredHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

// Handle implements slog.Handler
func (h *ColoredHandler) Handle(_ context.Context, r slog.Record) error {
	buf := &strings.Builder{}

	// Time
	if !r.Time.IsZero() {
		timeStr := r.Time.Format(h.opts.TimeFormat)
		buf.WriteString(Gray("%s", timeStr))
		buf.WriteString(" ")
	}

	// Level with color
	levelStr := h.colorizeLevel(r.Level)
	buf.WriteString(levelStr)
	buf.WriteString(" ")

	// Message
	buf.WriteString(h.colorizeMessage(r.Level, r.Message))

	// Attributes
	r.Attrs(func(a slog.Attr) bool {
		buf.WriteString(" ")
		buf.WriteString(h.formatAttr(a))
		return true
	})

	// Group attributes
	for _, attr := range h.attrs {
		buf.WriteString(" ")
		buf.WriteString(h.formatAttr(attr))
	}

	buf.WriteString("\n")

	_, err := h.opts.Writer.Write([]byte(buf.String()))
	return err
}

// WithAttrs implements slog.Handler
func (h *ColoredHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)

	return &ColoredHandler{
		opts:  h.opts,
		level: h.level,
		attrs: newAttrs,
		group: h.group,
	}
}

// WithGroup implements slog.Handler
func (h *ColoredHandler) WithGroup(name string) slog.Handler {
	return &ColoredHandler{
		opts:  h.opts,
		level: h.level,
		attrs: h.attrs,
		group: name,
	}
}

// colorizeLevel returns colored level string
func (h *ColoredHandler) colorizeLevel(level slog.Level) string {
	if h.opts.NoColor {
		return fmt.Sprintf("[%s]", level.String())
	}

	switch level {
	case slog.LevelDebug:
		return Colorize("[DEBUG]", ColorBrightBlack, Bold)
	case slog.LevelInfo:
		return Colorize("[INFO]", ColorBrightCyan, Bold)
	case slog.LevelWarn:
		return Colorize("[WARN]", ColorBrightYellow, Bold)
	case slog.LevelError:
		return Colorize("[ERROR]", ColorBrightRed, Bold)
	default:
		return Colorize(fmt.Sprintf("[%s]", level.String()), ColorWhite, Bold)
	}
}

// colorizeMessage returns colored message based on level
func (h *ColoredHandler) colorizeMessage(level slog.Level, msg string) string {
	if h.opts.NoColor {
		return msg
	}

	switch level {
	case slog.LevelDebug:
		return Colorize(msg, ColorBrightBlack)
	case slog.LevelInfo:
		return Colorize(msg, ColorWhite)
	case slog.LevelWarn:
		return Colorize(msg, ColorYellow)
	case slog.LevelError:
		return Colorize(msg, ColorRed, Bold)
	default:
		return msg
	}
}

// formatAttr formats an attribute with colors
func (h *ColoredHandler) formatAttr(a slog.Attr) string {
	if h.opts.NoColor {
		return fmt.Sprintf("%s=%v", a.Key, a.Value)
	}

	// Fix: Don't use a.Key as format string
	key := Cyan("%s", a.Key)
	value := h.formatValue(a.Value)

	if h.group != "" {
		return fmt.Sprintf("%s.%s=%s", h.group, key, value)
	}
	return fmt.Sprintf("%s=%s", key, value)
}

// formatValue formats a slog.Value with appropriate color
func (h *ColoredHandler) formatValue(v slog.Value) string {
	if h.opts.NoColor {
		return v.String()
	}

	switch v.Kind() {
	case slog.KindString:
		return Green("%q", v.String())
	case slog.KindInt64:
		return Magenta("%d", v.Int64())
	case slog.KindUint64:
		return Magenta("%d", v.Uint64())
	case slog.KindFloat64:
		return Magenta("%.2f", v.Float64())
	case slog.KindBool:
		if v.Bool() {
			return BrightGreen("true")
		}
		return BrightRed("false")
	case slog.KindDuration:
		return Yellow("%s", v.Duration())
	case slog.KindTime:
		return Gray("%s", v.Time().Format("15:04:05"))
	default:
		return fmt.Sprintf("%v", v.Any())
	}
}
