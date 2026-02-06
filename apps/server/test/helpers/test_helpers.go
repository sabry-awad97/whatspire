package helpers

import (
	"whatspire/internal/infrastructure/config"
	"whatspire/internal/infrastructure/logger"
	httpHandler "whatspire/internal/presentation/http"

	"github.com/gin-gonic/gin"
)

// ==================== Test Logger Helper ====================

// CreateTestLogger creates a logger for testing with minimal output
func CreateTestLogger() *logger.Logger {
	return logger.New(config.LogConfig{Level: "error", Format: "json"})
}

// ==================== Test Handler Helpers ====================

// NewTestHandlerBuilder creates a HandlerBuilder with a test logger for testing
func NewTestHandlerBuilder() *httpHandler.HandlerBuilder {
	log := CreateTestLogger()
	return httpHandler.NewHandlerBuilder(log)
}

// CreateTestRouter creates a router with the specified handler and config
func CreateTestRouter(handler *httpHandler.Handler, config httpHandler.RouterConfig) *gin.Engine {
	gin.SetMode(gin.TestMode)
	if config.Logger == nil {
		config.Logger = CreateTestLogger()
	}
	return httpHandler.NewRouter(handler, config)
}

// CreateTestRouterWithDefaults creates a router with default config
func CreateTestRouterWithDefaults(handler *httpHandler.Handler) *gin.Engine {
	config := httpHandler.DefaultRouterConfig()
	config.Logger = CreateTestLogger()
	return CreateTestRouter(handler, config)
}
