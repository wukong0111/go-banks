package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

// NewJSONFileHandler creates a JSON handler that writes to a file
func NewJSONFileHandler(filename string, opts *slog.HandlerOptions) (slog.Handler, error) {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}

	// Ensure directory exists
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return nil, err
	}

	return slog.NewJSONHandler(file, opts), nil
}

// NewTextFileHandler creates a text handler that writes to a file
func NewTextFileHandler(filename string, opts *slog.HandlerOptions) (slog.Handler, error) {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}

	// Ensure directory exists
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return nil, err
	}

	return slog.NewTextHandler(file, opts), nil
}

// FileHandlerWithRotation creates a file handler with basic rotation support
type FileHandlerWithRotation struct {
	handler  slog.Handler
	filename string
	file     *os.File
	maxSize  int64 // Maximum file size in bytes
}

// NewFileHandlerWithRotation creates a file handler that rotates when size limit is reached
func NewFileHandlerWithRotation(filename string, maxSize int64, opts *slog.HandlerOptions) (*FileHandlerWithRotation, error) {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}

	// Ensure directory exists
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return nil, err
	}

	handler := slog.NewJSONHandler(file, opts)

	return &FileHandlerWithRotation{
		handler:  handler,
		filename: filename,
		file:     file,
		maxSize:  maxSize,
	}, nil
}

func (fh *FileHandlerWithRotation) Enabled(ctx context.Context, level slog.Level) bool {
	return fh.handler.Enabled(ctx, level)
}

func (fh *FileHandlerWithRotation) Handle(ctx context.Context, record slog.Record) error {
	// Check if rotation is needed
	if err := fh.rotateIfNeeded(); err != nil {
		// Log rotation error but continue with logging
		slog.Default().Error("failed to rotate log file", "error", err)
	}

	return fh.handler.Handle(ctx, record)
}

func (fh *FileHandlerWithRotation) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &FileHandlerWithRotation{
		handler:  fh.handler.WithAttrs(attrs),
		filename: fh.filename,
		file:     fh.file,
		maxSize:  fh.maxSize,
	}
}

func (fh *FileHandlerWithRotation) WithGroup(name string) slog.Handler {
	return &FileHandlerWithRotation{
		handler:  fh.handler.WithGroup(name),
		filename: fh.filename,
		file:     fh.file,
		maxSize:  fh.maxSize,
	}
}

// rotateIfNeeded checks file size and rotates if necessary
func (fh *FileHandlerWithRotation) rotateIfNeeded() error {
	stat, err := fh.file.Stat()
	if err != nil {
		return err
	}

	if stat.Size() < fh.maxSize {
		return nil // No rotation needed
	}

	// Close current file
	if err := fh.file.Close(); err != nil {
		return err
	}

	// Rename current file to backup
	backupName := fh.filename + ".old"
	if err := os.Rename(fh.filename, backupName); err != nil {
		return err
	}

	// Create new file
	newFile, err := os.OpenFile(fh.filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}

	// Update file reference
	fh.file = newFile

	// Create new handler with new file
	opts := &slog.HandlerOptions{
		Level:     slog.LevelDebug, // We'll filter at MultiLogger level
		AddSource: true,
	}
	fh.handler = slog.NewJSONHandler(newFile, opts)

	return nil
}

// Close closes the file handler
func (fh *FileHandlerWithRotation) Close() error {
	if fh.file != nil {
		return fh.file.Close()
	}
	return nil
}

// MultiWriterHandler writes to multiple writers
type MultiWriterHandler struct {
}

// NewMultiWriterHandler creates a handler that writes to multiple writers
func NewMultiWriterHandler(writers []io.Writer, opts *slog.HandlerOptions) slog.Handler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}

	multiWriter := io.MultiWriter(writers...)
	return slog.NewJSONHandler(multiWriter, opts)
}
