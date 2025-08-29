package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorGray   = "\033[37m"
	colorCyan   = "\033[36m"
	colorGreen  = "\033[32m"
)

// NewConsoleHandler creates a colorized console handler
func NewConsoleHandler(opts *slog.HandlerOptions) slog.Handler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}

	// Use colored output if stdout is a terminal and not on Windows
	useColor := isTerminal(os.Stdout) && runtime.GOOS != "windows"

	return slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:       opts.Level,
		AddSource:   opts.AddSource,
		ReplaceAttr: colorizeAttr(useColor),
	})
}

// NewConsoleHandlerWithWriter creates a console handler with custom writer
func NewConsoleHandlerWithWriter(w io.Writer, opts *slog.HandlerOptions) slog.Handler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}

	useColor := isTerminal(w) && runtime.GOOS != "windows"

	return slog.NewTextHandler(w, &slog.HandlerOptions{
		Level:       opts.Level,
		AddSource:   opts.AddSource,
		ReplaceAttr: colorizeAttr(useColor),
	})
}

// colorizeAttr returns a function that colorizes log attributes
func colorizeAttr(useColor bool) func([]string, slog.Attr) slog.Attr {
	return func(_ []string, a slog.Attr) slog.Attr {
		if !useColor {
			return a
		}

		switch a.Key {
		case slog.LevelKey:
			level := a.Value.Any().(slog.Level)
			coloredLevel := colorizeLevel(level)
			return slog.Attr{Key: a.Key, Value: slog.StringValue(coloredLevel)}
		case slog.MessageKey:
			return slog.Attr{Key: a.Key, Value: slog.StringValue(colorCyan + a.Value.String() + colorReset)}
		case slog.TimeKey:
			return slog.Attr{Key: a.Key, Value: slog.StringValue(colorGray + a.Value.String() + colorReset)}
		}
		return a
	}
}

// colorizeLevel returns a colorized level string
func colorizeLevel(level slog.Level) string {
	switch level {
	case slog.LevelDebug:
		return colorGray + "DEBUG" + colorReset
	case slog.LevelInfo:
		return colorGreen + "INFO " + colorReset
	case slog.LevelWarn:
		return colorYellow + "WARN " + colorReset
	case slog.LevelError:
		return colorRed + "ERROR" + colorReset
	default:
		return level.String()
	}
}

// isTerminal checks if the writer is a terminal
func isTerminal(w io.Writer) bool {
	switch v := w.(type) {
	case *os.File:
		return isTerminalFile(v)
	default:
		return false
	}
}

// isTerminalFile checks if a file is a terminal
func isTerminalFile(f *os.File) bool {
	// Simple check for common terminal file descriptors
	fd := f.Fd()
	return fd == 0 || fd == 1 || fd == 2 // stdin, stdout, stderr
}

// devHandler creates a development-friendly handler with enhanced formatting
type devHandler struct {
	handler slog.Handler
}

// NewDevHandler creates a handler optimized for development with enhanced readability
func NewDevHandler(opts *slog.HandlerOptions) slog.Handler {
	if opts == nil {
		opts = &slog.HandlerOptions{Level: slog.LevelDebug}
	}

	baseHandler := NewConsoleHandler(opts)
	return &devHandler{handler: baseHandler}
}

func (h *devHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *devHandler) Handle(ctx context.Context, record slog.Record) error {
	// Add some spacing for better readability in development
	_, _ = fmt.Fprintln(os.Stdout)
	err := h.handler.Handle(ctx, record)
	_, _ = fmt.Fprintln(os.Stdout)
	return err
}

func (h *devHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &devHandler{handler: h.handler.WithAttrs(attrs)}
}

func (h *devHandler) WithGroup(name string) slog.Handler {
	return &devHandler{handler: h.handler.WithGroup(name)}
}
