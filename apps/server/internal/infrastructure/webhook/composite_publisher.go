package webhook

import (
	"context"

	"whatspire/internal/domain/entity"
	"whatspire/internal/domain/repository"
	"whatspire/internal/infrastructure/logger"
)

// CompositeEventPublisher publishes events to multiple destinations (WebSocket + Webhook)
type CompositeEventPublisher struct {
	websocketPublisher repository.EventPublisher
	webhookPublisher   *WebhookPublisher
	logger             *logger.Logger
}

// NewCompositeEventPublisher creates a new composite event publisher
func NewCompositeEventPublisher(
	websocketPublisher repository.EventPublisher,
	webhookPublisher *WebhookPublisher,
	log *logger.Logger,
) *CompositeEventPublisher {
	return &CompositeEventPublisher{
		websocketPublisher: websocketPublisher,
		webhookPublisher:   webhookPublisher,
		logger:             log,
	}
}

// Publish sends an event to both WebSocket and Webhook destinations
func (p *CompositeEventPublisher) Publish(ctx context.Context, event *entity.Event) error {
	// Publish to WebSocket (primary channel)
	if err := p.websocketPublisher.Publish(ctx, event); err != nil {
		p.logger.WithError(err).WithStr("event_type", string(event.Type)).Warn("Failed to publish event to WebSocket")
		// Continue to webhook even if WebSocket fails
	}

	// Publish to Webhook (secondary channel) - don't block on webhook failures
	if p.webhookPublisher != nil {
		go func() {
			// Use background context to avoid cancellation
			bgCtx := context.Background()
			if err := p.webhookPublisher.Publish(bgCtx, event); err != nil {
				p.logger.WithError(err).WithStr("event_type", string(event.Type)).Warn("Failed to publish event to webhook")
			}
		}()
	}

	return nil
}

// Connect establishes the WebSocket connection (webhooks don't need connection)
func (p *CompositeEventPublisher) Connect(ctx context.Context) error {
	return p.websocketPublisher.Connect(ctx)
}

// Disconnect closes the WebSocket connection (webhooks don't need disconnection)
func (p *CompositeEventPublisher) Disconnect(ctx context.Context) error {
	return p.websocketPublisher.Disconnect(ctx)
}

// IsConnected checks if the WebSocket publisher is connected
func (p *CompositeEventPublisher) IsConnected() bool {
	return p.websocketPublisher.IsConnected()
}

// QueueSize returns the number of events waiting in the WebSocket queue
func (p *CompositeEventPublisher) QueueSize() int {
	return p.websocketPublisher.QueueSize()
}
