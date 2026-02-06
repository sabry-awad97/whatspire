package integration

import (
	"context"
	"testing"
	"time"

	"whatspire/internal/domain/entity"
	"whatspire/internal/infrastructure/webhook"
	"whatspire/test/helpers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== E2E Test: Webhook Retry Logic ====================

func TestE2E_WebhookRetry_ExponentialBackoff(t *testing.T) {
	// Setup webhook server that fails initially
	webhookServer := newMockWebhookServer()
	defer webhookServer.Close()

	// Set to fail initially
	webhookServer.SetStatusCode(500)

	// Setup webhook publisher
	webhookConfig := webhook.WebhookConfig{
		URL:    webhookServer.GetURL(),
		Secret: "test-secret",
		Events: []string{"message.received"},
	}
	testLogger := helpers.CreateTestLogger()
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
	webhookServer.SetStatusCode(400)

	// Setup webhook publisher
	webhookConfig := webhook.WebhookConfig{
		URL:    webhookServer.GetURL(),
		Secret: "test-secret",
		Events: []string{"message.received"},
	}
	testLogger := helpers.CreateTestLogger()
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
	testLogger := helpers.CreateTestLogger()
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
	testLogger := helpers.CreateTestLogger()
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
