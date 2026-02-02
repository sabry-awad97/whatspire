package property

import (
	"context"
	"testing"

	"whatspire/internal/application/dto"
	"whatspire/internal/application/usecase"
	"whatspire/internal/infrastructure/persistence"
	"whatspire/test/mocks"

	"github.com/google/uuid"
	"pgregory.net/rapid"
)

// TestProperty9_ReadReceiptAtomicity tests that sending read receipts for multiple messages
// is atomic - either all messages are marked as read or none are
// **Validates: Requirements 3.1, 3.5**
// Feature: whatsapp-http-api-enhancement, Property 9: Read Receipt Atomicity
func TestProperty9_ReadReceiptAtomicity(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Setup
		ctx := context.Background()
		sessionID := uuid.New().String()
		chatJID := rapid.StringMatching("[0-9]{10,15}@s\\.whatsapp\\.net").Draw(t, "chat_jid")

		// Generate multiple message IDs (2-10 messages)
		numMessages := rapid.IntRange(2, 10).Draw(t, "num_messages")
		messageIDs := make([]string, numMessages)
		for i := 0; i < numMessages; i++ {
			messageIDs[i] = uuid.New().String()
		}

		// Create GORM repository
		db := setupTestDBForRapid(t)
		receiptRepo := persistence.NewReceiptRepository(db)

		// Create mock WhatsApp client
		mockWAClient := mocks.NewWhatsAppClientMock()
		mockWAClient.Connected[sessionID] = true
		mockWAClient.JIDMappings[sessionID] = chatJID

		// Create mock event publisher
		mockPublisher := mocks.NewEventPublisherMock()

		// Create use case
		uc := usecase.NewReceiptUseCase(mockWAClient, receiptRepo, mockPublisher)

		// Send read receipts for all messages
		req := dto.SendReceiptRequest{
			SessionID:  sessionID,
			ChatJID:    chatJID,
			MessageIDs: messageIDs,
		}

		err := uc.SendReadReceipt(ctx, req)
		if err != nil {
			t.Fatalf("Failed to send read receipts: %v", err)
		}

		// Property: All messages should have receipts saved
		for _, msgID := range messageIDs {
			receipts, err := receiptRepo.FindByMessageID(ctx, msgID)
			if err != nil {
				t.Fatalf("Failed to find receipts for message %s: %v", msgID, err)
			}
			if len(receipts) == 0 {
				t.Fatalf("Expected receipt for message %s, got none", msgID)
			}

			// Verify receipt is of type "read"
			receipt := receipts[0]
			if !receipt.IsRead() {
				t.Fatalf("Expected read receipt, got %s", receipt.Type)
			}
		}

		// Verify WhatsApp client was called exactly once with all message IDs
		if len(mockWAClient.SentReadReceipts) != 1 {
			t.Fatalf("Expected WhatsApp client to be called once, got %d calls", len(mockWAClient.SentReadReceipts))
		}

		sentReceipt := mockWAClient.SentReadReceipts[0]
		if len(sentReceipt.MessageIDs) != numMessages {
			t.Fatalf("Expected %d message IDs sent to WhatsApp, got %d", numMessages, len(sentReceipt.MessageIDs))
		}
	})
}

// TestReceiptUseCaseDisconnectedSession tests that operations on disconnected sessions fail
// **Validates: Requirements 3.1, 4.5**
// Feature: whatsapp-http-api-enhancement, Property 11: Disconnected Session Error Handling
func TestReceiptUseCaseDisconnectedSession(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		ctx := context.Background()
		sessionID := uuid.New().String()
		chatJID := rapid.StringMatching("[0-9]{10,15}@s\\.whatsapp\\.net").Draw(t, "chat_jid")

		// Generate message IDs
		numMessages := rapid.IntRange(1, 5).Draw(t, "num_messages")
		messageIDs := make([]string, numMessages)
		for i := 0; i < numMessages; i++ {
			messageIDs[i] = uuid.New().String()
		}

		// Create mock WhatsApp client that returns disconnected
		mockWAClient := mocks.NewWhatsAppClientMock()
		// Don't set Connected[sessionID] = true, so it's disconnected

		// Create use case
		uc := usecase.NewReceiptUseCase(mockWAClient, nil, nil)

		// Try to send read receipts
		req := dto.SendReceiptRequest{
			SessionID:  sessionID,
			ChatJID:    chatJID,
			MessageIDs: messageIDs,
		}

		err := uc.SendReadReceipt(ctx, req)

		// Property: Disconnected session should return error
		if err == nil {
			t.Fatalf("Expected error for disconnected session, got nil")
		}
	})
}

// TestReceiptUseCaseEmptyMessageIDs tests that empty message ID list is rejected
// **Validates: Requirements 3.1**
// Feature: whatsapp-http-api-enhancement, Property 9: Read Receipt Atomicity
func TestReceiptUseCaseEmptyMessageIDs(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		ctx := context.Background()
		sessionID := uuid.New().String()
		chatJID := rapid.StringMatching("[0-9]{10,15}@s\\.whatsapp\\.net").Draw(t, "chat_jid")

		// Create mock WhatsApp client
		mockWAClient := mocks.NewWhatsAppClientMock()
		mockWAClient.Connected[sessionID] = true
		mockWAClient.JIDMappings[sessionID] = chatJID

		// Create use case
		uc := usecase.NewReceiptUseCase(mockWAClient, nil, nil)

		// Try to send read receipts with empty message IDs
		req := dto.SendReceiptRequest{
			SessionID:  sessionID,
			ChatJID:    chatJID,
			MessageIDs: []string{}, // Empty list
		}

		err := uc.SendReadReceipt(ctx, req)

		// Property: Empty message ID list should return validation error
		if err == nil {
			t.Fatalf("Expected validation error for empty message IDs, got nil")
		}
	})
}
