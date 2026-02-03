package http

import (
	"net/http"

	"whatspire/internal/application/dto"
	"whatspire/pkg/validator"

	"github.com/gin-gonic/gin"
)

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
