package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"whatspire/internal/application/dto"
	"whatspire/internal/application/usecase"
	"whatspire/internal/domain/entity"

	"github.com/stretchr/testify/assert"
)

// ==================== E2E Test: Complete Flow with Multiple Operations ====================

func TestE2E_CompleteFlow_MultipleOperations(t *testing.T) {
	// Setup webhook server
	webhookServer := newMockWebhookServer()
	defer webhookServer.Close()

	// Setup mocks
	waClient := NewWhatsAppClientMock()
	publisher := NewEventPublisherMock()
	sessionRepo := NewSessionRepositoryMock()

	// Create session
	session := entity.NewSession(testSessionID, "Test Session")
	session.SetStatus(entity.StatusConnected)
	session.SetJID("1234567890@s.whatsapp.net")
	sessionRepo.Sessions[testSessionID] = session
	waClient.Connected[testSessionID] = true
	waClient.SetSessionJIDMapping(testSessionID, "1234567890@s.whatsapp.net")

	// Setup use cases
	reactionUC := usecase.NewReactionUseCase(waClient, nil, publisher)
	receiptUC := usecase.NewReceiptUseCase(waClient, nil, publisher)
	presenceUC := usecase.NewPresenceUseCase(waClient, nil, publisher)
	contactUC := usecase.NewContactUseCase(waClient)

	// Setup router
	router := setupE2ERouter(nil, nil, reactionUC, receiptUC, presenceUC, contactUC)

	// 1. Send typing presence
	presenceReq := dto.SendPresenceRequest{
		SessionID: testSessionID,
		ChatJID:   "recipient@s.whatsapp.net",
		State:     "typing",
	}
	body, _ := json.Marshal(presenceReq)
	req := httptest.NewRequest(http.MethodPost, "/api/presence", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 2. Send reaction
	reactionReq := dto.SendReactionRequest{
		SessionID: testSessionID,
		ChatJID:   "recipient@s.whatsapp.net",
		MessageID: "msg-123",
		Emoji:     "👍",
	}
	body, _ = json.Marshal(reactionReq)
	req = httptest.NewRequest(http.MethodPost, "/api/messages/msg-123/reactions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 3. Send read receipts
	receiptReq := dto.SendReceiptRequest{
		SessionID:  testSessionID,
		ChatJID:    "recipient@s.whatsapp.net",
		MessageIDs: []string{"msg-1", "msg-2"},
	}
	body, _ = json.Marshal(receiptReq)
	req = httptest.NewRequest(http.MethodPost, "/api/messages/receipts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 4. Check phone number
	req = httptest.NewRequest(http.MethodGet, "/api/contacts/check?phone=+1234567890&session_id=550e8400-e29b-41d4-a716-446655440000", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 5. List contacts
	req = httptest.NewRequest(http.MethodGet, "/api/sessions/550e8400-e29b-41d4-a716-446655440000/contacts", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify all events were published
	// 1 presence + 1 reaction + 2 receipts = 4 events
	assert.Len(t, publisher.Events, 4)

	// Verify event types
	eventTypes := make(map[entity.EventType]int)
	for _, event := range publisher.Events {
		eventTypes[event.Type]++
	}
	assert.Equal(t, 1, eventTypes[entity.EventTypePresenceUpdate])
	assert.Equal(t, 1, eventTypes[entity.EventTypeMessageReaction])
	assert.Equal(t, 2, eventTypes[entity.EventTypeMessageRead])
}
