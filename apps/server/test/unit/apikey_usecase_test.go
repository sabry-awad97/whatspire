package unit

import (
	"context"
	"testing"

	"whatspire/internal/application/usecase"
	"whatspire/internal/domain/entity"
	"whatspire/internal/domain/errors"
	"whatspire/internal/domain/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== Mock APIKeyRepository ====================

type APIKeyRepositoryMock struct {
	SaveFn           func(ctx context.Context, apiKey *entity.APIKey) error
	FindByKeyHashFn  func(ctx context.Context, keyHash string) (*entity.APIKey, error)
	FindByIDFn       func(ctx context.Context, id string) (*entity.APIKey, error)
	UpdateLastUsedFn func(ctx context.Context, keyHash string) error
	DeleteFn         func(ctx context.Context, id string) error
	ListFn           func(ctx context.Context, limit, offset int, role *string, isActive *bool) ([]*entity.APIKey, error)
	UpdateFn         func(ctx context.Context, apiKey *entity.APIKey) error
	CountFn          func(ctx context.Context, role *string, isActive *bool) (int64, error)
	APIKeys          map[string]*entity.APIKey // In-memory storage for testing
}

func NewAPIKeyRepositoryMock() *APIKeyRepositoryMock {
	return &APIKeyRepositoryMock{
		APIKeys: make(map[string]*entity.APIKey),
	}
}

func (m *APIKeyRepositoryMock) Save(ctx context.Context, apiKey *entity.APIKey) error {
	if m.SaveFn != nil {
		return m.SaveFn(ctx, apiKey)
	}
	m.APIKeys[apiKey.ID] = apiKey
	return nil
}

func (m *APIKeyRepositoryMock) FindByKeyHash(ctx context.Context, keyHash string) (*entity.APIKey, error) {
	if m.FindByKeyHashFn != nil {
		return m.FindByKeyHashFn(ctx, keyHash)
	}
	for _, apiKey := range m.APIKeys {
		if apiKey.KeyHash == keyHash {
			return apiKey, nil
		}
	}
	return nil, errors.ErrNotFound
}

func (m *APIKeyRepositoryMock) FindByID(ctx context.Context, id string) (*entity.APIKey, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(ctx, id)
	}
	apiKey, exists := m.APIKeys[id]
	if !exists {
		return nil, errors.ErrNotFound
	}
	return apiKey, nil
}

func (m *APIKeyRepositoryMock) UpdateLastUsed(ctx context.Context, keyHash string) error {
	if m.UpdateLastUsedFn != nil {
		return m.UpdateLastUsedFn(ctx, keyHash)
	}
	for _, apiKey := range m.APIKeys {
		if apiKey.KeyHash == keyHash {
			apiKey.UpdateLastUsed()
			return nil
		}
	}
	return errors.ErrNotFound
}

func (m *APIKeyRepositoryMock) Delete(ctx context.Context, id string) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, id)
	}
	if _, exists := m.APIKeys[id]; !exists {
		return errors.ErrNotFound
	}
	delete(m.APIKeys, id)
	return nil
}

func (m *APIKeyRepositoryMock) List(ctx context.Context, limit, offset int, role *string, isActive *bool) ([]*entity.APIKey, error) {
	if m.ListFn != nil {
		return m.ListFn(ctx, limit, offset, role, isActive)
	}

	// Collect all keys first (maps don't preserve order, but for testing this is fine)
	var allKeys []*entity.APIKey
	for _, apiKey := range m.APIKeys {
		allKeys = append(allKeys, apiKey)
	}

	// Apply filters
	var result []*entity.APIKey
	for _, apiKey := range allKeys {
		if role != nil && apiKey.Role != *role {
			continue
		}
		if isActive != nil && apiKey.IsActive != *isActive {
			continue
		}
		result = append(result, apiKey)
	}

	// Apply pagination
	if offset >= len(result) {
		return []*entity.APIKey{}, nil
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end], nil
}

func (m *APIKeyRepositoryMock) Update(ctx context.Context, apiKey *entity.APIKey) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, apiKey)
	}
	if _, exists := m.APIKeys[apiKey.ID]; !exists {
		return errors.ErrNotFound
	}
	m.APIKeys[apiKey.ID] = apiKey
	return nil
}

func (m *APIKeyRepositoryMock) Count(ctx context.Context, role *string, isActive *bool) (int64, error) {
	if m.CountFn != nil {
		return m.CountFn(ctx, role, isActive)
	}

	count := int64(0)
	for _, apiKey := range m.APIKeys {
		// Apply filters
		if role != nil && apiKey.Role != *role {
			continue
		}
		if isActive != nil && apiKey.IsActive != *isActive {
			continue
		}
		count++
	}
	return count, nil
}

// ==================== Mock AuditLogger ====================

type AuditLoggerMock struct {
	LogAPIKeyUsageFn     func(ctx context.Context, event repository.APIKeyUsageEvent)
	LogAPIKeyCreatedFn   func(ctx context.Context, event repository.APIKeyCreatedEvent)
	LogAPIKeyRevokedFn   func(ctx context.Context, event repository.APIKeyRevokedEvent)
	LogSessionActionFn   func(ctx context.Context, event repository.SessionActionEvent)
	LogMessageSentFn     func(ctx context.Context, event repository.MessageSentEvent)
	LogAuthFailureFn     func(ctx context.Context, event repository.AuthFailureEvent)
	LogWebhookDeliveryFn func(ctx context.Context, event repository.WebhookDeliveryEvent)
	CreatedEvents        []repository.APIKeyCreatedEvent
	RevokedEvents        []repository.APIKeyRevokedEvent
}

func NewAuditLoggerMock() *AuditLoggerMock {
	return &AuditLoggerMock{
		CreatedEvents: []repository.APIKeyCreatedEvent{},
		RevokedEvents: []repository.APIKeyRevokedEvent{},
	}
}

func (m *AuditLoggerMock) LogAPIKeyUsage(ctx context.Context, event repository.APIKeyUsageEvent) {
	if m.LogAPIKeyUsageFn != nil {
		m.LogAPIKeyUsageFn(ctx, event)
	}
}

func (m *AuditLoggerMock) LogAPIKeyCreated(ctx context.Context, event repository.APIKeyCreatedEvent) {
	if m.LogAPIKeyCreatedFn != nil {
		m.LogAPIKeyCreatedFn(ctx, event)
	}
	m.CreatedEvents = append(m.CreatedEvents, event)
}

func (m *AuditLoggerMock) LogAPIKeyRevoked(ctx context.Context, event repository.APIKeyRevokedEvent) {
	if m.LogAPIKeyRevokedFn != nil {
		m.LogAPIKeyRevokedFn(ctx, event)
	}
	m.RevokedEvents = append(m.RevokedEvents, event)
}

func (m *AuditLoggerMock) LogSessionAction(ctx context.Context, event repository.SessionActionEvent) {
	if m.LogSessionActionFn != nil {
		m.LogSessionActionFn(ctx, event)
	}
}

func (m *AuditLoggerMock) LogMessageSent(ctx context.Context, event repository.MessageSentEvent) {
	if m.LogMessageSentFn != nil {
		m.LogMessageSentFn(ctx, event)
	}
}

func (m *AuditLoggerMock) LogAuthFailure(ctx context.Context, event repository.AuthFailureEvent) {
	if m.LogAuthFailureFn != nil {
		m.LogAuthFailureFn(ctx, event)
	}
}

func (m *AuditLoggerMock) LogWebhookDelivery(ctx context.Context, event repository.WebhookDeliveryEvent) {
	if m.LogWebhookDeliveryFn != nil {
		m.LogWebhookDeliveryFn(ctx, event)
	}
}

// ==================== CreateAPIKey Tests ====================

func TestAPIKeyUseCase_CreateAPIKey_Success(t *testing.T) {
	repo := NewAPIKeyRepositoryMock()
	auditLogger := NewAuditLoggerMock()
	uc := usecase.NewAPIKeyUseCase(repo, auditLogger)

	description := "Test API Key"
	plainKey, apiKey, err := uc.CreateAPIKey(context.Background(), "read", &description, "admin@example.com")

	require.NoError(t, err)
	require.NotNil(t, apiKey)
	assert.NotEmpty(t, plainKey)
	assert.Equal(t, "read", apiKey.Role)
	assert.Equal(t, &description, apiKey.Description)
	assert.True(t, apiKey.IsActive)
	assert.False(t, apiKey.IsRevoked())

	// Verify key was saved
	assert.Len(t, repo.APIKeys, 1)

	// Verify audit log
	assert.Len(t, auditLogger.CreatedEvents, 1)
	assert.Equal(t, apiKey.ID, auditLogger.CreatedEvents[0].APIKeyID)
	assert.Equal(t, "read", auditLogger.CreatedEvents[0].Role)
}

func TestAPIKeyUseCase_CreateAPIKey_InvalidRole(t *testing.T) {
	repo := NewAPIKeyRepositoryMock()
	auditLogger := NewAuditLoggerMock()
	uc := usecase.NewAPIKeyUseCase(repo, auditLogger)

	plainKey, apiKey, err := uc.CreateAPIKey(context.Background(), "invalid", nil, "admin@example.com")

	assert.Nil(t, apiKey)
	assert.Empty(t, plainKey)
	assert.ErrorIs(t, err, errors.ErrValidationFailed)
	assert.Contains(t, err.Error(), "invalid role")

	// Verify nothing was saved
	assert.Len(t, repo.APIKeys, 0)
	assert.Len(t, auditLogger.CreatedEvents, 0)
}

func TestAPIKeyUseCase_CreateAPIKey_AllRoles(t *testing.T) {
	repo := NewAPIKeyRepositoryMock()
	auditLogger := NewAuditLoggerMock()
	uc := usecase.NewAPIKeyUseCase(repo, auditLogger)

	roles := []string{"read", "write", "admin"}

	for _, role := range roles {
		plainKey, apiKey, err := uc.CreateAPIKey(context.Background(), role, nil, "admin@example.com")

		require.NoError(t, err, "Role %s should be valid", role)
		require.NotNil(t, apiKey)
		assert.NotEmpty(t, plainKey)
		assert.Equal(t, role, apiKey.Role)
	}

	assert.Len(t, repo.APIKeys, 3)
}

func TestAPIKeyUseCase_CreateAPIKey_NilDescription(t *testing.T) {
	repo := NewAPIKeyRepositoryMock()
	auditLogger := NewAuditLoggerMock()
	uc := usecase.NewAPIKeyUseCase(repo, auditLogger)

	plainKey, apiKey, err := uc.CreateAPIKey(context.Background(), "write", nil, "admin@example.com")

	require.NoError(t, err)
	require.NotNil(t, apiKey)
	assert.NotEmpty(t, plainKey)
	assert.Nil(t, apiKey.Description)
}

func TestAPIKeyUseCase_CreateAPIKey_RepositoryError(t *testing.T) {
	repo := NewAPIKeyRepositoryMock()
	repo.SaveFn = func(ctx context.Context, apiKey *entity.APIKey) error {
		return errors.ErrDatabaseError
	}
	auditLogger := NewAuditLoggerMock()
	uc := usecase.NewAPIKeyUseCase(repo, auditLogger)

	plainKey, apiKey, err := uc.CreateAPIKey(context.Background(), "read", nil, "admin@example.com")

	assert.Nil(t, apiKey)
	assert.Empty(t, plainKey)
	assert.ErrorIs(t, err, errors.ErrDatabaseError)
}

func TestAPIKeyUseCase_CreateAPIKey_UniqueKeys(t *testing.T) {
	repo := NewAPIKeyRepositoryMock()
	auditLogger := NewAuditLoggerMock()
	uc := usecase.NewAPIKeyUseCase(repo, auditLogger)

	// Create multiple keys
	keys := make(map[string]bool)
	for i := 0; i < 10; i++ {
		plainKey, _, err := uc.CreateAPIKey(context.Background(), "read", nil, "admin@example.com")
		require.NoError(t, err)

		// Verify key is unique
		assert.False(t, keys[plainKey], "Generated key should be unique")
		keys[plainKey] = true
	}
}

// ==================== RevokeAPIKey Tests ====================

func TestAPIKeyUseCase_RevokeAPIKey_Success(t *testing.T) {
	repo := NewAPIKeyRepositoryMock()
	auditLogger := NewAuditLoggerMock()
	uc := usecase.NewAPIKeyUseCase(repo, auditLogger)

	// Create an API key first
	_, apiKey, err := uc.CreateAPIKey(context.Background(), "read", nil, "admin@example.com")
	require.NoError(t, err)

	// Revoke it
	reason := "Security breach"
	revokedKey, err := uc.RevokeAPIKey(context.Background(), apiKey.ID, "admin2@example.com", &reason)

	require.NoError(t, err)
	require.NotNil(t, revokedKey)
	assert.True(t, revokedKey.IsRevoked())
	assert.False(t, revokedKey.IsActive)
	assert.Equal(t, "admin2@example.com", *revokedKey.RevokedBy)
	assert.Equal(t, reason, *revokedKey.RevocationReason)

	// Verify audit log
	assert.Len(t, auditLogger.RevokedEvents, 1)
	assert.Equal(t, apiKey.ID, auditLogger.RevokedEvents[0].APIKeyID)
	assert.Equal(t, "admin2@example.com", auditLogger.RevokedEvents[0].RevokedBy)
}

func TestAPIKeyUseCase_RevokeAPIKey_NotFound(t *testing.T) {
	repo := NewAPIKeyRepositoryMock()
	auditLogger := NewAuditLoggerMock()
	uc := usecase.NewAPIKeyUseCase(repo, auditLogger)

	revokedKey, err := uc.RevokeAPIKey(context.Background(), "non-existent", "admin@example.com", nil)

	assert.Nil(t, revokedKey)
	assert.ErrorIs(t, err, errors.ErrNotFound)
	assert.Len(t, auditLogger.RevokedEvents, 0)
}

func TestAPIKeyUseCase_RevokeAPIKey_AlreadyRevoked(t *testing.T) {
	repo := NewAPIKeyRepositoryMock()
	auditLogger := NewAuditLoggerMock()
	uc := usecase.NewAPIKeyUseCase(repo, auditLogger)

	// Create and revoke an API key
	_, apiKey, err := uc.CreateAPIKey(context.Background(), "read", nil, "admin@example.com")
	require.NoError(t, err)

	_, err = uc.RevokeAPIKey(context.Background(), apiKey.ID, "admin@example.com", nil)
	require.NoError(t, err)

	// Try to revoke again
	revokedKey, err := uc.RevokeAPIKey(context.Background(), apiKey.ID, "admin2@example.com", nil)

	assert.Nil(t, revokedKey)
	assert.ErrorIs(t, err, errors.ErrAlreadyRevoked)
	assert.Contains(t, err.Error(), "already revoked")

	// Should only have one revocation event
	assert.Len(t, auditLogger.RevokedEvents, 1)
}

func TestAPIKeyUseCase_RevokeAPIKey_NilReason(t *testing.T) {
	repo := NewAPIKeyRepositoryMock()
	auditLogger := NewAuditLoggerMock()
	uc := usecase.NewAPIKeyUseCase(repo, auditLogger)

	_, apiKey, err := uc.CreateAPIKey(context.Background(), "read", nil, "admin@example.com")
	require.NoError(t, err)

	revokedKey, err := uc.RevokeAPIKey(context.Background(), apiKey.ID, "admin@example.com", nil)

	require.NoError(t, err)
	require.NotNil(t, revokedKey)
	assert.True(t, revokedKey.IsRevoked())
	assert.Nil(t, revokedKey.RevocationReason)
}

func TestAPIKeyUseCase_RevokeAPIKey_RepositoryError(t *testing.T) {
	repo := NewAPIKeyRepositoryMock()
	auditLogger := NewAuditLoggerMock()
	uc := usecase.NewAPIKeyUseCase(repo, auditLogger)

	_, apiKey, err := uc.CreateAPIKey(context.Background(), "read", nil, "admin@example.com")
	require.NoError(t, err)

	// Make Update fail
	repo.UpdateFn = func(ctx context.Context, apiKey *entity.APIKey) error {
		return errors.ErrDatabaseError
	}

	revokedKey, err := uc.RevokeAPIKey(context.Background(), apiKey.ID, "admin@example.com", nil)

	assert.Nil(t, revokedKey)
	assert.ErrorIs(t, err, errors.ErrDatabaseError)
}

// ==================== ListAPIKeys Tests ====================

func TestAPIKeyUseCase_ListAPIKeys_Success(t *testing.T) {
	repo := NewAPIKeyRepositoryMock()
	auditLogger := NewAuditLoggerMock()
	uc := usecase.NewAPIKeyUseCase(repo, auditLogger)

	// Create multiple API keys
	for i := 0; i < 5; i++ {
		_, _, err := uc.CreateAPIKey(context.Background(), "read", nil, "admin@example.com")
		require.NoError(t, err)
	}

	// List all keys
	keys, total, err := uc.ListAPIKeys(context.Background(), 1, 10, nil, nil)

	require.NoError(t, err)
	assert.Len(t, keys, 5)
	assert.Equal(t, int64(5), total)
}

func TestAPIKeyUseCase_ListAPIKeys_Pagination(t *testing.T) {
	repo := NewAPIKeyRepositoryMock()
	auditLogger := NewAuditLoggerMock()
	uc := usecase.NewAPIKeyUseCase(repo, auditLogger)

	// Create 10 API keys
	for i := 0; i < 10; i++ {
		_, _, err := uc.CreateAPIKey(context.Background(), "read", nil, "admin@example.com")
		require.NoError(t, err)
	}

	// Get first page (5 items)
	keys, total, err := uc.ListAPIKeys(context.Background(), 1, 5, nil, nil)
	require.NoError(t, err)
	assert.Len(t, keys, 5)
	assert.Equal(t, int64(10), total)

	// Get second page (5 items)
	keys, total, err = uc.ListAPIKeys(context.Background(), 2, 5, nil, nil)
	require.NoError(t, err)
	assert.Len(t, keys, 5)
	assert.Equal(t, int64(10), total)

	// Get third page (0 items)
	keys, total, err = uc.ListAPIKeys(context.Background(), 3, 5, nil, nil)
	require.NoError(t, err)
	assert.Len(t, keys, 0)
	assert.Equal(t, int64(10), total)
}

func TestAPIKeyUseCase_ListAPIKeys_FilterByRole(t *testing.T) {
	repo := NewAPIKeyRepositoryMock()
	auditLogger := NewAuditLoggerMock()
	uc := usecase.NewAPIKeyUseCase(repo, auditLogger)

	// Create keys with different roles
	_, _, err := uc.CreateAPIKey(context.Background(), "read", nil, "admin@example.com")
	require.NoError(t, err)
	_, _, err = uc.CreateAPIKey(context.Background(), "write", nil, "admin@example.com")
	require.NoError(t, err)
	_, _, err = uc.CreateAPIKey(context.Background(), "admin", nil, "admin@example.com")
	require.NoError(t, err)

	// Filter by read role
	readRole := "read"
	keys, total, err := uc.ListAPIKeys(context.Background(), 1, 10, &readRole, nil)
	require.NoError(t, err)
	assert.Len(t, keys, 1)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "read", keys[0].Role)
}

func TestAPIKeyUseCase_ListAPIKeys_FilterByStatus(t *testing.T) {
	repo := NewAPIKeyRepositoryMock()
	auditLogger := NewAuditLoggerMock()
	uc := usecase.NewAPIKeyUseCase(repo, auditLogger)

	// Create keys
	_, key1, err := uc.CreateAPIKey(context.Background(), "read", nil, "admin@example.com")
	require.NoError(t, err)
	_, _, err = uc.CreateAPIKey(context.Background(), "write", nil, "admin@example.com")
	require.NoError(t, err)

	// Revoke one key
	_, err = uc.RevokeAPIKey(context.Background(), key1.ID, "admin@example.com", nil)
	require.NoError(t, err)

	// Filter by active status
	activeStatus := "active"
	keys, total, err := uc.ListAPIKeys(context.Background(), 1, 10, nil, &activeStatus)
	require.NoError(t, err)
	assert.Len(t, keys, 1)
	assert.Equal(t, int64(1), total)
	assert.True(t, keys[0].IsActive)

	// Filter by revoked status
	revokedStatus := "revoked"
	keys, total, err = uc.ListAPIKeys(context.Background(), 1, 10, nil, &revokedStatus)
	require.NoError(t, err)
	assert.Len(t, keys, 1)
	assert.Equal(t, int64(1), total)
	assert.False(t, keys[0].IsActive)
}

func TestAPIKeyUseCase_ListAPIKeys_InvalidRole(t *testing.T) {
	repo := NewAPIKeyRepositoryMock()
	auditLogger := NewAuditLoggerMock()
	uc := usecase.NewAPIKeyUseCase(repo, auditLogger)

	invalidRole := "invalid"
	keys, total, err := uc.ListAPIKeys(context.Background(), 1, 10, &invalidRole, nil)

	assert.Nil(t, keys)
	assert.Equal(t, int64(0), total)
	assert.ErrorIs(t, err, errors.ErrValidationFailed)
	assert.Contains(t, err.Error(), "invalid role filter")
}

func TestAPIKeyUseCase_ListAPIKeys_InvalidStatus(t *testing.T) {
	repo := NewAPIKeyRepositoryMock()
	auditLogger := NewAuditLoggerMock()
	uc := usecase.NewAPIKeyUseCase(repo, auditLogger)

	invalidStatus := "invalid"
	keys, total, err := uc.ListAPIKeys(context.Background(), 1, 10, nil, &invalidStatus)

	assert.Nil(t, keys)
	assert.Equal(t, int64(0), total)
	assert.ErrorIs(t, err, errors.ErrValidationFailed)
	assert.Contains(t, err.Error(), "invalid status filter")
}

func TestAPIKeyUseCase_ListAPIKeys_DefaultPagination(t *testing.T) {
	repo := NewAPIKeyRepositoryMock()
	auditLogger := NewAuditLoggerMock()
	uc := usecase.NewAPIKeyUseCase(repo, auditLogger)

	// Create 5 keys
	for range 5 {
		_, _, err := uc.CreateAPIKey(context.Background(), "read", nil, "admin@example.com")
		require.NoError(t, err)
	}

	// Invalid page/limit should use defaults
	keys, total, err := uc.ListAPIKeys(context.Background(), 0, 0, nil, nil)
	require.NoError(t, err)
	assert.Len(t, keys, 5)
	assert.Equal(t, int64(5), total)
}

func TestAPIKeyUseCase_ListAPIKeys_MaxLimit(t *testing.T) {
	repo := NewAPIKeyRepositoryMock()
	auditLogger := NewAuditLoggerMock()
	uc := usecase.NewAPIKeyUseCase(repo, auditLogger)

	// Create 5 keys
	for range 5 {
		_, _, err := uc.CreateAPIKey(context.Background(), "read", nil, "admin@example.com")
		require.NoError(t, err)
	}

	// Limit > 100 should be capped to 50 (default)
	keys, total, err := uc.ListAPIKeys(context.Background(), 1, 200, nil, nil)
	require.NoError(t, err)
	assert.Len(t, keys, 5)
	assert.Equal(t, int64(5), total)
}

// ==================== GetAPIKeyDetails Tests ====================

func TestAPIKeyUseCase_GetAPIKeyDetails_Success(t *testing.T) {
	repo := NewAPIKeyRepositoryMock()
	auditLogger := NewAuditLoggerMock()
	uc := usecase.NewAPIKeyUseCase(repo, auditLogger)

	// Create an API key
	_, apiKey, err := uc.CreateAPIKey(context.Background(), "read", nil, "admin@example.com")
	require.NoError(t, err)

	// Get details
	details, totalRequests, last7Days, err := uc.GetAPIKeyDetails(context.Background(), apiKey.ID)

	require.NoError(t, err)
	require.NotNil(t, details)
	assert.Equal(t, apiKey.ID, details.ID)
	assert.Equal(t, 0, totalRequests) // Placeholder value
	assert.Equal(t, 0, last7Days)     // Placeholder value
}

func TestAPIKeyUseCase_GetAPIKeyDetails_NotFound(t *testing.T) {
	repo := NewAPIKeyRepositoryMock()
	auditLogger := NewAuditLoggerMock()
	uc := usecase.NewAPIKeyUseCase(repo, auditLogger)

	details, totalRequests, last7Days, err := uc.GetAPIKeyDetails(context.Background(), "non-existent")

	assert.Nil(t, details)
	assert.Equal(t, 0, totalRequests)
	assert.Equal(t, 0, last7Days)
	assert.ErrorIs(t, err, errors.ErrNotFound)
}

func TestAPIKeyUseCase_GetAPIKeyDetails_RevokedKey(t *testing.T) {
	repo := NewAPIKeyRepositoryMock()
	auditLogger := NewAuditLoggerMock()
	uc := usecase.NewAPIKeyUseCase(repo, auditLogger)

	// Create and revoke an API key
	_, apiKey, err := uc.CreateAPIKey(context.Background(), "read", nil, "admin@example.com")
	require.NoError(t, err)

	reason := "Test revocation"
	_, err = uc.RevokeAPIKey(context.Background(), apiKey.ID, "admin@example.com", &reason)
	require.NoError(t, err)

	// Get details of revoked key
	details, _, _, err := uc.GetAPIKeyDetails(context.Background(), apiKey.ID)

	require.NoError(t, err)
	require.NotNil(t, details)
	assert.True(t, details.IsRevoked())
	assert.Equal(t, reason, *details.RevocationReason)
}
