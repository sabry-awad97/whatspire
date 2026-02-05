package unit

import (
	"encoding/json"
	"testing"
	"time"

	"whatspire/internal/domain/entity"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== APIKey Entity Tests ====================

func TestNewAPIKey(t *testing.T) {
	id := "key_123"
	keyHash := "abc123hash"
	role := "read"
	description := "Test API Key"

	apiKey := entity.NewAPIKey(id, keyHash, role, &description)

	require.NotNil(t, apiKey)
	assert.Equal(t, id, apiKey.ID)
	assert.Equal(t, keyHash, apiKey.KeyHash)
	assert.Equal(t, role, apiKey.Role)
	assert.Equal(t, &description, apiKey.Description)
	assert.True(t, apiKey.IsActive)
	assert.Nil(t, apiKey.LastUsedAt)
	assert.Nil(t, apiKey.RevokedAt)
	assert.Nil(t, apiKey.RevokedBy)
	assert.Nil(t, apiKey.RevocationReason)
	assert.False(t, apiKey.CreatedAt.IsZero())
}

func TestNewAPIKey_NilDescription(t *testing.T) {
	apiKey := entity.NewAPIKey("key_123", "abc123hash", "write", nil)

	require.NotNil(t, apiKey)
	assert.Nil(t, apiKey.Description)
}

// ==================== UpdateLastUsed Tests ====================

func TestAPIKey_UpdateLastUsed(t *testing.T) {
	apiKey := entity.NewAPIKey("key_123", "abc123hash", "read", nil)

	// Initially LastUsedAt should be nil
	assert.Nil(t, apiKey.LastUsedAt)

	// Update last used
	beforeUpdate := time.Now()
	apiKey.UpdateLastUsed()
	afterUpdate := time.Now()

	// LastUsedAt should now be set
	require.NotNil(t, apiKey.LastUsedAt)
	assert.True(t, apiKey.LastUsedAt.After(beforeUpdate) || apiKey.LastUsedAt.Equal(beforeUpdate))
	assert.True(t, apiKey.LastUsedAt.Before(afterUpdate) || apiKey.LastUsedAt.Equal(afterUpdate))
}

func TestAPIKey_UpdateLastUsed_Multiple(t *testing.T) {
	apiKey := entity.NewAPIKey("key_123", "abc123hash", "read", nil)

	// First update
	apiKey.UpdateLastUsed()
	firstUpdate := *apiKey.LastUsedAt

	// Wait a bit to ensure different timestamp
	time.Sleep(10 * time.Millisecond)

	// Second update
	apiKey.UpdateLastUsed()
	secondUpdate := *apiKey.LastUsedAt

	// Second update should be after first
	assert.True(t, secondUpdate.After(firstUpdate))
}

// ==================== Activate/Deactivate Tests ====================

func TestAPIKey_Deactivate(t *testing.T) {
	apiKey := entity.NewAPIKey("key_123", "abc123hash", "read", nil)

	// Initially active
	assert.True(t, apiKey.IsActive)

	// Deactivate
	apiKey.Deactivate()

	// Should now be inactive
	assert.False(t, apiKey.IsActive)
}

func TestAPIKey_Activate(t *testing.T) {
	apiKey := entity.NewAPIKey("key_123", "abc123hash", "read", nil)

	// Deactivate first
	apiKey.Deactivate()
	assert.False(t, apiKey.IsActive)

	// Activate
	apiKey.Activate()

	// Should now be active
	assert.True(t, apiKey.IsActive)
}

// ==================== Revoke Tests ====================

func TestAPIKey_Revoke(t *testing.T) {
	apiKey := entity.NewAPIKey("key_123", "abc123hash", "read", nil)
	revokedBy := "admin@example.com"
	reason := "Security breach"

	// Initially not revoked
	assert.False(t, apiKey.IsRevoked())
	assert.True(t, apiKey.IsActive)
	assert.Nil(t, apiKey.RevokedAt)
	assert.Nil(t, apiKey.RevokedBy)
	assert.Nil(t, apiKey.RevocationReason)

	// Revoke the key
	beforeRevoke := time.Now()
	apiKey.Revoke(revokedBy, &reason)
	afterRevoke := time.Now()

	// Should now be revoked
	assert.True(t, apiKey.IsRevoked())
	assert.False(t, apiKey.IsActive)
	require.NotNil(t, apiKey.RevokedAt)
	require.NotNil(t, apiKey.RevokedBy)
	require.NotNil(t, apiKey.RevocationReason)

	// Verify revocation details
	assert.Equal(t, revokedBy, *apiKey.RevokedBy)
	assert.Equal(t, reason, *apiKey.RevocationReason)
	assert.True(t, apiKey.RevokedAt.After(beforeRevoke) || apiKey.RevokedAt.Equal(beforeRevoke))
	assert.True(t, apiKey.RevokedAt.Before(afterRevoke) || apiKey.RevokedAt.Equal(afterRevoke))
}

func TestAPIKey_Revoke_NilReason(t *testing.T) {
	apiKey := entity.NewAPIKey("key_123", "abc123hash", "read", nil)
	revokedBy := "admin@example.com"

	// Revoke without reason
	apiKey.Revoke(revokedBy, nil)

	// Should be revoked
	assert.True(t, apiKey.IsRevoked())
	assert.False(t, apiKey.IsActive)
	require.NotNil(t, apiKey.RevokedAt)
	require.NotNil(t, apiKey.RevokedBy)
	assert.Nil(t, apiKey.RevocationReason)
	assert.Equal(t, revokedBy, *apiKey.RevokedBy)
}

func TestAPIKey_Revoke_EmptyReason(t *testing.T) {
	apiKey := entity.NewAPIKey("key_123", "abc123hash", "read", nil)
	revokedBy := "admin@example.com"
	reason := ""

	// Revoke with empty reason
	apiKey.Revoke(revokedBy, &reason)

	// Should be revoked with empty reason
	assert.True(t, apiKey.IsRevoked())
	require.NotNil(t, apiKey.RevocationReason)
	assert.Equal(t, "", *apiKey.RevocationReason)
}

// ==================== IsRevoked Tests ====================

func TestAPIKey_IsRevoked_False(t *testing.T) {
	apiKey := entity.NewAPIKey("key_123", "abc123hash", "read", nil)

	// New key should not be revoked
	assert.False(t, apiKey.IsRevoked())
}

func TestAPIKey_IsRevoked_True(t *testing.T) {
	apiKey := entity.NewAPIKey("key_123", "abc123hash", "read", nil)
	revokedBy := "admin@example.com"

	// Revoke the key
	apiKey.Revoke(revokedBy, nil)

	// Should be revoked
	assert.True(t, apiKey.IsRevoked())
}

func TestAPIKey_IsRevoked_ManuallySet(t *testing.T) {
	apiKey := entity.NewAPIKey("key_123", "abc123hash", "read", nil)
	now := time.Now()

	// Manually set RevokedAt (simulating database load)
	apiKey.RevokedAt = &now

	// Should be considered revoked
	assert.True(t, apiKey.IsRevoked())
}

// ==================== JSON Marshaling Tests ====================

func TestAPIKey_MarshalJSON(t *testing.T) {
	description := "Test API Key"
	apiKey := entity.NewAPIKey("key_123", "abc123hash", "read", &description)

	// Marshal to JSON
	jsonData, err := apiKey.MarshalJSON()
	require.NoError(t, err)
	require.NotNil(t, jsonData)

	// Should contain expected fields
	jsonStr := string(jsonData)
	assert.Contains(t, jsonStr, `"id":"key_123"`)
	assert.Contains(t, jsonStr, `"key_hash":"abc123hash"`)
	assert.Contains(t, jsonStr, `"role":"read"`)
	assert.Contains(t, jsonStr, `"description":"Test API Key"`)
	assert.Contains(t, jsonStr, `"is_active":true`)
	assert.Contains(t, jsonStr, `"created_at"`)
}

func TestAPIKey_MarshalJSON_WithLastUsed(t *testing.T) {
	apiKey := entity.NewAPIKey("key_123", "abc123hash", "read", nil)
	apiKey.UpdateLastUsed()

	// Marshal to JSON
	jsonData, err := apiKey.MarshalJSON()
	require.NoError(t, err)

	// Should contain last_used_at
	jsonStr := string(jsonData)
	assert.Contains(t, jsonStr, `"last_used_at"`)
}

func TestAPIKey_MarshalJSON_Revoked(t *testing.T) {
	apiKey := entity.NewAPIKey("key_123", "abc123hash", "read", nil)
	revokedBy := "admin@example.com"
	reason := "Security breach"
	apiKey.Revoke(revokedBy, &reason)

	// Marshal to JSON
	jsonData, err := apiKey.MarshalJSON()
	require.NoError(t, err)

	// Should contain revocation fields
	jsonStr := string(jsonData)
	assert.Contains(t, jsonStr, `"is_active":false`)
	assert.Contains(t, jsonStr, `"revoked_at"`)
	assert.Contains(t, jsonStr, `"revoked_by":"admin@example.com"`)
	assert.Contains(t, jsonStr, `"revocation_reason":"Security breach"`)
}

func TestAPIKey_MarshalJSON_RFC3339Format(t *testing.T) {
	apiKey := entity.NewAPIKey("key_123", "abc123hash", "read", nil)

	// Marshal to JSON
	jsonData, err := apiKey.MarshalJSON()
	require.NoError(t, err)

	// Parse JSON to verify timestamp format
	var result map[string]interface{}
	err = json.Unmarshal(jsonData, &result)
	require.NoError(t, err)

	// Verify created_at is in RFC3339 format
	createdAt, ok := result["created_at"].(string)
	require.True(t, ok, "created_at should be a string")

	// Parse as RFC3339
	_, err = time.Parse(time.RFC3339, createdAt)
	assert.NoError(t, err, "created_at should be in RFC3339 format")
}

// ==================== Edge Cases ====================

func TestAPIKey_RevokeAlreadyRevoked(t *testing.T) {
	apiKey := entity.NewAPIKey("key_123", "abc123hash", "read", nil)

	// First revocation
	firstRevokedBy := "admin1@example.com"
	firstReason := "First reason"
	apiKey.Revoke(firstRevokedBy, &firstReason)
	firstRevokedAt := *apiKey.RevokedAt

	// Wait a bit
	time.Sleep(10 * time.Millisecond)

	// Second revocation (should overwrite)
	secondRevokedBy := "admin2@example.com"
	secondReason := "Second reason"
	apiKey.Revoke(secondRevokedBy, &secondReason)

	// Should have second revocation details
	assert.Equal(t, secondRevokedBy, *apiKey.RevokedBy)
	assert.Equal(t, secondReason, *apiKey.RevocationReason)
	assert.True(t, apiKey.RevokedAt.After(firstRevokedAt))
}

func TestAPIKey_DeactivateDoesNotRevoke(t *testing.T) {
	apiKey := entity.NewAPIKey("key_123", "abc123hash", "read", nil)

	// Deactivate (not revoke)
	apiKey.Deactivate()

	// Should be inactive but not revoked
	assert.False(t, apiKey.IsActive)
	assert.False(t, apiKey.IsRevoked())
	assert.Nil(t, apiKey.RevokedAt)
	assert.Nil(t, apiKey.RevokedBy)
	assert.Nil(t, apiKey.RevocationReason)
}

func TestAPIKey_RevokeDeactivates(t *testing.T) {
	apiKey := entity.NewAPIKey("key_123", "abc123hash", "read", nil)
	revokedBy := "admin@example.com"

	// Revoke should also deactivate
	apiKey.Revoke(revokedBy, nil)

	// Should be both revoked and inactive
	assert.True(t, apiKey.IsRevoked())
	assert.False(t, apiKey.IsActive)
}

func TestAPIKey_ActivateAfterRevoke(t *testing.T) {
	apiKey := entity.NewAPIKey("key_123", "abc123hash", "read", nil)
	revokedBy := "admin@example.com"

	// Revoke
	apiKey.Revoke(revokedBy, nil)
	assert.True(t, apiKey.IsRevoked())
	assert.False(t, apiKey.IsActive)

	// Activate (should set IsActive but not clear revocation)
	apiKey.Activate()

	// IsActive should be true, but still revoked
	assert.True(t, apiKey.IsActive)
	assert.True(t, apiKey.IsRevoked())
	assert.NotNil(t, apiKey.RevokedAt)
}
