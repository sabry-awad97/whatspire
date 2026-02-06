package whatsapp

import (
	"context"

	"whatspire/internal/domain/entity"
	"whatspire/internal/domain/repository"
	"whatspire/internal/infrastructure/logger"

	"github.com/google/uuid"
)

// ReactionHandler handles incoming WhatsApp reaction events
type ReactionHandler struct {
	reactionRepo   repository.ReactionRepository
	eventPublisher repository.EventPublisher
	logger         *logger.Logger
}

// NewReactionHandler creates a new reaction handler
func NewReactionHandler(
	reactionRepo repository.ReactionRepository,
	eventPublisher repository.EventPublisher,
	log *logger.Logger,
) *ReactionHandler {
	return &ReactionHandler{
		reactionRepo:   reactionRepo,
		eventPublisher: eventPublisher,
		logger:         log,
	}
}

// HandleIncomingReaction processes an incoming WhatsApp reaction from a parsed message
func (h *ReactionHandler) HandleIncomingReaction(
	ctx context.Context,
	sessionID string,
	parsedMsg *ParsedMessage,
) error {
	// Validate this is a reaction message
	if parsedMsg.MessageType != ParsedMessageTypeReaction {
		return nil
	}

	// Extract reaction data
	if parsedMsg.ReactionMessageID == nil || parsedMsg.ReactionEmoji == nil {
		h.logger.Warnf("Reaction message missing required fields")
		return nil
	}

	// Create reaction entity
	reaction := entity.NewReactionBuilder(uuid.New().String(), *parsedMsg.ReactionMessageID, sessionID).
		From(parsedMsg.SenderJID).
		To(parsedMsg.ChatJID).
		WithEmoji(*parsedMsg.ReactionEmoji).
		Build()

	// Validate reaction
	if !reaction.IsValid() {
		h.logger.Warnf("Invalid reaction received: %+v", reaction)
		return nil // Don't fail, just skip invalid reactions
	}

	// Save reaction to repository
	if h.reactionRepo != nil {
		if err := h.reactionRepo.Save(ctx, reaction); err != nil {
			h.logger.Warnf("Failed to save reaction: %v", err)
			// Continue to publish event even if save fails
		}
	}

	// Publish reaction event
	if h.eventPublisher != nil && h.eventPublisher.IsConnected() {
		event, err := entity.NewEventWithPayload(
			uuid.New().String(),
			entity.EventTypeMessageReaction,
			sessionID,
			reaction,
		)
		if err == nil {
			if err := h.eventPublisher.Publish(ctx, event); err != nil {
				h.logger.Warnf("Failed to publish reaction event: %v", err)
			}
		}
	}

	return nil
}
