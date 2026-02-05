package property

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"whatspire/internal/infrastructure/config"
	httpPresentation "whatspire/internal/presentation/http"
	"whatspire/test/helpers"

	"github.com/gin-gonic/gin"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/stretchr/testify/require"
)

// TestRoleHierarchy_PropertyBased verifies role hierarchy properties using property-based testing
func TestRoleHierarchy_PropertyBased(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: Admin role can access any endpoint regardless of required role
	properties.Property("Admin can access any endpoint", prop.ForAll(
		func(requiredRole config.Role) bool {
			apiKeyRepo := helpers.NewMockAPIKeyRepository()
			adminKey := helpers.CreateTestAPIKey(t, apiKeyRepo, "admin", nil)

			apiKeyConfig := &config.APIKeyConfig{
				Enabled: true,
				Header:  "X-API-Key",
			}

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.Use(httpPresentation.APIKeyMiddleware(*apiKeyConfig, nil, apiKeyRepo))
			router.Use(httpPresentation.RoleAuthorizationMiddleware(requiredRole, apiKeyConfig))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("X-API-Key", adminKey.PlainText)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			return w.Code == http.StatusOK
		},
		genRole(),
	))

	// Property: Write role can access read and write endpoints but not admin
	properties.Property("Write role permissions", prop.ForAll(
		func(requiredRole config.Role) bool {
			apiKeyRepo := helpers.NewMockAPIKeyRepository()
			writeKey := helpers.CreateTestAPIKey(t, apiKeyRepo, "write", nil)

			apiKeyConfig := &config.APIKeyConfig{
				Enabled: true,
				Header:  "X-API-Key",
			}

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.Use(httpPresentation.APIKeyMiddleware(*apiKeyConfig, nil, apiKeyRepo))
			router.Use(httpPresentation.RoleAuthorizationMiddleware(requiredRole, apiKeyConfig))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("X-API-Key", writeKey.PlainText)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			// Write can access read and write, but not admin
			expectedStatus := http.StatusOK
			if requiredRole == config.RoleAdmin {
				expectedStatus = http.StatusForbidden
			}

			return w.Code == expectedStatus
		},
		genRole(),
	))

	// Property: Read role can only access read endpoints
	properties.Property("Read role can only access read endpoints", prop.ForAll(
		func(requiredRole config.Role) bool {
			apiKeyRepo := helpers.NewMockAPIKeyRepository()
			readKey := helpers.CreateTestAPIKey(t, apiKeyRepo, "read", nil)

			apiKeyConfig := &config.APIKeyConfig{
				Enabled: true,
				Header:  "X-API-Key",
			}

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.Use(httpPresentation.APIKeyMiddleware(*apiKeyConfig, nil, apiKeyRepo))
			router.Use(httpPresentation.RoleAuthorizationMiddleware(requiredRole, apiKeyConfig))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("X-API-Key", readKey.PlainText)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			// Read can only access read endpoints
			expectedStatus := http.StatusForbidden
			if requiredRole == config.RoleRead {
				expectedStatus = http.StatusOK
			}

			return w.Code == expectedStatus
		},
		genRole(),
	))

	// Property: Role hierarchy is transitive (if A >= B and B >= C, then A >= C)
	properties.Property("Role hierarchy is transitive", prop.ForAll(
		func() bool {
			// Admin >= Write >= Read
			// If admin can access write endpoints, and write can access read endpoints,
			// then admin can access read endpoints

			apiKeyRepo := helpers.NewMockAPIKeyRepository()
			adminKey := helpers.CreateTestAPIKey(t, apiKeyRepo, "admin", nil)

			apiKeyConfig := &config.APIKeyConfig{
				Enabled: true,
				Header:  "X-API-Key",
			}

			// Test admin accessing read endpoint (transitivity)
			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.Use(httpPresentation.APIKeyMiddleware(*apiKeyConfig, nil, apiKeyRepo))
			router.Use(httpPresentation.RoleAuthorizationMiddleware(config.RoleRead, apiKeyConfig))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("X-API-Key", adminKey.PlainText)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			return w.Code == http.StatusOK
		},
	))

	// Property: Invalid API key always results in unauthorized, regardless of role
	properties.Property("Invalid API key always unauthorized", prop.ForAll(
		func(requiredRole config.Role) bool {
			apiKeyRepo := helpers.NewMockAPIKeyRepository()
			// Create a valid key but don't use it
			_ = helpers.CreateTestAPIKey(t, apiKeyRepo, "admin", nil)

			apiKeyConfig := &config.APIKeyConfig{
				Enabled: true,
				Header:  "X-API-Key",
			}

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.Use(httpPresentation.APIKeyMiddleware(*apiKeyConfig, nil, apiKeyRepo))
			router.Use(httpPresentation.RoleAuthorizationMiddleware(requiredRole, apiKeyConfig))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("X-API-Key", "invalid-key-not-in-database")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			return w.Code == http.StatusUnauthorized
		},
		genRole(),
	))

	// Property: Revoked keys are always rejected regardless of role
	properties.Property("Revoked keys always rejected", prop.ForAll(
		func(role string, requiredRole config.Role) bool {
			apiKeyRepo := helpers.NewMockAPIKeyRepository()
			revokedKey := helpers.CreateRevokedTestAPIKey(t, apiKeyRepo, role)

			apiKeyConfig := &config.APIKeyConfig{
				Enabled: true,
				Header:  "X-API-Key",
			}

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.Use(httpPresentation.APIKeyMiddleware(*apiKeyConfig, nil, apiKeyRepo))
			router.Use(httpPresentation.RoleAuthorizationMiddleware(requiredRole, apiKeyConfig))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("X-API-Key", revokedKey.PlainText)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			return w.Code == http.StatusUnauthorized
		},
		genRoleString(),
		genRole(),
	))

	properties.TestingRun(t)
}

// TestRolePermissionMatrix_PropertyBased tests the complete permission matrix
func TestRolePermissionMatrix_PropertyBased(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50

	properties := gopter.NewProperties(parameters)

	// Property: Permission matrix is consistent
	properties.Property("Permission matrix consistency", prop.ForAll(
		func(userRole, requiredRole config.Role) bool {
			apiKeyRepo := helpers.NewMockAPIKeyRepository()

			// Convert role to string for CreateTestAPIKey
			roleStr := string(userRole)
			testKey := helpers.CreateTestAPIKey(t, apiKeyRepo, roleStr, nil)

			apiKeyConfig := &config.APIKeyConfig{
				Enabled: true,
				Header:  "X-API-Key",
			}

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.Use(httpPresentation.APIKeyMiddleware(*apiKeyConfig, nil, apiKeyRepo))
			router.Use(httpPresentation.RoleAuthorizationMiddleware(requiredRole, apiKeyConfig))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("X-API-Key", testKey.PlainText)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			// Calculate expected result based on role hierarchy
			expectedStatus := calculateExpectedStatus(userRole, requiredRole)

			return w.Code == expectedStatus
		},
		genRole(),
		genRole(),
	))

	properties.TestingRun(t)
}

// TestRoleContextPropagation_PropertyBased verifies role is correctly stored in context
func TestRoleContextPropagation_PropertyBased(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50

	properties := gopter.NewProperties(parameters)

	// Property: Role is correctly propagated to context when authorized
	properties.Property("Role propagated to context", prop.ForAll(
		func(role config.Role) bool {
			apiKeyRepo := helpers.NewMockAPIKeyRepository()
			roleStr := string(role)
			testKey := helpers.CreateTestAPIKey(t, apiKeyRepo, roleStr, nil)

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
			req.Header.Set("X-API-Key", testKey.PlainText)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			// All roles can access read endpoints, so check if role was captured
			if w.Code == http.StatusOK {
				return capturedRole == role
			}

			return true // If forbidden, role propagation doesn't matter
		},
		genRole(),
	))

	properties.TestingRun(t)
}

// Helper functions

// genRole generates random valid roles
func genRole() gopter.Gen {
	return gen.OneConstOf(
		config.RoleAdmin,
		config.RoleWrite,
		config.RoleRead,
	)
}

// genRoleString generates random valid role strings
func genRoleString() gopter.Gen {
	return gen.OneConstOf("admin", "write", "read")
}

// calculateExpectedStatus determines the expected HTTP status based on role hierarchy
func calculateExpectedStatus(userRole, requiredRole config.Role) int {
	// Admin can access everything
	if userRole == config.RoleAdmin {
		return http.StatusOK
	}

	// Write can access write and read
	if userRole == config.RoleWrite && (requiredRole == config.RoleWrite || requiredRole == config.RoleRead) {
		return http.StatusOK
	}

	// Read can only access read
	if userRole == config.RoleRead && requiredRole == config.RoleRead {
		return http.StatusOK
	}

	// All other combinations are forbidden
	return http.StatusForbidden
}

// TestRoleAuthorizationDisabled_PropertyBased verifies behavior when auth is disabled
func TestRoleAuthorizationDisabled_PropertyBased(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 20

	properties := gopter.NewProperties(parameters)

	// Property: When auth is disabled, all requests succeed regardless of role
	properties.Property("Disabled auth allows all requests", prop.ForAll(
		func(requiredRole config.Role) bool {
			apiKeyConfig := &config.APIKeyConfig{
				Enabled: false,
				Header:  "X-API-Key",
			}

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.Use(httpPresentation.RoleAuthorizationMiddleware(requiredRole, apiKeyConfig))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			// No API key
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			return w.Code == http.StatusOK
		},
		genRole(),
	))

	properties.TestingRun(t)
}

// TestAPIKeyCreation_PropertyBased verifies API key creation and validation
func TestAPIKeyCreation_PropertyBased(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50

	properties := gopter.NewProperties(parameters)

	// Property: Created API keys can be used for authentication
	properties.Property("Created keys are valid", prop.ForAll(
		func(role string) bool {
			apiKeyRepo := helpers.NewMockAPIKeyRepository()
			testKey := helpers.CreateTestAPIKey(t, apiKeyRepo, role, nil)

			// Verify the key was created
			require.NotNil(t, testKey)
			require.NotEmpty(t, testKey.PlainText)
			require.NotNil(t, testKey.Entity)

			// Verify the key can be used for authentication
			apiKeyConfig := &config.APIKeyConfig{
				Enabled: true,
				Header:  "X-API-Key",
			}

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.Use(httpPresentation.APIKeyMiddleware(*apiKeyConfig, nil, apiKeyRepo))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("X-API-Key", testKey.PlainText)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			return w.Code == http.StatusOK
		},
		genRoleString(),
	))

	properties.TestingRun(t)
}
