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
	"whatspire/internal/domain/entity"
	"whatspire/internal/infrastructure/webhook"
	"whatspire/test/helpers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== E2E Test: Message Reception Flow ====================

func TestE2E_MessageReception_ProcessingAndWebhookDelivery(t *testing.T) {
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

	// Setup webhook publisher
	webhookConfig := webhook.WebhookConfig{
		URL:    webhookServer.GetURL(),
		Secret: "test-secret",
		Events: []string{"message.received"},
	}
	testLogger := helpers.CreateTestLogger()
	webhookPublisher := webhook.NewWebhookPublisher(webhookConfig, testLogger, nil)

	// Simulate incoming message event
	messageEvent, err := entity.NewEventWithPayload(
		testEventID,
		entity.EventTypeMessageReceived,
		testSessionID,
		map[string]interface{}{
			"message_id": "msg-123",
			"from":       "sender@s.whatsapp.net",
			"to":         "1234567890@s.whatsapp.net",
			"type":       "text",
			"text":       "Hello, World!",
			"timestamp":  time.Now().Unix(),
		},
	)
	require.NoError(t, err)

	// Publish to event hub (simulated)
	err = publisher.Publish(context.Background(), messageEvent)
	require.NoError(t, err)

	// Deliver to webhook
	err = webhookPublisher.Publish(context.Background(), messageEvent)
	require.NoError(t, err)

	// Wait for webhook delivery
	time.Sleep(100 * time.Millisecond)

	// Verify webhook received the event
	events := webhookServer.GetReceivedEvents()
	assert.Len(t, events, 1)
	assert.Equal(t, entity.EventTypeMessageReceived, events[0].Type)
	assert.Equal(t, testSessionID, events[0].SessionID)

	// Verify HMAC signature is present
	headers := webhookServer.GetHeaders()
	assert.NotEmpty(t, headers["X-Webhook-Signature"])
	assert.NotEmpty(t, headers["X-Webhook-Timestamp"])
	assert.Equal(t, "application/json", headers["Content-Type"])
}

// ==================== E2E Test: Reaction Send Flow ====================

func TestE2E_ReactionSend_WhatsAppAndEventPublishing(t *testing.T) {
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

	// Setup router
	router := setupE2ERouter(nil, nil, reactionUC, nil, nil, nil)

	// Send reaction request
	reqBody := dto.SendReactionRequest{
		SessionID: testSessionID,
		ChatJID:   "recipient@s.whatsapp.net",
		MessageID: "msg-123",
		Emoji:     "👍",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/messages/msg-123/reactions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Verify HTTP response
	if w.Code != http.StatusOK {
		var errResponse dto.APIResponse[interface{}]
		_ = json.Unmarshal(w.Body.Bytes(), &errResponse)
		t.Logf("Error response: %+v", errResponse)
	}
	require.Equal(t, http.StatusOK, w.Code, "Response body: %s", w.Body.String())

	var response dto.APIResponse[dto.ReactionResponse]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "msg-123", response.Data.MessageID)
	assert.Equal(t, "👍", response.Data.Emoji)

	// Verify event was published
	require.Len(t, publisher.Events, 1)
	assert.Equal(t, entity.EventTypeMessageReaction, publisher.Events[0].Type)
	assert.Equal(t, testSessionID, publisher.Events[0].SessionID)
}

// ==================== E2E Test: Reaction Removal ====================

func TestE2E_ReactionRemoval_EmptyEmoji(t *testing.T) {
	// Setup mocks
	waClient := NewWhatsAppClientMock()
	publisher := NewEventPublisherMock()

	// Create connected session
	waClient.Connected[testSessionID] = true
	waClient.SetSessionJIDMapping(testSessionID, "1234567890@s.whatsapp.net")

	// Setup use cases
	reactionUC := usecase.NewReactionUseCase(waClient, nil, publisher)

	// Setup router
	router := setupE2ERouter(nil, nil, reactionUC, nil, nil, nil)

	// Remove reaction
	reqBody := dto.RemoveReactionRequest{
		SessionID: testSessionID,
		ChatJID:   "recipient@s.whatsapp.net",
		MessageID: "msg-123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodDelete, "/api/messages/msg-123/reactions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Verify HTTP response
	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.APIResponse[map[string]string]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "Reaction removed successfully", response.Data["message"])

	// Verify event was published
	assert.Len(t, publisher.Events, 1)
	assert.Equal(t, entity.EventTypeMessageReaction, publisher.Events[0].Type)
}

// ==================== E2E Test: Receipt Send Flow ====================

func TestE2E_ReceiptSend_WhatsAppAndEventPublishing(t *testing.T) {
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
	receiptUC := usecase.NewReceiptUseCase(waClient, nil, publisher)

	// Setup router
	router := setupE2ERouter(nil, nil, nil, receiptUC, nil, nil)

	// Send receipt request
	reqBody := dto.SendReceiptRequest{
		SessionID:  testSessionID,
		ChatJID:    "recipient@s.whatsapp.net",
		MessageIDs: []string{"msg-1", "msg-2", "msg-3"},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/messages/receipts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Verify HTTP response
	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.APIResponse[dto.ReceiptResponse]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, 3, response.Data.ProcessedCount)

	// Verify WhatsApp client received the call
	assert.Len(t, waClient.SentReadReceipts, 1)
	assert.Equal(t, testSessionID, waClient.SentReadReceipts[0].SessionID)
	assert.Equal(t, "recipient@s.whatsapp.net", waClient.SentReadReceipts[0].ChatJID)
	assert.Equal(t, []string{"msg-1", "msg-2", "msg-3"}, waClient.SentReadReceipts[0].MessageIDs)

	// Verify events were published (one per message)
	assert.Len(t, publisher.Events, 3)
	for _, event := range publisher.Events {
		assert.Equal(t, entity.EventTypeMessageRead, event.Type)
		assert.Equal(t, testSessionID, event.SessionID)
	}
}

// ==================== E2E Test: Presence Send Flow ====================

func TestE2E_PresenceSend_WhatsAppAndEventPublishing(t *testing.T) {
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
	presenceUC := usecase.NewPresenceUseCase(waClient, nil, publisher)

	// Setup router
	router := setupE2ERouter(nil, nil, nil, nil, presenceUC, nil)

	// Send presence request
	reqBody := dto.SendPresenceRequest{
		SessionID: testSessionID,
		ChatJID:   "recipient@s.whatsapp.net",
		State:     "typing",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/presence", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Verify HTTP response
	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.APIResponse[dto.PresenceResponse]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "recipient@s.whatsapp.net", response.Data.ChatJID)
	assert.Equal(t, "typing", response.Data.State)

	// Verify event was published
	assert.Len(t, publisher.Events, 1)
	assert.Equal(t, entity.EventTypePresenceUpdate, publisher.Events[0].Type)
	assert.Equal(t, testSessionID, publisher.Events[0].SessionID)
}
