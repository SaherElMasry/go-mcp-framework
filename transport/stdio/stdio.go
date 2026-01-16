package stdio

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/SaherElMasry/go-mcp-framework/transport"
)

// StdioTransport implements MCP over standard input/output
type StdioTransport struct {
	handler transport.Handler
	logger  *slog.Logger
	reader  *bufio.Reader
	writer  *bufio.Writer
}

// NewStdioTransport creates a new stdio transport
func NewStdioTransport(handler transport.Handler, logger *slog.Logger) *StdioTransport {
	if logger == nil {
		logger = slog.Default()
	}

	return &StdioTransport{
		handler: handler,
		logger:  logger,
		reader:  bufio.NewReader(os.Stdin),
		writer:  bufio.NewWriter(os.Stdout),
	}
}

// Run starts the stdio transport loop
func (t *StdioTransport) Run(ctx context.Context) error {
	t.logger.Info("stdio transport started")

	for {
		select {
		case <-ctx.Done():
			t.logger.Info("stdio transport shutting down")
			return ctx.Err()
		default:
		}

		line, err := t.reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				t.logger.Info("client disconnected")
				return nil
			}
			t.logger.Error("read error", "error", err)
			return fmt.Errorf("read error: %w", err)
		}

		if len(line) == 0 || (len(line) == 1 && line[0] == '\n') {
			continue
		}

		t.logger.Debug("received message", "size", len(line))

		response, err := t.handler.Handle(ctx, line, "stdio")
		if err != nil {
			t.logger.Error("handler error", "error", err)
		}

		if len(response) > 0 {
			if _, err := t.writer.Write(response); err != nil {
				return fmt.Errorf("write error: %w", err)
			}

			if err := t.writer.WriteByte('\n'); err != nil {
				return fmt.Errorf("write error: %w", err)
			}

			if err := t.writer.Flush(); err != nil {
				return fmt.Errorf("flush error: %w", err)
			}

			t.logger.Debug("sent response", "size", len(response))
		}
	}
}
