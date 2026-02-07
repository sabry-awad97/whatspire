package persistence

import (
	"context"

	"whatspire/internal/domain/entity"
	domainErrors "whatspire/internal/domain/errors"
	"whatspire/internal/infrastructure/persistence/models"

	"gorm.io/gorm"
)

// ReceiptRepository implements ReceiptRepository with GORM
type ReceiptRepository struct {
	db *gorm.DB
}

// NewReceiptRepository creates a new GORM receipt repository
func NewReceiptRepository(db *gorm.DB) *ReceiptRepository {
	return &ReceiptRepository{db: db}
}

// Save stores a receipt in the repository
func (r *ReceiptRepository) Save(ctx context.Context, receipt *entity.Receipt) error {
	model := &models.Receipt{
		ID:        receipt.ID,
		MessageID: receipt.MessageID,
		SessionID: receipt.SessionID,
		FromJID:   receipt.From,
		ToJID:     receipt.To,
		Type:      receipt.Type.String(),
		CreatedAt: receipt.Timestamp,
	}

	result := r.db.WithContext(ctx).Create(model)
	if result.Error != nil {
		return domainErrors.ErrDatabase.WithCause(result.Error)
	}

	return nil
}

// FindByMessageID retrieves all receipts for a specific message
func (r *ReceiptRepository) FindByMessageID(ctx context.Context, messageID string) ([]*entity.Receipt, error) {
	var modelReceipts []models.Receipt

	result := r.db.WithContext(ctx).
		Where("message_id = ?", messageID).
		Order("created_at DESC").
		Find(&modelReceipts)

	if result.Error != nil {
		return nil, domainErrors.ErrDatabase.WithCause(result.Error)
	}

	// Convert models to domain entities
	receipts := make([]*entity.Receipt, 0, len(modelReceipts))
	for _, model := range modelReceipts {
		receipt := &entity.Receipt{
			ID:        model.ID,
			MessageID: model.MessageID,
			SessionID: model.SessionID,
			From:      model.FromJID,
			To:        model.ToJID,
			Type:      entity.ReceiptType(model.Type),
			Timestamp: model.CreatedAt,
		}
		receipts = append(receipts, receipt)
	}

	return receipts, nil
}

// FindBySessionID retrieves all receipts for a specific session with pagination
func (r *ReceiptRepository) FindBySessionID(ctx context.Context, sessionID string, limit, offset int) ([]*entity.Receipt, error) {
	var modelReceipts []models.Receipt

	result := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&modelReceipts)

	if result.Error != nil {
		return nil, domainErrors.ErrDatabase.WithCause(result.Error)
	}

	// Convert models to domain entities
	receipts := make([]*entity.Receipt, 0, len(modelReceipts))
	for _, model := range modelReceipts {
		receipt := &entity.Receipt{
			ID:        model.ID,
			MessageID: model.MessageID,
			SessionID: model.SessionID,
			From:      model.FromJID,
			To:        model.ToJID,
			Type:      entity.ReceiptType(model.Type),
			Timestamp: model.CreatedAt,
		}
		receipts = append(receipts, receipt)
	}

	return receipts, nil
}

// Delete removes a receipt by its ID
func (r *ReceiptRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&models.Receipt{}, "id = ?", id)

	if result.Error != nil {
		return domainErrors.ErrDatabase.WithCause(result.Error)
	}

	if result.RowsAffected == 0 {
		return domainErrors.ErrNotFound
	}

	return nil
}
