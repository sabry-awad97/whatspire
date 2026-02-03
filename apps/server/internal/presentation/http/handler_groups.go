package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

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
