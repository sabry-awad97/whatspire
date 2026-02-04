package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"whatspire/internal/application/dto"
	"whatspire/internal/infrastructure/config"
	httpPresentation "whatspire/internal/presentation/http"
	"whatspire/test/helpers"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== Test Setup ====================

func setupRoleTestRouter(apiKeyConfig *config.APIKeyConfig, apiKeyRepo *helpers.MockAPIKeyRepository, requiredRole config.Role) *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(httpPresentation.APIKeyMiddleware(*apiKeyConfig, nil, apiKeyRepo))
	router.Use(httpPresentation.RoleAuthorizationMiddleware(requiredRole, apiKeyConfig))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	return router
}

// ==================== Role Authorization Tests ====================

func TestRoleAuthorization_AdminCanAccessAll(t *testing.T) {
	apiKeyRepo := helpers.NewMockAPIKeyRepository()
	adminKey := helpers.CreateTestAPIKey(t, apiKeyRepo, "admin", nil)

	apiKeyConfig := &config.APIKeyConfig{
		Enabled: true,
		Header:  "X-API-Key",
	}

	tests := []struct {
		name         string
		requiredRole config.Role
	}{
		{"Admin accessing read endpoint", config.RoleRead},
		{"Admin accessing write endpoint", config.RoleWrite},
		{"Admin accessing admin endpoint", config.RoleAdmin},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupRoleTestRouter(apiKeyConfig, apiKeyRepo, tt.requiredRole)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("X-API-Key", adminKey.PlainText)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestRoleAuthorization_WriteCanAccessReadAndWrite(t *testing.T) {
	apiKeyRepo := helpers.NewMockAPIKeyRepository()
	writeKey := helpers.CreateTestAPIKey(t, apiKeyRepo, "write", nil)

	apiKeyConfig := &config.APIKeyConfig{
		Enabled: true,
		Header:  "X-API-Key",
	}

	tests := []struct {
		name         string
		requiredRole config.Role
		expectStatus int
	}{
		{"Write accessing read endpoint", config.RoleRead, http.StatusOK},
		{"Write accessing write endpoint", config.RoleWrite, http.StatusOK},
		{"Write accessing admin endpoint", config.RoleAdmin, http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupRoleTestRouter(apiKeyConfig, apiKeyRepo, tt.requiredRole)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("X-API-Key", writeKey.PlainText)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectStatus, w.Code)
		})
	}
}

func TestRoleAuthorization_ReadCanOnlyAccessRead(t *testing.T) {
	apiKeyRepo := helpers.NewMockAPIKeyRepository()
	readKey := helpers.CreateTestAPIKey(t, apiKeyRepo, "read", nil)

	apiKeyConfig := &config.APIKeyConfig{
		Enabled: true,
		Header:  "X-API-Key",
	}

	tests := []struct {
		name         string
		requiredRole config.Role
		expectStatus int
	}{
		{"Read accessing read endpoint", config.RoleRead, http.StatusOK},
		{"Read accessing write endpoint", config.RoleWrite, http.StatusForbidden},
		{"Read accessing admin endpoint", config.RoleAdmin, http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupRoleTestRouter(apiKeyConfig, apiKeyRepo, tt.requiredRole)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("X-API-Key", readKey.PlainText)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectStatus, w.Code)
		})
	}
}

func TestRoleAuthorization_ForbiddenResponse(t *testing.T) {
	apiKeyRepo := helpers.NewMockAPIKeyRepository()
	readKey := helpers.CreateTestAPIKey(t, apiKeyRepo, "read", nil)

	apiKeyConfig := &config.APIKeyConfig{
		Enabled: true,
		Header:  "X-API-Key",
	}

	router := setupRoleTestRouter(apiKeyConfig, apiKeyRepo, config.RoleAdmin)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", readKey.PlainText)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response dto.APIResponse[any]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Equal(t, "FORBIDDEN", response.Error.Code)
	assert.Contains(t, response.Error.Message, "Insufficient permissions")
}

func TestRoleAuthorization_NoAPIKeyInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	apiKeyConfig := &config.APIKeyConfig{
		Enabled: true,
		Header:  "X-API-Key",
	}

	router := gin.New()
	// Skip APIKeyMiddleware to simulate missing API key in context
	router.Use(httpPresentation.RoleAuthorizationMiddleware(config.RoleRead, apiKeyConfig))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response dto.APIResponse[any]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Equal(t, "UNAUTHORIZED", response.Error.Code)
}

func TestRoleAuthorization_DisabledAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	apiKeyConfig := &config.APIKeyConfig{
		Enabled: false, // Disabled
		Header:  "X-API-Key",
	}

	router := gin.New()
	router.Use(httpPresentation.RoleAuthorizationMiddleware(config.RoleAdmin, apiKeyConfig))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	// No API key
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should succeed when auth is disabled
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRoleAuthorization_MultipleRoles(t *testing.T) {
	apiKeyRepo := helpers.NewMockAPIKeyRepository()
	adminKey := helpers.CreateTestAPIKey(t, apiKeyRepo, "admin", nil)
	writeKey := helpers.CreateTestAPIKey(t, apiKeyRepo, "write", nil)
	readKey := helpers.CreateTestAPIKey(t, apiKeyRepo, "read", nil)

	apiKeyConfig := &config.APIKeyConfig{
		Enabled: true,
		Header:  "X-API-Key",
	}

	tests := []struct {
		name         string
		apiKey       *helpers.TestAPIKey
		requiredRole config.Role
		expectStatus int
	}{
		// Admin tests
		{"Admin can read", adminKey, config.RoleRead, http.StatusOK},
		{"Admin can write", adminKey, config.RoleWrite, http.StatusOK},
		{"Admin can admin", adminKey, config.RoleAdmin, http.StatusOK},

		// Write tests
		{"Write can read", writeKey, config.RoleRead, http.StatusOK},
		{"Write can write", writeKey, config.RoleWrite, http.StatusOK},
		{"Write cannot admin", writeKey, config.RoleAdmin, http.StatusForbidden},

		// Read tests
		{"Read can read", readKey, config.RoleRead, http.StatusOK},
		{"Read cannot write", readKey, config.RoleWrite, http.StatusForbidden},
		{"Read cannot admin", readKey, config.RoleAdmin, http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupRoleTestRouter(apiKeyConfig, apiKeyRepo, tt.requiredRole)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("X-API-Key", tt.apiKey.PlainText)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectStatus, w.Code)
		})
	}
}

func TestRoleAuthorization_StoresRoleInContext(t *testing.T) {
	apiKeyRepo := helpers.NewMockAPIKeyRepository()
	adminKey := helpers.CreateTestAPIKey(t, apiKeyRepo, "admin", nil)

	apiKeyConfig := &config.APIKeyConfig{
		Enabled: true,
		Header:  "X-API-Key",
	}

	var capturedRole config.Role

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(httpPresentation.APIKeyMiddleware(*apiKeyConfig, nil, apiKeyRepo))
	router.Use(httpPresentation.RoleAuthorizationMiddleware(config.RoleRead, apiKeyConfig))
	router.GET("/test", func(c *gin.Context) {
		capturedRole = httpPresentation.GetUserRole(c)
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", adminKey.PlainText)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, config.RoleAdmin, capturedRole)
}
