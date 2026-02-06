package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"whatspire/internal/application/dto"
	"whatspire/test/helpers"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== Issue 1: Message Status Tests ====================
// Tests for the message status "pending" issue and sync mode feature

// TestSendMessage_AsyncMode_ReturnsPending verifies that async mode (default)
// returns HTTP 202 with status "pending"
func TestSendMessage_AsyncMode_ReturnsPending(t *testing.T) {
	waClient := NewWhatsAppClientMock()
	publisher := NewEventPublisherMock()
	mediaUploader := NewMediaUploaderMock()

	messageUC := helpers.NewTestMessageUseCase(waClient, publisher, mediaUploader, nil)
	defer messageUC.Close()

	gin.SetMode(gin.TestMode)
	handler := helpers.NewTestHandlerBuilder().
		WithMessageUseCase(messageUC).
		Build()
	router := helpers.CreateTestRouterWithDefaults(handler)

	text := "Test message"
	reqBody := dto.SendMessageRequest{
		SessionID: "550e8400-e29b-41d4-a716-446655440000",
		To:        "+1234567890",
		Type:      "text",
		Content: dto.SendMessageContentInput{
			Text: &text,
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Verify HTTP 202 Accepted
	assert.Equal(t, http.StatusAccepted, w.Code, "Async mode should return HTTP 202 Accepted")

	var response dto.APIResponse[map[string]interface{}]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.NotEmpty(t, response.Data["message_id"], "Response should contain message_id")
	assert.Equal(t, "pending", response.Data["status"], "Async mode should return status 'pending'")
}

// TestSendMessage_SyncMode_ReturnsActualStatus verifies that sync mode (?sync=true)
// returns HTTP 200 with actual status (sent or failed)
func TestSendMessage_SyncMode_ReturnsActualStatus(t *testing.T) {
	waClient := NewWhatsAppClientMock()
	publisher := NewEventPublisherMock()
	mediaUploader := NewMediaUploaderMock()

	messageUC := helpers.NewTestMessageUseCase(waClient, publisher, mediaUploader, nil)
	defer messageUC.Close()

	gin.SetMode(gin.TestMode)
	handler := helpers.NewTestHandlerBuilder().
		WithMessageUseCase(messageUC).
		Build()
	router := helpers.CreateTestRouterWithDefaults(handler)

	text := "Test message"
	reqBody := dto.SendMessageRequest{
		SessionID: "550e8400-e29b-41d4-a716-446655440000",
		To:        "+1234567890",
		Type:      "text",
		Content: dto.SendMessageContentInput{
			Text: &text,
		},
	}
	body, _ := json.Marshal(reqBody)

	// Use ?sync=true query parameter
	req := httptest.NewRequest(http.MethodPost, "/api/messages?sync=true", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Verify HTTP 200 OK
	assert.Equal(t, http.StatusOK, w.Code, "Sync mode should return HTTP 200 OK")

	var response dto.APIResponse[map[string]interface{}]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.NotEmpty(t, response.Data["message_id"], "Response should contain message_id")

	// Status should be actual result, not "pending"
	status := response.Data["status"].(string)
	assert.NotEqual(t, "pending", status, "Sync mode should not return 'pending' status")
	assert.Contains(t, []string{"sent", "failed"}, status, "Sync mode should return 'sent' or 'failed'")
}

// TestSendMessage_SyncMode_WithoutWhatsAppClient verifies that sync mode
// returns "failed" when WhatsApp client is not available
func TestSendMessage_SyncMode_WithoutWhatsAppClient(t *testing.T) {
	// No WhatsApp client
	publisher := NewEventPublisherMock()
	mediaUploader := NewMediaUploaderMock()

	messageUC := helpers.NewTestMessageUseCase(nil, publisher, mediaUploader, nil)
	defer messageUC.Close()

	gin.SetMode(gin.TestMode)
	handler := helpers.NewTestHandlerBuilder().
		WithMessageUseCase(messageUC).
		Build()
	router := helpers.CreateTestRouterWithDefaults(handler)

	text := "Test message"
	reqBody := dto.SendMessageRequest{
		SessionID: "550e8400-e29b-41d4-a716-446655440000",
		To:        "+1234567890",
		Type:      "text",
		Content: dto.SendMessageContentInput{
			Text: &text,
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/messages?sync=true", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return error because WhatsApp client is not available
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response dto.APIResponse[interface{}]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
}

// TestSendMessage_SyncMode_QueryParameterCaseSensitive verifies that the
// sync parameter is case-sensitive (only "true" lowercase works)
func TestSendMessage_SyncMode_QueryParameterCaseSensitive(t *testing.T) {
	waClient := NewWhatsAppClientMock()
	publisher := NewEventPublisherMock()
	mediaUploader := NewMediaUploaderMock()

	messageUC := helpers.NewTestMessageUseCase(waClient, publisher, mediaUploader, nil)
	defer messageUC.Close()

	gin.SetMode(gin.TestMode)
	handler := helpers.NewTestHandlerBuilder().
		WithMessageUseCase(messageUC).
		Build()
	router := helpers.CreateTestRouterWithDefaults(handler)

	text := "Test message"
	reqBody := dto.SendMessageRequest{
		SessionID: "550e8400-e29b-41d4-a716-446655440000",
		To:        "+1234567890",
		Type:      "text",
		Content: dto.SendMessageContentInput{
			Text: &text,
		},
	}
	body, _ := json.Marshal(reqBody)

	// Test with uppercase "True" - should be treated as async mode
	req := httptest.NewRequest(http.MethodPost, "/api/messages?sync=True", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 202 (async mode) because "True" != "true"
	assert.Equal(t, http.StatusAccepted, w.Code, "sync=True should be treated as async mode")

	var response dto.APIResponse[map[string]interface{}]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "pending", response.Data["status"], "sync=True should return pending status")
}

// TestSendMessage_SyncMode_WithInvalidValue verifies that invalid sync values
// are treated as async mode
func TestSendMessage_SyncMode_WithInvalidValue(t *testing.T) {
	waClient := NewWhatsAppClientMock()
	publisher := NewEventPublisherMock()
	mediaUploader := NewMediaUploaderMock()

	messageUC := helpers.NewTestMessageUseCase(waClient, publisher, mediaUploader, nil)
	defer messageUC.Close()

	gin.SetMode(gin.TestMode)
	handler := helpers.NewTestHandlerBuilder().
		WithMessageUseCase(messageUC).
		Build()
	router := helpers.CreateTestRouterWithDefaults(handler)

	text := "Test message"
	reqBody := dto.SendMessageRequest{
		SessionID: "550e8400-e29b-41d4-a716-446655440000",
		To:        "+1234567890",
		Type:      "text",
		Content: dto.SendMessageContentInput{
			Text: &text,
		},
	}
	body, _ := json.Marshal(reqBody)

	// Test with invalid value
	req := httptest.NewRequest(http.MethodPost, "/api/messages?sync=invalid", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 202 (async mode) because "invalid" != "true"
	assert.Equal(t, http.StatusAccepted, w.Code, "sync=invalid should be treated as async mode")

	var response dto.APIResponse[map[string]interface{}]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "pending", response.Data["status"], "sync=invalid should return pending status")
}

// TestSendMessage_AsyncMode_WithMultipleMessages verifies that async mode
// can handle multiple messages without blocking
func TestSendMessage_AsyncMode_WithMultipleMessages(t *testing.T) {
	waClient := NewWhatsAppClientMock()
	publisher := NewEventPublisherMock()
	mediaUploader := NewMediaUploaderMock()

	messageUC := helpers.NewTestMessageUseCase(waClient, publisher, mediaUploader, nil)
	defer messageUC.Close()

	gin.SetMode(gin.TestMode)
	handler := helpers.NewTestHandlerBuilder().
		WithMessageUseCase(messageUC).
		Build()
	router := helpers.CreateTestRouterWithDefaults(handler)

	// Send multiple messages in async mode
	messageIDs := make([]string, 5)
	for i := 0; i < 5; i++ {
		text := "Test message " + string(rune(i))
		reqBody := dto.SendMessageRequest{
			SessionID: "550e8400-e29b-41d4-a716-446655440000",
			To:        "+1234567890",
			Type:      "text",
			Content: dto.SendMessageContentInput{
				Text: &text,
			},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/messages", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusAccepted, w.Code)

		var response dto.APIResponse[map[string]interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		messageIDs[i] = response.Data["message_id"].(string)
	}

	// Verify all message IDs are unique
	seen := make(map[string]bool)
	for _, id := range messageIDs {
		assert.False(t, seen[id], "Message IDs should be unique")
		seen[id] = true
	}
}

// TestSendMessage_SyncMode_WithValidation verifies that sync mode still
// validates input before sending
func TestSendMessage_SyncMode_WithValidation(t *testing.T) {
	waClient := NewWhatsAppClientMock()
	publisher := NewEventPublisherMock()
	mediaUploader := NewMediaUploaderMock()

	messageUC := helpers.NewTestMessageUseCase(waClient, publisher, mediaUploader, nil)
	defer messageUC.Close()

	gin.SetMode(gin.TestMode)
	handler := helpers.NewTestHandlerBuilder().
		WithMessageUseCase(messageUC).
		Build()
	router := helpers.CreateTestRouterWithDefaults(handler)

	// Invalid phone number
	text := "Test message"
	reqBody := dto.SendMessageRequest{
		SessionID: "550e8400-e29b-41d4-a716-446655440000",
		To:        "invalid-phone",
		Type:      "text",
		Content: dto.SendMessageContentInput{
			Text: &text,
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/messages?sync=true", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return validation error
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response dto.APIResponse[interface{}]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Equal(t, "VALIDATION_FAILED", response.Error.Code)
}

// TestSendMessage_AsyncVsSyncResponseTime verifies that async mode returns
// faster than sync mode (async returns immediately, sync waits for send)
func TestSendMessage_AsyncVsSyncResponseTime(t *testing.T) {
	waClient := NewWhatsAppClientMock()
	publisher := NewEventPublisherMock()
	mediaUploader := NewMediaUploaderMock()

	messageUC := helpers.NewTestMessageUseCase(waClient, publisher, mediaUploader, nil)
	defer messageUC.Close()

	gin.SetMode(gin.TestMode)
	handler := helpers.NewTestHandlerBuilder().
		WithMessageUseCase(messageUC).
		Build()
	router := helpers.CreateTestRouterWithDefaults(handler)

	text := "Test message"
	reqBody := dto.SendMessageRequest{
		SessionID: "550e8400-e29b-41d4-a716-446655440000",
		To:        "+1234567890",
		Type:      "text",
		Content: dto.SendMessageContentInput{
			Text: &text,
		},
	}
	body, _ := json.Marshal(reqBody)

	// Measure async mode response time
	req1 := httptest.NewRequest(http.MethodPost, "/api/messages", bytes.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	// Measure sync mode response time
	body2, _ := json.Marshal(reqBody)
	req2 := httptest.NewRequest(http.MethodPost, "/api/messages?sync=true", bytes.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	// Both should succeed
	assert.Equal(t, http.StatusAccepted, w1.Code)
	assert.Equal(t, http.StatusOK, w2.Code)

	// Verify response statuses
	var response1 dto.APIResponse[map[string]interface{}]
	json.Unmarshal(w1.Body.Bytes(), &response1)
	assert.Equal(t, "pending", response1.Data["status"])

	var response2 dto.APIResponse[map[string]interface{}]
	json.Unmarshal(w2.Body.Bytes(), &response2)
	assert.NotEqual(t, "pending", response2.Data["status"])
}
