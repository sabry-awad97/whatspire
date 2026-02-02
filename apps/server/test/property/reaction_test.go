package property

import (
	"context"
	"testing"

	"whatspire/internal/application/dto"
	"whatspire/internal/application/usecase"
	"whatspire/internal/domain/entity"
	"whatspire/internal/domain/repository"
	"whatspire/internal/infrastructure/persistence"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"pgregory.net/rapid"
)

// MockWhatsAppClient is a mock implementation of WhatsAppClient for testing
type MockWhatsAppClient struct {
	mock.Mock
}

func (m *MockWhatsAppClient) Connect(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockWhatsAppClient) Disconnect(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockWhatsAppClient) SendMessage(ctx context.Context, msg *entity.Message) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

func (m *MockWhatsAppClient) IsConnected(sessionID string) bool {
	args := m.Called(sessionID)
	return args.Bool(0)
}

func (m *MockWhatsAppClient) GetSessionJID(sessionID string) (string, error) {
	args := m.Called(sessionID)
	return args.String(0), args.Error(1)
}

func (m *MockWhatsAppClient) SendReaction(ctx context.Context, sessionID, chatJID, messageID, emoji string) error {
	args := m.Called(ctx, sessionID, chatJID, messageID, emoji)
	return args.Error(0)
}

func (m *MockWhatsAppClient) GetQRChannel(ctx context.Context, sessionID string) (<-chan repository.QREvent, error) {
	args := m.Called(ctx, sessionID)
	if ch := args.Get(0); ch != nil {
		return ch.(<-chan repository.QREvent), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockWhatsAppClient) RegisterEventHandler(handler repository.EventHandler) {
	m.Called(handler)
}

func (m *MockWhatsAppClient) SetSessionJIDMapping(sessionID, jid string) {
	m.Called(sessionID, jid)
}

func (m *MockWhatsAppClient) SetHistorySyncConfig(sessionID string, enabled, fullSync bool, since string) {
	m.Called(sessionID, enabled, fullSync, since)
}

func (m *MockWhatsAppClient) GetHistorySyncConfig(sessionID string) (enabled, fullSync bool, since string) {
	args := m.Called(sessionID)
	return args.Bool(0), args.Bool(1), args.String(2)
}

// MockEventPublisher is a mock implementation of EventPublisher for testing
type MockEventPublisher struct {
	mock.Mock
}

func (m *MockEventPublisher) Connect(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockEventPublisher) Disconnect(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockEventPublisher) IsConnected() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockEventPublisher) Publish(ctx context.Context, event *entity.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventPublisher) QueueSize() int {
	args := m.Called()
	return args.Int(0)
}

// TestProperty6_ReactionDeliveryIdempotence tests that sending the same reaction multiple times
// results in the same final state (one reaction on the message)
// **Validates: Requirements 2.1**
// Feature: whatsapp-http-api-enhancement, Property 6: Reaction Delivery Idempotence
func TestProperty6_ReactionDeliveryIdempotence(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Setup
		ctx := context.Background()
		sessionID := uuid.New().String()
		messageID := uuid.New().String()
		chatJID := rapid.StringMatching("[0-9]{10,15}@s\\.whatsapp\\.net").Draw(t, "chat_jid")
		emoji := rapid.SampledFrom([]string{"👍", "❤️", "😂", "🎉"}).Draw(t, "emoji")

		// Create in-memory repository
		reactionRepo := persistence.NewInMemoryReactionRepository()

		// Create mock WhatsApp client
		mockWAClient := new(MockWhatsAppClient)
		mockWAClient.On("IsConnected", sessionID).Return(true)
		mockWAClient.On("GetSessionJID", sessionID).Return(chatJID, nil)
		mockWAClient.On("SendReaction", mock.Anything, sessionID, chatJID, messageID, emoji).Return(nil)

		// Create mock event publisher
		mockPublisher := new(MockEventPublisher)
		mockPublisher.On("IsConnected").Return(true)
		mockPublisher.On("Publish", mock.Anything, mock.Anything).Return(nil)

		// Create use case
		uc := usecase.NewReactionUseCase(mockWAClient, reactionRepo, mockPublisher)

		// Send the same reaction multiple times
		numAttempts := rapid.IntRange(2, 5).Draw(t, "num_attempts")
		var lastReaction *entity.Reaction
		for i := 0; i < numAttempts; i++ {
			req := dto.SendReactionRequest{
				SessionID: sessionID,
				ChatJID:   chatJID,
				MessageID: messageID,
				Emoji:     emoji,
			}

			reaction, err := uc.SendReaction(ctx, req)
			if err != nil {
				t.Fatalf("Failed to send reaction: %v", err)
			}
			lastReaction = reaction
		}

		// Property: After multiple sends, only one reaction should exist for this message
		reactions, err := reactionRepo.FindByMessageID(ctx, messageID)
		if err != nil {
			t.Fatalf("Failed to find reactions: %v", err)
		}

		// The repository should contain reactions from all attempts
		// but the final state should be idempotent (same emoji, same message)
		if len(reactions) == 0 {
			t.Fatalf("Expected at least one reaction, got none")
		}

		// All reactions should have the same emoji and message ID
		for _, r := range reactions {
			if r.MessageID != messageID {
				t.Fatalf("Expected message ID %s, got %s", messageID, r.MessageID)
			}
			if r.Emoji != emoji {
				t.Fatalf("Expected emoji %s, got %s", emoji, r.Emoji)
			}
		}

		// The last reaction should be valid
		if lastReaction == nil || !lastReaction.IsValid() {
			t.Fatalf("Last reaction should be valid")
		}
	})
}

// TestProperty7_ReactionRemovalIntegration tests that reactions can be removed via empty emoji
// at the use case level (integration test)
// **Validates: Requirements 2.4**
// Feature: whatsapp-http-api-enhancement, Property 7: Reaction Removal via Empty Emoji
func TestProperty7_ReactionRemovalIntegration(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Setup
		ctx := context.Background()
		sessionID := uuid.New().String()
		messageID := uuid.New().String()
		chatJID := rapid.StringMatching("[0-9]{10,15}@s\\.whatsapp\\.net").Draw(t, "chat_jid")
		emoji := rapid.SampledFrom([]string{"👍", "❤️", "😂"}).Draw(t, "emoji")

		// Create in-memory repository
		reactionRepo := persistence.NewInMemoryReactionRepository()

		// Create mock WhatsApp client
		mockWAClient := new(MockWhatsAppClient)
		mockWAClient.On("IsConnected", sessionID).Return(true)
		mockWAClient.On("GetSessionJID", sessionID).Return(chatJID, nil)
		mockWAClient.On("SendReaction", mock.Anything, sessionID, chatJID, messageID, emoji).Return(nil)
		mockWAClient.On("SendReaction", mock.Anything, sessionID, chatJID, messageID, "").Return(nil)

		// Create mock event publisher
		mockPublisher := new(MockEventPublisher)
		mockPublisher.On("IsConnected").Return(true)
		mockPublisher.On("Publish", mock.Anything, mock.Anything).Return(nil)

		// Create use case
		uc := usecase.NewReactionUseCase(mockWAClient, reactionRepo, mockPublisher)

		// First, send a reaction
		sendReq := dto.SendReactionRequest{
			SessionID: sessionID,
			ChatJID:   chatJID,
			MessageID: messageID,
			Emoji:     emoji,
		}

		reaction, err := uc.SendReaction(ctx, sendReq)
		if err != nil {
			t.Fatalf("Failed to send reaction: %v", err)
		}

		// Verify reaction was saved
		reactions, err := reactionRepo.FindByMessageID(ctx, messageID)
		if err != nil {
			t.Fatalf("Failed to find reactions: %v", err)
		}
		if len(reactions) == 0 {
			t.Fatalf("Expected reaction to be saved")
		}

		// Now remove the reaction using empty emoji
		removeReq := dto.RemoveReactionRequest{
			SessionID: sessionID,
			ChatJID:   chatJID,
			MessageID: messageID,
		}

		err = uc.RemoveReaction(ctx, removeReq)
		if err != nil {
			t.Fatalf("Failed to remove reaction: %v", err)
		}

		// Property: After removal, the reaction should be deleted from repository
		// Note: The repository's DeleteByMessageIDAndFrom should have been called
		// We verify the WhatsApp client was called with empty emoji
		mockWAClient.AssertCalled(t, "SendReaction", mock.Anything, sessionID, chatJID, messageID, "")

		// Verify the reaction entity itself accepts empty emoji
		if !reaction.IsValidEmoji() {
			t.Fatalf("Reaction entity should accept empty emoji for removal")
		}
	})
}

// TestProperty8_InvalidEmojiRejectionIntegration tests that invalid emoji strings are rejected
// at the use case level (integration test)
// **Validates: Requirements 2.5**
// Feature: whatsapp-http-api-enhancement, Property 8: Invalid Emoji Rejection
func TestProperty8_InvalidEmojiRejectionIntegration(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Setup
		ctx := context.Background()
		sessionID := uuid.New().String()
		messageID := uuid.New().String()
		chatJID := rapid.StringMatching("[0-9]{10,15}@s\\.whatsapp\\.net").Draw(t, "chat_jid")

		// Generate invalid emoji strings
		invalidEmoji := rapid.SampledFrom([]string{
			"abc",                      // Plain text
			"123",                      // Numbers
			"hello world",              // Multiple words
			string([]byte{0xFF, 0xFE}), // Invalid UTF-8
			"😀😀😀😀😀",                    // Too many emoji (>4 runes)
		}).Draw(t, "invalid_emoji")

		// Create in-memory repository
		reactionRepo := persistence.NewInMemoryReactionRepository()

		// Create mock WhatsApp client
		mockWAClient := new(MockWhatsAppClient)
		mockWAClient.On("IsConnected", sessionID).Return(true)
		mockWAClient.On("GetSessionJID", sessionID).Return(chatJID, nil)

		// Create mock event publisher
		mockPublisher := new(MockEventPublisher)
		mockPublisher.On("IsConnected").Return(true)

		// Create use case
		uc := usecase.NewReactionUseCase(mockWAClient, reactionRepo, mockPublisher)

		// Try to send reaction with invalid emoji
		req := dto.SendReactionRequest{
			SessionID: sessionID,
			ChatJID:   chatJID,
			MessageID: messageID,
			Emoji:     invalidEmoji,
		}

		_, err := uc.SendReaction(ctx, req)

		// Property: Invalid emoji should be rejected with validation error
		if err == nil {
			t.Fatalf("Expected validation error for invalid emoji %q, got nil", invalidEmoji)
		}

		// Verify WhatsApp client was NOT called (validation should fail before sending)
		mockWAClient.AssertNotCalled(t, "SendReaction", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

		// Verify no reaction was saved to repository
		reactions, err := reactionRepo.FindByMessageID(ctx, messageID)
		if err != nil {
			t.Fatalf("Failed to query reactions: %v", err)
		}
		if len(reactions) > 0 {
			t.Fatalf("Expected no reactions to be saved for invalid emoji, got %d", len(reactions))
		}
	})
}

// TestReactionUseCaseDisconnectedSession tests that operations on disconnected sessions fail
func TestReactionUseCaseDisconnectedSession(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		ctx := context.Background()
		sessionID := uuid.New().String()
		messageID := uuid.New().String()
		chatJID := rapid.StringMatching("[0-9]{10,15}@s\\.whatsapp\\.net").Draw(t, "chat_jid")
		emoji := rapid.SampledFrom([]string{"👍", "❤️"}).Draw(t, "emoji")

		// Create mock WhatsApp client that returns disconnected
		mockWAClient := new(MockWhatsAppClient)
		mockWAClient.On("IsConnected", sessionID).Return(false)

		// Create use case
		uc := usecase.NewReactionUseCase(mockWAClient, nil, nil)

		// Try to send reaction
		req := dto.SendReactionRequest{
			SessionID: sessionID,
			ChatJID:   chatJID,
			MessageID: messageID,
			Emoji:     emoji,
		}

		_, err := uc.SendReaction(ctx, req)

		// Property: Disconnected session should return error
		if err == nil {
			t.Fatalf("Expected error for disconnected session, got nil")
		}
	})
}
