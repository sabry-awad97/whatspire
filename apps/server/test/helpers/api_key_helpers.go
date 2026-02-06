package helpers

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"whatspire/internal/domain/entity"
	"whatspire/internal/domain/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestAPIKey represents an API key with its plain-text value for testing
type TestAPIKey struct {
	Entity    *entity.APIKey
	PlainText string // The actual API key value (not stored in database)
}

// GenerateTestAPIKey creates a new API key for testing with a random key value
func GenerateTestAPIKey(role string, description *string) *TestAPIKey {
	// Generate a random API key
	keyBytes := make([]byte, 32)
	_, _ = rand.Read(keyBytes)
	plainTextKey := fmt.Sprintf("key_%s", hex.EncodeToString(keyBytes))

	// Hash the key
	hash := sha256.Sum256([]byte(plainTextKey))
	keyHash := fmt.Sprintf("%x", hash)

	// Create the entity
	id := fmt.Sprintf("key_%s", uuid.New().String())
	apiKey := entity.NewAPIKey(id, keyHash, role, description)

	return &TestAPIKey{
		Entity:    apiKey,
		PlainText: plainTextKey,
	}
}

// CreateTestAPIKey creates and saves an API key to the repository for testing
func CreateTestAPIKey(t *testing.T, repo repository.APIKeyRepository, role string, description *string) *TestAPIKey {
	testKey := GenerateTestAPIKey(role, description)

	err := repo.Save(context.Background(), testKey.Entity)
	require.NoError(t, err, "Failed to save test API key")

	return testKey
}

// CreateTestAPIKeyWithPlainText creates an API key with a specific plain-text value
func CreateTestAPIKeyWithPlainText(t *testing.T, repo repository.APIKeyRepository, plainTextKey, role string, description *string) *TestAPIKey {
	// Hash the key
	hash := sha256.Sum256([]byte(plainTextKey))
	keyHash := fmt.Sprintf("%x", hash)

	// Create the entity
	id := fmt.Sprintf("key_%s", uuid.New().String())
	apiKey := entity.NewAPIKey(id, keyHash, role, description)

	err := repo.Save(context.Background(), apiKey)
	require.NoError(t, err, "Failed to save test API key")

	return &TestAPIKey{
		Entity:    apiKey,
		PlainText: plainTextKey,
	}
}

// CreateRevokedTestAPIKey creates a revoked API key for testing
func CreateRevokedTestAPIKey(t *testing.T, repo repository.APIKeyRepository, role string) *TestAPIKey {
	testKey := GenerateTestAPIKey(role, nil)

	// Revoke the key
	reason := "Test revocation"
	testKey.Entity.Revoke("test-admin", &reason)

	err := repo.Save(context.Background(), testKey.Entity)
	require.NoError(t, err, "Failed to save revoked test API key")

	return testKey
}

// CreateInactiveTestAPIKey creates an inactive API key for testing
func CreateInactiveTestAPIKey(t *testing.T, repo repository.APIKeyRepository, role string) *TestAPIKey {
	testKey := GenerateTestAPIKey(role, nil)

	// Deactivate the key
	testKey.Entity.Deactivate()

	err := repo.Save(context.Background(), testKey.Entity)
	require.NoError(t, err, "Failed to save inactive test API key")

	return testKey
}

// UpdateTestAPIKeyLastUsed updates the last used timestamp for a test API key
func UpdateTestAPIKeyLastUsed(t *testing.T, repo repository.APIKeyRepository, testKey *TestAPIKey) {
	hash := sha256.Sum256([]byte(testKey.PlainText))
	keyHash := fmt.Sprintf("%x", hash)

	err := repo.UpdateLastUsed(context.Background(), keyHash)
	require.NoError(t, err, "Failed to update last used timestamp")

	// Refresh the entity from database
	updated, err := repo.FindByID(context.Background(), testKey.Entity.ID)
	require.NoError(t, err, "Failed to refresh API key entity")
	testKey.Entity = updated
}

// MockAPIKeyRepository is a simple in-memory implementation for testing
type MockAPIKeyRepository struct {
	keys map[string]*entity.APIKey // keyHash -> APIKey
}

// NewMockAPIKeyRepository creates a new mock API key repository
func NewMockAPIKeyRepository() *MockAPIKeyRepository {
	return &MockAPIKeyRepository{
		keys: make(map[string]*entity.APIKey),
	}
}

// Save stores an API key
func (m *MockAPIKeyRepository) Save(ctx context.Context, apiKey *entity.APIKey) error {
	m.keys[apiKey.KeyHash] = apiKey
	return nil
}

// FindByKeyHash retrieves an API key by its hash
func (m *MockAPIKeyRepository) FindByKeyHash(ctx context.Context, keyHash string) (*entity.APIKey, error) {
	if key, exists := m.keys[keyHash]; exists {
		return key, nil
	}
	return nil, fmt.Errorf("API key not found")
}

// FindByID retrieves an API key by its ID
func (m *MockAPIKeyRepository) FindByID(ctx context.Context, id string) (*entity.APIKey, error) {
	for _, key := range m.keys {
		if key.ID == id {
			return key, nil
		}
	}
	return nil, fmt.Errorf("API key not found")
}

// UpdateLastUsed updates the last used timestamp
func (m *MockAPIKeyRepository) UpdateLastUsed(ctx context.Context, keyHash string) error {
	if key, exists := m.keys[keyHash]; exists {
		now := time.Now()
		key.LastUsedAt = &now
		return nil
	}
	return fmt.Errorf("API key not found")
}

// Update updates an existing API key
func (m *MockAPIKeyRepository) Update(ctx context.Context, apiKey *entity.APIKey) error {
	if _, exists := m.keys[apiKey.KeyHash]; exists {
		m.keys[apiKey.KeyHash] = apiKey
		return nil
	}
	return fmt.Errorf("API key not found")
}

// Delete removes an API key
func (m *MockAPIKeyRepository) Delete(ctx context.Context, id string) error {
	for hash, key := range m.keys {
		if key.ID == id {
			delete(m.keys, hash)
			return nil
		}
	}
	return fmt.Errorf("API key not found")
}

// List retrieves all API keys with filters
func (m *MockAPIKeyRepository) List(ctx context.Context, limit, offset int, role *string, isActive *bool) ([]*entity.APIKey, error) {
	var result []*entity.APIKey
	for _, key := range m.keys {
		// Apply filters
		if role != nil && key.Role != *role {
			continue
		}
		if isActive != nil && key.IsActive != *isActive {
			continue
		}
		result = append(result, key)
	}

	// Apply pagination
	if offset >= len(result) {
		return []*entity.APIKey{}, nil
	}
	end := min(offset+limit, len(result))
	return result[offset:end], nil
}

// Count returns the total number of API keys matching filters
func (m *MockAPIKeyRepository) Count(ctx context.Context, role *string, isActive *bool) (int64, error) {
	count := int64(0)
	for _, key := range m.keys {
		// Apply filters
		if role != nil && key.Role != *role {
			continue
		}
		if isActive != nil && key.IsActive != *isActive {
			continue
		}
		count++
	}
	return count, nil
}
