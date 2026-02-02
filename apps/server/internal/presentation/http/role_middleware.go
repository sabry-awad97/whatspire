package http

import (
	"net/http"

	"whatspire/internal/application/dto"
	"whatspire/internal/infrastructure/config"

	"github.com/gin-gonic/gin"
)

// RoleAuthorizationMiddleware creates a middleware that enforces role-based authorization
// It checks if the authenticated API key has sufficient permissions for the requested operation
func RoleAuthorizationMiddleware(requiredRole config.Role, apiKeyConfig *config.APIKeyConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip if API key authentication is disabled
		if apiKeyConfig == nil || !apiKeyConfig.Enabled {
			c.Next()
			return
		}

		// Get API key from context (set by APIKeyMiddleware)
		apiKey := GetAPIKey(c)
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, dto.NewErrorResponse[interface{}](
				"UNAUTHORIZED",
				"API key is required",
				nil,
			))
			c.Abort()
			return
		}

		// Get role for API key
		userRole := apiKeyConfig.GetRoleForKey(apiKey)
		if userRole == "" {
			c.JSON(http.StatusUnauthorized, dto.NewErrorResponse[interface{}](
				"INVALID_API_KEY",
				"Invalid API key",
				nil,
			))
			c.Abort()
			return
		}

		// Check if role has permission
		if !hasPermission(userRole, requiredRole) {
			c.JSON(http.StatusForbidden, dto.NewErrorResponse[interface{}](
				"FORBIDDEN",
				"Insufficient permissions for this operation",
				nil,
			))
			c.Abort()
			return
		}

		// Store user role in context for potential use by handlers
		c.Set("user_role", userRole)

		c.Next()
	}
}

// hasPermission checks if a user role has permission for a required role
// Permission hierarchy: admin > write > read
func hasPermission(userRole, requiredRole config.Role) bool {
	// Admin can do everything
	if userRole == config.RoleAdmin {
		return true
	}

	// Write can do write and read
	if userRole == config.RoleWrite && (requiredRole == config.RoleWrite || requiredRole == config.RoleRead) {
		return true
	}

	// Read can only do read
	if userRole == config.RoleRead && requiredRole == config.RoleRead {
		return true
	}

	return false
}

// GetUserRole retrieves the user role from the Gin context
func GetUserRole(c *gin.Context) config.Role {
	if role, exists := c.Get("user_role"); exists {
		if r, ok := role.(config.Role); ok {
			return r
		}
	}
	return ""
}
