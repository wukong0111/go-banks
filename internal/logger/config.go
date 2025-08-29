package logger

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// Config holds logger configuration
type Config struct {
	Level       string   `json:"level" env:"LOG_LEVEL"`
	Outputs     []string `json:"outputs" env:"LOG_OUTPUTS"`
	JSONFile    string   `json:"json_file,omitempty" env:"LOG_JSON_FILE"`
	TextFile    string   `json:"text_file,omitempty" env:"LOG_TEXT_FILE"`
	Format      string   `json:"format" env:"LOG_FORMAT"`
	AddSource   bool     `json:"add_source" env:"LOG_ADD_SOURCE"`
	MaxFileSize int64    `json:"max_file_size" env:"LOG_MAX_FILE_SIZE"`
}

// DefaultConfig returns default logger configuration
func DefaultConfig() *Config {
	return &Config{
		Level:       "info",
		Outputs:     []string{"console"},
		Format:      "text",
		AddSource:   false,
		MaxFileSize: 100 * 1024 * 1024, // 100MB default
	}
}

// SetupLogger creates and configures a multi-logger based on config
func SetupLogger(cfg *Config) (*MultiLogger, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	var handlers []slog.Handler

	level := parseLevel(cfg.Level)
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: cfg.AddSource,
	}

	// Process each output type
	for _, output := range cfg.Outputs {
		handler, err := createHandler(strings.ToLower(output), cfg, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to create %s handler: %w", output, err)
		}
		if handler != nil {
			handlers = append(handlers, handler)
		}
	}

	// If no handlers were created, add console as fallback
	if len(handlers) == 0 {
		handlers = append(handlers, NewConsoleHandler(opts))
	}

	return NewMultiLogger(handlers...), nil
}

// SetupDevLogger creates a development-friendly logger
func SetupDevLogger() *MultiLogger {
	opts := &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}

	devHandler := NewDevHandler(opts)
	return NewMultiLogger(devHandler)
}

// SetupProdLogger creates a production-optimized logger
func SetupProdLogger(cfg *Config) (*MultiLogger, error) {
	if cfg == nil {
		cfg = &Config{
			Level:     "info",
			Outputs:   []string{"json"},
			JSONFile:  "/var/log/go-banks/app.log",
			AddSource: false,
		}
	}

	return SetupLogger(cfg)
}

// createHandler creates a specific handler based on type
func createHandler(handlerType string, cfg *Config, opts *slog.HandlerOptions) (slog.Handler, error) {
	switch handlerType {
	case "console", "stdout":
		return NewConsoleHandler(opts), nil

	case "dev", "development":
		return NewDevHandler(opts), nil

	case "json":
		if cfg.JSONFile == "" {
			return nil, errors.New("json_file must be specified for json output")
		}
		return NewJSONFileHandler(cfg.JSONFile, opts)

	case "file", "text":
		if cfg.TextFile == "" {
			return nil, errors.New("text_file must be specified for text file output")
		}
		return NewTextFileHandler(cfg.TextFile, opts)

	case "json-rotating", "json_rotating":
		if cfg.JSONFile == "" {
			return nil, errors.New("json_file must be specified for rotating json output")
		}
		return NewFileHandlerWithRotation(cfg.JSONFile, cfg.MaxFileSize, opts)

	case "stderr":
		return slog.NewTextHandler(os.Stderr, opts), nil

	case "discard", "none":
		return slog.NewTextHandler(discardWriter{}, opts), nil

	default:
		return nil, fmt.Errorf("unsupported handler type: %s", handlerType)
	}
}

// parseLevel converts string level to slog.Level
func parseLevel(levelStr string) slog.Level {
	switch strings.ToLower(levelStr) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo // Default to info
	}
}

// discardWriter is a writer that discards all data
type discardWriter struct{}

func (discardWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

// LoggerFromConfig creates logger from environment variables
func LoggerFromConfig() (*MultiLogger, error) {
	cfg := &Config{
		Level:       getEnvOrDefault("LOG_LEVEL", "info"),
		Outputs:     parseOutputs(getEnvOrDefault("LOG_OUTPUTS", "console")),
		JSONFile:    os.Getenv("LOG_JSON_FILE"),
		TextFile:    os.Getenv("LOG_TEXT_FILE"),
		Format:      getEnvOrDefault("LOG_FORMAT", "text"),
		AddSource:   getEnvOrDefault("LOG_ADD_SOURCE", "false") == "true",
		MaxFileSize: parseMaxFileSize(getEnvOrDefault("LOG_MAX_FILE_SIZE", "104857600")), // 100MB
	}

	return SetupLogger(cfg)
}

// Helper functions
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseOutputs(outputsStr string) []string {
	if outputsStr == "" {
		return []string{"console"}
	}

	outputs := strings.Split(outputsStr, ",")
	for i, output := range outputs {
		outputs[i] = strings.TrimSpace(output)
	}
	return outputs
}

func parseMaxFileSize(sizeStr string) int64 {
	// Simple parsing - in production you might want more sophisticated parsing
	switch sizeStr {
	case "":
		return 100 * 1024 * 1024 // 100MB
	default:
		// For now, just return default. In production, parse units like "10MB", "1GB"
		return 100 * 1024 * 1024
	}
}
