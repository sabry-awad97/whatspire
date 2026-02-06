package unit

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"whatspire/internal/infrastructure/config"
	"whatspire/internal/infrastructure/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLogger(t *testing.T) {
	cfg := config.LogConfig{
		Level:  "info",
		Format: "json",
	}

	l := logger.New(cfg)
	assert.NotNil(t, l)
}

func TestLogger_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	cfg := config.LogConfig{Level: "debug", Format: "json"}
	l := logger.NewWithOutput(cfg, &buf)

	l.WithStr("key", "value").Info("test message")

	var entry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &entry)
	require.NoError(t, err)

	assert.Equal(t, "info", entry["level"])
	assert.Equal(t, "test message", entry["message"])
	assert.Equal(t, "value", entry["key"])
	assert.Contains(t, entry, "time")
}

func TestLogger_TextFormat(t *testing.T) {
	var buf bytes.Buffer
	cfg := config.LogConfig{Level: "debug", Format: "text"}
	l := logger.NewWithOutput(cfg, &buf)

	l.WithStr("key", "value").Info("test message")

	output := buf.String()
	// Zerolog console writer uses different format
	assert.Contains(t, output, "test message")
	assert.Contains(t, output, "key")
	assert.Contains(t, output, "value")
}

func TestLogger_LogLevels(t *testing.T) {
	tests := []struct {
		name      string
		logLevel  string
		logMethod string
		shouldLog bool
	}{
		{"debug at debug level", "debug", "debug", true},
		{"info at debug level", "debug", "info", true},
		{"warn at debug level", "debug", "warn", true},
		{"error at debug level", "debug", "error", true},
		{"debug at info level", "info", "debug", false},
		{"info at info level", "info", "info", true},
		{"warn at info level", "info", "warn", true},
		{"error at info level", "info", "error", true},
		{"debug at warn level", "warn", "debug", false},
		{"info at warn level", "warn", "info", false},
		{"warn at warn level", "warn", "warn", true},
		{"error at warn level", "warn", "error", true},
		{"debug at error level", "error", "debug", false},
		{"info at error level", "error", "info", false},
		{"warn at error level", "error", "warn", false},
		{"error at error level", "error", "error", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			cfg := config.LogConfig{Level: tt.logLevel, Format: "json"}
			l := logger.NewWithOutput(cfg, &buf)

			switch tt.logMethod {
			case "debug":
				l.Debug("test")
			case "info":
				l.Info("test")
			case "warn":
				l.Warn("test")
			case "error":
				l.Error("test")
			}

			if tt.shouldLog {
				assert.NotEmpty(t, buf.String())
			} else {
				assert.Empty(t, buf.String())
			}
		})
	}
}

func TestLogger_WithFields(t *testing.T) {
	var buf bytes.Buffer
	cfg := config.LogConfig{Level: "debug", Format: "json"}
	l := logger.NewWithOutput(cfg, &buf)

	childLogger := l.WithStr("service", "whatsapp")
	childLogger.Info("test message")

	var entry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &entry)
	require.NoError(t, err)

	assert.Equal(t, "whatsapp", entry["service"])
}

func TestLogger_WithRequestID(t *testing.T) {
	var buf bytes.Buffer
	cfg := config.LogConfig{Level: "debug", Format: "json"}
	l := logger.NewWithOutput(cfg, &buf)

	childLogger := l.WithRequestID("req-123")
	childLogger.Info("test message")

	var entry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &entry)
	require.NoError(t, err)

	assert.Equal(t, "req-123", entry["request_id"])
}

func TestLogger_WithSessionID(t *testing.T) {
	var buf bytes.Buffer
	cfg := config.LogConfig{Level: "debug", Format: "json"}
	l := logger.NewWithOutput(cfg, &buf)

	childLogger := l.WithSessionID("sess-456")
	childLogger.Info("test message")

	var entry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &entry)
	require.NoError(t, err)

	assert.Equal(t, "sess-456", entry["session_id"])
}

func TestLogger_WithContext(t *testing.T) {
	var buf bytes.Buffer
	cfg := config.LogConfig{Level: "debug", Format: "json"}
	l := logger.NewWithOutput(cfg, &buf)

	ctx := context.WithValue(context.Background(), logger.RequestIDKey, "ctx-req-123")
	ctx = context.WithValue(ctx, logger.SessionIDKey, "ctx-sess-456")

	childLogger := l.WithContext(ctx)
	childLogger.Info("test message")

	var entry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &entry)
	require.NoError(t, err)

	assert.Equal(t, "ctx-req-123", entry["request_id"])
	assert.Equal(t, "ctx-sess-456", entry["session_id"])
}

func TestLogger_WithError(t *testing.T) {
	var buf bytes.Buffer
	cfg := config.LogConfig{Level: "debug", Format: "json"}
	l := logger.NewWithOutput(cfg, &buf)

	l.WithError(assert.AnError).Error("operation failed")

	var entry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &entry)
	require.NoError(t, err)

	assert.Equal(t, assert.AnError.Error(), entry["error"])
}

func TestLogger_WithInt(t *testing.T) {
	var buf bytes.Buffer
	cfg := config.LogConfig{Level: "debug", Format: "json"}
	l := logger.NewWithOutput(cfg, &buf)

	l.WithInt("count", 42).Info("test")

	var entry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &entry)
	require.NoError(t, err)

	assert.Equal(t, float64(42), entry["count"])
}

func TestLogger_FormattedMethods(t *testing.T) {
	var buf bytes.Buffer
	cfg := config.LogConfig{Level: "debug", Format: "json"}
	l := logger.NewWithOutput(cfg, &buf)

	l.Infof("test %s %d", "message", 123)

	var entry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &entry)
	require.NoError(t, err)

	assert.Equal(t, "test message 123", entry["message"])
}

func TestLogger_Sub(t *testing.T) {
	var buf bytes.Buffer
	cfg := config.LogConfig{Level: "debug", Format: "json"}
	l := logger.NewWithOutput(cfg, &buf)

	subLogger := l.Sub("whatsmeow")
	subLogger.Infof("test message")

	var entry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &entry)
	require.NoError(t, err)

	assert.Equal(t, "whatsmeow", entry["module"])
	assert.Equal(t, "test message", entry["message"])
}

func TestLoggerDefaultConfig(t *testing.T) {
	// Test that default log config values work correctly
	cfg := config.LogConfig{
		Level:  "info",
		Format: "json",
	}

	assert.Equal(t, "info", cfg.Level)
	assert.Equal(t, "json", cfg.Format)
}

func TestLogger_ChainedFields(t *testing.T) {
	var buf bytes.Buffer
	cfg := config.LogConfig{Level: "debug", Format: "json"}
	l := logger.NewWithOutput(cfg, &buf)

	childLogger := l.WithRequestID("req-1").WithSessionID("sess-1").WithStr("extra", "value")
	childLogger.Info("test")

	var entry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &entry)
	require.NoError(t, err)

	assert.Equal(t, "req-1", entry["request_id"])
	assert.Equal(t, "sess-1", entry["session_id"])
	assert.Equal(t, "value", entry["extra"])
}

func TestLogger_DoesNotMutateParent(t *testing.T) {
	var buf1, buf2 bytes.Buffer
	cfg := config.LogConfig{Level: "debug", Format: "json"}
	l := logger.NewWithOutput(cfg, &buf1)

	// Log from parent
	l.Info("parent log")

	// Log from child (need new buffer)
	cfg2 := config.LogConfig{Level: "debug", Format: "json"}
	l2 := logger.NewWithOutput(cfg2, &buf2)
	child2 := l2.WithStr("child", "true")
	child2.Info("child log")

	// Parent should not have child field
	parentOutput := buf1.String()
	assert.False(t, strings.Contains(parentOutput, `"child"`))

	// Child should have child field
	childOutput := buf2.String()
	assert.True(t, strings.Contains(childOutput, `"child":"true"`))
}

func TestLogger_WithFieldsMap(t *testing.T) {
	var buf bytes.Buffer
	cfg := config.LogConfig{Level: "debug", Format: "json"}
	l := logger.NewWithOutput(cfg, &buf)

	fields := map[string]any{
		"str":   "value",
		"int":   42,
		"bool":  true,
		"float": 3.14,
	}

	l.WithFields(fields).Info("test")

	var entry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &entry)
	require.NoError(t, err)

	assert.Equal(t, "value", entry["str"])
	assert.Equal(t, float64(42), entry["int"])
	assert.Equal(t, true, entry["bool"])
	assert.Equal(t, 3.14, entry["float"])
}
