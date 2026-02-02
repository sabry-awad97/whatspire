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
	"whatspire/internal/infrastructure/logger"
	"whatspire/internal/infrastructure/webhook"
	httpHandler "whatspire/internal/presentation/http"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== Test Setup ====================

// Test constants
const (
	testSessionID = "550e8400-e29b-41d4-a716-446655440000"
	testEventID   = "660e8400-e29b-41d4-a716-446655440001"
)

// noOpLogger is a no-op logger for testing
type noOpLogger struct{}

func (l *noOpLogger) Info(msg string, fields ...logger.Field)         {}
func (l *noOpLogger) Warn(msg string, fields ...logger.Field)         {}
func (l *noOpLogger) Error(msg string, fields ...logger.Field)        {}
func (l *noOpLogger) Debug(msg string, fields ...logger.Field)        {}
func (l *noOpLogger) WithContext(ctx context.Context) logger.Logger   { return l }
func (l *noOpLogger) WithFields(fields ...logger.Field) logger.Logger { return l }
func (l *noOpLogger) WithRequestID(requestID string) logger.Logger    { return l }
func (l *noOpLogger) WithSessionID(sessionID string) logger.Logger    { return l }

// setupE2ERouter creates a router with all use cases configured
func setupE2ERouter(
	sessionUC *usecase.SessionUseCase,
	messageUC *usecase.MessageUseCase,
	reactionUC *usecase.ReactionUseCase,
	receiptUC *usecase.ReceiptUseCase,
	presenceUC *usecase.PresenceUseCase,
	contactUC *usecase.ContactUseCase,
) *gin.Engine {
	gin.SetMode(gin.TestMode)
	handler := httpHandler.NewHandler(
		sessionUC,
		messageUC,
		nil, // healthUC
		nil, // groupsUC
		reactionUC,
		receiptUC,
		presenceUC,
		contactUC,
	)
	return httpHandler.NewRouter(handler, httpHandler.DefaultRouterConfig())
}

// mockWebhookServer creates a test HTTP server that captures webhook deliveries
type mockWebhookServer struct {
	server          *httptest.Server
	receivedEvents  []*entity.Event
	receivedHeaders map[string]string
	statusCode      int
}

func newMockWebhookServer() *mockWebhookServer {
	mws := &mockWebhookServer{
		receivedEvents:  make([]*entity.Event, 0),
		receivedHeaders: make(map[string]string),
		statusCode:      http.StatusOK,
	}

	mws.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Capture headers
		mws.receivedHeaders["X-Webhook-Signature"] = r.Header.Get("X-Webhook-Signature")
		mws.receivedHeaders["X-Webhook-Timestamp"] = r.Header.Get("X-Webhook-Timestamp")
		mws.receivedHeaders["Content-Type"] = r.Header.Get("Content-Type")

		// Decode event
		var event entity.Event
		if err := json.NewDecoder(r.Body).Decode(&event); err == nil {
			mws.receivedEvents = append(mws.receivedEvents, &event)
		}

		w.WriteHeader(mws.statusCode)
	}))

	return mws
}

func (mws *mockWebhookServer) Close() {
	mws.server.Close()
}

func (mws *mockWebhookServer) GetURL() string {
	return mws.server.URL
}

func (mws *mockWebhookServer) SetStatusCode(code int) {
	mws.statusCode = code
}

func (mws *mockWebhookServer) GetReceivedEvents() []*entity.Event {
	return mws.receivedEvents
}

func (mws *mockWebhookServer) GetHeaders() map[string]string {
	return mws.receivedHeaders
}

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
	testLogger := &noOpLogger{}
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
		json.Unmarshal(w.Body.Bytes(), &errResponse)
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

// ==================== E2E Test: Contact Operations ====================

func TestE2E_ContactOperations_CheckPhoneNumber(t *testing.T) {
	// Setup mocks
	waClient := NewWhatsAppClientMock()
	sessionRepo := NewSessionRepositoryMock()

	// Create session
	session := entity.NewSession(testSessionID, "Test Session")
	session.SetStatus(entity.StatusConnected)
	sessionRepo.Sessions[testSessionID] = session
	waClient.Connected[testSessionID] = true

	// Setup use cases
	contactUC := usecase.NewContactUseCase(waClient)

	// Setup router
	router := setupE2ERouter(nil, nil, nil, nil, nil, contactUC)

	// Check phone number
	req := httptest.NewRequest(http.MethodGet, "/api/contacts/check?phone=+1234567890&session_id=550e8400-e29b-41d4-a716-446655440000", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Verify HTTP response
	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.APIResponse[dto.ContactResponse]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)
	assert.True(t, response.Data.IsOnWhatsApp)
	assert.NotEmpty(t, response.Data.JID)
}

func TestE2E_ContactOperations_GetUserProfile(t *testing.T) {
	// Setup mocks
	waClient := NewWhatsAppClientMock()
	sessionRepo := NewSessionRepositoryMock()

	// Create session
	session := entity.NewSession(testSessionID, "Test Session")
	session.SetStatus(entity.StatusConnected)
	sessionRepo.Sessions[testSessionID] = session
	waClient.Connected[testSessionID] = true

	// Setup use cases
	contactUC := usecase.NewContactUseCase(waClient)

	// Setup router
	router := setupE2ERouter(nil, nil, nil, nil, nil, contactUC)

	// Get user profile
	req := httptest.NewRequest(http.MethodGet, "/api/contacts/1234567890@s.whatsapp.net/profile?session_id=550e8400-e29b-41d4-a716-446655440000", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Verify HTTP response
	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.APIResponse[dto.ContactResponse]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "1234567890@s.whatsapp.net", response.Data.JID)
	assert.Equal(t, "Test User", response.Data.Name)
	assert.NotNil(t, response.Data.AvatarURL)
	assert.NotNil(t, response.Data.Status)
}

func TestE2E_ContactOperations_ListContacts(t *testing.T) {
	// Setup mocks
	waClient := NewWhatsAppClientMock()
	sessionRepo := NewSessionRepositoryMock()

	// Create session
	session := entity.NewSession(testSessionID, "Test Session")
	session.SetStatus(entity.StatusConnected)
	sessionRepo.Sessions[testSessionID] = session
	waClient.Connected[testSessionID] = true

	// Setup use cases
	contactUC := usecase.NewContactUseCase(waClient)

	// Setup router
	router := setupE2ERouter(nil, nil, nil, nil, nil, contactUC)

	// List contacts
	req := httptest.NewRequest(http.MethodGet, "/api/sessions/550e8400-e29b-41d4-a716-446655440000/contacts", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Verify HTTP response
	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.APIResponse[dto.ContactListResponse]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)
	assert.Len(t, response.Data.Contacts, 2)
	assert.Equal(t, "1234567890@s.whatsapp.net", response.Data.Contacts[0].JID)
	assert.Equal(t, "Contact 1", response.Data.Contacts[0].Name)
}

func TestE2E_ContactOperations_ListChats(t *testing.T) {
	// Setup mocks
	waClient := NewWhatsAppClientMock()
	sessionRepo := NewSessionRepositoryMock()

	// Create session
	session := entity.NewSession(testSessionID, "Test Session")
	session.SetStatus(entity.StatusConnected)
	sessionRepo.Sessions[testSessionID] = session
	waClient.Connected[testSessionID] = true

	// Setup use cases
	contactUC := usecase.NewContactUseCase(waClient)

	// Setup router
	router := setupE2ERouter(nil, nil, nil, nil, nil, contactUC)

	// List chats
	req := httptest.NewRequest(http.MethodGet, "/api/sessions/550e8400-e29b-41d4-a716-446655440000/chats", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Verify HTTP response
	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.APIResponse[dto.ChatListResponse]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)
	assert.Len(t, response.Data.Chats, 2)
	assert.Equal(t, "1234567890@s.whatsapp.net", response.Data.Chats[0].JID)
	assert.Equal(t, "Chat 1", response.Data.Chats[0].Name)
	assert.False(t, response.Data.Chats[0].IsGroup)
	assert.Equal(t, "group123@g.us", response.Data.Chats[1].JID)
	assert.True(t, response.Data.Chats[1].IsGroup)
}

// ==================== E2E Test: Webhook Retry Logic ====================

func TestE2E_WebhookRetry_ExponentialBackoff(t *testing.T) {
	// Setup webhook server that fails initially
	webhookServer := newMockWebhookServer()
	defer webhookServer.Close()

	// Set to fail initially
	webhookServer.SetStatusCode(http.StatusInternalServerError)

	// Setup webhook publisher
	webhookConfig := webhook.WebhookConfig{
		URL:    webhookServer.GetURL(),
		Secret: "test-secret",
		Events: []string{"message.received"},
	}
	testLogger := &noOpLogger{}
	webhookPublisher := webhook.NewWebhookPublisher(webhookConfig, testLogger, nil)

	// Create event
	messageEvent, err := entity.NewEventWithPayload(
		testEventID,
		entity.EventTypeMessageReceived,
		testSessionID,
		map[string]interface{}{
			"message_id": "msg-123",
			"text":       "Test message",
		},
	)
	require.NoError(t, err)

	// Attempt to publish (should fail and retry)
	startTime := time.Now()
	err = webhookPublisher.Publish(context.Background(), messageEvent)
	duration := time.Since(startTime)

	// Should fail after retries
	assert.Error(t, err)

	// Should have taken at least 3 seconds (1s + 2s backoff for 3 total attempts)
	// Allow some tolerance for test execution time
	assert.GreaterOrEqual(t, duration.Seconds(), 2.5)

	// Should have received 3 attempts total
	assert.Equal(t, 3, len(webhookServer.GetReceivedEvents()))
}

func TestE2E_WebhookRetry_NoRetryOn4xx(t *testing.T) {
	// Setup webhook server that returns 400
	webhookServer := newMockWebhookServer()
	defer webhookServer.Close()

	// Set to return 400 (client error)
	webhookServer.SetStatusCode(http.StatusBadRequest)

	// Setup webhook publisher
	webhookConfig := webhook.WebhookConfig{
		URL:    webhookServer.GetURL(),
		Secret: "test-secret",
		Events: []string{"message.received"},
	}
	testLogger := &noOpLogger{}
	webhookPublisher := webhook.NewWebhookPublisher(webhookConfig, testLogger, nil)

	// Create event
	messageEvent, err := entity.NewEventWithPayload(
		testEventID,
		entity.EventTypeMessageReceived,
		testSessionID,
		map[string]interface{}{
			"message_id": "msg-123",
			"text":       "Test message",
		},
	)
	require.NoError(t, err)

	// Attempt to publish (should fail immediately without retry)
	startTime := time.Now()
	err = webhookPublisher.Publish(context.Background(), messageEvent)
	duration := time.Since(startTime)

	// Should fail
	assert.Error(t, err)

	// Should have taken less than 1 second (no retries)
	assert.Less(t, duration.Seconds(), 1.0)

	// Should have received only 1 attempt
	assert.Len(t, webhookServer.GetReceivedEvents(), 1)
}

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

func TestE2E_DisconnectedSession_ContactOperationsFail(t *testing.T) {
	// Setup mocks
	waClient := NewWhatsAppClientMock()

	// Create disconnected session
	waClient.Connected[testSessionID] = false

	// Setup use cases
	contactUC := usecase.NewContactUseCase(waClient)

	// Setup router
	router := setupE2ERouter(nil, nil, nil, nil, nil, contactUC)

	// Test check phone number
	req := httptest.NewRequest(http.MethodGet, "/api/contacts/check?phone=+1234567890&session_id=550e8400-e29b-41d4-a716-446655440000", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

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

// ==================== E2E Test: HMAC Signature Verification ====================

func TestE2E_WebhookHMAC_SignatureVerification(t *testing.T) {
	// Setup webhook server
	webhookServer := newMockWebhookServer()
	defer webhookServer.Close()

	// Setup webhook publisher with secret
	webhookConfig := webhook.WebhookConfig{
		URL:    webhookServer.GetURL(),
		Secret: "test-secret-key",
		Events: []string{"message.received"},
	}
	testLogger := &noOpLogger{}
	webhookPublisher := webhook.NewWebhookPublisher(webhookConfig, testLogger, nil)

	// Create event
	messageEvent, err := entity.NewEventWithPayload(
		testEventID,
		entity.EventTypeMessageReceived,
		testSessionID,
		map[string]interface{}{
			"message_id": "msg-123",
			"text":       "Test message",
		},
	)
	require.NoError(t, err)

	// Publish to webhook
	err = webhookPublisher.Publish(context.Background(), messageEvent)
	require.NoError(t, err)

	// Wait for delivery
	time.Sleep(100 * time.Millisecond)

	// Verify signature is present
	headers := webhookServer.GetHeaders()
	assert.NotEmpty(t, headers["X-Webhook-Signature"])
	assert.NotEmpty(t, headers["X-Webhook-Timestamp"])

	// Verify signature format (should be hex-encoded)
	signature := headers["X-Webhook-Signature"]
	assert.Len(t, signature, 64) // SHA256 hex = 64 characters
}

func TestE2E_WebhookHMAC_NoSignatureWithoutSecret(t *testing.T) {
	// Setup webhook server
	webhookServer := newMockWebhookServer()
	defer webhookServer.Close()

	// Setup webhook publisher WITHOUT secret
	webhookConfig := webhook.WebhookConfig{
		URL:    webhookServer.GetURL(),
		Secret: "", // No secret
		Events: []string{"message.received"},
	}
	testLogger := &noOpLogger{}
	webhookPublisher := webhook.NewWebhookPublisher(webhookConfig, testLogger, nil)

	// Create event
	messageEvent, err := entity.NewEventWithPayload(
		testEventID,
		entity.EventTypeMessageReceived,
		testSessionID,
		map[string]interface{}{
			"message_id": "msg-123",
			"text":       "Test message",
		},
	)
	require.NoError(t, err)

	// Publish to webhook
	err = webhookPublisher.Publish(context.Background(), messageEvent)
	require.NoError(t, err)

	// Wait for delivery
	time.Sleep(100 * time.Millisecond)

	// Verify signature is NOT present
	headers := webhookServer.GetHeaders()
	assert.Empty(t, headers["X-Webhook-Signature"])
	assert.NotEmpty(t, headers["X-Webhook-Timestamp"]) // Timestamp should still be present
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
