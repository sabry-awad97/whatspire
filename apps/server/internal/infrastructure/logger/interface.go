package logger

import (
	"context"
	"io"
	"os"
	"time"

	"whatspire/internal/infrastructure/config"

	"github.com/rs/zerolog"
	waLog "go.mau.fi/whatsmeow/util/log"
)

// Logger wraps zerolog.Logger with our application-specific methods
type Logger struct {
	zlog zerolog.Logger
}

// New creates a new Logger instance
func New(cfg config.LogConfig) *Logger {
	// Configure output format
	var output io.Writer = os.Stdout

	if cfg.Format == "text" {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
			NoColor:    false,
		}
	}

	// Parse log level
	level := parseZerologLevel(cfg.Level)

	// Create logger with timestamp
	zlog := zerolog.New(output).
		Level(level).
		With().
		Timestamp().
		Logger()

	return &Logger{zlog: zlog}
}

// NewWithOutput creates a logger with custom output (for testing)
func NewWithOutput(cfg config.LogConfig, output io.Writer) *Logger {
	level := parseZerologLevel(cfg.Level)

	zlog := zerolog.New(output).
		Level(level).
		With().
		Timestamp().
		Logger()

	return &Logger{zlog: zlog}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string) {
	l.zlog.Debug().Msg(msg)
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, args ...any) {
	l.zlog.Debug().Msgf(format, args...)
}

// Info logs an info message
func (l *Logger) Info(msg string) {
	l.zlog.Info().Msg(msg)
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, args ...any) {
	l.zlog.Info().Msgf(format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string) {
	l.zlog.Warn().Msg(msg)
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(format string, args ...any) {
	l.zlog.Warn().Msgf(format, args...)
}

// Error logs an error message
func (l *Logger) Error(msg string) {
	l.zlog.Error().Msg(msg)
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, args ...any) {
	l.zlog.Error().Msgf(format, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string) {
	l.zlog.Fatal().Msg(msg)
}

// Fatalf logs a formatted fatal message and exits
func (l *Logger) Fatalf(format string, args ...any) {
	l.zlog.Fatal().Msgf(format, args...)
}

// WithContext returns a logger with context fields extracted
func (l *Logger) WithContext(ctx context.Context) *Logger {
	zlog := l.zlog

	if requestID, ok := ctx.Value(RequestIDKey).(string); ok && requestID != "" {
		zlog = zlog.With().Str("request_id", requestID).Logger()
	}
	if sessionID, ok := ctx.Value(SessionIDKey).(string); ok && sessionID != "" {
		zlog = zlog.With().Str("session_id", sessionID).Logger()
	}

	return &Logger{zlog: zlog}
}

// WithStr returns a logger with a string field
func (l *Logger) WithStr(key, value string) *Logger {
	return &Logger{zlog: l.zlog.With().Str(key, value).Logger()}
}

// WithInt returns a logger with an int field
func (l *Logger) WithInt(key string, value int) *Logger {
	return &Logger{zlog: l.zlog.With().Int(key, value).Logger()}
}

// WithError returns a logger with an error field
func (l *Logger) WithError(err error) *Logger {
	return &Logger{zlog: l.zlog.With().Err(err).Logger()}
}

// WithFields returns a logger with multiple fields
func (l *Logger) WithFields(fields map[string]any) *Logger {
	zlog := l.zlog.With()
	for k, v := range fields {
		zlog = zlog.Interface(k, v)
	}
	return &Logger{zlog: zlog.Logger()}
}

// WithRequestID returns a logger with request ID field
func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{zlog: l.zlog.With().Str("request_id", requestID).Logger()}
}

// WithSessionID returns a logger with session ID field
func (l *Logger) WithSessionID(sessionID string) *Logger {
	return &Logger{zlog: l.zlog.With().Str("session_id", sessionID).Logger()}
}

// Sub creates a sub-logger with a module name (for whatsmeow compatibility)
func (l *Logger) Sub(module string) waLog.Logger {
	return &Logger{zlog: l.zlog.With().Str("module", module).Logger()}
}

// GetZerolog returns the underlying zerolog.Logger for advanced usage
func (l *Logger) GetZerolog() zerolog.Logger {
	return l.zlog
}

// parseZerologLevel converts string level to zerolog.Level
func parseZerologLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}

// Context keys for extracting values
type contextKey string

const (
	RequestIDKey contextKey = "request_id"
	SessionIDKey contextKey = "session_id"
)
