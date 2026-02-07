package http

import (
	"whatspire/internal/application/usecase"
	"whatspire/internal/infrastructure/logger"
)

// Handler defines HTTP handlers for the WhatsApp service
type Handler struct {
	sessionUC  *usecase.SessionUseCase
	messageUC  *usecase.MessageUseCase
	healthUC   *usecase.HealthUseCase
	groupsUC   *usecase.GroupsUseCase
	reactionUC *usecase.ReactionUseCase
	receiptUC  *usecase.ReceiptUseCase
	presenceUC *usecase.PresenceUseCase
	contactUC  *usecase.ContactUseCase
	eventUC    *usecase.EventUseCase
	apikeyUC   *usecase.APIKeyUseCase
	webhookUC  *usecase.WebhookUseCase
	logger     *logger.Logger
}

// HandlerBuilder provides a builder pattern for creating Handler instances
type HandlerBuilder struct {
	handler *Handler
}

// NewHandlerBuilder creates a new HandlerBuilder with a logger
func NewHandlerBuilder(log *logger.Logger) *HandlerBuilder {
	return &HandlerBuilder{
		handler: &Handler{
			logger: log,
		},
	}
}

// WithSessionUseCase sets the session use case
func (b *HandlerBuilder) WithSessionUseCase(uc *usecase.SessionUseCase) *HandlerBuilder {
	b.handler.sessionUC = uc
	return b
}

// WithMessageUseCase sets the message use case
func (b *HandlerBuilder) WithMessageUseCase(uc *usecase.MessageUseCase) *HandlerBuilder {
	b.handler.messageUC = uc
	return b
}

// WithHealthUseCase sets the health use case
func (b *HandlerBuilder) WithHealthUseCase(uc *usecase.HealthUseCase) *HandlerBuilder {
	b.handler.healthUC = uc
	return b
}

// WithGroupsUseCase sets the groups use case
func (b *HandlerBuilder) WithGroupsUseCase(uc *usecase.GroupsUseCase) *HandlerBuilder {
	b.handler.groupsUC = uc
	return b
}

// WithReactionUseCase sets the reaction use case
func (b *HandlerBuilder) WithReactionUseCase(uc *usecase.ReactionUseCase) *HandlerBuilder {
	b.handler.reactionUC = uc
	return b
}

// WithReceiptUseCase sets the receipt use case
func (b *HandlerBuilder) WithReceiptUseCase(uc *usecase.ReceiptUseCase) *HandlerBuilder {
	b.handler.receiptUC = uc
	return b
}

// WithPresenceUseCase sets the presence use case
func (b *HandlerBuilder) WithPresenceUseCase(uc *usecase.PresenceUseCase) *HandlerBuilder {
	b.handler.presenceUC = uc
	return b
}

// WithContactUseCase sets the contact use case
func (b *HandlerBuilder) WithContactUseCase(uc *usecase.ContactUseCase) *HandlerBuilder {
	b.handler.contactUC = uc
	return b
}

// WithEventUseCase sets the event use case
func (b *HandlerBuilder) WithEventUseCase(uc *usecase.EventUseCase) *HandlerBuilder {
	b.handler.eventUC = uc
	return b
}

// WithAPIKeyUseCase sets the API key use case
func (b *HandlerBuilder) WithAPIKeyUseCase(uc *usecase.APIKeyUseCase) *HandlerBuilder {
	b.handler.apikeyUC = uc
	return b
}

// WithWebhookUseCase sets the webhook use case
func (b *HandlerBuilder) WithWebhookUseCase(uc *usecase.WebhookUseCase) *HandlerBuilder {
	b.handler.webhookUC = uc
	return b
}

// Build returns the constructed Handler
func (b *HandlerBuilder) Build() *Handler {
	return b.handler
}
