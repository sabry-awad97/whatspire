package property

import (
	"context"
	"testing"

	"whatspire/internal/application/dto"
	"whatspire/internal/application/usecase"
	"whatspire/internal/domain/entity"
	"whatspire/internal/infrastructure/persistence"
	"whatspire/test/mocks"

	"github.com/google/uuid"
	"pgregory.net/rapid"
)

// TestProperty10_PresenceStateTransitions tests that presence state transitions work correctly
// **Validates: Requirements 4.1, 4.2**
// Feature: whatsapp-http-api-enhancement, Property 10: Presence State Transitions
func TestProperty10_PresenceStateTransitions(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Setup
		ctx := context.Background()
		sessionID := uuid.New().String()
		chatJID := rapid.StringMatching("[0-9]{10,15}@s\\.whatsapp\\.net").Draw(t, "chat_jid")

		// Create GORM repository
		db := setupTestDBForRapid(t)
		presenceRepo := persistence.NewPresenceRepository(db)

		// Create mock WhatsApp client
		mockWAClient := mocks.NewWhatsAppClientMock()
		mockWAClient.Connected[sessionID] = true
		mockWAClient.JIDMappings[sessionID] = chatJID

		// Create mock event publisher
		mockPublisher := mocks.NewEventPublisherMock()

		// Create use case
		uc := usecase.NewPresenceUseCase(mockWAClient, presenceRepo, mockPublisher)

		// Send "typing" presence
		typingReq := dto.SendPresenceRequest{
			SessionID: sessionID,
			ChatJID:   chatJID,
			State:     "typing",
		}

		err := uc.SendPresence(ctx, typingReq)
		if err != nil {
			t.Fatalf("Failed to send typing presence: %v", err)
		}

		// Send "paused" presence
		pausedReq := dto.SendPresenceRequest{
			SessionID: sessionID,
			ChatJID:   chatJID,
			State:     "paused",
		}

		err = uc.SendPresence(ctx, pausedReq)
		if err != nil {
			t.Fatalf("Failed to send paused presence: %v", err)
		}

		// Property: The final state should be "paused"
		presences, err := presenceRepo.FindBySessionID(ctx, sessionID, 10, 0)
		if err != nil {
			t.Fatalf("Failed to find presences: %v", err)
		}

		if len(presences) < 2 {
			t.Fatalf("Expected at least 2 presence records, got %d", len(presences))
		}

		// Get the latest presence (last one added)
		latest := presences[len(presences)-1]

		// Verify the final state is "paused"
		if latest.State != entity.PresenceStatePaused {
			t.Fatalf("Expected final state to be 'paused', got %s", latest.State)
		}
	})
}

// TestProperty11_PresenceDisconnectedSession tests that operations on disconnected sessions fail
// **Validates: Requirements 4.5**
// Feature: whatsapp-http-api-enhancement, Property 11: Disconnected Session Error Handling
func TestProperty11_PresenceDisconnectedSession(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		ctx := context.Background()
		sessionID := uuid.New().String()
		chatJID := rapid.StringMatching("[0-9]{10,15}@s\\.whatsapp\\.net").Draw(t, "chat_jid")
		state := rapid.SampledFrom([]string{"typing", "paused", "online", "offline"}).Draw(t, "state")

		// Create mock WhatsApp client that returns disconnected
		mockWAClient := mocks.NewWhatsAppClientMock()
		// Don't set Connected[sessionID] = true, so it's disconnected

		// Create use case
		uc := usecase.NewPresenceUseCase(mockWAClient, nil, nil)

		// Try to send presence
		req := dto.SendPresenceRequest{
			SessionID: sessionID,
			ChatJID:   chatJID,
			State:     state,
		}

		err := uc.SendPresence(ctx, req)

		// Property: Disconnected session should return error
		if err == nil {
			t.Fatalf("Expected error for disconnected session, got nil")
		}
	})
}

// TestPresenceUseCaseInvalidState tests that invalid presence states are rejected
// **Validates: Requirements 4.1, 4.2**
// Feature: whatsapp-http-api-enhancement, Property 10: Presence State Transitions
func TestPresenceUseCaseInvalidState(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		ctx := context.Background()
		sessionID := uuid.New().String()
		chatJID := rapid.StringMatching("[0-9]{10,15}@s\\.whatsapp\\.net").Draw(t, "chat_jid")

		// Generate invalid presence states
		invalidState := rapid.SampledFrom([]string{
			"invalid",
			"unknown",
			"",
			"TYPING", // Wrong case
			"away",
		}).Draw(t, "invalid_state")

		// Create mock WhatsApp client
		mockWAClient := mocks.NewWhatsAppClientMock()
		mockWAClient.Connected[sessionID] = true
		mockWAClient.JIDMappings[sessionID] = chatJID

		// Create use case
		uc := usecase.NewPresenceUseCase(mockWAClient, nil, nil)

		// Try to send presence with invalid state
		req := dto.SendPresenceRequest{
			SessionID: sessionID,
			ChatJID:   chatJID,
			State:     invalidState,
		}

		err := uc.SendPresence(ctx, req)

		// Property: Invalid state should be rejected with validation error
		if err == nil {
			t.Fatalf("Expected validation error for invalid state %q, got nil", invalidState)
		}
	})
}

// TestPresenceUseCaseValidStates tests that all valid presence states are accepted
// **Validates: Requirements 4.1, 4.2**
// Feature: whatsapp-http-api-enhancement, Property 10: Presence State Transitions
func TestPresenceUseCaseValidStates(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		ctx := context.Background()
		sessionID := uuid.New().String()
		chatJID := rapid.StringMatching("[0-9]{10,15}@s\\.whatsapp\\.net").Draw(t, "chat_jid")

		// Test all valid presence states
		validState := rapid.SampledFrom([]string{
			"typing",
			"paused",
			"online",
			"offline",
		}).Draw(t, "valid_state")

		// Create GORM repository
		db := setupTestDBForRapid(t)
		presenceRepo := persistence.NewPresenceRepository(db)

		// Create mock WhatsApp client
		mockWAClient := mocks.NewWhatsAppClientMock()
		mockWAClient.Connected[sessionID] = true
		mockWAClient.JIDMappings[sessionID] = chatJID

		// Create mock event publisher
		mockPublisher := mocks.NewEventPublisherMock()

		// Create use case
		uc := usecase.NewPresenceUseCase(mockWAClient, presenceRepo, mockPublisher)

		// Send presence with valid state
		req := dto.SendPresenceRequest{
			SessionID: sessionID,
			ChatJID:   chatJID,
			State:     validState,
		}

		err := uc.SendPresence(ctx, req)

		// Property: Valid state should be accepted without error
		if err != nil {
			t.Fatalf("Expected no error for valid state %q, got %v", validState, err)
		}

		// Verify presence was saved
		presences, err := presenceRepo.FindBySessionID(ctx, sessionID, 10, 0)
		if err != nil {
			t.Fatalf("Failed to find presences: %v", err)
		}

		if len(presences) == 0 {
			t.Fatalf("Expected presence to be saved for valid state %q", validState)
		}
	})
}
