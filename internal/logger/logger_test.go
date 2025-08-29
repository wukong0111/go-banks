package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMultiLogger(t *testing.T) {
	// Create buffers to capture output
	var buf1, buf2 bytes.Buffer

	// Create handlers that write to buffers
	handler1 := slog.NewTextHandler(&buf1, nil)
	handler2 := slog.NewJSONHandler(&buf2, nil)

	// Create multi-logger
	logger := NewMultiLogger(handler1, handler2)

	// Test logging
	logger.Info("test message", "key", "value")

	// Check that both handlers received the message
	assert.Contains(t, buf1.String(), "test message")
	assert.Contains(t, buf1.String(), "key=value")

	// Parse JSON from second buffer
	lines := strings.Split(strings.TrimSpace(buf2.String()), "\n")
	var logEntry map[string]any
	err := json.Unmarshal([]byte(lines[0]), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, "test message", logEntry["msg"])
	assert.Equal(t, "value", logEntry["key"])
}

func TestLoggerWithContext(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, nil)
	logger := NewMultiLogger(handler)

	type traceIDKey string
	ctx := context.WithValue(context.Background(), traceIDKey("trace_id"), "123456")
	ctxLogger := logger.WithContext(ctx)

	ctxLogger.Info("context test")

	// Basic test to ensure context logger works
	assert.Contains(t, buf.String(), "context test")
}

func TestLoggerWith(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, nil)
	logger := NewMultiLogger(handler)

	enrichedLogger := logger.With("user_id", 42, "service", "test")
	enrichedLogger.Info("enriched message")

	// Parse JSON output
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	var logEntry map[string]any
	err := json.Unmarshal([]byte(lines[0]), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, "enriched message", logEntry["msg"])
	assert.Equal(t, float64(42), logEntry["user_id"]) // JSON numbers are float64
	assert.Equal(t, "test", logEntry["service"])
}

func TestSetupLogger(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		want   bool // true if should succeed
	}{
		{
			name: "console output",
			config: &Config{
				Level:   "info",
				Outputs: []string{"console"},
			},
			want: true,
		},
		{
			name: "multiple outputs",
			config: &Config{
				Level:   "debug",
				Outputs: []string{"console", "stderr"},
			},
			want: true,
		},
		{
			name: "file output with missing file",
			config: &Config{
				Level:   "info",
				Outputs: []string{"json"},
				// Missing JSONFile
			},
			want: false,
		},
		{
			name:   "nil config uses defaults",
			config: nil,
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := SetupLogger(tt.config)

			if tt.want {
				assert.NoError(t, err)
				assert.NotNil(t, logger)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestConsoleHandler(t *testing.T) {
	var buf bytes.Buffer
	handler := NewConsoleHandlerWithWriter(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	logger := slog.New(handler)
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	output := buf.String()
	assert.Contains(t, output, "debug message")
	assert.Contains(t, output, "info message")
	assert.Contains(t, output, "warn message")
	assert.Contains(t, output, "error message")
}

func TestJSONFileHandler(t *testing.T) {
	// Create temporary file
	tmpFile, err := os.CreateTemp("", "test-log-*.json")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	_ = tmpFile.Close()

	// Create handler
	handler, err := NewJSONFileHandler(tmpFile.Name(), &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	require.NoError(t, err)

	// Log some messages
	logger := slog.New(handler)
	logger.Info("test message", "key", "value")
	logger.Error("error message", "error", "something went wrong")

	// Read file contents
	content, err := os.ReadFile(tmpFile.Name())
	require.NoError(t, err)

	// Verify JSON format
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	assert.Len(t, lines, 2)

	var logEntry1, logEntry2 map[string]any
	err = json.Unmarshal([]byte(lines[0]), &logEntry1)
	require.NoError(t, err)
	err = json.Unmarshal([]byte(lines[1]), &logEntry2)
	require.NoError(t, err)

	assert.Equal(t, "test message", logEntry1["msg"])
	assert.Equal(t, "value", logEntry1["key"])
	assert.Equal(t, "error message", logEntry2["msg"])
	assert.Equal(t, "something went wrong", logEntry2["error"])
}

func TestDevLogger(t *testing.T) {
	logger := SetupDevLogger()
	assert.NotNil(t, logger)

	// Test that it doesn't panic
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input string
		want  slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"DEBUG", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"INFO", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"warning", slog.LevelWarn},
		{"error", slog.LevelError},
		{"ERROR", slog.LevelError},
		{"invalid", slog.LevelInfo}, // defaults to info
		{"", slog.LevelInfo},        // defaults to info
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseLevel(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCreateHandler(t *testing.T) {
	tests := []struct {
		name        string
		handlerType string
		config      *Config
		wantErr     bool
	}{
		{"console", "console", &Config{}, false},
		{"dev", "dev", &Config{}, false},
		{"stderr", "stderr", &Config{}, false},
		{"discard", "discard", &Config{}, false},
		{"json with file", "json", &Config{JSONFile: "/tmp/test.log"}, false},
		{"json without file", "json", &Config{}, true},
		{"unsupported", "unsupported", &Config{}, true},
	}

	opts := &slog.HandlerOptions{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, err := createHandler(tt.handlerType, tt.config, opts)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, handler)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, handler)
			}
		})
	}
}

// TestLoggerFromConfig tests environment variable configuration
func TestLoggerFromConfig(t *testing.T) {
	// Set environment variables
	_ = os.Setenv("LOG_LEVEL", "debug")
	_ = os.Setenv("LOG_OUTPUTS", "console,stderr")
	defer func() {
		_ = os.Unsetenv("LOG_LEVEL")
		_ = os.Unsetenv("LOG_OUTPUTS")
	}()

	logger, err := LoggerFromConfig()
	assert.NoError(t, err)
	assert.NotNil(t, logger)

	// Test logging
	logger.Debug("test debug message")
	logger.Info("test info message")
}

// BenchmarkMultiLogger tests performance of multi-logger
func BenchmarkMultiLogger(b *testing.B) {
	handler1 := slog.NewTextHandler(io.Discard, nil)
	handler2 := slog.NewJSONHandler(io.Discard, nil)
	logger := NewMultiLogger(handler1, handler2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message", "iteration", i)
	}
}

// BenchmarkSingleLogger compares with single handler
func BenchmarkSingleLogger(b *testing.B) {
	handler := slog.NewJSONHandler(io.Discard, nil)
	logger := NewMultiLogger(handler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message", "iteration", i)
	}
}
