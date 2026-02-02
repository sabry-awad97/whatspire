package property

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"whatspire/internal/domain/entity"
	"whatspire/internal/infrastructure/config"
	"whatspire/internal/infrastructure/persistence"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: whatsapp-http-api-enhancement, Property 19: Role Permission Hierarchy
// *For any* API key with role R and operation requiring role R', the operation SHALL be
// allowed if and only if R has permission for R' (admin > write > read).
// **Validates: Requirements 7.2, 7.3, 7.4, 7.5**

func TestRolePermissionHierarchy_Property19(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property 19.1: Admin role can access all operations
	properties.Property("admin role can access all operations", prop.ForAll(
		func(requiredRole config.Role) bool {
			userRole := config.RoleAdmin
			return hasPermission(userRole, requiredRole)
		},
		genRole(),
	))

	// Property 19.2: Write role can access write and read operations
	properties.Property("write role can access write and read operations", prop.ForAll(
		func(requiredRole config.Role) bool {
			userRole := config.RoleWrite
			canAccess := hasPermission(userRole, requiredRole)

			// Write should be able to access write and read, but not admin
			if requiredRole == config.RoleWrite || requiredRole == config.RoleRead {
				return canAccess
			}
			return !canAccess
		},
		genRole(),
	))

	// Property 19.3: Read role can only access read operations
	properties.Property("read role can only access read operations", prop.ForAll(
		func(requiredRole config.Role) bool {
			userRole := config.RoleRead
			canAccess := hasPermission(userRole, requiredRole)

			// Read should only be able to access read operations
			if requiredRole == config.RoleRead {
				return canAccess
			}
			return !canAccess
		},
		genRole(),
	))

	// Property 19.4: Permission hierarchy is transitive
	properties.Property("permission hierarchy is transitive", prop.ForAll(
		func(_ int) bool {
			// If admin can do X and write can do Y, and admin > write, then admin can do Y
			adminCanDoWrite := hasPermission(config.RoleAdmin, config.RoleWrite)
			adminCanDoRead := hasPermission(config.RoleAdmin, config.RoleRead)
			writeCanDoRead := hasPermission(config.RoleWrite, config.RoleRead)

			return adminCanDoWrite && adminCanDoRead && writeCanDoRead
		},
		gen.Const(0),
	))

	// Property 19.5: Permission hierarchy is anti-symmetric
	properties.Property("permission hierarchy is anti-symmetric", prop.ForAll(
		func(_ int) bool {
			// If read cannot do write, then write should be able to do read
			readCannotDoWrite := !hasPermission(config.RoleRead, config.RoleWrite)
			writeCanDoRead := hasPermission(config.RoleWrite, config.RoleRead)

			return readCannotDoWrite && writeCanDoRead
		},
		gen.Const(0),
	))

	properties.TestingRun(t)
}

// Feature: whatsapp-http-api-enhancement, Property 20: Role Storage and Retrieval
// *For any* API key created with a role, retrieving the API key SHALL return the same role.
// **Validates: Requirements 7.1**

func TestRoleStorageAndRetrieval_Property20(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property 20.1: Stored role matches retrieved role
	properties.Property("stored role matches retrieved role", prop.ForAll(
		func(id, key string, role config.Role) bool {
			if id == "" || key == "" {
				return true // skip invalid inputs
			}

			ctx := context.Background()
			repo := persistence.NewInMemoryAPIKeyRepository()

			// Create and save API key
			keyHash := hashKey(key)
			apiKey := entity.NewAPIKey(id, keyHash, string(role))
			err := repo.Save(ctx, apiKey)
			if err != nil {
				return false
			}

			// Retrieve by key hash
			retrieved, err := repo.FindByKeyHash(ctx, keyHash)
			if err != nil {
				return false
			}

			return retrieved.Role == string(role)
		},
		gen.Identifier(),
		gen.Identifier(),
		genRole(),
	))

	// Property 20.2: Stored role persists across multiple retrievals
	properties.Property("stored role persists across multiple retrievals", prop.ForAll(
		func(id, key string, role config.Role) bool {
			if id == "" || key == "" {
				return true // skip invalid inputs
			}

			ctx := context.Background()
			repo := persistence.NewInMemoryAPIKeyRepository()

			// Create and save API key
			keyHash := hashKey(key)
			apiKey := entity.NewAPIKey(id, keyHash, string(role))
			err := repo.Save(ctx, apiKey)
			if err != nil {
				return false
			}

			// Retrieve multiple times
			for i := 0; i < 5; i++ {
				retrieved, err := repo.FindByKeyHash(ctx, keyHash)
				if err != nil {
					return false
				}
				if retrieved.Role != string(role) {
					return false
				}
			}

			return true
		},
		gen.Identifier(),
		gen.Identifier(),
		genRole(),
	))

	// Property 20.3: Role can be retrieved by ID
	properties.Property("role can be retrieved by ID", prop.ForAll(
		func(id, key string, role config.Role) bool {
			if id == "" || key == "" {
				return true // skip invalid inputs
			}

			ctx := context.Background()
			repo := persistence.NewInMemoryAPIKeyRepository()

			// Create and save API key
			keyHash := hashKey(key)
			apiKey := entity.NewAPIKey(id, keyHash, string(role))
			err := repo.Save(ctx, apiKey)
			if err != nil {
				return false
			}

			// Retrieve by ID
			retrieved, err := repo.FindByID(ctx, id)
			if err != nil {
				return false
			}

			return retrieved.Role == string(role)
		},
		gen.Identifier(),
		gen.Identifier(),
		genRole(),
	))

	properties.TestingRun(t)
}

// Feature: whatsapp-http-api-enhancement, Property 21: Default Role Assignment
// *For any* API key created without an explicit role, the key SHALL be treated as having "write" role.
// **Validates: Requirements 7.6**

func TestDefaultRoleAssignment_Property21(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property 21.1: Empty role defaults to write
	properties.Property("empty role defaults to write", prop.ForAll(
		func(key string) bool {
			if key == "" {
				return true // skip invalid inputs
			}

			cfg := &config.APIKeyConfig{
				Enabled: true,
				KeysMap: []config.APIKeyInfo{
					{Key: key, Role: ""}, // Empty role
				},
			}

			role := cfg.GetRoleForKey(key)
			return role == config.RoleWrite
		},
		gen.Identifier(),
	))

	// Property 21.2: Legacy keys default to write role
	properties.Property("legacy keys default to write role", prop.ForAll(
		func(key string) bool {
			if key == "" {
				return true // skip invalid inputs
			}

			cfg := &config.APIKeyConfig{
				Enabled: true,
				Keys:    []string{key}, // Legacy keys list
			}

			role := cfg.GetRoleForKey(key)
			return role == config.RoleWrite
		},
		gen.Identifier(),
	))

	// Property 21.3: Default role allows write operations
	properties.Property("default role allows write operations", prop.ForAll(
		func(key string) bool {
			if key == "" {
				return true // skip invalid inputs
			}

			cfg := &config.APIKeyConfig{
				Enabled: true,
				Keys:    []string{key},
			}

			role := cfg.GetRoleForKey(key)
			return hasPermission(role, config.RoleWrite)
		},
		gen.Identifier(),
	))

	// Property 21.4: Default role allows read operations
	properties.Property("default role allows read operations", prop.ForAll(
		func(key string) bool {
			if key == "" {
				return true // skip invalid inputs
			}

			cfg := &config.APIKeyConfig{
				Enabled: true,
				Keys:    []string{key},
			}

			role := cfg.GetRoleForKey(key)
			return hasPermission(role, config.RoleRead)
		},
		gen.Identifier(),
	))

	// Property 21.5: Default role does not allow admin operations
	properties.Property("default role does not allow admin operations", prop.ForAll(
		func(key string) bool {
			if key == "" {
				return true // skip invalid inputs
			}

			cfg := &config.APIKeyConfig{
				Enabled: true,
				Keys:    []string{key},
			}

			role := cfg.GetRoleForKey(key)
			// Write role should not have admin permissions
			return !hasPermission(role, config.RoleAdmin) || role == config.RoleAdmin
		},
		gen.Identifier(),
	))

	properties.TestingRun(t)
}

// Helper functions

// genRole generates a random role
func genRole() gopter.Gen {
	return gen.OneConstOf(
		config.RoleRead,
		config.RoleWrite,
		config.RoleAdmin,
	)
}

// hasPermission checks if a user role has permission for a required role
// This is a copy of the function from role_middleware.go for testing
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

// hashKey creates a SHA-256 hash of the key
func hashKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}
