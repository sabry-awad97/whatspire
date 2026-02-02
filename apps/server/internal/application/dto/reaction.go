package dto

import "whatspire/internal/domain/entity"

// SendReactionRequest represents a request to send a reaction to a message
type SendReactionRequest struct {
	SessionID string `json:"session_id" validate:"required,uuid"`
	ChatJID   string `json:"chat_jid" validate:"required"`
	MessageID string `json:"message_id" validate:"required"`
	Emoji     string `json:"emoji" validate:"required"`
}

// RemoveReactionRequest represents a request to remove a reaction from a message
type RemoveReactionRequest struct {
	SessionID string `json:"session_id" validate:"required,uuid"`
	ChatJID   string `json:"chat_jid" validate:"required"`
	MessageID string `json:"message_id" validate:"required"`
}

// ReactionResponse represents the response after sending a reaction
type ReactionResponse struct {
	MessageID string `json:"message_id"`
	Emoji     string `json:"emoji"`
	Timestamp string `json:"timestamp"`
}

// NewReactionResponse creates a ReactionResponse from a Reaction entity
func NewReactionResponse(reaction *entity.Reaction) *ReactionResponse {
	return &ReactionResponse{
		MessageID: reaction.MessageID,
		Emoji:     reaction.Emoji,
		Timestamp: reaction.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
	}
}
