package logging

import (
	"fmt"
	"log/slog"
	"os"
)

// Logger is our custom logger interface
type Logger interface {
	Debug(msg string, args ...any)
	Debugf(msg string, args ...any)
	Info(msg string, args ...any)
	Infof(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// CustomLogger implements the Logger interface
type CustomLogger struct {
	slogger *slog.Logger
}

// NewLogger creates a new CustomLogger
func NewLogger(useJSON bool, level slog.Level) Logger {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level: level,
	}

	if useJSON {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return &CustomLogger{
		slogger: slog.New(handler),
	}
}

// Debug logs a debug message
func (l *CustomLogger) Debug(msg string, args ...any) {
	l.slogger.Debug(msg, args...)
}

// Debugf logs a debug message with formatting.
func (l *CustomLogger) Debugf(format string, args ...interface{}) {
	formattedMessage := fmt.Sprintf(format, args...)
	l.slogger.Debug(formattedMessage)
}

// Info logs an info message
func (l *CustomLogger) Info(msg string, args ...any) {
	l.slogger.Info(msg, args...)
}

// Infof logs an info message with formatting.
func (l *CustomLogger) Infof(format string, args ...interface{}) {
	formattedMessage := fmt.Sprintf(format, args...)
	l.slogger.Info(formattedMessage)
}

// Warn logs a warning message
func (l *CustomLogger) Warn(msg string, args ...any) {
	l.slogger.Warn(msg, args...)
}

// Error logs an error message
func (l *CustomLogger) Error(msg string, args ...any) {
	l.slogger.Error(msg, args...)
}

// Global logger instance
var globalLogger Logger

// InitLogger initializes the global logger
func InitLogger(useJSON bool, level slog.Level) {
	globalLogger = NewLogger(useJSON, level)
}

// GetLogger returns the global logger instance
func GetLogger() Logger {
	if globalLogger == nil {
		// Default to a console logger with Info level if not initialized
		globalLogger = NewLogger(false, slog.LevelInfo)
	}
	return globalLogger
}
