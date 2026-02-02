package persistence

import (
	"context"
	"sync"

	"whatspire/internal/domain/entity"
	"whatspire/internal/domain/errors"
)

// InMemoryReceiptRepository implements ReceiptRepository with in-memory storage
type InMemoryReceiptRepository struct {
	receipts        map[string]*entity.Receipt // ID -> Receipt
	messageReceipts map[string][]string        // MessageID -> []ReceiptID
	sessionReceipts map[string][]string        // SessionID -> []ReceiptID
	mu              sync.RWMutex
}

// NewInMemoryReceiptRepository creates a new in-memory receipt repository
func NewInMemoryReceiptRepository() *InMemoryReceiptRepository {
	return &InMemoryReceiptRepository{
		receipts:        make(map[string]*entity.Receipt),
		messageReceipts: make(map[string][]string),
		sessionReceipts: make(map[string][]string),
	}
}

// Save stores a receipt in the repository
func (r *InMemoryReceiptRepository) Save(ctx context.Context, receipt *entity.Receipt) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Store a copy to prevent external mutation
	receiptCopy := *receipt
	r.receipts[receipt.ID] = &receiptCopy

	// Index by message ID
	if _, exists := r.messageReceipts[receipt.MessageID]; !exists {
		r.messageReceipts[receipt.MessageID] = []string{}
	}
	// Check if already indexed
	found := false
	for _, id := range r.messageReceipts[receipt.MessageID] {
		if id == receipt.ID {
			found = true
			break
		}
	}
	if !found {
		r.messageReceipts[receipt.MessageID] = append(r.messageReceipts[receipt.MessageID], receipt.ID)
	}

	// Index by session ID
	if _, exists := r.sessionReceipts[receipt.SessionID]; !exists {
		r.sessionReceipts[receipt.SessionID] = []string{}
	}
	// Check if already indexed
	found = false
	for _, id := range r.sessionReceipts[receipt.SessionID] {
		if id == receipt.ID {
			found = true
			break
		}
	}
	if !found {
		r.sessionReceipts[receipt.SessionID] = append(r.sessionReceipts[receipt.SessionID], receipt.ID)
	}

	return nil
}

// FindByMessageID retrieves all receipts for a specific message
func (r *InMemoryReceiptRepository) FindByMessageID(ctx context.Context, messageID string) ([]*entity.Receipt, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	receiptIDs, exists := r.messageReceipts[messageID]
	if !exists {
		return []*entity.Receipt{}, nil
	}

	receipts := make([]*entity.Receipt, 0, len(receiptIDs))
	for _, id := range receiptIDs {
		if receipt, exists := r.receipts[id]; exists {
			receiptCopy := *receipt
			receipts = append(receipts, &receiptCopy)
		}
	}

	return receipts, nil
}

// FindBySessionID retrieves all receipts for a specific session with pagination
func (r *InMemoryReceiptRepository) FindBySessionID(ctx context.Context, sessionID string, limit, offset int) ([]*entity.Receipt, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	receiptIDs, exists := r.sessionReceipts[sessionID]
	if !exists {
		return []*entity.Receipt{}, nil
	}

	// Apply pagination
	start := offset
	if start >= len(receiptIDs) {
		return []*entity.Receipt{}, nil
	}

	end := start + limit
	if end > len(receiptIDs) {
		end = len(receiptIDs)
	}

	receipts := make([]*entity.Receipt, 0, end-start)
	for i := start; i < end; i++ {
		if receipt, exists := r.receipts[receiptIDs[i]]; exists {
			receiptCopy := *receipt
			receipts = append(receipts, &receiptCopy)
		}
	}

	return receipts, nil
}

// Delete removes a receipt by its ID
func (r *InMemoryReceiptRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	receipt, exists := r.receipts[id]
	if !exists {
		return errors.ErrNotFound
	}

	// Remove from message index
	if receiptIDs, exists := r.messageReceipts[receipt.MessageID]; exists {
		newIDs := make([]string, 0, len(receiptIDs))
		for _, rid := range receiptIDs {
			if rid != id {
				newIDs = append(newIDs, rid)
			}
		}
		r.messageReceipts[receipt.MessageID] = newIDs
	}

	// Remove from session index
	if receiptIDs, exists := r.sessionReceipts[receipt.SessionID]; exists {
		newIDs := make([]string, 0, len(receiptIDs))
		for _, rid := range receiptIDs {
			if rid != id {
				newIDs = append(newIDs, rid)
			}
		}
		r.sessionReceipts[receipt.SessionID] = newIDs
	}

	// Remove from main storage
	delete(r.receipts, id)
	return nil
}

// Clear removes all receipts (for testing)
func (r *InMemoryReceiptRepository) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.receipts = make(map[string]*entity.Receipt)
	r.messageReceipts = make(map[string][]string)
	r.sessionReceipts = make(map[string][]string)
}
