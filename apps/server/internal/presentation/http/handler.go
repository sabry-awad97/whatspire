package http

import (
	"log"
	"net/http"
	"time"

	"whatspire/internal/application/dto"
	"whatspire/internal/application/usecase"
	"whatspire/internal/domain/entity"
	"whatspire/internal/domain/errors"
	"whatspire/pkg/validator"

	"github.com/gin-gonic/gin"
)

// Handler defines HTTP handlers for the WhatsApp service
type Handler struct {
	sessionUC  *usecase.SessionUseCase
	messageUC  *usecase.MessageUseCase
	healthUC   *usecase.HealthUseCase
	groupsUC   *usecase.GroupsUseCase
	reactionUC *usecase.ReactionUseCase
	receiptUC  *usecase.ReceiptUseCase
	presenceUC *usecase.PresenceUseCase
	contactUC  *usecase.ContactUseCase
}

// NewHandler creates a new Handler with all use cases
func NewHandler(sessionUC *usecase.SessionUseCase, messageUC *usecase.MessageUseCase, healthUC *usecase.HealthUseCase, groupsUC *usecase.GroupsUseCase, reactionUC *usecase.ReactionUseCase, receiptUC *usecase.ReceiptUseCase, presenceUC *usecase.PresenceUseCase, contactUC *usecase.ContactUseCase) *Handler {
	return &Handler{
		sessionUC:  sessionUC,
		messageUC:  messageUC,
		healthUC:   healthUC,
		groupsUC:   groupsUC,
		reactionUC: reactionUC,
		receiptUC:  receiptUC,
		presenceUC: presenceUC,
		contactUC:  contactUC,
	}
}

// SyncGroups handles POST /api/sessions/:id/groups/sync
func (h *Handler) SyncGroups(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		respondWithError(c, http.StatusBadRequest, "INVALID_ID", "Session ID is required", nil)
		return
	}

	if h.groupsUC == nil {
		respondWithError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Groups use case not configured", nil)
		return
	}

	result, err := h.groupsUC.SyncGroups(c.Request.Context(), sessionID)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, result)
}

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

// RegisterSession handles POST /api/internal/sessions/register
// Called by Node.js API when a new session is created
func (h *Handler) RegisterSession(c *gin.Context) {
	var req struct {
		ID   string `json:"id" binding:"required"`
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, http.StatusBadRequest, "INVALID_JSON", "Invalid request body", nil)
		return
	}

	// Create session in local repository for WhatsApp client tracking
	session, err := h.sessionUC.CreateSessionWithID(c.Request.Context(), req.ID, req.Name)
	if err != nil {
		// Session might already exist, which is fine
		handleDomainError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusCreated, dto.NewSessionResponse(session))
}

// UnregisterSession handles POST /api/internal/sessions/:id/unregister
// Called by Node.js API when a session is deleted
func (h *Handler) UnregisterSession(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		respondWithError(c, http.StatusBadRequest, "INVALID_ID", "Session ID is required", nil)
		return
	}

	// Disconnect and cleanup WhatsApp client resources
	if err := h.sessionUC.DeleteSession(c.Request.Context(), id); err != nil {
		// Ignore not found errors - session might not exist locally
		if !errors.IsNotFound(err) {
			handleDomainError(c, err)
			return
		}
	}

	respondWithSuccess(c, http.StatusOK, map[string]string{"message": "Session unregistered successfully"})
}

// UpdateSessionStatus handles POST /api/internal/sessions/:id/status
// Called internally when WhatsApp connection status changes
func (h *Handler) UpdateSessionStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		respondWithError(c, http.StatusBadRequest, "INVALID_ID", "Session ID is required", nil)
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
		JID    string `json:"jid,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, http.StatusBadRequest, "INVALID_JSON", "Invalid request body", nil)
		return
	}

	status := entity.Status(req.Status)
	if !status.IsValid() {
		respondWithError(c, http.StatusBadRequest, "INVALID_STATUS", "Invalid status value", nil)
		return
	}

	if err := h.sessionUC.UpdateSessionStatus(c.Request.Context(), id, status); err != nil {
		handleDomainError(c, err)
		return
	}

	// Update JID if provided
	if req.JID != "" {
		if err := h.sessionUC.UpdateSessionJID(c.Request.Context(), id, req.JID); err != nil {
			handleDomainError(c, err)
			return
		}
	}

	respondWithSuccess(c, http.StatusOK, map[string]string{"message": "Status updated successfully"})
}

// ReconnectSession handles POST /api/internal/sessions/:id/reconnect
// Attempts to reconnect a session using stored WhatsApp credentials
func (h *Handler) ReconnectSession(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		respondWithError(c, http.StatusBadRequest, "INVALID_ID", "Session ID is required", nil)
		return
	}

	// Parse optional JID from request body
	var req struct {
		JID string `json:"jid,omitempty"`
	}
	// Ignore binding errors - JID is optional
	_ = c.ShouldBindJSON(&req)

	if err := h.sessionUC.ReconnectSessionWithJID(c.Request.Context(), id, req.JID); err != nil {
		handleDomainError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, map[string]any{
		"success": true,
		"message": "Session reconnected successfully",
	})
}

// DisconnectSession handles POST /api/internal/sessions/:id/disconnect
// Disconnects a session without deleting it (keeps credentials for reconnect)
func (h *Handler) DisconnectSession(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		respondWithError(c, http.StatusBadRequest, "INVALID_ID", "Session ID is required", nil)
		return
	}

	if err := h.sessionUC.DisconnectSession(c.Request.Context(), id); err != nil {
		handleDomainError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, map[string]any{
		"success": true,
		"message": "Session disconnected successfully",
	})
}

// ConfigureHistorySync handles POST /api/internal/sessions/:id/history-sync
// Configures history sync settings for a session
func (h *Handler) ConfigureHistorySync(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		respondWithError(c, http.StatusBadRequest, "INVALID_ID", "Session ID is required", nil)
		return
	}

	var req struct {
		Enabled  bool   `json:"enabled"`
		FullSync bool   `json:"full_sync"`
		Since    string `json:"since,omitempty"` // ISO 8601 timestamp for incremental sync
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, http.StatusBadRequest, "INVALID_JSON", "Invalid request body", nil)
		return
	}

	if err := h.sessionUC.ConfigureHistorySync(c.Request.Context(), id, req.Enabled, req.FullSync, req.Since); err != nil {
		handleDomainError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, map[string]any{
		"success": true,
		"message": "History sync configured successfully",
	})
}

// Health handles GET /health (liveness probe)
func (h *Handler) Health(c *gin.Context) {
	if h.healthUC != nil {
		response := h.healthUC.CheckLiveness(c.Request.Context())
		if !response.Alive {
			c.JSON(http.StatusServiceUnavailable, map[string]any{
				"status":  "unhealthy",
				"message": response.Message,
			})
			return
		}

		// Include additional details if requested
		if c.Query("details") == "true" {
			details := h.healthUC.GetHealthDetails(c.Request.Context())
			c.JSON(http.StatusOK, map[string]any{
				"status":  "healthy",
				"message": response.Message,
				"details": details,
			})
			return
		}
	}

	respondWithSuccess(c, http.StatusOK, map[string]string{
		"status": "healthy",
	})
}

// Ready handles GET /ready (readiness probe)
func (h *Handler) Ready(c *gin.Context) {
	if h.healthUC != nil {
		response := h.healthUC.CheckReadiness(c.Request.Context())

		statusCode := http.StatusOK
		status := "ready"
		if !response.Ready {
			statusCode = http.StatusServiceUnavailable
			status = "not_ready"
		}

		// Use respondWithSuccess for consistent API response format
		if statusCode == http.StatusOK {
			respondWithSuccess(c, statusCode, map[string]any{
				"status":     status,
				"components": response.Components,
			})
		} else {
			c.JSON(statusCode, map[string]any{
				"success": false,
				"error": map[string]any{
					"code":    "NOT_READY",
					"message": "Service is not ready",
				},
				"data": map[string]any{
					"status":     status,
					"components": response.Components,
				},
			})
		}
		return
	}

	// Fallback if health use case is not configured
	respondWithSuccess(c, http.StatusOK, map[string]string{
		"status": "ready",
	})
}

// handleDomainError converts domain errors to HTTP responses
func handleDomainError(c *gin.Context, err error) {
	domainErr := errors.GetDomainError(err)
	if domainErr == nil {
		// Log full error details for unexpected errors
		requestID, _ := c.Get(RequestIDKey)
		log.Printf("[ERROR] [%v] Unexpected error: %+v", requestID, err)

		// Return generic message to client
		respondWithError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "An internal error occurred", nil)
		return
	}

	statusCode := mapErrorToHTTPStatus(domainErr.Code)

	// Log internal errors with full details
	if statusCode == http.StatusInternalServerError {
		requestID, _ := c.Get(RequestIDKey)
		log.Printf("[ERROR] [%v] Domain error: code=%s, message=%s, cause=%+v",
			requestID, domainErr.Code, domainErr.Message, domainErr.Cause)
	}

	respondWithError(c, statusCode, domainErr.Code, domainErr.Message, nil)
}

// mapErrorToHTTPStatus maps domain error codes to HTTP status codes
func mapErrorToHTTPStatus(code string) int {
	switch code {
	// Not Found errors (404)
	case "SESSION_NOT_FOUND", "MESSAGE_NOT_FOUND", "NOT_FOUND", "CONTACT_NOT_FOUND", "CHAT_NOT_FOUND":
		return http.StatusNotFound

	// Conflict errors (409)
	case "SESSION_EXISTS", "DUPLICATE":
		return http.StatusConflict

	// Bad Request errors (400)
	case "INVALID_PHONE", "VALIDATION_FAILED", "INVALID_INPUT", "EMPTY_CONTENT", "INVALID_MESSAGE_TYPE",
		"INVALID_MEDIA_SIZE", "MEDIA_TOO_LARGE", "UNSUPPORTED_MEDIA_TYPE", "INVALID_MIME_TYPE", "UNSUPPORTED_MIME_TYPE",
		"DISCONNECTED", "SESSION_INVALID", "INVALID_EMOJI", "INVALID_REACTION", "INVALID_RECEIPT_TYPE",
		"INVALID_PRESENCE_STATE", "INVALID_JID", "INVALID_STATUS":
		return http.StatusBadRequest

	// Timeout errors (408)
	case "QR_TIMEOUT", "AUTH_TIMEOUT":
		return http.StatusRequestTimeout

	// Service Unavailable errors (503)
	case "CONNECTION_FAILED", "RECONNECT_FAILED", "CIRCUIT_OPEN", "WHATSAPP_UNAVAILABLE":
		return http.StatusServiceUnavailable

	// Internal Server errors (500)
	case "DATABASE_ERROR", "INTERNAL_ERROR", "QR_GENERATION_FAILED", "AUTH_FAILED",
		"MESSAGE_SEND_FAILED", "MEDIA_DOWNLOAD_FAILED", "MEDIA_UPLOAD_FAILED",
		"REACTION_SEND_FAILED", "RECEIPT_SEND_FAILED", "PRESENCE_SEND_FAILED",
		"CONFIG_MISSING", "CONFIG_INVALID", "WHATSAPP_ERROR":
		return http.StatusInternalServerError

	// Unauthorized errors (401)
	case "UNAUTHORIZED", "MISSING_API_KEY", "INVALID_API_KEY":
		return http.StatusUnauthorized

	// Forbidden errors (403)
	case "FORBIDDEN", "INSUFFICIENT_PERMISSIONS":
		return http.StatusForbidden

	// Rate Limit errors (429)
	case "RATE_LIMIT_EXCEEDED":
		return http.StatusTooManyRequests

	// Default to Internal Server Error for unknown codes
	default:
		return http.StatusInternalServerError
	}
}

// respondWithSuccess sends a successful JSON response
func respondWithSuccess(c *gin.Context, statusCode int, data any) {
	c.JSON(statusCode, dto.NewSuccessResponse(data))
}

// respondWithError sends an error JSON response
func respondWithError(c *gin.Context, statusCode int, code, message string, details map[string]string) {
	c.JSON(statusCode, dto.NewErrorResponse[any](code, message, details))
}

// CheckPhoneNumber handles GET /api/contacts/check
func (h *Handler) CheckPhoneNumber(c *gin.Context) {
	phone := c.Query("phone")
	sessionID := c.Query("session_id")

	if phone == "" || sessionID == "" {
		respondWithError(c, http.StatusBadRequest, "VALIDATION_FAILED", "phone and session_id are required", nil)
		return
	}

	req := dto.CheckPhoneRequest{
		SessionID: sessionID,
		Phone:     phone,
	}

	if err := validator.Validate(req); err != nil {
		details := validator.ValidationErrors(err)
		respondWithError(c, http.StatusBadRequest, "VALIDATION_FAILED", "Validation failed", details)
		return
	}

	if h.contactUC == nil {
		respondWithError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Contact use case not configured", nil)
		return
	}

	contact, err := h.contactUC.CheckPhoneNumber(c.Request.Context(), req)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, dto.NewContactResponse(contact))
}

// GetUserProfile handles GET /api/contacts/:jid/profile
func (h *Handler) GetUserProfile(c *gin.Context) {
	jid := c.Param("jid")
	sessionID := c.Query("session_id")

	if jid == "" || sessionID == "" {
		respondWithError(c, http.StatusBadRequest, "VALIDATION_FAILED", "jid and session_id are required", nil)
		return
	}

	req := dto.GetProfileRequest{
		SessionID: sessionID,
		JID:       jid,
	}

	if err := validator.Validate(req); err != nil {
		details := validator.ValidationErrors(err)
		respondWithError(c, http.StatusBadRequest, "VALIDATION_FAILED", "Validation failed", details)
		return
	}

	if h.contactUC == nil {
		respondWithError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Contact use case not configured", nil)
		return
	}

	contact, err := h.contactUC.GetUserProfile(c.Request.Context(), req)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, dto.NewContactResponse(contact))
}

// ListContacts handles GET /api/sessions/:id/contacts
func (h *Handler) ListContacts(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		respondWithError(c, http.StatusBadRequest, "INVALID_ID", "Session ID is required", nil)
		return
	}

	if h.contactUC == nil {
		respondWithError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Contact use case not configured", nil)
		return
	}

	contacts, err := h.contactUC.ListContacts(c.Request.Context(), sessionID)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, dto.NewContactListResponse(contacts))
}

// ListChats handles GET /api/sessions/:id/chats
func (h *Handler) ListChats(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		respondWithError(c, http.StatusBadRequest, "INVALID_ID", "Session ID is required", nil)
		return
	}

	if h.contactUC == nil {
		respondWithError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Contact use case not configured", nil)
		return
	}

	chats, err := h.contactUC.ListChats(c.Request.Context(), sessionID)
	if err != nil {
		handleDomainError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, dto.NewChatListResponse(chats))
}
