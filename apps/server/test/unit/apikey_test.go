package unit

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"whatspire/internal/infrastructure/config"
	httpPresentation "whatspire/internal/presentation/http"
	"whatspire/test/helpers"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIKeyMiddleware_ValidKey(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock repository and test API keys
	repo := helpers.NewMockAPIKeyRepository()
	testKey1 := helpers.CreateTestAPIKey(t, repo, "admin", nil)
	testKey2 := helpers.CreateTestAPIKey(t, repo, "write", nil)

	apiKeyConfig := config.APIKeyConfig{
		Enabled: true,
		Header:  "X-API-Key",
	}

	router := gin.New()
	router.Use(httpPresentation.APIKeyMiddleware(apiKeyConfig, nil, repo))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	tests := []struct {
		name   string
		apiKey string
	}{
		{"First valid key", testKey1.PlainText},
		{"Second valid key", testKey2.PlainText},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("X-API-Key", tt.apiKey)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestAPIKeyMiddleware_InvalidKey(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock repository with one valid key
	repo := helpers.NewMockAPIKeyRepository()
	_ = helpers.CreateTestAPIKey(t, repo, "admin", nil)

	apiKeyConfig := config.APIKeyConfig{
		Enabled: true,
		Header:  "X-API-Key",
	}

	router := gin.New()
	router.Use(httpPresentation.APIKeyMiddleware(apiKeyConfig, nil, repo))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "invalid-key-that-does-not-exist")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "INVALID_API_KEY")
}

func TestAPIKeyMiddleware_MissingKey(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := helpers.NewMockAPIKeyRepository()
	_ = helpers.CreateTestAPIKey(t, repo, "admin", nil)

	apiKeyConfig := config.APIKeyConfig{
		Enabled: true,
		Header:  "X-API-Key",
	}

	router := gin.New()
	router.Use(httpPresentation.APIKeyMiddleware(apiKeyConfig, nil, repo))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	// No API key header
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "MISSING_API_KEY")
}

func TestAPIKeyMiddleware_Disabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := helpers.NewMockAPIKeyRepository()

	apiKeyConfig := config.APIKeyConfig{
		Enabled: false,
		Header:  "X-API-Key",
	}

	router := gin.New()
	router.Use(httpPresentation.APIKeyMiddleware(apiKeyConfig, nil, repo))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	// Request without API key should succeed when disabled
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAPIKeyMiddleware_CustomHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := helpers.NewMockAPIKeyRepository()
	testKey := helpers.CreateTestAPIKey(t, repo, "admin", nil)

	apiKeyConfig := config.APIKeyConfig{
		Enabled: true,
		Header:  "Authorization",
	}

	router := gin.New()
	router.Use(httpPresentation.APIKeyMiddleware(apiKeyConfig, nil, repo))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	// Using custom header
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", testKey.PlainText)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Using default header should fail
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", testKey.PlainText)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAPIKeyMiddleware_DefaultHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := helpers.NewMockAPIKeyRepository()
	testKey := helpers.CreateTestAPIKey(t, repo, "admin", nil)

	apiKeyConfig := config.APIKeyConfig{
		Enabled: true,
		Header:  "", // Empty header should default to X-API-Key
	}

	router := gin.New()
	router.Use(httpPresentation.APIKeyMiddleware(apiKeyConfig, nil, repo))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", testKey.PlainText)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAPIKeyMiddleware_BearerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := helpers.NewMockAPIKeyRepository()
	testKey := helpers.CreateTestAPIKey(t, repo, "admin", nil)

	apiKeyConfig := config.APIKeyConfig{
		Enabled: true,
		Header:  "X-API-Key",
	}

	router := gin.New()
	router.Use(httpPresentation.APIKeyMiddleware(apiKeyConfig, nil, repo))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	// Test Bearer token format
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+testKey.PlainText)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAPIKeyMiddleware_StoresKeyInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := helpers.NewMockAPIKeyRepository()
	testKey := helpers.CreateTestAPIKey(t, repo, "admin", nil)

	apiKeyConfig := config.APIKeyConfig{
		Enabled: true,
		Header:  "X-API-Key",
	}

	var capturedKey string
	var capturedRole string
	var capturedID string

	router := gin.New()
	router.Use(httpPresentation.APIKeyMiddleware(apiKeyConfig, nil, repo))
	router.GET("/test", func(c *gin.Context) {
		capturedKey = httpPresentation.GetAPIKey(c)
		if role, exists := c.Get("api_key_role"); exists {
			capturedRole = role.(string)
		}
		if id, exists := c.Get("api_key_id"); exists {
			capturedID = id.(string)
		}
		c.String(http.StatusOK, "OK")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", testKey.PlainText)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, testKey.PlainText, capturedKey)
	assert.Equal(t, "admin", capturedRole)
	assert.Equal(t, testKey.Entity.ID, capturedID)
}

func TestAPIKeyMiddleware_RevokedKey(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := helpers.NewMockAPIKeyRepository()
	testKey := helpers.CreateRevokedTestAPIKey(t, repo, "admin")

	apiKeyConfig := config.APIKeyConfig{
		Enabled: true,
		Header:  "X-API-Key",
	}

	router := gin.New()
	router.Use(httpPresentation.APIKeyMiddleware(apiKeyConfig, nil, repo))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", testKey.PlainText)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "REVOKED_API_KEY")
}

func TestAPIKeyMiddleware_InactiveKey(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := helpers.NewMockAPIKeyRepository()
	testKey := helpers.CreateInactiveTestAPIKey(t, repo, "admin")

	apiKeyConfig := config.APIKeyConfig{
		Enabled: true,
		Header:  "X-API-Key",
	}

	router := gin.New()
	router.Use(httpPresentation.APIKeyMiddleware(apiKeyConfig, nil, repo))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", testKey.PlainText)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "REVOKED_API_KEY")
}

func TestAPIKeyMiddleware_NoRepositoryWhenEnabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	apiKeyConfig := config.APIKeyConfig{
		Enabled: true,
		Header:  "X-API-Key",
	}

	router := gin.New()
	router.Use(httpPresentation.APIKeyMiddleware(apiKeyConfig, nil, nil)) // nil repository
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "some-key")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "CONFIGURATION_ERROR")
}

func TestAPIKeyMiddleware_UpdatesLastUsed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := helpers.NewMockAPIKeyRepository()
	testKey := helpers.CreateTestAPIKey(t, repo, "admin", nil)

	// Verify initial state
	require.Nil(t, testKey.Entity.LastUsedAt)

	apiKeyConfig := config.APIKeyConfig{
		Enabled: true,
		Header:  "X-API-Key",
	}

	router := gin.New()
	router.Use(httpPresentation.APIKeyMiddleware(apiKeyConfig, nil, repo))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", testKey.PlainText)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Note: LastUsedAt is updated asynchronously in a goroutine,
	// so we can't reliably test it in a unit test without adding delays
	// This is tested in integration tests instead
}

func TestGetAPIKey_NoKey(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var capturedKey string

	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		capturedKey = httpPresentation.GetAPIKey(c)
		c.String(http.StatusOK, "OK")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, capturedKey)
}
