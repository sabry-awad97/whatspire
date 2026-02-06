package helpers

import (
	"whatspire/internal/application/usecase"
	"whatspire/internal/domain/repository"
	"whatspire/internal/infrastructure/logger"
	httpHandler "whatspire/internal/presentation/http"

	"github.com/gin-gonic/gin"
)

// ==================== Test Logger Helper ====================

// CreateTestLogger creates a logger for testing with minimal output
func CreateTestLogger() *logger.Logger {
	return logger.New("error", "json")
}

// ==================== Test MessageUseCase Helper ====================

// NewTestMessageUseCaseBuilder creates a MessageUseCaseBuilder with test logger and default config
func NewTestMessageUseCaseBuilder() *usecase.MessageUseCaseBuilder {
	config := usecase.DefaultMessageUseCaseConfig()
	log := CreateTestLogger()
	return usecase.NewMessageUseCaseBuilder(config, log)
}

// NewTestMessageUseCase creates a MessageUseCase with the provided dependencies for testing
func NewTestMessageUseCase(
	waClient repository.WhatsAppClient,
	publisher repository.EventPublisher,
	mediaUploader repository.MediaUploader,
	auditLogger repository.AuditLogger,
) *usecase.MessageUseCase {
	return NewTestMessageUseCaseBuilder().
		WithWhatsAppClient(waClient).
		WithEventPublisher(publisher).
		WithMediaUploader(mediaUploader).
		WithAuditLogger(auditLogger).
		Build()
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
