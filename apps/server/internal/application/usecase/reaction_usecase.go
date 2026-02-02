package usecase

import (
	"context"
	"strings"

	"whatspire/internal/application/dto"
	"whatspire/internal/domain/entity"
	"whatspire/internal/domain/errors"
	"whatspire/internal/domain/repository"

	"github.com/google/uuid"
)

// ReactionUseCase handles reaction business logic
type ReactionUseCase struct {
	waClient       repository.WhatsAppClient
	reactionRepo   repository.ReactionRepository
	eventPublisher repository.EventPublisher
}

// NewReactionUseCase creates a new ReactionUseCase
func NewReactionUseCase(
	waClient repository.WhatsAppClient,
	reactionRepo repository.ReactionRepository,
	eventPublisher repository.EventPublisher,
) *ReactionUseCase {
	return &ReactionUseCase{
		waClient:       waClient,
		reactionRepo:   reactionRepo,
		eventPublisher: eventPublisher,
	}
}

// SendReaction sends a reaction to a message
func (uc *ReactionUseCase) SendReaction(ctx context.Context, req dto.SendReactionRequest) (*entity.Reaction, error) {
	// Check if session is connected
	if !uc.waClient.IsConnected(req.SessionID) {
		return nil, errors.ErrDisconnected.WithMessage("session is not connected")
	}

	// Create reaction entity for validation
	reaction := entity.NewReaction(
		uuid.New().String(),
		req.MessageID,
		req.SessionID,
		"", // From will be set after sending
		req.ChatJID,
		req.Emoji,
	)

	// Validate emoji
	if !reaction.IsValidEmoji() {
		return nil, errors.ErrValidationFailed.WithMessage("invalid emoji format")
	}

	// Get session JID
	sessionJID, err := uc.waClient.GetSessionJID(req.SessionID)
	if err != nil {
		return nil, errors.ErrSessionNotFound.WithCause(err)
	}

	// Extract user part from session JID
	fromJID := sessionJID
	if idx := strings.Index(sessionJID, "@"); idx > 0 {
		fromJID = sessionJID[:idx]
	}
	reaction.From = fromJID

	// Send reaction via WhatsApp client
	err = uc.waClient.SendReaction(ctx, req.SessionID, req.ChatJID, req.MessageID, req.Emoji)
	if err != nil {
		return nil, errors.ErrMessageSendFailed.WithMessage("failed to send reaction").WithCause(err)
	}

	// Save reaction to repository
	if uc.reactionRepo != nil {
		if err := uc.reactionRepo.Save(ctx, reaction); err != nil {
			// Log error but don't fail the request
			// The reaction was sent successfully
		}
	}

	// Publish reaction event
	if uc.eventPublisher != nil && uc.eventPublisher.IsConnected() {
		event, err := entity.NewEventWithPayload(
			uuid.New().String(),
			entity.EventTypeMessageReaction,
			req.SessionID,
			reaction,
		)
		if err == nil {
			_ = uc.eventPublisher.Publish(ctx, event)
		}
	}

	return reaction, nil
}

// RemoveReaction removes a reaction from a message
func (uc *ReactionUseCase) RemoveReaction(ctx context.Context, req dto.RemoveReactionRequest) error {
	// Check if session is connected
	if !uc.waClient.IsConnected(req.SessionID) {
		return errors.ErrDisconnected.WithMessage("session is not connected")
	}

	// Get session JID
	sessionJID, err := uc.waClient.GetSessionJID(req.SessionID)
	if err != nil {
		return errors.ErrSessionNotFound.WithCause(err)
	}

	// Extract user part from session JID
	fromJID := sessionJID
	if idx := strings.Index(sessionJID, "@"); idx > 0 {
		fromJID = sessionJID[:idx]
	}

	// Send reaction removal via WhatsApp client
	err = uc.waClient.SendReaction(ctx, req.SessionID, req.ChatJID, req.MessageID, "")
	if err != nil {
		return errors.ErrMessageSendFailed.WithMessage("failed to remove reaction").WithCause(err)
	}

	// Delete reaction from repository
	if uc.reactionRepo != nil {
		if err := uc.reactionRepo.DeleteByMessageIDAndFrom(ctx, req.MessageID, fromJID); err != nil {
			// Log error but don't fail the request
		}
	}

	// Publish reaction removal event
	if uc.eventPublisher != nil && uc.eventPublisher.IsConnected() {
		event, err := entity.NewEventWithPayload(
			uuid.New().String(),
			entity.EventTypeMessageReaction,
			req.SessionID,
			map[string]interface{}{
				"message_id": req.MessageID,
				"from":       fromJID,
				"to":         req.ChatJID,
				"emoji":      "",
				"removed":    true,
			},
		)
		if err == nil {
			_ = uc.eventPublisher.Publish(ctx, event)
		}
	}

	return nil
}
