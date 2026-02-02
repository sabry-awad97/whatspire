package usecase

import (
	"context"

	"whatspire/internal/application/dto"
	"whatspire/internal/domain/entity"
	"whatspire/internal/domain/errors"
	"whatspire/internal/domain/repository"
)

// ContactUseCase handles contact operations business logic
type ContactUseCase struct {
	waClient repository.WhatsAppClient
}

// NewContactUseCase creates a new ContactUseCase
func NewContactUseCase(
	waClient repository.WhatsAppClient,
) *ContactUseCase {
	return &ContactUseCase{
		waClient: waClient,
	}
}

// CheckPhoneNumber checks if a phone number is registered on WhatsApp
func (uc *ContactUseCase) CheckPhoneNumber(ctx context.Context, req dto.CheckPhoneRequest) (*entity.Contact, error) {
	// Check if session is connected
	if !uc.waClient.IsConnected(req.SessionID) {
		return nil, errors.ErrDisconnected.WithMessage("session is not connected")
	}

	// Validate phone number format (basic validation)
	if req.Phone == "" {
		return nil, errors.ErrValidationFailed.WithMessage("phone number is required")
	}

	// Check phone number via WhatsApp client
	contact, err := uc.waClient.CheckPhoneNumber(ctx, req.SessionID, req.Phone)
	if err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to check phone number").WithCause(err)
	}

	return contact, nil
}

// GetUserProfile retrieves the profile information for a user
func (uc *ContactUseCase) GetUserProfile(ctx context.Context, req dto.GetProfileRequest) (*entity.Contact, error) {
	// Check if session is connected
	if !uc.waClient.IsConnected(req.SessionID) {
		return nil, errors.ErrDisconnected.WithMessage("session is not connected")
	}

	// Validate JID
	if req.JID == "" {
		return nil, errors.ErrValidationFailed.WithMessage("jid is required")
	}

	// Get user profile via WhatsApp client
	contact, err := uc.waClient.GetUserProfile(ctx, req.SessionID, req.JID)
	if err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to get user profile").WithCause(err)
	}

	return contact, nil
}

// ListContacts retrieves all contacts for a session
func (uc *ContactUseCase) ListContacts(ctx context.Context, sessionID string) ([]*entity.Contact, error) {
	// Check if session is connected
	if !uc.waClient.IsConnected(sessionID) {
		return nil, errors.ErrDisconnected.WithMessage("session is not connected")
	}

	// List contacts via WhatsApp client
	contacts, err := uc.waClient.ListContacts(ctx, sessionID)
	if err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to list contacts").WithCause(err)
	}

	return contacts, nil
}

// ListChats retrieves all chats for a session
func (uc *ContactUseCase) ListChats(ctx context.Context, sessionID string) ([]*entity.Chat, error) {
	// Check if session is connected
	if !uc.waClient.IsConnected(sessionID) {
		return nil, errors.ErrDisconnected.WithMessage("session is not connected")
	}

	// List chats via WhatsApp client
	chats, err := uc.waClient.ListChats(ctx, sessionID)
	if err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to list chats").WithCause(err)
	}

	return chats, nil
}
