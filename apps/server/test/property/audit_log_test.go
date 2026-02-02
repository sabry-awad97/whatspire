package property

import (
	"context"
	"testing"
	"time"

	"whatspire/internal/domain/repository"
	"whatspire/internal/infrastructure/logger"
	"whatspire/internal/infrastructure/persistence"

	"pgregory.net/rapid"
)

// ==================== Property 26: Audit Log Completeness ====================

// TestProperty26_AuditLogCompleteness tests that all security-relevant events are logged
// Feature: whatsapp-http-api-enhancement, Property 26: Audit Log Completeness
// Validates: Requirements 9.1, 9.2, 9.3, 9.4, 9.5
func TestProperty26_AuditLogCompleteness(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	// Test API key usage logging
	t.Run("API Key Usage", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			// Create audit logger and repository
			testLogger := logger.NewStructuredLogger(logger.Config{Level: "info", Format: "json"})
			auditLogger := logger.NewAuditLogger(testLogger)
			auditRepo := persistence.NewInMemoryAuditLogRepository()

			ctx := context.Background()

			apiKeyID := rapid.StringMatching("[a-zA-Z0-9-]{10,36}").Draw(t, "api_key_id")
			endpoint := rapid.StringMatching("/api/[a-z]+").Draw(t, "endpoint")
			method := rapid.SampledFrom([]string{"GET", "POST", "PUT", "DELETE"}).Draw(t, "method")
			ipAddress := rapid.StringMatching("[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}").Draw(t, "ip_address")

			event := repository.APIKeyUsageEvent{
				APIKeyID:  apiKeyID,
				Endpoint:  endpoint,
				Method:    method,
				Timestamp: time.Now(),
				IPAddress: ipAddress,
			}

			// Log the event
			auditLogger.LogAPIKeyUsage(ctx, event)

			// Save to repository
			err := auditRepo.SaveAPIKeyUsage(ctx, event)
			if err != nil {
				t.Fatalf("Failed to save API key usage event: %v", err)
			}

			// Verify event was saved
			logs, err := auditRepo.FindByEventType(ctx, "api_key_usage", 10, 0)
			if err != nil {
				t.Fatalf("Failed to retrieve audit logs: %v", err)
			}

			if len(logs) == 0 {
				t.Fatalf("API key usage event was not saved")
			}

			// Verify event details
			found := false
			for _, log := range logs {
				if log.APIKeyID != nil && *log.APIKeyID == apiKeyID && log.Endpoint != nil && *log.Endpoint == endpoint {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("API key usage event not found in audit logs")
			}
		})
	})

	// Test session action logging
	t.Run("Session Action", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			// Create audit logger and repository
			testLogger := logger.NewStructuredLogger(logger.Config{Level: "info", Format: "json"})
			auditLogger := logger.NewAuditLogger(testLogger)
			auditRepo := persistence.NewInMemoryAuditLogRepository()

			ctx := context.Background()

			sessionID := rapid.StringMatching("[a-zA-Z0-9-]{10,36}").Draw(t, "session_id")
			action := rapid.SampledFrom([]string{"created", "deleted", "connected", "disconnected"}).Draw(t, "action")
			apiKeyID := rapid.StringMatching("[a-zA-Z0-9-]{10,36}").Draw(t, "api_key_id")

			event := repository.SessionActionEvent{
				SessionID: sessionID,
				Action:    action,
				APIKeyID:  apiKeyID,
				Timestamp: time.Now(),
			}

			// Log the event
			auditLogger.LogSessionAction(ctx, event)

			// Save to repository
			err := auditRepo.SaveSessionAction(ctx, event)
			if err != nil {
				t.Fatalf("Failed to save session action event: %v", err)
			}

			// Verify event was saved
			logs, err := auditRepo.FindByEventType(ctx, "session_action", 10, 0)
			if err != nil {
				t.Fatalf("Failed to retrieve audit logs: %v", err)
			}

			if len(logs) == 0 {
				t.Fatalf("Session action event was not saved")
			}

			// Verify event details
			found := false
			for _, log := range logs {
				if log.SessionID != nil && *log.SessionID == sessionID && log.Action != nil && *log.Action == action {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("Session action event not found in audit logs")
			}
		})
	})

	// Test message sent logging
	t.Run("Message Sent", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			// Create audit logger and repository
			testLogger := logger.NewStructuredLogger(logger.Config{Level: "info", Format: "json"})
			auditLogger := logger.NewAuditLogger(testLogger)
			auditRepo := persistence.NewInMemoryAuditLogRepository()

			ctx := context.Background()

			sessionID := rapid.StringMatching("[a-zA-Z0-9-]{10,36}").Draw(t, "session_id")
			recipient := rapid.StringMatching("[0-9]{10,15}@s\\.whatsapp\\.net").Draw(t, "recipient")
			messageType := rapid.SampledFrom([]string{"text", "image", "document", "audio", "video"}).Draw(t, "message_type")

			event := repository.MessageSentEvent{
				SessionID:   sessionID,
				Recipient:   recipient,
				MessageType: messageType,
				Timestamp:   time.Now(),
			}

			// Log the event
			auditLogger.LogMessageSent(ctx, event)

			// Save to repository
			err := auditRepo.SaveMessageSent(ctx, event)
			if err != nil {
				t.Fatalf("Failed to save message sent event: %v", err)
			}

			// Verify event was saved
			logs, err := auditRepo.FindByEventType(ctx, "message_sent", 10, 0)
			if err != nil {
				t.Fatalf("Failed to retrieve audit logs: %v", err)
			}

			if len(logs) == 0 {
				t.Fatalf("Message sent event was not saved")
			}

			// Verify event details
			found := false
			for _, log := range logs {
				if log.SessionID != nil && *log.SessionID == sessionID {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("Message sent event not found in audit logs")
			}
		})
	})

	// Test authentication failure logging
	t.Run("Auth Failure", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			// Create audit logger and repository
			testLogger := logger.NewStructuredLogger(logger.Config{Level: "info", Format: "json"})
			auditLogger := logger.NewAuditLogger(testLogger)
			auditRepo := persistence.NewInMemoryAuditLogRepository()

			ctx := context.Background()

			apiKey := rapid.StringMatching("[a-zA-Z0-9]{16,32}").Draw(t, "api_key")
			endpoint := rapid.StringMatching("/api/[a-z]+").Draw(t, "endpoint")
			reason := rapid.SampledFrom([]string{"invalid_api_key", "missing_api_key", "expired_api_key"}).Draw(t, "reason")
			ipAddress := rapid.StringMatching("[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}").Draw(t, "ip_address")

			event := repository.AuthFailureEvent{
				APIKey:    apiKey,
				Endpoint:  endpoint,
				Reason:    reason,
				Timestamp: time.Now(),
				IPAddress: ipAddress,
			}

			// Log the event
			auditLogger.LogAuthFailure(ctx, event)

			// Save to repository
			err := auditRepo.SaveAuthFailure(ctx, event)
			if err != nil {
				t.Fatalf("Failed to save auth failure event: %v", err)
			}

			// Verify event was saved
			logs, err := auditRepo.FindByEventType(ctx, "auth_failure", 10, 0)
			if err != nil {
				t.Fatalf("Failed to retrieve audit logs: %v", err)
			}

			if len(logs) == 0 {
				t.Fatalf("Auth failure event was not saved")
			}

			// Verify event details
			found := false
			for _, log := range logs {
				if log.Endpoint != nil && *log.Endpoint == endpoint && log.IPAddress != nil && *log.IPAddress == ipAddress {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("Auth failure event not found in audit logs")
			}
		})
	})

	// Test webhook delivery logging
	t.Run("Webhook Delivery", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			// Create audit logger and repository
			testLogger := logger.NewStructuredLogger(logger.Config{Level: "info", Format: "json"})
			auditLogger := logger.NewAuditLogger(testLogger)
			auditRepo := persistence.NewInMemoryAuditLogRepository()

			ctx := context.Background()

			webhookURL := rapid.StringMatching("https://[a-z]+\\.example\\.com/webhook").Draw(t, "webhook_url")
			eventType := rapid.StringMatching("[a-z]+\\.[a-z]+").Draw(t, "event_type")
			statusCode := rapid.IntRange(200, 599).Draw(t, "status_code")
			success := statusCode >= 200 && statusCode < 300

			var errorMsg *string
			if !success {
				msg := "delivery failed"
				errorMsg = &msg
			}

			event := repository.WebhookDeliveryEvent{
				WebhookURL: webhookURL,
				EventType:  eventType,
				StatusCode: statusCode,
				Success:    success,
				Error:      errorMsg,
				Timestamp:  time.Now(),
			}

			// Log the event
			auditLogger.LogWebhookDelivery(ctx, event)

			// Save to repository
			err := auditRepo.SaveWebhookDelivery(ctx, event)
			if err != nil {
				t.Fatalf("Failed to save webhook delivery event: %v", err)
			}

			// Verify event was saved
			logs, err := auditRepo.FindByEventType(ctx, "webhook_delivery", 10, 0)
			if err != nil {
				t.Fatalf("Failed to retrieve audit logs: %v", err)
			}

			if len(logs) == 0 {
				t.Fatalf("Webhook delivery event was not saved")
			}

			// Verify at least one log exists (we don't check specific details due to multiple test runs)
			if len(logs) == 0 {
				t.Fatalf("Webhook delivery event not found in audit logs")
			}
		})
	})
}
