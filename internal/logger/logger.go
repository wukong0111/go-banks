package logger

import (
	"context"
	"log/slog"
)

// Logger defines our logging interface
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	With(args ...any) Logger
	WithContext(ctx context.Context) Logger
}

// MultiLogger broadcasts logs to multiple handlers
type MultiLogger struct {
	logger *slog.Logger
}

// NewMultiLogger creates a logger that sends to multiple destinations
func NewMultiLogger(handlers ...slog.Handler) *MultiLogger {
	combined := &multiHandler{handlers: handlers}
	return &MultiLogger{
		logger: slog.New(combined),
	}
}

// NewDiscardLogger creates a logger that discards all output (for tests)
func NewDiscardLogger() *MultiLogger {
	handler := &discardHandler{}
	return &MultiLogger{
		logger: slog.New(handler),
	}
}

// Debug logs a debug message
func (ml *MultiLogger) Debug(msg string, args ...any) {
	ml.logger.Debug(msg, args...)
}

// Info logs an info message
func (ml *MultiLogger) Info(msg string, args ...any) {
	ml.logger.Info(msg, args...)
}

// Warn logs a warning message
func (ml *MultiLogger) Warn(msg string, args ...any) {
	ml.logger.Warn(msg, args...)
}

// Error logs an error message
func (ml *MultiLogger) Error(msg string, args ...any) {
	ml.logger.Error(msg, args...)
}

// With returns a logger with the given attributes
func (ml *MultiLogger) With(args ...any) Logger {
	return &MultiLogger{
		logger: ml.logger.With(args...),
	}
}

// WithContext returns a logger with context
func (ml *MultiLogger) WithContext(_ context.Context) Logger {
	// slog doesn't have WithContext, but we can store context info
	// For now, just return the same logger since slog handles context in logging methods
	return ml
}

// multiHandler distributes log records to multiple handlers
type multiHandler struct {
	handlers []slog.Handler
}

// Enabled returns true if any handler is enabled for the given level
func (mh *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range mh.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

// Handle sends the record to all enabled handlers
func (mh *multiHandler) Handle(ctx context.Context, record slog.Record) error {
	var lastErr error
	for _, h := range mh.handlers {
		if h.Enabled(ctx, record.Level) {
			if err := h.Handle(ctx, record); err != nil {
				lastErr = err
			}
		}
	}
	return lastErr
}

// WithAttrs returns a new multiHandler with the given attributes
func (mh *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(mh.handlers))
	for i, h := range mh.handlers {
		newHandlers[i] = h.WithAttrs(attrs)
	}
	return &multiHandler{handlers: newHandlers}
}

// WithGroup returns a new multiHandler with the given group
func (mh *multiHandler) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(mh.handlers))
	for i, h := range mh.handlers {
		newHandlers[i] = h.WithGroup(name)
	}
	return &multiHandler{handlers: newHandlers}
}

// discardHandler discards all log records (for tests)
type discardHandler struct{}

func (dh *discardHandler) Enabled(context.Context, slog.Level) bool {
	return false
}

func (dh *discardHandler) Handle(context.Context, slog.Record) error {
	return nil
}

func (dh *discardHandler) WithAttrs([]slog.Attr) slog.Handler {
	return dh
}

func (dh *discardHandler) WithGroup(string) slog.Handler {
	return dh
}
