package entity

import (
	"encoding/json"
	"time"
)

// ReceiptType represents the type of receipt
type ReceiptType string

const (
	ReceiptTypeDelivered ReceiptType = "delivered"
	ReceiptTypeRead      ReceiptType = "read"
)

// IsValid checks if the receipt type is valid
func (rt ReceiptType) IsValid() bool {
	switch rt {
	case ReceiptTypeDelivered, ReceiptTypeRead:
		return true
	}
	return false
}

// String returns the string representation of the receipt type
func (rt ReceiptType) String() string {
	return string(rt)
}

// Receipt represents a WhatsApp message receipt (delivered or read)
type Receipt struct {
	ID        string      `json:"id"`
	MessageID string      `json:"message_id"`
	SessionID string      `json:"session_id"`
	From      string      `json:"from"`
	To        string      `json:"to"`
	Type      ReceiptType `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
}

// ReceiptBuilder provides a builder pattern for creating Receipt instances
type ReceiptBuilder struct {
	receipt *Receipt
}

// NewReceiptBuilder creates a new ReceiptBuilder with required fields
func NewReceiptBuilder(id, messageID, sessionID string) *ReceiptBuilder {
	return &ReceiptBuilder{
		receipt: &Receipt{
			ID:        id,
			MessageID: messageID,
			SessionID: sessionID,
			Timestamp: time.Now(),
		},
	}
}

// From sets the sender
func (b *ReceiptBuilder) From(from string) *ReceiptBuilder {
	b.receipt.From = from
	return b
}

// To sets the recipient
func (b *ReceiptBuilder) To(to string) *ReceiptBuilder {
	b.receipt.To = to
	return b
}

// WithType sets the receipt type
func (b *ReceiptBuilder) WithType(receiptType ReceiptType) *ReceiptBuilder {
	b.receipt.Type = receiptType
	return b
}

// WithTimestamp sets a custom timestamp (optional, defaults to now)
func (b *ReceiptBuilder) WithTimestamp(timestamp time.Time) *ReceiptBuilder {
	b.receipt.Timestamp = timestamp
	return b
}

// Build returns the constructed Receipt
func (b *ReceiptBuilder) Build() *Receipt {
	return b.receipt
}

// IsValid checks if the receipt is valid
func (r *Receipt) IsValid() bool {
	if r.ID == "" || r.MessageID == "" || r.SessionID == "" {
		return false
	}
	if r.From == "" || r.To == "" {
		return false
	}
	return r.Type.IsValid()
}

// IsDelivered returns true if this is a delivered receipt
func (r *Receipt) IsDelivered() bool {
	return r.Type == ReceiptTypeDelivered
}

// IsRead returns true if this is a read receipt
func (r *Receipt) IsRead() bool {
	return r.Type == ReceiptTypeRead
}

// MarshalJSON implements json.Marshaler
func (r *Receipt) MarshalJSON() ([]byte, error) {
	type Alias Receipt
	return json.Marshal(&struct {
		*Alias
		Timestamp string `json:"timestamp"`
	}{
		Alias:     (*Alias)(r),
		Timestamp: r.Timestamp.Format(time.RFC3339),
	})
}
