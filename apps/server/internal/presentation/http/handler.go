package http

import (
	"whatspire/internal/application/usecase"
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
}

// NewHandler creates a new Handler with all use cases
func NewHandler(sessionUC *usecase.SessionUseCase, messageUC *usecase.MessageUseCase, healthUC *usecase.HealthUseCase, groupsUC *usecase.GroupsUseCase, reactionUC *usecase.ReactionUseCase, receiptUC *usecase.ReceiptUseCase, presenceUC *usecase.PresenceUseCase, contactUC *usecase.ContactUseCase, eventUC *usecase.EventUseCase) *Handler {
	return &Handler{
		sessionUC:  sessionUC,
		messageUC:  messageUC,
		healthUC:   healthUC,
		groupsUC:   groupsUC,
		reactionUC: reactionUC,
		receiptUC:  receiptUC,
		presenceUC: presenceUC,
		contactUC:  contactUC,
		eventUC:    eventUC,
	}
}
