package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

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
