package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"whatspire/internal/application/dto"
	"whatspire/internal/application/usecase"
	"whatspire/internal/infrastructure/config"
	httpHandler "whatspire/internal/presentation/http"
	"whatspire/test/helpers"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== Issue 2: API Key Revocation Tests ====================
// Tests for the revoked API keys still working issue

// TestAPIKeyRevocation_RevokedKeyRejected verifies that a revoked API key
// is properly rejected with 401 REVOKED_API_KEY error
func TestAPIKeyRevocation_RevokedKeyRejected(t *testing.T) {
	repo := NewSessionRepositoryMock()
	sessionUC := usecase.NewSessionUseCase(repo, nil, nil, nil)

	apiKeyRepo := helpers.NewMockAPIKeyRepository()
	testKey := helpers.CreateRevokedTestAPIKey(t, apiKeyRepo, "admin")

	apiKeyConfig := &config.APIKeyConfig{
		Enabled: true,
		Header:  "X-API-Key",
	}

	gin.SetMode(gin.TestMode)
	routerConfig := httpHandler.DefaultRouterConfig()
	routerConfig.APIKeyConfig = apiKeyConfig
	routerConfig.APIKeyRepository = apiKeyRepo
	routerConfig.AuditLogger = &MockAuditLogger{}

	handler := httpHandler.NewHandler(sessionUC, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	router := httpHandler.NewRouter(handler, routerConfig)

	// Try to use revoked key
	req := httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
	req.Header.Set("X-API-Key", testKey.PlainText)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 401 Unauthorized
	assert.Equal(t, http.StatusUnauthorized, w.Code, "Revoked key should return 401 Unauthorized")

	var response dto.APIResponse[interface{}]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Equal(t, "REVOKED_API_KEY", response.Error.Code, "Error code should be REVOKED_API_KEY")
	assert.Equal(t, "This API key has been revoked", response.Error.Message)
}

// TestAPIKeyRevocation_ActiveKeyAccepted verifies that an active (not revoked)
// API key is properly accepted
func TestAPIKeyRevocation_ActiveKeyAccepted(t *testing.T) {
	repo := NewSessionRepositoryMock()
	sessionUC := usecase.NewSessionUseCase(repo, nil, nil, nil)

	apiKeyRepo := helpers.NewMockAPIKeyRepository()
	testKey := helpers.CreateTestAPIKey(t, apiKeyRepo, "admin", nil)

	apiKeyConfig := &config.APIKeyConfig{
		Enabled: true,
		Header:  "X-API-Key",
	}

	gin.SetMode(gin.TestMode)
	routerConfig := httpHandler.DefaultRouterConfig()
	routerConfig.APIKeyConfig = apiKeyConfig
	routerConfig.APIKeyRepository = apiKeyRepo
	routerConfig.AuditLogger = &MockAuditLogger{}

	handler := httpHandler.NewHandler(sessionUC, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	router := httpHandler.NewRouter(handler, routerConfig)

	// Use active key
	req := httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
	req.Header.Set("X-API-Key", testKey.PlainText)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should not return 401 (might be 200 or other error, but not 401 for auth)
	assert.NotEqual(t, http.StatusUnauthorized, w.Code, "Active key should not return 401")
}

// TestAPIKeyRevocation_RevokedKeyWithBearerToken verifies that revoked keys
// are rejected even when using Bearer token format
func TestAPIKeyRevocation_RevokedKeyWithBearerToken(t *testing.T) {
	repo := NewSessionRepositoryMock()
	sessionUC := usecase.NewSessionUseCase(repo, nil, nil, nil)

	apiKeyRepo := helpers.NewMockAPIKeyRepository()
	testKey := helpers.CreateRevokedTestAPIKey(t, apiKeyRepo, "admin")

	apiKeyConfig := &config.APIKeyConfig{
		Enabled: true,
		Header:  "X-API-Key",
	}

	gin.SetMode(gin.TestMode)
	routerConfig := httpHandler.DefaultRouterConfig()
	routerConfig.APIKeyConfig = apiKeyConfig
	routerConfig.APIKeyRepository = apiKeyRepo
	routerConfig.AuditLogger = &MockAuditLogger{}

	handler := httpHandler.NewHandler(sessionUC, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	router := httpHandler.NewRouter(handler, routerConfig)

	// Try to use revoked key with Bearer token format
	req := httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
	req.Header.Set("Authorization", "Bearer "+testKey.PlainText)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 401 Unauthorized
	assert.Equal(t, http.StatusUnauthorized, w.Code, "Revoked key with Bearer token should return 401")

	var response dto.APIResponse[interface{}]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Equal(t, "REVOKED_API_KEY", response.Error.Code)
}

// TestAPIKeyRevocation_InactiveKeyRejected verifies that inactive keys
// (is_active = false) are rejected with REVOKED_API_KEY error
func TestAPIKeyRevocation_InactiveKeyRejected(t *testing.T) {
	repo := NewSessionRepositoryMock()
	sessionUC := usecase.NewSessionUseCase(repo, nil, nil, nil)

	apiKeyRepo := helpers.NewMockAPIKeyRepository()
	testKey := helpers.CreateInactiveTestAPIKey(t, apiKeyRepo, "admin")

	apiKeyConfig := &config.APIKeyConfig{
		Enabled: true,
		Header:  "X-API-Key",
	}

	gin.SetMode(gin.TestMode)
	routerConfig := httpHandler.DefaultRouterConfig()
	routerConfig.APIKeyConfig = apiKeyConfig
	routerConfig.APIKeyRepository = apiKeyRepo
	routerConfig.AuditLogger = &MockAuditLogger{}

	handler := httpHandler.NewHandler(sessionUC, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	router := httpHandler.NewRouter(handler, routerConfig)

	// Try to use inactive key
	req := httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
	req.Header.Set("X-API-Key", testKey.PlainText)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 401 Unauthorized
	assert.Equal(t, http.StatusUnauthorized, w.Code, "Inactive key should return 401")

	var response dto.APIResponse[interface{}]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Equal(t, "REVOKED_API_KEY", response.Error.Code)
}

// TestAPIKeyRevocation_MultipleRevokedKeys verifies that multiple revoked keys
// are all properly rejected
func TestAPIKeyRevocation_MultipleRevokedKeys(t *testing.T) {
	repo := NewSessionRepositoryMock()
	sessionUC := usecase.NewSessionUseCase(repo, nil, nil, nil)

	apiKeyRepo := helpers.NewMockAPIKeyRepository()

	// Create multiple revoked keys
	revokedKeys := make([]*helpers.TestAPIKey, 3)
	for i := 0; i < 3; i++ {
		revokedKeys[i] = helpers.CreateRevokedTestAPIKey(t, apiKeyRepo, "admin")
	}

	apiKeyConfig := &config.APIKeyConfig{
		Enabled: true,
		Header:  "X-API-Key",
	}

	gin.SetMode(gin.TestMode)
	routerConfig := httpHandler.DefaultRouterConfig()
	routerConfig.APIKeyConfig = apiKeyConfig
	routerConfig.APIKeyRepository = apiKeyRepo
	routerConfig.AuditLogger = &MockAuditLogger{}

	handler := httpHandler.NewHandler(sessionUC, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	router := httpHandler.NewRouter(handler, routerConfig)

	// Try each revoked key
	for i, testKey := range revokedKeys {
		req := httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
		req.Header.Set("X-API-Key", testKey.PlainText)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code, "Revoked key %d should return 401", i)

		var response dto.APIResponse[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "REVOKED_API_KEY", response.Error.Code, "Revoked key %d should have REVOKED_API_KEY error", i)
	}
}

// TestAPIKeyRevocation_RevokedKeyCannotAccessAnyEndpoint verifies that revoked keys
// are rejected on all protected endpoints
func TestAPIKeyRevocation_RevokedKeyCannotAccessAnyEndpoint(t *testing.T) {
	repo := NewSessionRepositoryMock()
	sessionUC := usecase.NewSessionUseCase(repo, nil, nil, nil)

	apiKeyRepo := helpers.NewMockAPIKeyRepository()
	testKey := helpers.CreateRevokedTestAPIKey(t, apiKeyRepo, "admin")

	apiKeyConfig := &config.APIKeyConfig{
		Enabled: true,
		Header:  "X-API-Key",
	}

	gin.SetMode(gin.TestMode)
	routerConfig := httpHandler.DefaultRouterConfig()
	routerConfig.APIKeyConfig = apiKeyConfig
	routerConfig.APIKeyRepository = apiKeyRepo
	routerConfig.AuditLogger = &MockAuditLogger{}

	handler := httpHandler.NewHandler(sessionUC, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	router := httpHandler.NewRouter(handler, routerConfig)

	// Test multiple endpoints with correct HTTP methods
	// Format: [method, endpoint]
	endpoints := [][]string{
		{http.MethodGet, "/api/sessions"},
		{http.MethodGet, "/api/contacts/check"},
		{http.MethodGet, "/api/events"},
		{http.MethodGet, "/api/apikeys"},
	}

	for _, endpointInfo := range endpoints {
		method := endpointInfo[0]
		endpoint := endpointInfo[1]

		req := httptest.NewRequest(method, endpoint, nil)
		req.Header.Set("X-API-Key", testKey.PlainText)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code, "Revoked key should be rejected on %s %s", method, endpoint)

		var response dto.APIResponse[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "REVOKED_API_KEY", response.Error.Code, "Error on %s %s should be REVOKED_API_KEY", method, endpoint)
	}
}

// TestAPIKeyRevocation_RevokedKeyWithDifferentRoles verifies that revoked keys
// are rejected regardless of their role
func TestAPIKeyRevocation_RevokedKeyWithDifferentRoles(t *testing.T) {
	repo := NewSessionRepositoryMock()
	sessionUC := usecase.NewSessionUseCase(repo, nil, nil, nil)

	apiKeyRepo := helpers.NewMockAPIKeyRepository()

	// Create revoked keys with different roles
	roles := []string{"admin", "write", "read"}
	revokedKeys := make([]*helpers.TestAPIKey, len(roles))

	for i, role := range roles {
		revokedKeys[i] = helpers.CreateRevokedTestAPIKey(t, apiKeyRepo, role)
	}

	apiKeyConfig := &config.APIKeyConfig{
		Enabled: true,
		Header:  "X-API-Key",
	}

	gin.SetMode(gin.TestMode)
	routerConfig := httpHandler.DefaultRouterConfig()
	routerConfig.APIKeyConfig = apiKeyConfig
	routerConfig.APIKeyRepository = apiKeyRepo
	routerConfig.AuditLogger = &MockAuditLogger{}

	handler := httpHandler.NewHandler(sessionUC, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	router := httpHandler.NewRouter(handler, routerConfig)

	// Try each revoked key with different role
	for i, testKey := range revokedKeys {
		req := httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
		req.Header.Set("X-API-Key", testKey.PlainText)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code, "Revoked %s key should return 401", roles[i])

		var response dto.APIResponse[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "REVOKED_API_KEY", response.Error.Code, "Revoked %s key should have REVOKED_API_KEY error", roles[i])
	}
}

// TestAPIKeyRevocation_RevokedKeyDoesNotBypassRoleCheck verifies that even if
// a revoked key somehow passed the is_active check, it would still fail on role check
func TestAPIKeyRevocation_RevokedKeyDoesNotBypassRoleCheck(t *testing.T) {
	repo := NewSessionRepositoryMock()
	sessionUC := usecase.NewSessionUseCase(repo, nil, nil, nil)

	apiKeyRepo := helpers.NewMockAPIKeyRepository()
	// Create a revoked key with read role
	testKey := helpers.CreateRevokedTestAPIKey(t, apiKeyRepo, "read")

	apiKeyConfig := &config.APIKeyConfig{
		Enabled: true,
		Header:  "X-API-Key",
	}

	gin.SetMode(gin.TestMode)
	routerConfig := httpHandler.DefaultRouterConfig()
	routerConfig.APIKeyConfig = apiKeyConfig
	routerConfig.APIKeyRepository = apiKeyRepo
	routerConfig.AuditLogger = &MockAuditLogger{}

	handler := httpHandler.NewHandler(sessionUC, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	router := httpHandler.NewRouter(handler, routerConfig)

	// Try to access endpoint that requires write role with revoked read key
	req := httptest.NewRequest(http.MethodPost, "/api/messages", nil)
	req.Header.Set("X-API-Key", testKey.PlainText)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should fail on revocation check first (401 REVOKED_API_KEY)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response dto.APIResponse[interface{}]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "REVOKED_API_KEY", response.Error.Code)
}

// TestAPIKeyRevocation_RevokeAndRetry verifies that after revoking a key,
// it can no longer be used
func TestAPIKeyRevocation_RevokeAndRetry(t *testing.T) {
	repo := NewSessionRepositoryMock()
	sessionUC := usecase.NewSessionUseCase(repo, nil, nil, nil)

	apiKeyRepo := helpers.NewMockAPIKeyRepository()
	testKey := helpers.CreateTestAPIKey(t, apiKeyRepo, "admin", nil)

	apiKeyConfig := &config.APIKeyConfig{
		Enabled: true,
		Header:  "X-API-Key",
	}

	gin.SetMode(gin.TestMode)
	routerConfig := httpHandler.DefaultRouterConfig()
	routerConfig.APIKeyConfig = apiKeyConfig
	routerConfig.APIKeyRepository = apiKeyRepo
	routerConfig.AuditLogger = &MockAuditLogger{}

	handler := httpHandler.NewHandler(sessionUC, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	router := httpHandler.NewRouter(handler, routerConfig)

	// First request should succeed
	req1 := httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
	req1.Header.Set("X-API-Key", testKey.PlainText)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.NotEqual(t, http.StatusUnauthorized, w1.Code, "Active key should work initially")

	// Revoke the key
	reason := "Test revocation"
	testKey.Entity.Revoke("test-admin", &reason)
	err := apiKeyRepo.Update(context.Background(), testKey.Entity)
	require.NoError(t, err)

	// Second request should fail
	req2 := httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
	req2.Header.Set("X-API-Key", testKey.PlainText)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusUnauthorized, w2.Code, "Revoked key should not work after revocation")

	var response dto.APIResponse[interface{}]
	err = json.Unmarshal(w2.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "REVOKED_API_KEY", response.Error.Code)
}

// TestAPIKeyRevocation_AuthenticationEnabledByDefault verifies that when
// apikey.enabled is true, authentication is enforced
func TestAPIKeyRevocation_AuthenticationEnabledByDefault(t *testing.T) {
	repo := NewSessionRepositoryMock()
	sessionUC := usecase.NewSessionUseCase(repo, nil, nil, nil)

	apiKeyRepo := helpers.NewMockAPIKeyRepository()

	apiKeyConfig := &config.APIKeyConfig{
		Enabled: true, // Authentication enabled
		Header:  "X-API-Key",
	}

	gin.SetMode(gin.TestMode)
	routerConfig := httpHandler.DefaultRouterConfig()
	routerConfig.APIKeyConfig = apiKeyConfig
	routerConfig.APIKeyRepository = apiKeyRepo
	routerConfig.AuditLogger = &MockAuditLogger{}

	handler := httpHandler.NewHandler(sessionUC, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	router := httpHandler.NewRouter(handler, routerConfig)

	// Request without API key should fail
	req := httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code, "Request without API key should fail when auth is enabled")

	var response dto.APIResponse[interface{}]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "MISSING_API_KEY", response.Error.Code)
}

// TestAPIKeyRevocation_AuthenticationDisabledAllowsAccess verifies that when
// apikey.enabled is false, requests are allowed without authentication
func TestAPIKeyRevocation_AuthenticationDisabledAllowsAccess(t *testing.T) {
	repo := NewSessionRepositoryMock()
	sessionUC := usecase.NewSessionUseCase(repo, nil, nil, nil)

	apiKeyConfig := &config.APIKeyConfig{
		Enabled: false, // Authentication disabled
		Header:  "X-API-Key",
	}

	gin.SetMode(gin.TestMode)
	routerConfig := httpHandler.DefaultRouterConfig()
	routerConfig.APIKeyConfig = apiKeyConfig
	routerConfig.APIKeyRepository = nil
	routerConfig.AuditLogger = &MockAuditLogger{}

	handler := httpHandler.NewHandler(sessionUC, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	router := httpHandler.NewRouter(handler, routerConfig)

	// Request without API key should succeed when auth is disabled
	req := httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.NotEqual(t, http.StatusUnauthorized, w.Code, "Request without API key should succeed when auth is disabled")
}

// TestAPIKeyRevocation_RevokedKeyErrorMessage verifies that the error message
// for revoked keys is clear and informative
func TestAPIKeyRevocation_RevokedKeyErrorMessage(t *testing.T) {
	repo := NewSessionRepositoryMock()
	sessionUC := usecase.NewSessionUseCase(repo, nil, nil, nil)

	apiKeyRepo := helpers.NewMockAPIKeyRepository()
	testKey := helpers.CreateRevokedTestAPIKey(t, apiKeyRepo, "admin")

	apiKeyConfig := &config.APIKeyConfig{
		Enabled: true,
		Header:  "X-API-Key",
	}

	gin.SetMode(gin.TestMode)
	routerConfig := httpHandler.DefaultRouterConfig()
	routerConfig.APIKeyConfig = apiKeyConfig
	routerConfig.APIKeyRepository = apiKeyRepo
	routerConfig.AuditLogger = &MockAuditLogger{}

	handler := httpHandler.NewHandler(sessionUC, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	router := httpHandler.NewRouter(handler, routerConfig)

	req := httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
	req.Header.Set("X-API-Key", testKey.PlainText)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	var response dto.APIResponse[interface{}]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify error message is clear
	assert.Equal(t, "REVOKED_API_KEY", response.Error.Code)
	assert.Equal(t, "This API key has been revoked", response.Error.Message)
	assert.NotEmpty(t, response.Error.Message, "Error message should not be empty")
}
