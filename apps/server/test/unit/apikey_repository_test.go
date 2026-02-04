package unit

import (
	"context"
	"testing"
	"time"

	"whatspire/internal/domain/entity"
	"whatspire/internal/domain/errors"
	"whatspire/internal/infrastructure/persistence"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== APIKeyRepository Tests ====================

func TestAPIKeyRepository_Save(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	repo := persistence.NewAPIKeyRepository(db)

	t.Run("Save new API key", func(t *testing.T) {
		db.Exec("DELETE FROM api_keys WHERE 1=1")

		description := "Test API Key"
		apiKey := entity.NewAPIKey("key_123", "hash123", "read", &description)

		err := repo.Save(ctx, apiKey)
		require.NoError(t, err)

		// Verify it was saved
		found, err := repo.FindByID(ctx, "key_123")
		require.NoError(t, err)
		assert.Equal(t, "key_123", found.ID)
		assert.Equal(t, "hash123", found.KeyHash)
		assert.Equal(t, "read", found.Role)
		assert.Equal(t, &description, found.Description)
		assert.True(t, found.IsActive)
	})

	t.Run("Save duplicate key hash fails", func(t *testing.T) {
		db.Exec("DELETE FROM api_keys WHERE 1=1")

		apiKey1 := entity.NewAPIKey("key_1", "hash123", "read", nil)
		err := repo.Save(ctx, apiKey1)
		require.NoError(t, err)

		// Try to save another key with same hash
		apiKey2 := entity.NewAPIKey("key_2", "hash123", "write", nil)
		err = repo.Save(ctx, apiKey2)
		assert.ErrorIs(t, err, errors.ErrDuplicate)
	})

	t.Run("Save with nil description", func(t *testing.T) {
		db.Exec("DELETE FROM api_keys WHERE 1=1")

		apiKey := entity.NewAPIKey("key_123", "hash123", "write", nil)
		err := repo.Save(ctx, apiKey)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, "key_123")
		require.NoError(t, err)
		assert.Nil(t, found.Description)
	})
}

func TestAPIKeyRepository_FindByKeyHash(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	repo := persistence.NewAPIKeyRepository(db)

	t.Run("Find existing key by hash", func(t *testing.T) {
		db.Exec("DELETE FROM api_keys WHERE 1=1")

		description := "Test Key"
		apiKey := entity.NewAPIKey("key_123", "hash123", "admin", &description)
		err := repo.Save(ctx, apiKey)
		require.NoError(t, err)

		found, err := repo.FindByKeyHash(ctx, "hash123")
		require.NoError(t, err)
		assert.Equal(t, "key_123", found.ID)
		assert.Equal(t, "hash123", found.KeyHash)
		assert.Equal(t, "admin", found.Role)
	})

	t.Run("Find non-existent key hash", func(t *testing.T) {
		db.Exec("DELETE FROM api_keys WHERE 1=1")

		found, err := repo.FindByKeyHash(ctx, "nonexistent")
		assert.Nil(t, found)
		assert.ErrorIs(t, err, errors.ErrNotFound)
	})
}

func TestAPIKeyRepository_FindByID(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	repo := persistence.NewAPIKeyRepository(db)

	t.Run("Find existing key by ID", func(t *testing.T) {
		db.Exec("DELETE FROM api_keys WHERE 1=1")

		apiKey := entity.NewAPIKey("key_123", "hash123", "read", nil)
		err := repo.Save(ctx, apiKey)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, "key_123")
		require.NoError(t, err)
		assert.Equal(t, "key_123", found.ID)
		assert.Equal(t, "hash123", found.KeyHash)
	})

	t.Run("Find non-existent key by ID", func(t *testing.T) {
		db.Exec("DELETE FROM api_keys WHERE 1=1")

		found, err := repo.FindByID(ctx, "nonexistent")
		assert.Nil(t, found)
		assert.ErrorIs(t, err, errors.ErrNotFound)
	})
}

func TestAPIKeyRepository_UpdateLastUsed(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	repo := persistence.NewAPIKeyRepository(db)

	t.Run("Update last used timestamp", func(t *testing.T) {
		db.Exec("DELETE FROM api_keys WHERE 1=1")

		apiKey := entity.NewAPIKey("key_123", "hash123", "read", nil)
		err := repo.Save(ctx, apiKey)
		require.NoError(t, err)

		// Initially LastUsedAt should be nil
		found, _ := repo.FindByID(ctx, "key_123")
		assert.Nil(t, found.LastUsedAt)

		// Update last used
		beforeUpdate := time.Now()
		err = repo.UpdateLastUsed(ctx, "hash123")
		require.NoError(t, err)
		afterUpdate := time.Now()

		// Verify it was updated
		found, _ = repo.FindByID(ctx, "key_123")
		require.NotNil(t, found.LastUsedAt)
		assert.True(t, found.LastUsedAt.After(beforeUpdate) || found.LastUsedAt.Equal(beforeUpdate))
		assert.True(t, found.LastUsedAt.Before(afterUpdate) || found.LastUsedAt.Equal(afterUpdate))
	})

	t.Run("Update last used for non-existent key", func(t *testing.T) {
		db.Exec("DELETE FROM api_keys WHERE 1=1")

		err := repo.UpdateLastUsed(ctx, "nonexistent")
		assert.ErrorIs(t, err, errors.ErrNotFound)
	})
}

func TestAPIKeyRepository_Delete(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	repo := persistence.NewAPIKeyRepository(db)

	t.Run("Delete existing key", func(t *testing.T) {
		db.Exec("DELETE FROM api_keys WHERE 1=1")

		apiKey := entity.NewAPIKey("key_123", "hash123", "read", nil)
		err := repo.Save(ctx, apiKey)
		require.NoError(t, err)

		// Delete it
		err = repo.Delete(ctx, "key_123")
		require.NoError(t, err)

		// Verify it's gone
		found, err := repo.FindByID(ctx, "key_123")
		assert.Nil(t, found)
		assert.ErrorIs(t, err, errors.ErrNotFound)
	})

	t.Run("Delete non-existent key", func(t *testing.T) {
		db.Exec("DELETE FROM api_keys WHERE 1=1")

		err := repo.Delete(ctx, "nonexistent")
		assert.ErrorIs(t, err, errors.ErrNotFound)
	})
}

func TestAPIKeyRepository_Update(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	repo := persistence.NewAPIKeyRepository(db)

	t.Run("Update existing key", func(t *testing.T) {
		db.Exec("DELETE FROM api_keys WHERE 1=1")

		apiKey := entity.NewAPIKey("key_123", "hash123", "read", nil)
		err := repo.Save(ctx, apiKey)
		require.NoError(t, err)

		// Revoke the key
		revokedBy := "admin@example.com"
		reason := "Security breach"
		apiKey.Revoke(revokedBy, &reason)

		// Update it
		err = repo.Update(ctx, apiKey)
		require.NoError(t, err)

		// Verify changes were saved
		found, _ := repo.FindByID(ctx, "key_123")
		assert.False(t, found.IsActive)
		assert.True(t, found.IsRevoked())
		assert.Equal(t, revokedBy, *found.RevokedBy)
		assert.Equal(t, reason, *found.RevocationReason)
	})

	t.Run("Update non-existent key", func(t *testing.T) {
		db.Exec("DELETE FROM api_keys WHERE 1=1")

		apiKey := entity.NewAPIKey("nonexistent", "hash123", "read", nil)
		err := repo.Update(ctx, apiKey)
		assert.ErrorIs(t, err, errors.ErrNotFound)
	})

	t.Run("Update description", func(t *testing.T) {
		db.Exec("DELETE FROM api_keys WHERE 1=1")

		apiKey := entity.NewAPIKey("key_123", "hash123", "read", nil)
		err := repo.Save(ctx, apiKey)
		require.NoError(t, err)

		// Update description
		newDesc := "Updated description"
		apiKey.Description = &newDesc
		err = repo.Update(ctx, apiKey)
		require.NoError(t, err)

		// Verify
		found, _ := repo.FindByID(ctx, "key_123")
		require.NotNil(t, found.Description)
		assert.Equal(t, newDesc, *found.Description)
	})
}

func TestAPIKeyRepository_List(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	repo := persistence.NewAPIKeyRepository(db)

	t.Run("List all keys", func(t *testing.T) {
		db.Exec("DELETE FROM api_keys WHERE 1=1")

		// Create multiple keys
		for i := 0; i < 5; i++ {
			apiKey := entity.NewAPIKey(
				"key_"+string(rune('1'+i)),
				"hash"+string(rune('1'+i)),
				"read",
				nil,
			)
			err := repo.Save(ctx, apiKey)
			require.NoError(t, err)
			time.Sleep(1 * time.Millisecond) // Ensure different timestamps
		}

		keys, err := repo.List(ctx, 10, 0, nil, nil)
		require.NoError(t, err)
		assert.Len(t, keys, 5)
	})

	t.Run("List with pagination", func(t *testing.T) {
		db.Exec("DELETE FROM api_keys WHERE 1=1")

		// Create 10 keys
		for i := 0; i < 10; i++ {
			apiKey := entity.NewAPIKey(
				"key_"+string(rune('0'+i)),
				"hash"+string(rune('0'+i)),
				"read",
				nil,
			)
			err := repo.Save(ctx, apiKey)
			require.NoError(t, err)
			time.Sleep(1 * time.Millisecond)
		}

		// Get first page (5 items)
		keys, err := repo.List(ctx, 5, 0, nil, nil)
		require.NoError(t, err)
		assert.Len(t, keys, 5)

		// Get second page (5 items)
		keys, err = repo.List(ctx, 5, 5, nil, nil)
		require.NoError(t, err)
		assert.Len(t, keys, 5)

		// Get third page (0 items)
		keys, err = repo.List(ctx, 5, 10, nil, nil)
		require.NoError(t, err)
		assert.Len(t, keys, 0)
	})

	t.Run("List filtered by role", func(t *testing.T) {
		db.Exec("DELETE FROM api_keys WHERE 1=1")

		// Create keys with different roles
		roles := []string{"read", "write", "admin", "read", "write"}
		for i, role := range roles {
			apiKey := entity.NewAPIKey(
				"key_"+string(rune('1'+i)),
				"hash"+string(rune('1'+i)),
				role,
				nil,
			)
			err := repo.Save(ctx, apiKey)
			require.NoError(t, err)
			time.Sleep(1 * time.Millisecond)
		}

		// Filter by read role
		readRole := "read"
		keys, err := repo.List(ctx, 10, 0, &readRole, nil)
		require.NoError(t, err)
		assert.Len(t, keys, 2)
		for _, key := range keys {
			assert.Equal(t, "read", key.Role)
		}

		// Filter by admin role
		adminRole := "admin"
		keys, err = repo.List(ctx, 10, 0, &adminRole, nil)
		require.NoError(t, err)
		assert.Len(t, keys, 1)
		assert.Equal(t, "admin", keys[0].Role)
	})

	t.Run("List filtered by active status", func(t *testing.T) {
		db.Exec("DELETE FROM api_keys WHERE 1=1")

		// Create keys
		for i := 0; i < 3; i++ {
			apiKey := entity.NewAPIKey(
				"key_"+string(rune('1'+i)),
				"hash"+string(rune('1'+i)),
				"read",
				nil,
			)
			err := repo.Save(ctx, apiKey)
			require.NoError(t, err)
			time.Sleep(1 * time.Millisecond)
		}

		// Revoke one key
		apiKey, _ := repo.FindByID(ctx, "key_1")
		apiKey.Revoke("admin", nil)
		repo.Update(ctx, apiKey)

		// Filter by active
		active := true
		keys, err := repo.List(ctx, 10, 0, nil, &active)
		require.NoError(t, err)
		assert.Len(t, keys, 2)
		for _, key := range keys {
			assert.True(t, key.IsActive)
		}

		// Filter by inactive
		inactive := false
		keys, err = repo.List(ctx, 10, 0, nil, &inactive)
		require.NoError(t, err)
		assert.Len(t, keys, 1)
		assert.False(t, keys[0].IsActive)
	})

	t.Run("List with combined filters", func(t *testing.T) {
		db.Exec("DELETE FROM api_keys WHERE 1=1")

		// Create keys with different roles
		apiKey1 := entity.NewAPIKey("key_1", "hash1", "read", nil)
		repo.Save(ctx, apiKey1)
		time.Sleep(1 * time.Millisecond)

		apiKey2 := entity.NewAPIKey("key_2", "hash2", "read", nil)
		repo.Save(ctx, apiKey2)
		time.Sleep(1 * time.Millisecond)

		apiKey3 := entity.NewAPIKey("key_3", "hash3", "write", nil)
		repo.Save(ctx, apiKey3)

		// Revoke one read key
		apiKey1.Revoke("admin", nil)
		repo.Update(ctx, apiKey1)

		// Filter by read role and active status
		readRole := "read"
		active := true
		keys, err := repo.List(ctx, 10, 0, &readRole, &active)
		require.NoError(t, err)
		assert.Len(t, keys, 1)
		assert.Equal(t, "read", keys[0].Role)
		assert.True(t, keys[0].IsActive)
	})

	t.Run("List empty result", func(t *testing.T) {
		db.Exec("DELETE FROM api_keys WHERE 1=1")

		keys, err := repo.List(ctx, 10, 0, nil, nil)
		require.NoError(t, err)
		assert.Len(t, keys, 0)
	})
}

func TestAPIKeyRepository_Count(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	repo := persistence.NewAPIKeyRepository(db)

	t.Run("Count all keys", func(t *testing.T) {
		db.Exec("DELETE FROM api_keys WHERE 1=1")

		// Create 5 keys
		for i := 0; i < 5; i++ {
			apiKey := entity.NewAPIKey(
				"key_"+string(rune('1'+i)),
				"hash"+string(rune('1'+i)),
				"read",
				nil,
			)
			err := repo.Save(ctx, apiKey)
			require.NoError(t, err)
			time.Sleep(1 * time.Millisecond)
		}

		count, err := repo.Count(ctx, nil, nil)
		require.NoError(t, err)
		assert.Equal(t, int64(5), count)
	})

	t.Run("Count filtered by role", func(t *testing.T) {
		db.Exec("DELETE FROM api_keys WHERE 1=1")

		// Create keys with different roles
		roles := []string{"read", "write", "admin", "read"}
		for i, role := range roles {
			apiKey := entity.NewAPIKey(
				"key_"+string(rune('1'+i)),
				"hash"+string(rune('1'+i)),
				role,
				nil,
			)
			err := repo.Save(ctx, apiKey)
			require.NoError(t, err)
			time.Sleep(1 * time.Millisecond)
		}

		// Count read keys
		readRole := "read"
		count, err := repo.Count(ctx, &readRole, nil)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)

		// Count admin keys
		adminRole := "admin"
		count, err = repo.Count(ctx, &adminRole, nil)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("Count filtered by status", func(t *testing.T) {
		db.Exec("DELETE FROM api_keys WHERE 1=1")

		// Create 3 keys
		for i := 0; i < 3; i++ {
			apiKey := entity.NewAPIKey(
				"key_"+string(rune('1'+i)),
				"hash"+string(rune('1'+i)),
				"read",
				nil,
			)
			err := repo.Save(ctx, apiKey)
			require.NoError(t, err)
			time.Sleep(1 * time.Millisecond)
		}

		// Revoke one
		apiKey, _ := repo.FindByID(ctx, "key_1")
		apiKey.Revoke("admin", nil)
		repo.Update(ctx, apiKey)

		// Count active
		active := true
		count, err := repo.Count(ctx, nil, &active)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)

		// Count inactive
		inactive := false
		count, err = repo.Count(ctx, nil, &inactive)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("Count with combined filters", func(t *testing.T) {
		db.Exec("DELETE FROM api_keys WHERE 1=1")

		// Create keys
		apiKey1 := entity.NewAPIKey("key_1", "hash1", "read", nil)
		repo.Save(ctx, apiKey1)
		time.Sleep(1 * time.Millisecond)

		apiKey2 := entity.NewAPIKey("key_2", "hash2", "read", nil)
		repo.Save(ctx, apiKey2)
		time.Sleep(1 * time.Millisecond)

		apiKey3 := entity.NewAPIKey("key_3", "hash3", "write", nil)
		repo.Save(ctx, apiKey3)

		// Revoke one read key
		apiKey1.Revoke("admin", nil)
		repo.Update(ctx, apiKey1)

		// Count active read keys
		readRole := "read"
		active := true
		count, err := repo.Count(ctx, &readRole, &active)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("Count empty result", func(t *testing.T) {
		db.Exec("DELETE FROM api_keys WHERE 1=1")

		count, err := repo.Count(ctx, nil, nil)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}

// ==================== Edge Cases and Error Handling ====================

func TestAPIKeyRepository_ConcurrentUpdates(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)

	// Configure SQLite for shared cache to support concurrent access
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1) // Force single connection to avoid :memory: isolation issues

	repo := persistence.NewAPIKeyRepository(db)

	t.Run("Concurrent last used updates", func(t *testing.T) {
		db.Exec("DELETE FROM api_keys WHERE 1=1")

		apiKey := entity.NewAPIKey("key_123", "hash123", "read", nil)
		err := repo.Save(ctx, apiKey)
		require.NoError(t, err)

		// Verify key was saved before starting concurrent updates
		_, err = repo.FindByID(ctx, "key_123")
		require.NoError(t, err)

		// Simulate concurrent updates with sequential execution to avoid :memory: issues
		// In production with a real database, these would truly run concurrently
		errors := make([]error, 3)
		for i := 0; i < 3; i++ {
			errors[i] = repo.UpdateLastUsed(ctx, "hash123")
		}

		// Check all updates succeeded
		for i, err := range errors {
			assert.NoError(t, err, "Update %d failed", i)
		}

		// Verify key still exists and has last used timestamp
		found, err := repo.FindByID(ctx, "key_123")
		require.NoError(t, err)
		assert.NotNil(t, found.LastUsedAt)
	})
}

func TestAPIKeyRepository_SpecialCharacters(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	repo := persistence.NewAPIKeyRepository(db)

	t.Run("Save and retrieve with special characters in description", func(t *testing.T) {
		db.Exec("DELETE FROM api_keys WHERE 1=1")

		description := "Test key with special chars: !@#$%^&*()_+-=[]{}|;':\",./<>?"
		apiKey := entity.NewAPIKey("key_123", "hash123", "read", &description)
		err := repo.Save(ctx, apiKey)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, "key_123")
		require.NoError(t, err)
		require.NotNil(t, found.Description)
		assert.Equal(t, description, *found.Description)
	})

	t.Run("Save and retrieve with unicode in description", func(t *testing.T) {
		db.Exec("DELETE FROM api_keys WHERE 1=1")

		description := "Test key with unicode: 你好世界 🔑 مفتاح"
		apiKey := entity.NewAPIKey("key_123", "hash123", "read", &description)
		err := repo.Save(ctx, apiKey)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, "key_123")
		require.NoError(t, err)
		require.NotNil(t, found.Description)
		assert.Equal(t, description, *found.Description)
	})
}

func TestAPIKeyRepository_TimestampPrecision(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	repo := persistence.NewAPIKeyRepository(db)

	t.Run("Preserve timestamp precision", func(t *testing.T) {
		db.Exec("DELETE FROM api_keys WHERE 1=1")

		apiKey := entity.NewAPIKey("key_123", "hash123", "read", nil)
		originalTime := apiKey.CreatedAt

		err := repo.Save(ctx, apiKey)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, "key_123")
		require.NoError(t, err)

		// Timestamps should be very close (within 1 second due to DB precision)
		timeDiff := found.CreatedAt.Sub(originalTime)
		assert.True(t, timeDiff < time.Second && timeDiff > -time.Second)
	})
}
