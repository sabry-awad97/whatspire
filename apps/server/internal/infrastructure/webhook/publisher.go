package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"whatspire/internal/domain/entity"
	"whatspire/internal/infrastructure/logger"
)

// WebhookConfig defines configuration for webhook delivery
type WebhookConfig struct {
	URL    string   // Webhook endpoint URL
	Secret string   // Secret for HMAC signing (optional)
	Events []string // Event types to deliver (e.g., "message.received", "message.reaction")
}

// WebhookPublisher publishes events to external webhooks with HMAC security
type WebhookPublisher struct {
	config     WebhookConfig
	httpClient *http.Client
	logger     logger.Logger
}

// NewWebhookPublisher creates a new webhook publisher
func NewWebhookPublisher(config WebhookConfig, logger logger.Logger) *WebhookPublisher {
	return &WebhookPublisher{
		config: config,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// Publish sends an event to the configured webhook URL
func (wp *WebhookPublisher) Publish(ctx context.Context, event *entity.Event) error {
	// Check if this event type should be delivered
	if !wp.shouldPublish(event.Type) {
		return nil
	}

	// Serialize event to JSON
	payload, err := json.Marshal(event)
	if err != nil {
		wp.logger.Error("Failed to marshal event", logger.Err(err), logger.String("event_type", string(event.Type)))
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, wp.config.URL, bytes.NewReader(payload))
	if err != nil {
		wp.logger.Error("Failed to create webhook request", logger.Err(err), logger.String("url", wp.config.URL))
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Add timestamp header
	timestamp := time.Now().Unix()
	req.Header.Set("X-Webhook-Timestamp", fmt.Sprintf("%d", timestamp))

	// Add HMAC signature if secret is configured
	if wp.config.Secret != "" {
		signature := wp.computeHMAC(payload)
		req.Header.Set("X-Webhook-Signature", signature)
	}

	// Send with retry logic
	return wp.sendWithRetry(ctx, req, payload)
}

// computeHMAC computes HMAC-SHA256 signature for the payload
func (wp *WebhookPublisher) computeHMAC(payload []byte) string {
	mac := hmac.New(sha256.New, []byte(wp.config.Secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// shouldPublish checks if the event type should be delivered
func (wp *WebhookPublisher) shouldPublish(eventType entity.EventType) bool {
	// If no events configured, publish all
	if len(wp.config.Events) == 0 {
		return true
	}

	// Check if event type is in the configured list
	for _, configuredEvent := range wp.config.Events {
		if configuredEvent == string(eventType) {
			return true
		}
	}

	return false
}

// sendWithRetry sends the webhook request with exponential backoff retry logic
func (wp *WebhookPublisher) sendWithRetry(ctx context.Context, req *http.Request, payload []byte) error {
	// Exponential backoff: 1s, 2s, 4s
	backoff := []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second}
	maxAttempts := 3

	var lastErr error
	var lastStatusCode int

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Clone request for retry (body needs to be reset)
		reqClone, err := http.NewRequestWithContext(ctx, req.Method, req.URL.String(), bytes.NewReader(payload))
		if err != nil {
			return fmt.Errorf("failed to clone request: %w", err)
		}
		reqClone.Header = req.Header.Clone()

		// Send request
		resp, err := wp.httpClient.Do(reqClone)

		// Success case
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			resp.Body.Close()
			wp.logger.Info("Webhook delivered successfully",
				logger.String("url", wp.config.URL),
				logger.Int("status_code", resp.StatusCode),
				logger.Int("attempt", attempt+1))
			return nil
		}

		// Non-retryable error (4xx status codes)
		if err == nil && resp.StatusCode >= 400 && resp.StatusCode < 500 {
			resp.Body.Close()
			wp.logger.Warn("Webhook delivery failed with client error (no retry)",
				logger.String("url", wp.config.URL),
				logger.Int("status_code", resp.StatusCode),
				logger.Int("attempt", attempt+1))
			return fmt.Errorf("webhook delivery failed with status %d", resp.StatusCode)
		}

		// Retryable error (network error or 5xx status)
		if resp != nil {
			lastStatusCode = resp.StatusCode
			resp.Body.Close()
		}
		lastErr = err

		// Log retry attempt
		if attempt < maxAttempts-1 {
			fields := []logger.Field{
				logger.String("url", wp.config.URL),
				logger.Int("status_code", lastStatusCode),
				logger.Int("attempt", attempt+1),
				logger.Duration("retry_after", backoff[attempt].Seconds()*1000),
			}
			if lastErr != nil {
				fields = append(fields, logger.Err(lastErr))
			}
			wp.logger.Warn("Webhook delivery failed, retrying", fields...)

			// Wait before retry
			select {
			case <-time.After(backoff[attempt]):
				// Continue to next attempt
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	// All retries exhausted
	fields := []logger.Field{
		logger.String("url", wp.config.URL),
		logger.Int("status_code", lastStatusCode),
		logger.Int("attempts", maxAttempts),
	}
	if lastErr != nil {
		fields = append(fields, logger.Err(lastErr))
	}
	wp.logger.Error("Webhook delivery failed after all retries", fields...)

	if lastErr != nil {
		return fmt.Errorf("webhook delivery failed after %d attempts: %w", maxAttempts, lastErr)
	}
	return fmt.Errorf("webhook delivery failed after %d attempts with status %d", maxAttempts, lastStatusCode)
}
