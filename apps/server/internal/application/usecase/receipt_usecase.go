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

// ReceiptUseCase handles receipt business logic
type ReceiptUseCase struct {
	waClient       repository.WhatsAppClient
	receiptRepo    repository.ReceiptRepository
	eventPublisher repository.EventPublisher
}

// NewReceiptUseCase creates a new ReceiptUseCase
func NewReceiptUseCase(
	waClient repository.WhatsAppClient,
	receiptRepo repository.ReceiptRepository,
	eventPublisher repository.EventPublisher,
) *ReceiptUseCase {
	return &ReceiptUseCase{
		waClient:       waClient,
		receiptRepo:    receiptRepo,
		eventPublisher: eventPublisher,
	}
}

// SendReadReceipt sends read receipts for multiple messages atomically
func (uc *ReceiptUseCase) SendReadReceipt(ctx context.Context, req dto.SendReceiptRequest) error {
	// Check if session is connected
	if !uc.waClient.IsConnected(req.SessionID) {
		return errors.ErrDisconnected.WithMessage("session is not connected")
	}

	// Validate message IDs
	if len(req.MessageIDs) == 0 {
		return errors.ErrValidationFailed.WithMessage("at least one message ID is required")
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

	// Send read receipts atomically via WhatsApp client
	err = uc.waClient.SendReadReceipt(ctx, req.SessionID, req.ChatJID, req.MessageIDs)
	if err != nil {
		return errors.ErrMessageSendFailed.WithMessage("failed to send read receipts").WithCause(err)
	}

	// Save receipts to repository and publish events
	for _, messageID := range req.MessageIDs {
		receipt := entity.NewReceipt(
			uuid.New().String(),
			messageID,
			req.SessionID,
			fromJID,
			req.ChatJID,
			entity.ReceiptTypeRead,
		)

		// Save receipt to repository
		if uc.receiptRepo != nil {
			if err := uc.receiptRepo.Save(ctx, receipt); err != nil {
				// Log error but don't fail the request
				// The receipt was sent successfully
			}
		}

		// Publish receipt event
		if uc.eventPublisher != nil && uc.eventPublisher.IsConnected() {
			event, err := entity.NewEventWithPayload(
				uuid.New().String(),
				entity.EventTypeMessageRead,
				req.SessionID,
				receipt,
			)
			if err == nil {
				_ = uc.eventPublisher.Publish(ctx, event)
			}
		}
	}

	return nil
}
