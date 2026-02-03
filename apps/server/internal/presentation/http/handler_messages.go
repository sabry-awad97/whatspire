package http

import (
	"net/http"
	"time"

	"whatspire/internal/application/dto"
	"whatspire/pkg/validator"

	"github.com/gin-gonic/gin"
)

// SendMessage handles POST /api/messages
func (h *Handler) SendMessage(c *gin.Context) {
	var req dto.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, http.StatusBadRequest, "INVALID_JSON", "Invalid request body", nil)
		return
	}

	if err := validator.Validate(req); err != nil {
		details := validator.ValidationErrors(err)
		respondWithError(c, http.StatusBadRequest, "VALIDATION_FAILED", "Validation failed", details)
		return
	}

	// Additional validation for message content
	if err := req.Validate(); err != nil {
		respondWithError(c, http.StatusBadRequest, "VALIDATION_FAILED", err.Error(), nil)
		return
	}

	msg, err := h.messageUC.SendMessage(c.Request.Context(), req)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusAccepted, map[string]any{
		"message_id": msg.ID,
		"status":     msg.GetStatus().String(),
	})
}

// SendReaction handles POST /api/messages/:messageId/reactions
func (h *Handler) SendReaction(c *gin.Context) {
	messageID := c.Param("messageId")
	if messageID == "" {
		respondWithError(c, http.StatusBadRequest, "INVALID_ID", "Message ID is required", nil)
		return
	}

	var req dto.SendReactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, http.StatusBadRequest, "INVALID_JSON", "Invalid request body", nil)
		return
	}

	// Set message ID from URL parameter
	req.MessageID = messageID

	if err := validator.Validate(req); err != nil {
		details := validator.ValidationErrors(err)
		respondWithError(c, http.StatusBadRequest, "VALIDATION_FAILED", "Validation failed", details)
		return
	}

	if h.reactionUC == nil {
		respondWithError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Reaction use case not configured", nil)
		return
	}

	reaction, err := h.reactionUC.SendReaction(c.Request.Context(), req)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, dto.NewReactionResponse(reaction))
}

// RemoveReaction handles DELETE /api/messages/:messageId/reactions
func (h *Handler) RemoveReaction(c *gin.Context) {
	messageID := c.Param("messageId")
	if messageID == "" {
		respondWithError(c, http.StatusBadRequest, "INVALID_ID", "Message ID is required", nil)
		return
	}

	var req dto.RemoveReactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, http.StatusBadRequest, "INVALID_JSON", "Invalid request body", nil)
		return
	}

	// Set message ID from URL parameter
	req.MessageID = messageID

	if err := validator.Validate(req); err != nil {
		details := validator.ValidationErrors(err)
		respondWithError(c, http.StatusBadRequest, "VALIDATION_FAILED", "Validation failed", details)
		return
	}

	if h.reactionUC == nil {
		respondWithError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Reaction use case not configured", nil)
		return
	}

	err := h.reactionUC.RemoveReaction(c.Request.Context(), req)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, map[string]string{"message": "Reaction removed successfully"})
}

// SendReadReceipt handles POST /api/messages/receipts
func (h *Handler) SendReadReceipt(c *gin.Context) {
	var req dto.SendReceiptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, http.StatusBadRequest, "INVALID_JSON", "Invalid request body", nil)
		return
	}

	if err := validator.Validate(req); err != nil {
		details := validator.ValidationErrors(err)
		respondWithError(c, http.StatusBadRequest, "VALIDATION_FAILED", "Validation failed", details)
		return
	}

	if h.receiptUC == nil {
		respondWithError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Receipt use case not configured", nil)
		return
	}

	err := h.receiptUC.SendReadReceipt(c.Request.Context(), req)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	response := dto.NewReceiptResponse(len(req.MessageIDs), time.Now().Format(time.RFC3339))
	respondWithSuccess(c, http.StatusOK, response)
}

// SendPresence handles POST /api/presence
func (h *Handler) SendPresence(c *gin.Context) {
	var req dto.SendPresenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, http.StatusBadRequest, "INVALID_JSON", "Invalid request body", nil)
		return
	}

	if err := validator.Validate(req); err != nil {
		details := validator.ValidationErrors(err)
		respondWithError(c, http.StatusBadRequest, "VALIDATION_FAILED", "Validation failed", details)
		return
	}

	if h.presenceUC == nil {
		respondWithError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Presence use case not configured", nil)
		return
	}

	err := h.presenceUC.SendPresence(c.Request.Context(), req)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	response := dto.PresenceResponse{
		ChatJID:   req.ChatJID,
		State:     req.State,
		Timestamp: time.Now(),
	}
	respondWithSuccess(c, http.StatusOK, response)
}
