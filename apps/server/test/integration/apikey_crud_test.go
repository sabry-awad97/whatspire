package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"whatspire/internal/application/dto"
	"whatspire/internal/application/usecase"
	"whatspire/internal/domain/repository"
	"whatspire/internal/infrastructure/persistence"
	"whatspire/test/helpers"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ==================== Test Setup ====================

func setupAPIKeyCRUDTestRouter(t *testing.T) (*gin.Engine, *usecase.APIKeyUseCase, *gorm.DB) {
	gin.SetMode(gin.TestMode)

	// Create in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Run migrations
	err = persistence.RunAutoMigration(db)
	require.NoError(t, err)

	// Create repository and use case
	repo := persistence.NewAPIKeyRepository(db)
	auditLogRepo := persistence.NewAuditLogRepository(db)
	auditLogger := &AuditLoggerMock{} // Mock audit logger
	apiKeyUC := usecase.NewAPIKeyUseCase(repo, auditLogger, auditLogRepo)

	// Create handler and router
	handler := helpers.NewTestHandlerBuilder().
		WithAPIKeyUseCase(apiKeyUC).
		Build()
	router := helpers.CreateTestRouterWithDefaults(handler)

	return router, apiKeyUC, db
}

// AuditLoggerMock is a simple mock for audit logging
type AuditLoggerMock struct {
	Logs []string
}

func (m *AuditLoggerMock) LogAPIKeyUsage(ctx context.Context, event repository.APIKeyUsageEvent) {
	m.Logs = append(m.Logs, "used:"+event.APIKeyID)
}

func (m *AuditLoggerMock) LogAPIKeyCreated(ctx context.Context, event repository.APIKeyCreatedEvent) {
	m.Logs = append(m.Logs, "created:"+event.APIKeyID)
}

func (m *AuditLoggerMock) LogAPIKeyRevoked(ctx context.Context, event repository.APIKeyRevokedEvent) {
	m.Logs = append(m.Logs, "revoked:"+event.APIKeyID)
}

func (m *AuditLoggerMock) LogSessionAction(ctx context.Context, event repository.SessionActionEvent) {
	m.Logs = append(m.Logs, "session:"+event.SessionID)
}

func (m *AuditLoggerMock) LogMessageSent(ctx context.Context, event repository.MessageSentEvent) {
	m.Logs = append(m.Logs, "message:"+event.SessionID)
}

func (m *AuditLoggerMock) LogAuthFailure(ctx context.Context, event repository.AuthFailureEvent) {
	m.Logs = append(m.Logs, "auth_failure:"+event.APIKey)
}

func (m *AuditLoggerMock) LogWebhookDelivery(ctx context.Context, event repository.WebhookDeliveryEvent) {
	m.Logs = append(m.Logs, "webhook:"+event.WebhookURL)
}

// ==================== POST /api/apikeys Tests (Create) ====================

func TestCreateAPIKey_Success(t *testing.T) {
	router, _, db := setupAPIKeyCRUDTestRouter(t)
	defer db.Exec("DELETE FROM api_keys WHERE 1=1")

	description := "Test API Key"
	reqBody := dto.CreateAPIKeyRequest{
		Role:        "read",
		Description: &description,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/apikeys", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response dto.APIResponse[dto.CreateAPIKeyResponse]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.NotEmpty(t, response.Data.PlainKey)
	assert.NotEmpty(t, response.Data.APIKey.ID)
	assert.Equal(t, "read", response.Data.APIKey.Role)
	assert.Equal(t, &description, response.Data.APIKey.Description)
	assert.True(t, response.Data.APIKey.IsActive)
	assert.Nil(t, response.Data.APIKey.RevokedAt)
}

func TestCreateAPIKey_WithoutDescription(t *testing.T) {
	router, _, db := setupAPIKeyCRUDTestRouter(t)
	defer db.Exec("DELETE FROM api_keys WHERE 1=1")

	reqBody := dto.CreateAPIKeyRequest{
		Role: "write",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/apikeys", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response dto.APIResponse[dto.CreateAPIKeyResponse]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.NotEmpty(t, response.Data.PlainKey)
	assert.Nil(t, response.Data.APIKey.Description)
}

func TestCreateAPIKey_InvalidRole(t *testing.T) {
	router, _, db := setupAPIKeyCRUDTestRouter(t)
	defer db.Exec("DELETE FROM api_keys WHERE 1=1")

	reqBody := dto.CreateAPIKeyRequest{
		Role: "invalid_role",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/apikeys", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response dto.APIResponse[interface{}]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Equal(t, "VALIDATION_FAILED", response.Error.Code)
}

func TestCreateAPIKey_MissingRole(t *testing.T) {
	router, _, db := setupAPIKeyCRUDTestRouter(t)
	defer db.Exec("DELETE FROM api_keys WHERE 1=1")

	reqBody := dto.CreateAPIKeyRequest{
		Role: "",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/apikeys", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateAPIKey_InvalidJSON(t *testing.T) {
	router, _, db := setupAPIKeyCRUDTestRouter(t)
	defer db.Exec("DELETE FROM api_keys WHERE 1=1")

	req := httptest.NewRequest(http.MethodPost, "/api/apikeys", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response dto.APIResponse[interface{}]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Equal(t, "INVALID_JSON", response.Error.Code)
}

// ==================== DELETE /api/apikeys/:id Tests (Revoke) ====================

func TestRevokeAPIKey_Success(t *testing.T) {
	router, apiKeyUC, db := setupAPIKeyCRUDTestRouter(t)
	defer db.Exec("DELETE FROM api_keys WHERE 1=1")

	// Create an API key first
	_, apiKey, err := apiKeyUC.CreateAPIKey(context.Background(), "read", nil, "test-user")
	require.NoError(t, err)

	// Revoke it
	reason := "Security audit"
	reqBody := dto.RevokeAPIKeyRequest{
		Reason: &reason,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodDelete, "/api/apikeys/"+apiKey.ID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.APIResponse[dto.RevokeAPIKeyResponse]
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.Equal(t, apiKey.ID, response.Data.ID)
	assert.NotNil(t, response.Data.RevokedAt)
	assert.Equal(t, "system", response.Data.RevokedBy)
}

func TestRevokeAPIKey_WithoutReason(t *testing.T) {
	router, apiKeyUC, db := setupAPIKeyCRUDTestRouter(t)
	defer db.Exec("DELETE FROM api_keys WHERE 1=1")

	// Create an API key first
	_, apiKey, err := apiKeyUC.CreateAPIKey(context.Background(), "write", nil, "test-user")
	require.NoError(t, err)

	// Revoke without reason
	req := httptest.NewRequest(http.MethodDelete, "/api/apikeys/"+apiKey.ID, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.APIResponse[dto.RevokeAPIKeyResponse]
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.Equal(t, apiKey.ID, response.Data.ID)
}

func TestRevokeAPIKey_NotFound(t *testing.T) {
	router, _, db := setupAPIKeyCRUDTestRouter(t)
	defer db.Exec("DELETE FROM api_keys WHERE 1=1")

	req := httptest.NewRequest(http.MethodDelete, "/api/apikeys/nonexistent-id", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response dto.APIResponse[interface{}]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Equal(t, "NOT_FOUND", response.Error.Code)
}

func TestRevokeAPIKey_AlreadyRevoked(t *testing.T) {
	router, apiKeyUC, db := setupAPIKeyCRUDTestRouter(t)
	defer db.Exec("DELETE FROM api_keys WHERE 1=1")

	// Create and revoke an API key
	_, apiKey, err := apiKeyUC.CreateAPIKey(context.Background(), "admin", nil, "test-user")
	require.NoError(t, err)

	_, err = apiKeyUC.RevokeAPIKey(context.Background(), apiKey.ID, "test-user", nil)
	require.NoError(t, err)

	// Try to revoke again
	req := httptest.NewRequest(http.MethodDelete, "/api/apikeys/"+apiKey.ID, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response dto.APIResponse[interface{}]
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Equal(t, "ALREADY_REVOKED", response.Error.Code)
}

// ==================== GET /api/apikeys Tests (List) ====================

func TestListAPIKeys_Success(t *testing.T) {
	router, apiKeyUC, db := setupAPIKeyCRUDTestRouter(t)
	defer db.Exec("DELETE FROM api_keys WHERE 1=1")

	// Create multiple API keys
	for range 5 {
		_, _, err := apiKeyUC.CreateAPIKey(context.Background(), "read", nil, "test-user")
		require.NoError(t, err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/apikeys", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.APIResponse[dto.ListAPIKeysResponse]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.Len(t, response.Data.APIKeys, 5)
	assert.Equal(t, int64(5), response.Data.Pagination.Total)
	assert.Equal(t, 1, response.Data.Pagination.Page)
	assert.Equal(t, 50, response.Data.Pagination.Limit)
}

func TestListAPIKeys_WithPagination(t *testing.T) {
	router, apiKeyUC, db := setupAPIKeyCRUDTestRouter(t)
	defer db.Exec("DELETE FROM api_keys WHERE 1=1")

	// Create 10 API keys
	for range 10 {
		_, _, err := apiKeyUC.CreateAPIKey(context.Background(), "read", nil, "test-user")
		require.NoError(t, err)
	}

	// Get first page (5 items)
	req := httptest.NewRequest(http.MethodGet, "/api/apikeys?page=1&limit=5", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.APIResponse[dto.ListAPIKeysResponse]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.Len(t, response.Data.APIKeys, 5)
	assert.Equal(t, int64(10), response.Data.Pagination.Total)
	assert.Equal(t, 1, response.Data.Pagination.Page)
	assert.Equal(t, 2, response.Data.Pagination.TotalPages)

	// Get second page (5 items)
	req2 := httptest.NewRequest(http.MethodGet, "/api/apikeys?page=2&limit=5", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)

	var response2 dto.APIResponse[dto.ListAPIKeysResponse]
	err = json.Unmarshal(w2.Body.Bytes(), &response2)
	require.NoError(t, err)

	assert.True(t, response2.Success)
	assert.Len(t, response2.Data.APIKeys, 5)
	assert.Equal(t, 2, response2.Data.Pagination.Page)
}

func TestListAPIKeys_FilterByRole(t *testing.T) {
	router, apiKeyUC, db := setupAPIKeyCRUDTestRouter(t)
	defer db.Exec("DELETE FROM api_keys WHERE 1=1")

	// Create keys with different roles
	roles := []string{"read", "write", "admin", "read", "write"}
	for _, role := range roles {
		_, _, err := apiKeyUC.CreateAPIKey(context.Background(), role, nil, "test-user")
		require.NoError(t, err)
	}

	// Filter by read role
	req := httptest.NewRequest(http.MethodGet, "/api/apikeys?role=read", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.APIResponse[dto.ListAPIKeysResponse]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.Len(t, response.Data.APIKeys, 2)
	for _, key := range response.Data.APIKeys {
		assert.Equal(t, "read", key.Role)
	}
}

func TestListAPIKeys_FilterByStatus(t *testing.T) {
	router, apiKeyUC, db := setupAPIKeyCRUDTestRouter(t)
	defer db.Exec("DELETE FROM api_keys WHERE 1=1")

	// Create 3 keys
	var keyIDs []string
	for range 3 {
		_, apiKey, err := apiKeyUC.CreateAPIKey(context.Background(), "read", nil, "test-user")
		require.NoError(t, err)
		keyIDs = append(keyIDs, apiKey.ID)
	}

	// Revoke one key
	_, err := apiKeyUC.RevokeAPIKey(context.Background(), keyIDs[0], "test-user", nil)
	require.NoError(t, err)

	// Filter by active status
	req := httptest.NewRequest(http.MethodGet, "/api/apikeys?status=active", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.APIResponse[dto.ListAPIKeysResponse]
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.Len(t, response.Data.APIKeys, 2)
	for _, key := range response.Data.APIKeys {
		assert.True(t, key.IsActive)
	}

	// Filter by revoked status
	req2 := httptest.NewRequest(http.MethodGet, "/api/apikeys?status=revoked", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)

	var response2 dto.APIResponse[dto.ListAPIKeysResponse]
	err = json.Unmarshal(w2.Body.Bytes(), &response2)
	require.NoError(t, err)

	assert.True(t, response2.Success)
	assert.Len(t, response2.Data.APIKeys, 1)
	assert.False(t, response2.Data.APIKeys[0].IsActive)
}

func TestListAPIKeys_CombinedFilters(t *testing.T) {
	router, apiKeyUC, db := setupAPIKeyCRUDTestRouter(t)
	defer db.Exec("DELETE FROM api_keys WHERE 1=1")

	// Create keys with different roles
	_, key1, _ := apiKeyUC.CreateAPIKey(context.Background(), "read", nil, "test-user")
	_, key2, _ := apiKeyUC.CreateAPIKey(context.Background(), "read", nil, "test-user")
	_, _, _ = apiKeyUC.CreateAPIKey(context.Background(), "write", nil, "test-user")

	// Revoke one read key
	apiKeyUC.RevokeAPIKey(context.Background(), key1.ID, "test-user", nil)

	// Filter by read role and active status
	req := httptest.NewRequest(http.MethodGet, "/api/apikeys?role=read&status=active", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.APIResponse[dto.ListAPIKeysResponse]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.Len(t, response.Data.APIKeys, 1)
	assert.Equal(t, key2.ID, response.Data.APIKeys[0].ID)
	assert.Equal(t, "read", response.Data.APIKeys[0].Role)
	assert.True(t, response.Data.APIKeys[0].IsActive)
}

func TestListAPIKeys_EmptyResult(t *testing.T) {
	router, _, db := setupAPIKeyCRUDTestRouter(t)
	defer db.Exec("DELETE FROM api_keys WHERE 1=1")

	req := httptest.NewRequest(http.MethodGet, "/api/apikeys", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.APIResponse[dto.ListAPIKeysResponse]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.Len(t, response.Data.APIKeys, 0)
	assert.Equal(t, int64(0), response.Data.Pagination.Total)
}

func TestListAPIKeys_InvalidQueryParams(t *testing.T) {
	router, _, db := setupAPIKeyCRUDTestRouter(t)
	defer db.Exec("DELETE FROM api_keys WHERE 1=1")

	// Invalid page number
	req := httptest.NewRequest(http.MethodGet, "/api/apikeys?page=-1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ==================== GET /api/apikeys/:id Tests (Details) ====================

func TestGetAPIKeyDetails_Success(t *testing.T) {
	router, apiKeyUC, db := setupAPIKeyCRUDTestRouter(t)
	defer db.Exec("DELETE FROM api_keys WHERE 1=1")

	// Create an API key
	description := "Test Key for Details"
	_, apiKey, err := apiKeyUC.CreateAPIKey(context.Background(), "admin", &description, "test-user")
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/apikeys/"+apiKey.ID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.APIResponse[dto.APIKeyDetailsResponse]
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.Equal(t, apiKey.ID, response.Data.APIKey.ID)
	assert.Equal(t, "admin", response.Data.APIKey.Role)
	assert.Equal(t, &description, response.Data.APIKey.Description)
	assert.NotEmpty(t, response.Data.APIKey.MaskedKey)
	assert.True(t, response.Data.APIKey.IsActive)
	assert.NotNil(t, response.Data.UsageStats)
}

func TestGetAPIKeyDetails_NotFound(t *testing.T) {
	router, _, db := setupAPIKeyCRUDTestRouter(t)
	defer db.Exec("DELETE FROM api_keys WHERE 1=1")

	req := httptest.NewRequest(http.MethodGet, "/api/apikeys/nonexistent-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response dto.APIResponse[interface{}]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Equal(t, "NOT_FOUND", response.Error.Code)
}

func TestGetAPIKeyDetails_RevokedKey(t *testing.T) {
	router, apiKeyUC, db := setupAPIKeyCRUDTestRouter(t)
	defer db.Exec("DELETE FROM api_keys WHERE 1=1")

	// Create and revoke an API key
	reason := "Test revocation"
	_, apiKey, err := apiKeyUC.CreateAPIKey(context.Background(), "write", nil, "test-user")
	require.NoError(t, err)

	_, err = apiKeyUC.RevokeAPIKey(context.Background(), apiKey.ID, "admin-user", &reason)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/apikeys/"+apiKey.ID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.APIResponse[dto.APIKeyDetailsResponse]
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.False(t, response.Data.APIKey.IsActive)
	assert.NotNil(t, response.Data.APIKey.RevokedAt)
	assert.NotNil(t, response.Data.APIKey.RevokedBy)
	assert.Equal(t, "admin-user", *response.Data.APIKey.RevokedBy)
	assert.NotNil(t, response.Data.APIKey.RevocationReason)
	assert.Equal(t, reason, *response.Data.APIKey.RevocationReason)
}

// ==================== Complete Flow Tests ====================

func TestCompleteFlow_CreateUseRevokeVerify(t *testing.T) {
	router, _, db := setupAPIKeyCRUDTestRouter(t)
	defer db.Exec("DELETE FROM api_keys WHERE 1=1")

	// Step 1: Create an API key
	description := "Integration Test Key"
	reqBody := dto.CreateAPIKeyRequest{
		Role:        "read",
		Description: &description,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/apikeys", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var createResponse dto.APIResponse[dto.CreateAPIKeyResponse]
	err := json.Unmarshal(w.Body.Bytes(), &createResponse)
	require.NoError(t, err)

	plainKey := createResponse.Data.PlainKey
	keyID := createResponse.Data.APIKey.ID

	// Step 2: Verify key is active in list
	req2 := httptest.NewRequest(http.MethodGet, "/api/apikeys", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)

	var listResponse dto.APIResponse[dto.ListAPIKeysResponse]
	err = json.Unmarshal(w2.Body.Bytes(), &listResponse)
	require.NoError(t, err)

	assert.Len(t, listResponse.Data.APIKeys, 1)
	assert.True(t, listResponse.Data.APIKeys[0].IsActive)

	// Step 3: Simulate key usage (would normally be done by middleware)
	// For now, we'll just verify the key exists
	assert.NotEmpty(t, plainKey)

	// Step 4: Revoke the key
	revokeReason := "End of integration test"
	revokeBody := dto.RevokeAPIKeyRequest{
		Reason: &revokeReason,
	}
	revokeJSON, _ := json.Marshal(revokeBody)

	req3 := httptest.NewRequest(http.MethodDelete, "/api/apikeys/"+keyID, bytes.NewReader(revokeJSON))
	req3.Header.Set("Content-Type", "application/json")
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)

	assert.Equal(t, http.StatusOK, w3.Code)

	// Step 5: Verify key is revoked
	req4 := httptest.NewRequest(http.MethodGet, "/api/apikeys/"+keyID, nil)
	w4 := httptest.NewRecorder()
	router.ServeHTTP(w4, req4)

	assert.Equal(t, http.StatusOK, w4.Code)

	var detailsResponse dto.APIResponse[dto.APIKeyDetailsResponse]
	err = json.Unmarshal(w4.Body.Bytes(), &detailsResponse)
	require.NoError(t, err)

	assert.False(t, detailsResponse.Data.APIKey.IsActive)
	assert.NotNil(t, detailsResponse.Data.APIKey.RevokedAt)
	assert.Equal(t, &revokeReason, detailsResponse.Data.APIKey.RevocationReason)

	// Step 6: Verify revoked key appears in revoked filter
	req5 := httptest.NewRequest(http.MethodGet, "/api/apikeys?status=revoked", nil)
	w5 := httptest.NewRecorder()
	router.ServeHTTP(w5, req5)

	assert.Equal(t, http.StatusOK, w5.Code)

	var revokedListResponse dto.APIResponse[dto.ListAPIKeysResponse]
	err = json.Unmarshal(w5.Body.Bytes(), &revokedListResponse)
	require.NoError(t, err)

	assert.Len(t, revokedListResponse.Data.APIKeys, 1)
	assert.Equal(t, keyID, revokedListResponse.Data.APIKeys[0].ID)
	assert.False(t, revokedListResponse.Data.APIKeys[0].IsActive)
}

func TestCompleteFlow_MultipleKeysWithDifferentRoles(t *testing.T) {
	router, apiKeyUC, db := setupAPIKeyCRUDTestRouter(t)
	defer db.Exec("DELETE FROM api_keys WHERE 1=1")

	// Create keys with different roles
	roles := []string{"read", "write", "admin"}
	createdKeys := make(map[string]string) // role -> keyID

	for _, role := range roles {
		desc := role + " key"
		reqBody := dto.CreateAPIKeyRequest{
			Role:        role,
			Description: &desc,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/apikeys", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response dto.APIResponse[dto.CreateAPIKeyResponse]
		json.Unmarshal(w.Body.Bytes(), &response)
		createdKeys[role] = response.Data.APIKey.ID
	}

	// Verify all keys exist
	req := httptest.NewRequest(http.MethodGet, "/api/apikeys", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var listResponse dto.APIResponse[dto.ListAPIKeysResponse]
	json.Unmarshal(w.Body.Bytes(), &listResponse)
	assert.Len(t, listResponse.Data.APIKeys, 3)

	// Revoke the write key
	_, err := apiKeyUC.RevokeAPIKey(context.Background(), createdKeys["write"], "test-user", nil)
	require.NoError(t, err)

	// Verify active keys (should be 2)
	req2 := httptest.NewRequest(http.MethodGet, "/api/apikeys?status=active", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	var activeResponse dto.APIResponse[dto.ListAPIKeysResponse]
	json.Unmarshal(w2.Body.Bytes(), &activeResponse)
	assert.Len(t, activeResponse.Data.APIKeys, 2)

	// Verify read role filter
	req3 := httptest.NewRequest(http.MethodGet, "/api/apikeys?role=read", nil)
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)

	var readResponse dto.APIResponse[dto.ListAPIKeysResponse]
	json.Unmarshal(w3.Body.Bytes(), &readResponse)
	assert.Len(t, readResponse.Data.APIKeys, 1)
	assert.Equal(t, "read", readResponse.Data.APIKeys[0].Role)
}

// ==================== Usage Statistics Tests ====================

func TestGetAPIKeyDetails_WithUsageStatistics(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	defer db.Exec("DELETE FROM api_keys WHERE 1=1")
	defer db.Exec("DELETE FROM audit_logs WHERE 1=1")

	// Run migrations
	err = persistence.RunAutoMigration(db)
	require.NoError(t, err)

	repo := persistence.NewAPIKeyRepository(db)
	auditLogRepo := persistence.NewAuditLogRepository(db)
	auditLogger := &AuditLoggerMock{}
	uc := usecase.NewAPIKeyUseCase(repo, auditLogger, auditLogRepo)

	// Create an API key
	description := "Test key"
	_, apiKey, err := uc.CreateAPIKey(context.Background(), "read", &description, "admin@example.com")
	require.NoError(t, err)

	// Simulate API key usage by inserting audit log entries
	for i := range 10 {
		err := auditLogRepo.SaveAPIKeyUsage(context.Background(), repository.APIKeyUsageEvent{
			APIKeyID:  apiKey.ID,
			Endpoint:  "/api/sessions",
			Method:    "GET",
			IPAddress: "127.0.0.1",
			Timestamp: time.Now().Add(-time.Duration(i) * time.Hour),
		})
		require.NoError(t, err)
	}

	// Get API key details
	details, totalRequests, last7Days, err := uc.GetAPIKeyDetails(context.Background(), apiKey.ID)

	// Verify
	require.NoError(t, err)
	require.NotNil(t, details)
	assert.Equal(t, apiKey.ID, details.ID)
	assert.Equal(t, 10, totalRequests, "Should count all 10 usage events")
	assert.Equal(t, 10, last7Days, "All events are within last 7 days")
}

func TestGetAPIKeyDetails_UsageStatisticsLast7Days(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	defer db.Exec("DELETE FROM api_keys WHERE 1=1")
	defer db.Exec("DELETE FROM audit_logs WHERE 1=1")

	// Run migrations
	err = persistence.RunAutoMigration(db)
	require.NoError(t, err)

	repo := persistence.NewAPIKeyRepository(db)
	auditLogRepo := persistence.NewAuditLogRepository(db)
	auditLogger := &AuditLoggerMock{}
	uc := usecase.NewAPIKeyUseCase(repo, auditLogger, auditLogRepo)

	// Create an API key
	description := "Test key"
	_, apiKey, err := uc.CreateAPIKey(context.Background(), "read", &description, "admin@example.com")
	require.NoError(t, err)

	// Simulate API key usage - 5 recent, 5 old
	for i := range 5 {
		// Recent usage (within last 7 days)
		err := auditLogRepo.SaveAPIKeyUsage(context.Background(), repository.APIKeyUsageEvent{
			APIKeyID:  apiKey.ID,
			Endpoint:  "/api/sessions",
			Method:    "GET",
			IPAddress: "127.0.0.1",
			Timestamp: time.Now().Add(-time.Duration(i) * time.Hour),
		})
		require.NoError(t, err)

		// Old usage (more than 7 days ago)
		err = auditLogRepo.SaveAPIKeyUsage(context.Background(), repository.APIKeyUsageEvent{
			APIKeyID:  apiKey.ID,
			Endpoint:  "/api/sessions",
			Method:    "GET",
			IPAddress: "127.0.0.1",
			Timestamp: time.Now().Add(-time.Duration(8+i) * 24 * time.Hour),
		})
		require.NoError(t, err)
	}

	// Get API key details
	details, totalRequests, last7Days, err := uc.GetAPIKeyDetails(context.Background(), apiKey.ID)

	// Verify
	require.NoError(t, err)
	require.NotNil(t, details)
	assert.Equal(t, apiKey.ID, details.ID)
	assert.Equal(t, 10, totalRequests, "Should count all 10 usage events")
	assert.Equal(t, 5, last7Days, "Should only count 5 recent events")
}

func TestGetAPIKeyDetails_NoUsageStatistics(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	defer db.Exec("DELETE FROM api_keys WHERE 1=1")
	defer db.Exec("DELETE FROM audit_logs WHERE 1=1")

	// Run migrations
	err = persistence.RunAutoMigration(db)
	require.NoError(t, err)

	repo := persistence.NewAPIKeyRepository(db)
	auditLogRepo := persistence.NewAuditLogRepository(db)
	auditLogger := &AuditLoggerMock{}
	uc := usecase.NewAPIKeyUseCase(repo, auditLogger, auditLogRepo)

	// Create an API key
	description := "Test key"
	_, apiKey, err := uc.CreateAPIKey(context.Background(), "read", &description, "admin@example.com")
	require.NoError(t, err)

	// Don't create any usage events

	// Get API key details
	details, totalRequests, last7Days, err := uc.GetAPIKeyDetails(context.Background(), apiKey.ID)

	// Verify
	require.NoError(t, err)
	require.NotNil(t, details)
	assert.Equal(t, apiKey.ID, details.ID)
	assert.Equal(t, 0, totalRequests, "Should return 0 when no usage")
	assert.Equal(t, 0, last7Days, "Should return 0 when no usage")
}
