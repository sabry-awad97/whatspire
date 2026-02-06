package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"whatspire/internal/application/usecase"
	"whatspire/internal/domain/entity"
	"whatspire/test/helpers"

	"github.com/gin-gonic/gin"
)

// ==================== Shared Test Setup ====================
// This file contains shared setup code for E2E tests.
// Individual test cases are split across multiple files:
// - e2e_message_flows_test.go: Message, reaction, receipt, presence tests
// - e2e_contact_flows_test.go: Contact operations tests
// - e2e_webhook_flows_test.go: Webhook retry and HMAC tests
// - e2e_error_flows_test.go: Error handling and validation tests
// - e2e_complete_flows_test.go: Complete multi-operation flow test

// Test constants
const (
	testSessionID = "550e8400-e29b-41d4-a716-446655440000"
	testEventID   = "660e8400-e29b-41d4-a716-446655440001"
)

// setupE2ERouter creates a router with all use cases configured
func setupE2ERouter(
	sessionUC *usecase.SessionUseCase,
	messageUC *usecase.MessageUseCase,
	reactionUC *usecase.ReactionUseCase,
	receiptUC *usecase.ReceiptUseCase,
	presenceUC *usecase.PresenceUseCase,
	contactUC *usecase.ContactUseCase,
) *gin.Engine {
	handler := helpers.NewTestHandlerBuilder().
		WithSessionUseCase(sessionUC).
		WithMessageUseCase(messageUC).
		WithReactionUseCase(reactionUC).
		WithReceiptUseCase(receiptUC).
		WithPresenceUseCase(presenceUC).
		WithContactUseCase(contactUC).
		Build()
	return helpers.CreateTestRouterWithDefaults(handler)
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
