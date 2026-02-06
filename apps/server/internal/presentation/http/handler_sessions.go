package http

import (
	"net/http"

	"whatspire/internal/application/dto"
	"whatspire/internal/domain/entity"
	"whatspire/internal/domain/errors"
	"whatspire/pkg/validator"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateSession handles POST /api/sessions
// Public endpoint for creating a new WhatsApp session
func (h *Handler) CreateSession(c *gin.Context) {
	var req dto.CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, http.StatusBadRequest, "INVALID_JSON", "Invalid request body", nil)
		return
	}

	// Validate request
	if err := validator.Validate(req); err != nil {
		details := validator.ValidationErrors(err)
		respondWithError(c, http.StatusBadRequest, "VALIDATION_FAILED", "Validation failed", details)
		return
	}

	// Generate UUID for session ID
	sessionID := uuid.New().String()

	// Create session in local repository for WhatsApp client tracking
	session, err := h.sessionUC.CreateSessionWithID(c.Request.Context(), sessionID, req.Name)
	if err != nil {
		handleDomainError(c, err, h.logger)
		return
	}

	respondWithSuccess(c, http.StatusCreated, dto.NewSessionResponse(session))
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
		handleDomainError(c, err, h.logger)
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
			handleDomainError(c, err, h.logger)
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
		handleDomainError(c, err, h.logger)
		return
	}

	// Update JID if provided
	if req.JID != "" {
		if err := h.sessionUC.UpdateSessionJID(c.Request.Context(), id, req.JID); err != nil {
			handleDomainError(c, err, h.logger)
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
		handleDomainError(c, err, h.logger)
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
		handleDomainError(c, err, h.logger)
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
		handleDomainError(c, err, h.logger)
		return
	}

	respondWithSuccess(c, http.StatusOK, map[string]any{
		"success": true,
		"message": "History sync configured successfully",
	})
}

// ListSessions handles GET /api/sessions
// Returns all sessions with their current status
func (h *Handler) ListSessions(c *gin.Context) {
	sessions, err := h.sessionUC.ListSessions(c.Request.Context())
	if err != nil {
		handleDomainError(c, err, h.logger)
		return
	}

	// Convert to response DTOs
	sessionResponses := make([]map[string]any, 0, len(sessions))
	for _, s := range sessions {
		sessionResponses = append(sessionResponses, map[string]any{
			"id":         s.ID,
			"name":       s.Name,
			"status":     s.Status.String(),
			"jid":        s.JID,
			"created_at": s.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			"updated_at": s.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	respondWithSuccess(c, http.StatusOK, map[string]any{
		"sessions": sessionResponses,
	})
}

// GetSession handles GET /api/sessions/:id
// Returns a single session by ID
func (h *Handler) GetSession(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		respondWithError(c, http.StatusBadRequest, "INVALID_ID", "Session ID is required", nil)
		return
	}

	session, err := h.sessionUC.GetSession(c.Request.Context(), id)
	if err != nil {
		handleDomainError(c, err, h.logger)
		return
	}

	respondWithSuccess(c, http.StatusOK, map[string]any{
		"id":         session.ID,
		"name":       session.Name,
		"status":     session.Status.String(),
		"jid":        session.JID,
		"created_at": session.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		"updated_at": session.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// DeleteSession handles DELETE /api/sessions/:id
// Public endpoint for deleting a session
func (h *Handler) DeleteSession(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		respondWithError(c, http.StatusBadRequest, "INVALID_ID", "Session ID is required", nil)
		return
	}

	// Disconnect and cleanup WhatsApp client resources
	if err := h.sessionUC.DeleteSession(c.Request.Context(), id); err != nil {
		handleDomainError(c, err, h.logger)
		return
	}

	respondWithSuccess(c, http.StatusOK, map[string]string{"message": "Session deleted successfully"})
}

// isValidSessionID validates that a session ID contains only alphanumeric characters, hyphens, and underscores
func isValidSessionID(sessionID string) bool {
	if sessionID == "" {
		return false
	}
	for _, char := range sessionID {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '-' || char == '_') {
			return false
		}
	}
	return true
}
