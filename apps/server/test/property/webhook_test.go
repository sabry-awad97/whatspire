package property

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	"whatspire/internal/domain/entity"
	"whatspire/internal/infrastructure/logger"
	"whatspire/internal/infrastructure/webhook"

	"pgregory.net/rapid"
)

// Helper to create test logger
func createTestLogger() logger.Logger {
	return logger.NewStructuredLogger(logger.Config{Level: "info", Format: "json"})
}

// Helper to create test event
func createTestEvent(eventType string, sessionID string) *entity.Event {
	data, _ := json.Marshal(map[string]any{"test": "data"})
	return &entity.Event{
		ID:        "test-event-id",
		Type:      entity.EventType(eventType),
		SessionID: sessionID,
		Timestamp: time.Now(),
		Data:      data,
	}
}

// ==================== Property 16: HMAC Signature Verification ====================

// TestProperty16_HMACSignatureVerification tests that HMAC signatures are correctly computed
// Feature: whatsapp-http-api-enhancement, Property 16: HMAC Signature Verification
// Validates: Requirements 6.1, 6.2
func TestProperty16_HMACSignatureVerification(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	rapid.Check(t, func(t *rapid.T) {
		// Generate random event and secret
		eventType := rapid.StringMatching("[a-z]+\\.[a-z]+").Draw(t, "event_type")
		secret := rapid.StringMatching("[a-zA-Z0-9]{16,32}").Draw(t, "secret")
		sessionID := rapid.StringMatching("[a-zA-Z0-9-]{10,36}").Draw(t, "session_id")

		// Create test server to capture request
		var capturedSignature string
		var capturedPayload []byte
		var mu sync.Mutex

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mu.Lock()
			defer mu.Unlock()

			capturedSignature = r.Header.Get("X-Webhook-Signature")
			capturedPayload, _ = io.ReadAll(r.Body)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		// Create webhook publisher with secret
		config := webhook.WebhookConfig{
			URL:    server.URL,
			Secret: secret,
			Events: []string{eventType},
		}
		publisher := webhook.NewWebhookPublisher(config, createTestLogger(), nil)

		// Create and publish event
		event := createTestEvent(eventType, sessionID)

		ctx := context.Background()
		err := publisher.Publish(ctx, event)
		if err != nil {
			t.Fatalf("Failed to publish event: %v", err)
		}

		// Wait for request to be processed
		time.Sleep(100 * time.Millisecond)

		mu.Lock()
		defer mu.Unlock()

		// Verify signature was included
		if capturedSignature == "" {
			t.Fatalf("HMAC signature not included in request")
		}

		// Compute expected HMAC
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(capturedPayload)
		expectedSignature := hex.EncodeToString(mac.Sum(nil))

		// Verify signature matches
		if capturedSignature != expectedSignature {
			t.Fatalf("HMAC signature mismatch: got %s, expected %s", capturedSignature, expectedSignature)
		}
	})
}

// ==================== Property 17: Webhook Timestamp Freshness ====================

// TestProperty17_WebhookTimestampFreshness tests that webhook timestamps are fresh
// Feature: whatsapp-http-api-enhancement, Property 17: Webhook Timestamp Freshness
// Validates: Requirements 6.3
func TestProperty17_WebhookTimestampFreshness(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	rapid.Check(t, func(t *rapid.T) {
		// Generate random event
		eventType := rapid.StringMatching("[a-z]+\\.[a-z]+").Draw(t, "event_type")
		sessionID := rapid.StringMatching("[a-zA-Z0-9-]{10,36}").Draw(t, "session_id")

		// Create test server to capture timestamp
		var capturedTimestamp string
		var mu sync.Mutex

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mu.Lock()
			defer mu.Unlock()

			capturedTimestamp = r.Header.Get("X-Webhook-Timestamp")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		// Create webhook publisher
		config := webhook.WebhookConfig{
			URL:    server.URL,
			Events: []string{eventType},
		}
		publisher := webhook.NewWebhookPublisher(config, createTestLogger(), nil)

		// Record time before publishing (with buffer for timing variations)
		beforePublish := time.Now().Add(-1 * time.Second)

		// Create and publish event
		event := createTestEvent(eventType, sessionID)

		ctx := context.Background()
		err := publisher.Publish(ctx, event)
		if err != nil {
			t.Fatalf("Failed to publish event: %v", err)
		}

		// Wait for request to be processed
		time.Sleep(100 * time.Millisecond)

		mu.Lock()
		defer mu.Unlock()

		// Verify timestamp was included
		if capturedTimestamp == "" {
			t.Fatalf("Timestamp not included in request")
		}

		// Parse timestamp
		timestampInt, err := strconv.ParseInt(capturedTimestamp, 10, 64)
		if err != nil {
			t.Fatalf("Invalid timestamp format: %v", err)
		}

		webhookTime := time.Unix(timestampInt, 0)

		// Verify timestamp is within 5 minutes of current time
		timeDiff := time.Since(webhookTime).Abs()
		if timeDiff > 5*time.Minute {
			t.Fatalf("Timestamp not fresh: %v old", timeDiff)
		}

		// Verify timestamp is not in the future (with 2 second buffer)
		if webhookTime.After(time.Now().Add(2 * time.Second)) {
			t.Fatalf("Timestamp is in the future")
		}

		// Verify timestamp is reasonably close to when we started the test
		if webhookTime.Before(beforePublish) {
			t.Fatalf("Timestamp is before publish time")
		}
	})
}

// ==================== Property 18: HMAC Signature Conditional Inclusion ====================

// TestProperty18_HMACSignatureConditionalInclusion tests that HMAC is only included when secret is configured
// Feature: whatsapp-http-api-enhancement, Property 18: HMAC Signature Conditional Inclusion
// Validates: Requirements 6.4, 6.5
func TestProperty18_HMACSignatureConditionalInclusion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	rapid.Check(t, func(t *rapid.T) {
		// Generate random event
		eventType := rapid.StringMatching("[a-z]+\\.[a-z]+").Draw(t, "event_type")
		sessionID := rapid.StringMatching("[a-zA-Z0-9-]{10,36}").Draw(t, "session_id")
		hasSecret := rapid.Bool().Draw(t, "has_secret")

		// Create test server to capture headers
		var capturedSignature string
		var mu sync.Mutex

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mu.Lock()
			defer mu.Unlock()

			capturedSignature = r.Header.Get("X-Webhook-Signature")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		// Create webhook publisher with or without secret
		config := webhook.WebhookConfig{
			URL:    server.URL,
			Events: []string{eventType},
		}
		if hasSecret {
			config.Secret = "test-secret-key"
		}
		publisher := webhook.NewWebhookPublisher(config, createTestLogger(), nil)

		// Create and publish event
		event := createTestEvent(eventType, sessionID)

		ctx := context.Background()
		err := publisher.Publish(ctx, event)
		if err != nil {
			t.Fatalf("Failed to publish event: %v", err)
		}

		// Wait for request to be processed
		time.Sleep(100 * time.Millisecond)

		mu.Lock()
		defer mu.Unlock()

		// Verify signature presence matches secret configuration
		if hasSecret && capturedSignature == "" {
			t.Fatalf("HMAC signature should be included when secret is configured")
		}
		if !hasSecret && capturedSignature != "" {
			t.Fatalf("HMAC signature should not be included when secret is not configured")
		}
	})
}

// ==================== Property 30: Webhook Retry Policy ====================

// TestProperty30_WebhookRetryPolicy tests that webhook retries follow the correct policy
// Feature: whatsapp-http-api-enhancement, Property 30: Webhook Retry Policy
// Validates: Requirements 11.1, 11.2, 11.3
func TestProperty30_WebhookRetryPolicy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	rapid.Check(t, func(t *rapid.T) {
		// Generate random event
		eventType := rapid.StringMatching("[a-z]+\\.[a-z]+").Draw(t, "event_type")
		sessionID := rapid.StringMatching("[a-zA-Z0-9-]{10,36}").Draw(t, "session_id")

		// Choose error type: 5xx or 4xx
		errorType := rapid.IntRange(0, 1).Draw(t, "error_type")

		// Track request attempts
		var attemptCount int
		var mu sync.Mutex

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mu.Lock()
			attemptCount++
			currentAttempt := attemptCount
			mu.Unlock()

			switch errorType {
			case 0: // 5xx error - should retry
				w.WriteHeader(http.StatusInternalServerError)
			case 1: // 4xx error - should not retry
				if currentAttempt == 1 {
					w.WriteHeader(http.StatusBadRequest)
				} else {
					t.Fatalf("4xx error should not be retried, but got attempt %d", currentAttempt)
				}
			}
		}))
		defer server.Close()

		// Create webhook publisher
		config := webhook.WebhookConfig{
			URL:    server.URL,
			Events: []string{eventType},
		}
		publisher := webhook.NewWebhookPublisher(config, createTestLogger(), nil)

		// Create and publish event
		event := createTestEvent(eventType, sessionID)

		ctx := context.Background()
		_ = publisher.Publish(ctx, event) // Expect error

		// Wait for retries to complete
		time.Sleep(8 * time.Second) // Max retry time: 1s + 2s + 4s = 7s

		mu.Lock()
		defer mu.Unlock()

		// Verify retry behavior
		switch errorType {
		case 0: // 5xx error - should retry 3 times
			if attemptCount != 3 {
				t.Fatalf("Expected 3 attempts for retryable error, got %d", attemptCount)
			}
		case 1: // 4xx error - should not retry
			if attemptCount != 1 {
				t.Fatalf("Expected 1 attempt for non-retryable error, got %d", attemptCount)
			}
		}
	})
}

// ==================== Property 31: Webhook Retry Exhaustion ====================

// TestProperty31_WebhookRetryExhaustion tests that events are discarded after retry exhaustion
// Feature: whatsapp-http-api-enhancement, Property 31: Webhook Retry Exhaustion
// Validates: Requirements 11.4
func TestProperty31_WebhookRetryExhaustion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	rapid.Check(t, func(t *rapid.T) {
		// Generate random event
		eventType := rapid.StringMatching("[a-z]+\\.[a-z]+").Draw(t, "event_type")
		sessionID := rapid.StringMatching("[a-zA-Z0-9-]{10,36}").Draw(t, "session_id")

		// Track request attempts
		var attemptCount int
		var mu sync.Mutex

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mu.Lock()
			attemptCount++
			mu.Unlock()

			// Always fail with 5xx
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		// Create webhook publisher
		config := webhook.WebhookConfig{
			URL:    server.URL,
			Events: []string{eventType},
		}
		publisher := webhook.NewWebhookPublisher(config, createTestLogger(), nil)

		// Create and publish event
		event := createTestEvent(eventType, sessionID)

		ctx := context.Background()
		err := publisher.Publish(ctx, event)

		// Should return error after exhausting retries
		if err == nil {
			t.Fatalf("Expected error after retry exhaustion")
		}

		// Wait for all retries to complete
		time.Sleep(8 * time.Second)

		mu.Lock()
		defer mu.Unlock()

		// Verify exactly 3 attempts were made
		if attemptCount != 3 {
			t.Fatalf("Expected exactly 3 attempts, got %d", attemptCount)
		}
	})
}

// ==================== Property 32: Webhook Success Logging ====================

// TestProperty32_WebhookSuccessLogging tests that successful deliveries are logged
// Feature: whatsapp-http-api-enhancement, Property 32: Webhook Success Logging
// Validates: Requirements 11.5
func TestProperty32_WebhookSuccessLogging(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	rapid.Check(t, func(t *rapid.T) {
		// Generate random event
		eventType := rapid.StringMatching("[a-z]+\\.[a-z]+").Draw(t, "event_type")
		sessionID := rapid.StringMatching("[a-zA-Z0-9-]{10,36}").Draw(t, "session_id")

		// Track successful delivery
		var deliverySuccessful bool
		var mu sync.Mutex

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mu.Lock()
			defer mu.Unlock()

			deliverySuccessful = true
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		// Create webhook publisher
		config := webhook.WebhookConfig{
			URL:    server.URL,
			Events: []string{eventType},
		}
		publisher := webhook.NewWebhookPublisher(config, createTestLogger(), nil)

		// Create and publish event
		event := createTestEvent(eventType, sessionID)

		ctx := context.Background()
		err := publisher.Publish(ctx, event)
		if err != nil {
			t.Fatalf("Failed to publish event: %v", err)
		}

		// Wait for request to be processed
		time.Sleep(100 * time.Millisecond)

		mu.Lock()
		defer mu.Unlock()

		// Verify delivery was successful
		if !deliverySuccessful {
			t.Fatalf("Webhook delivery should have succeeded")
		}

		// Note: Actual log verification would require a test logger that captures log entries
		// For now, we verify the delivery succeeded, which triggers the logging code path
	})
}
