package repository

import (
	"context"

	"whatspire/internal/domain/entity"
)

// ReceiptRepository defines receipt persistence operations
type ReceiptRepository interface {
	// Save stores a receipt in the repository
	Save(ctx context.Context, receipt *entity.Receipt) error

	// FindByMessageID retrieves all receipts for a specific message
	FindByMessageID(ctx context.Context, messageID string) ([]*entity.Receipt, error)

	// FindBySessionID retrieves all receipts for a specific session
	FindBySessionID(ctx context.Context, sessionID string, limit, offset int) ([]*entity.Receipt, error)

	// Delete removes a receipt by its ID
	Delete(ctx context.Context, id string) error
}
