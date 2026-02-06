package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"whatspire/internal/application/dto"
	"whatspire/internal/application/usecase"
	"whatspire/internal/domain/repository"
	"whatspire/internal/infrastructure/config"
	httpHandler "whatspire/internal/presentation/http"
	"whatspire/test/helpers"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== Test Setup ====================

// MockAuditLogger is a no-op audit logger for testing
type MockAuditLogger struct{}

func (m *MockAuditLogger) LogAPIKeyUsage(ctx context.Context, event repository.APIKeyUsageEvent) {}

func (m *MockAuditLogger) LogAPIKeyCreated(ctx context.Context, event repository.APIKeyCreatedEvent) {
}

func (m *MockAuditLogger) LogAPIKeyRevoked(ctx context.Context, event repository.APIKeyRevokedEvent) {
}

func (m *MockAuditLogger) LogSessionAction(ctx context.Context, event repository.SessionActionEvent) {
}

func (m *MockAuditLogger) LogMessageSent(ctx context.Context, event repository.MessageSentEvent) {}

func (m *MockAuditLogger) LogAuthFailure(ctx context.Context, event repository.AuthFailureEvent) {}

func (m *MockAuditLogger) LogWebhookDelivery(ctx context.Context, event repository.WebhookDeliveryEvent) {
}

// MockAuditLogRepository is a no-op audit log repository for testing
type MockAuditLogRepository struct{}

func (m *MockAuditLogRepository) CountAPIKeyUsage(ctx context.Context, apiKeyID string) (int64, error) {
	return 0, nil
}

func (m *MockAuditLogRepository) CountAPIKeyUsageSince(ctx context.Context, apiKeyID string, since time.Time) (int64, error) {
	return 0, nil
}

func setupAPIKeyTestRouter(sessionUC *usecase.SessionUseCase, apiKeyConfig *config.APIKeyConfig, apiKeyRepo *helpers.MockAPIKeyRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)

	routerConfig := httpHandler.DefaultRouterConfig()
	routerConfig.APIKeyConfig = apiKeyConfig
	routerConfig.APIKeyRepository = apiKeyRepo    // Pass the repository to the router config
	routerConfig.AuditLogger = &MockAuditLogger{} // Pass the audit logger

	// Create API key use case with the mock repository and mock audit components
	apiKeyUC := usecase.NewAPIKeyUseCase(apiKeyRepo, &MockAuditLogger{}, &MockAuditLogRepository{})

	handler := helpers.NewTestHandlerBuilder().
		WithSessionUseCase(sessionUC).
		WithAPIKeyUseCase(apiKeyUC).
		Build()
	routerConfig.Logger = helpers.CreateTestLogger()
	return helpers.CreateTestRouter(handler, routerConfig)
}

// ==================== API Key Authentication Tests ====================

func TestAPIKey_ValidKey(t *testing.T) {
	repo := NewSessionRepositoryMock()
	sessionUC := usecase.NewSessionUseCase(repo, nil, nil, nil)

	// Create mock API key repository with a valid key
	apiKeyRepo := helpers.NewMockAPIKeyRepository()
	testKey := helpers.CreateTestAPIKey(t, apiKeyRepo, "admin", nil)

	apiKeyConfig := &config.APIKeyConfig{
		Enabled: true,
		Header:  "X-API-Key",
	}

	router := setupAPIKeyTestRouter(sessionUC, apiKeyConfig, apiKeyRepo)

	req := httptest.NewRequest(http.MethodPost, "/api/messages", nil)
	req.Header.Set("X-API-Key", testKey.PlainText)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// 400 is expected because we're not sending a valid message body,
	// but it proves the API key was accepted (not 401)
	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
}

func TestAPIKey_InvalidKey(t *testing.T) {
	repo := NewSessionRepositoryMock()
	sessionUC := usecase.NewSessionUseCase(repo, nil, nil, nil)

	// Create mock API key repository with a valid key
	apiKeyRepo := helpers.NewMockAPIKeyRepository()
	_ = helpers.CreateTestAPIKey(t, apiKeyRepo, "admin", nil)

	apiKeyConfig := &config.APIKeyConfig{
		Enabled: true,
		Header:  "X-API-Key",
	}

	router := setupAPIKeyTestRouter(sessionUC, apiKeyConfig, apiKeyRepo)

	req := httptest.NewRequest(http.MethodPost, "/api/messages", nil)
	req.Header.Set("X-API-Key", "invalid-api-key-not-in-database")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response dto.APIResponse[any]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Equal(t, "INVALID_API_KEY", response.Error.Code)
}

func TestAPIKey_MissingKey(t *testing.T) {
	repo := NewSessionRepositoryMock()
	sessionUC := usecase.NewSessionUseCase(repo, nil, nil, nil)

	apiKeyRepo := helpers.NewMockAPIKeyRepository()

	apiKeyConfig := &config.APIKeyConfig{
		Enabled: true,
		Header:  "X-API-Key",
	}

	router := setupAPIKeyTestRouter(sessionUC, apiKeyConfig, apiKeyRepo)

	req := httptest.NewRequest(http.MethodPost, "/api/messages", nil)
	// No API key header
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response dto.APIResponse[any]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Equal(t, "MISSING_API_KEY", response.Error.Code)
}

func TestAPIKey_Disabled(t *testing.T) {
	repo := NewSessionRepositoryMock()
	sessionUC := usecase.NewSessionUseCase(repo, nil, nil, nil)

	apiKeyRepo := helpers.NewMockAPIKeyRepository()

	apiKeyConfig := &config.APIKeyConfig{
		Enabled: false, // Disabled
		Header:  "X-API-Key",
	}

	router := setupAPIKeyTestRouter(sessionUC, apiKeyConfig, apiKeyRepo)

	// Request without API key should not get 401 when auth is disabled
	req := httptest.NewRequest(http.MethodPost, "/api/messages", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
}

func TestAPIKey_NoConfig(t *testing.T) {
	repo := NewSessionRepositoryMock()
	sessionUC := usecase.NewSessionUseCase(repo, nil, nil, nil)

	// No API key config
	router := setupAPIKeyTestRouter(sessionUC, nil, nil)

	// Request without API key should not get 401 when no config
	req := httptest.NewRequest(http.MethodPost, "/api/messages", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
}

func TestAPIKey_CustomHeader(t *testing.T) {
	repo := NewSessionRepositoryMock()
	sessionUC := usecase.NewSessionUseCase(repo, nil, nil, nil)

	apiKeyRepo := helpers.NewMockAPIKeyRepository()
	testKey := helpers.CreateTestAPIKey(t, apiKeyRepo, "admin", nil)

	apiKeyConfig := &config.APIKeyConfig{
		Enabled: true,
		Header:  "Authorization", // Custom header
	}

	router := setupAPIKeyTestRouter(sessionUC, apiKeyConfig, apiKeyRepo)

	// Request with custom header
	req := httptest.NewRequest(http.MethodPost, "/api/messages", nil)
	req.Header.Set("Authorization", testKey.PlainText)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Not 401 means API key was accepted
	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
}

func TestAPIKey_BearerToken(t *testing.T) {
	repo := NewSessionRepositoryMock()
	sessionUC := usecase.NewSessionUseCase(repo, nil, nil, nil)

	apiKeyRepo := helpers.NewMockAPIKeyRepository()
	testKey := helpers.CreateTestAPIKey(t, apiKeyRepo, "admin", nil)

	apiKeyConfig := &config.APIKeyConfig{
		Enabled: true,
		Header:  "X-API-Key",
	}

	router := setupAPIKeyTestRouter(sessionUC, apiKeyConfig, apiKeyRepo)

	// Request with Bearer token format
	req := httptest.NewRequest(http.MethodPost, "/api/messages", nil)
	req.Header.Set("Authorization", "Bearer "+testKey.PlainText)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Not 401 means API key was accepted
	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
}

func TestAPIKey_MultipleValidKeys(t *testing.T) {
	repo := NewSessionRepositoryMock()
	sessionUC := usecase.NewSessionUseCase(repo, nil, nil, nil)

	apiKeyRepo := helpers.NewMockAPIKeyRepository()
	testKey1 := helpers.CreateTestAPIKey(t, apiKeyRepo, "admin", nil)
	testKey2 := helpers.CreateTestAPIKey(t, apiKeyRepo, "write", nil)
	testKey3 := helpers.CreateTestAPIKey(t, apiKeyRepo, "read", nil)

	apiKeyConfig := &config.APIKeyConfig{
		Enabled: true,
		Header:  "X-API-Key",
	}

	router := setupAPIKeyTestRouter(sessionUC, apiKeyConfig, apiKeyRepo)

	// Test each valid key
	for _, testKey := range []*helpers.TestAPIKey{testKey1, testKey2, testKey3} {
		req := httptest.NewRequest(http.MethodPost, "/api/messages", nil)
		req.Header.Set("X-API-Key", testKey.PlainText)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.NotEqual(t, http.StatusUnauthorized, w.Code, "Key %s should be valid", testKey.Entity.ID)
	}
}

func TestAPIKey_HealthEndpointsNoAuth(t *testing.T) {
	repo := NewSessionRepositoryMock()
	sessionUC := usecase.NewSessionUseCase(repo, nil, nil, nil)

	apiKeyRepo := helpers.NewMockAPIKeyRepository()

	apiKeyConfig := &config.APIKeyConfig{
		Enabled: true,
		Header:  "X-API-Key",
	}

	router := setupAPIKeyTestRouter(sessionUC, apiKeyConfig, apiKeyRepo)

	// Health endpoints should not require API key
	healthEndpoints := []string{"/health", "/ready"}

	for _, endpoint := range healthEndpoints {
		req := httptest.NewRequest(http.MethodGet, endpoint, nil)
		// No API key
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Endpoint %s should not require API key", endpoint)
	}
}

func TestAPIKey_EmptyKey(t *testing.T) {
	repo := NewSessionRepositoryMock()
	sessionUC := usecase.NewSessionUseCase(repo, nil, nil, nil)

	apiKeyRepo := helpers.NewMockAPIKeyRepository()

	apiKeyConfig := &config.APIKeyConfig{
		Enabled: true,
		Header:  "X-API-Key",
	}

	router := setupAPIKeyTestRouter(sessionUC, apiKeyConfig, apiKeyRepo)

	// Request with empty API key
	req := httptest.NewRequest(http.MethodPost, "/api/messages", nil)
	req.Header.Set("X-API-Key", "")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response dto.APIResponse[any]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Equal(t, "MISSING_API_KEY", response.Error.Code)
}

func TestAPIKey_CaseSensitive(t *testing.T) {
	repo := NewSessionRepositoryMock()
	sessionUC := usecase.NewSessionUseCase(repo, nil, nil, nil)

	apiKeyRepo := helpers.NewMockAPIKeyRepository()

	// Create an API key with specific case
	testKey := helpers.CreateTestAPIKeyWithPlainText(t, apiKeyRepo, "ValidApiKey123", "write", nil)

	apiKeyConfig := &config.APIKeyConfig{
		Enabled: true,
		Header:  "X-API-Key",
	}

	router := setupAPIKeyTestRouter(sessionUC, apiKeyConfig, apiKeyRepo)

	// Exact case should work
	req1 := httptest.NewRequest(http.MethodPost, "/api/messages", nil)
	req1.Header.Set("X-API-Key", testKey.PlainText)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.NotEqual(t, http.StatusUnauthorized, w1.Code, "Exact case should authenticate successfully")

	// Different case should fail (different hash)
	req2 := httptest.NewRequest(http.MethodPost, "/api/messages", nil)
	req2.Header.Set("X-API-Key", "validapikey123") // lowercase version
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusUnauthorized, w2.Code, "Different case should fail authentication")
}

func TestAPIKey_RevokedKey(t *testing.T) {
	repo := NewSessionRepositoryMock()
	sessionUC := usecase.NewSessionUseCase(repo, nil, nil, nil)

	apiKeyRepo := helpers.NewMockAPIKeyRepository()
	testKey := helpers.CreateRevokedTestAPIKey(t, apiKeyRepo, "admin")

	apiKeyConfig := &config.APIKeyConfig{
		Enabled: true,
		Header:  "X-API-Key",
	}

	router := setupAPIKeyTestRouter(sessionUC, apiKeyConfig, apiKeyRepo)

	req := httptest.NewRequest(http.MethodPost, "/api/messages", nil)
	req.Header.Set("X-API-Key", testKey.PlainText)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response dto.APIResponse[any]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Equal(t, "REVOKED_API_KEY", response.Error.Code)
}

func TestAPIKey_InactiveKey(t *testing.T) {
	repo := NewSessionRepositoryMock()
	sessionUC := usecase.NewSessionUseCase(repo, nil, nil, nil)

	apiKeyRepo := helpers.NewMockAPIKeyRepository()
	testKey := helpers.CreateInactiveTestAPIKey(t, apiKeyRepo, "admin")

	apiKeyConfig := &config.APIKeyConfig{
		Enabled: true,
		Header:  "X-API-Key",
	}

	router := setupAPIKeyTestRouter(sessionUC, apiKeyConfig, apiKeyRepo)

	req := httptest.NewRequest(http.MethodPost, "/api/messages", nil)
	req.Header.Set("X-API-Key", testKey.PlainText)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response dto.APIResponse[any]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Equal(t, "REVOKED_API_KEY", response.Error.Code)
}
