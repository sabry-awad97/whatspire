package usecase

import (
	"context"

	"whatspire/internal/application/dto"
	"whatspire/internal/domain/entity"
	"whatspire/internal/domain/errors"
	"whatspire/internal/domain/repository"

	"github.com/google/uuid"
)

// PresenceUseCase handles presence business logic
type PresenceUseCase struct {
	waClient       repository.WhatsAppClient
	presenceRepo   repository.PresenceRepository
	eventPublisher repository.EventPublisher
}

// NewPresenceUseCase creates a new PresenceUseCase
func NewPresenceUseCase(
	waClient repository.WhatsAppClient,
	presenceRepo repository.PresenceRepository,
	eventPublisher repository.EventPublisher,
) *PresenceUseCase {
	return &PresenceUseCase{
		waClient:       waClient,
		presenceRepo:   presenceRepo,
		eventPublisher: eventPublisher,
	}
}

// SendPresence sends a presence update (typing, paused, etc.)
func (uc *PresenceUseCase) SendPresence(ctx context.Context, req dto.SendPresenceRequest) error {
	// Check if session is connected
	if !uc.waClient.IsConnected(req.SessionID) {
		return errors.ErrDisconnected.WithMessage("session is not connected")
	}

	// Validate presence state
	state := entity.PresenceState(req.State)
	if !state.IsValid() {
		return errors.ErrValidationFailed.WithMessage("invalid presence state")
	}

	// Get session JID
	sessionJID, err := uc.waClient.GetSessionJID(req.SessionID)
	if err != nil {
		return errors.ErrSessionNotFound.WithCause(err)
	}

	// Send presence via WhatsApp client
	err = uc.waClient.SendPresence(ctx, req.SessionID, req.ChatJID, req.State)
	if err != nil {
		return errors.ErrMessageSendFailed.WithMessage("failed to send presence").WithCause(err)
	}

	// Create presence entity
	presence := entity.NewPresence(
		uuid.New().String(),
		req.SessionID,
		sessionJID,
		req.ChatJID,
		state,
	)

	// Save presence to repository
	if uc.presenceRepo != nil {
		if err := uc.presenceRepo.Save(ctx, presence); err != nil {
			// Log error but don't fail the request
			// The presence was sent successfully
		}
	}

	// Publish presence event
	if uc.eventPublisher != nil && uc.eventPublisher.IsConnected() {
		event, err := entity.NewEventWithPayload(
			uuid.New().String(),
			entity.EventTypePresenceUpdate,
			req.SessionID,
			presence,
		)
		if err == nil {
			_ = uc.eventPublisher.Publish(ctx, event)
		}
	}

	return nil
}
