package dto

// SendReceiptRequest represents a request to send read receipts for messages
type SendReceiptRequest struct {
	SessionID  string   `json:"session_id" validate:"required,uuid"`
	ChatJID    string   `json:"chat_jid" validate:"required"`
	MessageIDs []string `json:"message_ids" validate:"required,min=1,dive,required"`
}

// ReceiptResponse represents the response after sending receipts
type ReceiptResponse struct {
	ProcessedCount int    `json:"processed_count"`
	Timestamp      string `json:"timestamp"`
}

// NewReceiptResponse creates a ReceiptResponse
func NewReceiptResponse(count int, timestamp string) *ReceiptResponse {
	return &ReceiptResponse{
		ProcessedCount: count,
		Timestamp:      timestamp,
	}
}
