package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"whatspire/internal/application/dto"
	"whatspire/internal/application/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== E2E Test: Disconnected Session Error Handling ====================

func TestE2E_DisconnectedSession_ReactionFails(t *testing.T) {
	// Setup mocks
	waClient := NewWhatsAppClientMock()
	publisher := NewEventPublisherMock()

	// Create disconnected session
	waClient.Connected[testSessionID] = false

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

	// Verify error response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response dto.APIResponse[interface{}]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "DISCONNECTED", response.Error.Code)
}

func TestE2E_DisconnectedSession_ReceiptFails(t *testing.T) {
	// Setup mocks
	waClient := NewWhatsAppClientMock()
	publisher := NewEventPublisherMock()

	// Create disconnected session
	waClient.Connected[testSessionID] = false

	// Setup use cases
	receiptUC := usecase.NewReceiptUseCase(waClient, nil, publisher)

	// Setup router
	router := setupE2ERouter(nil, nil, nil, receiptUC, nil, nil)

	// Send receipt request
	reqBody := dto.SendReceiptRequest{
		SessionID:  testSessionID,
		ChatJID:    "recipient@s.whatsapp.net",
		MessageIDs: []string{"msg-1"},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/messages/receipts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Verify error response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response dto.APIResponse[interface{}]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "DISCONNECTED", response.Error.Code)
}

func TestE2E_DisconnectedSession_PresenceFails(t *testing.T) {
	// Setup mocks
	waClient := NewWhatsAppClientMock()
	publisher := NewEventPublisherMock()

	// Create disconnected session
	waClient.Connected[testSessionID] = false

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

	// Verify error response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response dto.APIResponse[interface{}]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "DISCONNECTED", response.Error.Code)
}

// ==================== E2E Test: Validation Errors ====================

func TestE2E_Validation_InvalidEmoji(t *testing.T) {
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

	// Send reaction with invalid emoji
	reqBody := dto.SendReactionRequest{
		SessionID: testSessionID,
		ChatJID:   "recipient@s.whatsapp.net",
		MessageID: "msg-123",
		Emoji:     "not-an-emoji",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/messages/msg-123/reactions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Verify error response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response dto.APIResponse[interface{}]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "VALIDATION_FAILED", response.Error.Code)
}

func TestE2E_Validation_InvalidPresenceState(t *testing.T) {
	// Setup mocks
	waClient := NewWhatsAppClientMock()
	publisher := NewEventPublisherMock()

	// Create connected session
	waClient.Connected[testSessionID] = true
	waClient.SetSessionJIDMapping(testSessionID, "1234567890@s.whatsapp.net")

	// Setup use cases
	presenceUC := usecase.NewPresenceUseCase(waClient, nil, publisher)

	// Setup router
	router := setupE2ERouter(nil, nil, nil, nil, presenceUC, nil)

	// Send presence with invalid state
	reqBody := dto.SendPresenceRequest{
		SessionID: testSessionID,
		ChatJID:   "recipient@s.whatsapp.net",
		State:     "invalid-state",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/presence", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Verify error response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response dto.APIResponse[interface{}]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "VALIDATION_FAILED", response.Error.Code)
}
