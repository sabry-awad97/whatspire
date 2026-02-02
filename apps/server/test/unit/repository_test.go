package unit

import (
	"context"
	"testing"
	"time"

	"whatspire/internal/domain/entity"
	"whatspire/internal/infrastructure/persistence"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReactionRepository tests the reaction repository operations
func TestReactionRepository(t *testing.T) {
	ctx := context.Background()
	repo := persistence.NewInMemoryReactionRepository()

	t.Run("Save and FindByMessageID", func(t *testing.T) {
		repo.Clear()

		messageID := uuid.New().String()
		reaction1 := entity.NewReaction(uuid.New().String(), messageID, "session1", "user1", "user2", "😀")
		reaction2 := entity.NewReaction(uuid.New().String(), messageID, "session1", "user3", "user2", "👍")

		err := repo.Save(ctx, reaction1)
		require.NoError(t, err)

		err = repo.Save(ctx, reaction2)
		require.NoError(t, err)

		reactions, err := repo.FindByMessageID(ctx, messageID)
		require.NoError(t, err)
		assert.Len(t, reactions, 2)
	})

	t.Run("FindBySessionID with pagination", func(t *testing.T) {
		repo.Clear()

		sessionID := "session1"
		for i := 0; i < 5; i++ {
			reaction := entity.NewReaction(
				uuid.New().String(),
				uuid.New().String(),
				sessionID,
				"user1",
				"user2",
				"😀",
			)
			err := repo.Save(ctx, reaction)
			require.NoError(t, err)
		}

		// Get first 3
		reactions, err := repo.FindBySessionID(ctx, sessionID, 3, 0)
		require.NoError(t, err)
		assert.Len(t, reactions, 3)

		// Get next 2
		reactions, err = repo.FindBySessionID(ctx, sessionID, 3, 3)
		require.NoError(t, err)
		assert.Len(t, reactions, 2)
	})

	t.Run("Delete reaction", func(t *testing.T) {
		repo.Clear()

		reaction := entity.NewReaction(uuid.New().String(), uuid.New().String(), "session1", "user1", "user2", "😀")
		err := repo.Save(ctx, reaction)
		require.NoError(t, err)

		err = repo.Delete(ctx, reaction.ID)
		require.NoError(t, err)

		reactions, err := repo.FindByMessageID(ctx, reaction.MessageID)
		require.NoError(t, err)
		assert.Len(t, reactions, 0)
	})

	t.Run("DeleteByMessageIDAndFrom", func(t *testing.T) {
		repo.Clear()

		messageID := uuid.New().String()
		reaction1 := entity.NewReaction(uuid.New().String(), messageID, "session1", "user1", "user2", "😀")
		reaction2 := entity.NewReaction(uuid.New().String(), messageID, "session1", "user3", "user2", "👍")

		err := repo.Save(ctx, reaction1)
		require.NoError(t, err)
		err = repo.Save(ctx, reaction2)
		require.NoError(t, err)

		// Delete reaction from user1
		err = repo.DeleteByMessageIDAndFrom(ctx, messageID, "user1")
		require.NoError(t, err)

		reactions, err := repo.FindByMessageID(ctx, messageID)
		require.NoError(t, err)
		assert.Len(t, reactions, 1)
		assert.Equal(t, "user3", reactions[0].From)
	})
}

// TestReceiptRepository tests the receipt repository operations
func TestReceiptRepository(t *testing.T) {
	ctx := context.Background()
	repo := persistence.NewInMemoryReceiptRepository()

	t.Run("Save and FindByMessageID", func(t *testing.T) {
		repo.Clear()

		messageID := uuid.New().String()
		receipt1 := entity.NewReceipt(uuid.New().String(), messageID, "session1", "user1", "user2", entity.ReceiptTypeDelivered)
		receipt2 := entity.NewReceipt(uuid.New().String(), messageID, "session1", "user1", "user2", entity.ReceiptTypeRead)

		err := repo.Save(ctx, receipt1)
		require.NoError(t, err)

		err = repo.Save(ctx, receipt2)
		require.NoError(t, err)

		receipts, err := repo.FindByMessageID(ctx, messageID)
		require.NoError(t, err)
		assert.Len(t, receipts, 2)
	})

	t.Run("FindBySessionID with pagination", func(t *testing.T) {
		repo.Clear()

		sessionID := "session1"
		for i := 0; i < 5; i++ {
			receipt := entity.NewReceipt(
				uuid.New().String(),
				uuid.New().String(),
				sessionID,
				"user1",
				"user2",
				entity.ReceiptTypeDelivered,
			)
			err := repo.Save(ctx, receipt)
			require.NoError(t, err)
		}

		// Get first 3
		receipts, err := repo.FindBySessionID(ctx, sessionID, 3, 0)
		require.NoError(t, err)
		assert.Len(t, receipts, 3)

		// Get next 2
		receipts, err = repo.FindBySessionID(ctx, sessionID, 3, 3)
		require.NoError(t, err)
		assert.Len(t, receipts, 2)
	})

	t.Run("Delete receipt", func(t *testing.T) {
		repo.Clear()

		receipt := entity.NewReceipt(uuid.New().String(), uuid.New().String(), "session1", "user1", "user2", entity.ReceiptTypeDelivered)
		err := repo.Save(ctx, receipt)
		require.NoError(t, err)

		err = repo.Delete(ctx, receipt.ID)
		require.NoError(t, err)

		receipts, err := repo.FindByMessageID(ctx, receipt.MessageID)
		require.NoError(t, err)
		assert.Len(t, receipts, 0)
	})
}

// TestPresenceRepository tests the presence repository operations
func TestPresenceRepository(t *testing.T) {
	ctx := context.Background()
	repo := persistence.NewInMemoryPresenceRepository()

	t.Run("Save and FindBySessionID", func(t *testing.T) {
		repo.Clear()

		sessionID := "session1"
		presence1 := entity.NewPresence(uuid.New().String(), sessionID, "user1", "chat1", entity.PresenceStateTyping)
		presence2 := entity.NewPresence(uuid.New().String(), sessionID, "user2", "chat1", entity.PresenceStateOnline)

		err := repo.Save(ctx, presence1)
		require.NoError(t, err)

		err = repo.Save(ctx, presence2)
		require.NoError(t, err)

		presences, err := repo.FindBySessionID(ctx, sessionID, 10, 0)
		require.NoError(t, err)
		assert.Len(t, presences, 2)
	})

	t.Run("FindByUserJID with pagination", func(t *testing.T) {
		repo.Clear()

		userJID := "user1"
		for i := 0; i < 5; i++ {
			presence := entity.NewPresence(
				uuid.New().String(),
				"session1",
				userJID,
				"chat1",
				entity.PresenceStateTyping,
			)
			err := repo.Save(ctx, presence)
			require.NoError(t, err)
		}

		// Get first 3
		presences, err := repo.FindByUserJID(ctx, userJID, 3, 0)
		require.NoError(t, err)
		assert.Len(t, presences, 3)

		// Get next 2
		presences, err = repo.FindByUserJID(ctx, userJID, 3, 3)
		require.NoError(t, err)
		assert.Len(t, presences, 2)
	})

	t.Run("GetLatestByUserJID", func(t *testing.T) {
		repo.Clear()

		userJID := "user1"
		presence1 := entity.NewPresence(uuid.New().String(), "session1", userJID, "chat1", entity.PresenceStateTyping)

		err := repo.Save(ctx, presence1)
		require.NoError(t, err)

		// Create second presence with a later timestamp
		presence2 := entity.NewPresence(uuid.New().String(), "session1", userJID, "chat1", entity.PresenceStatePaused)
		presence2.Timestamp = presence1.Timestamp.Add(1 * time.Second)

		err = repo.Save(ctx, presence2)
		require.NoError(t, err)

		latest, err := repo.GetLatestByUserJID(ctx, userJID)
		require.NoError(t, err)
		assert.Equal(t, entity.PresenceStatePaused, latest.State)
	})

	t.Run("Delete presence", func(t *testing.T) {
		repo.Clear()

		presence := entity.NewPresence(uuid.New().String(), "session1", "user1", "chat1", entity.PresenceStateTyping)
		err := repo.Save(ctx, presence)
		require.NoError(t, err)

		err = repo.Delete(ctx, presence.ID)
		require.NoError(t, err)

		presences, err := repo.FindBySessionID(ctx, presence.SessionID, 10, 0)
		require.NoError(t, err)
		assert.Len(t, presences, 0)
	})
}
