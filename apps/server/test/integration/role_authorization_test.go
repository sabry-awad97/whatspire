package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"whatspire/internal/application/dto"
	"whatspire/internal/infrastructure/config"
	httpPres "whatspire/internal/presentation/http"
)

func TestRoleAuthorization_ReadRole(t *testing.T) {
	// Setup
	handler := httpPres.NewHandler(nil, nil, nil, nil, nil, nil, nil, nil, nil)

	apiKeyConfig := &config.APIKeyConfig{
		Enabled: true,
		KeysMap: []config.APIKeyInfo{
			{Key: "read-key", Role: config.RoleRead},
		},
	}

	routerConfig := httpPres.RouterConfig{
		APIKeyConfig: apiKeyConfig,
	}

	router := httpPres.NewRouter(handler, routerConfig)

	// Test: Read role can access GET endpoints
	t.Run("read role can access GET /api/contacts/check", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/contacts/check?phone=+1234567890&session_id=test", nil)
		req.Header.Set("X-API-Key", "read-key")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should not get 403 Forbidden (might get other errors due to missing use cases, but not 403)
		if w.Code == http.StatusForbidden {
			t.Errorf("Expected read role to access GET endpoint, got 403 Forbidden")
		}
	})

	// Test: Read role cannot access POST endpoints
	t.Run("read role cannot access POST /api/messages", func(t *testing.T) {
		text := "test"
		reqBody := dto.SendMessageRequest{
			SessionID: "test",
			To:        "+1234567890",
			Type:      "text",
			Content: dto.SendMessageContentInput{
				Text: &text,
			},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/messages", bytes.NewBuffer(body))
		req.Header.Set("X-API-Key", "read-key")
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected 403 Forbidden for read role on POST endpoint, got %d", w.Code)
		}
	})
}

func TestRoleAuthorization_WriteRole(t *testing.T) {
	// Setup
	handler := httpPres.NewHandler(nil, nil, nil, nil, nil, nil, nil, nil, nil)

	apiKeyConfig := &config.APIKeyConfig{
		Enabled: true,
		KeysMap: []config.APIKeyInfo{
			{Key: "write-key", Role: config.RoleWrite},
		},
	}

	routerConfig := httpPres.RouterConfig{
		APIKeyConfig: apiKeyConfig,
	}

	router := httpPres.NewRouter(handler, routerConfig)

	// Test: Write role can access GET endpoints
	t.Run("write role can access GET /api/contacts/check", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/contacts/check?phone=+1234567890&session_id=test", nil)
		req.Header.Set("X-API-Key", "write-key")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should not get 403 Forbidden
		if w.Code == http.StatusForbidden {
			t.Errorf("Expected write role to access GET endpoint, got 403 Forbidden")
		}
	})

	// Test: Write role can access POST endpoints
	t.Run("write role can access POST /api/messages", func(t *testing.T) {
		text := "test"
		reqBody := dto.SendMessageRequest{
			SessionID: "test",
			To:        "+1234567890",
			Type:      "text",
			Content: dto.SendMessageContentInput{
				Text: &text,
			},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/messages", bytes.NewBuffer(body))
		req.Header.Set("X-API-Key", "write-key")
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should not get 403 Forbidden (might get other errors due to missing use cases, but not 403)
		if w.Code == http.StatusForbidden {
			t.Errorf("Expected write role to access POST endpoint, got 403 Forbidden")
		}
	})

	// Test: Write role cannot access admin endpoints
	t.Run("write role cannot access POST /api/internal/sessions/register", func(t *testing.T) {
		reqBody := map[string]string{
			"id":   "test",
			"name": "test",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/internal/sessions/register", bytes.NewBuffer(body))
		req.Header.Set("X-API-Key", "write-key")
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected 403 Forbidden for write role on admin endpoint, got %d", w.Code)
		}
	})
}

func TestRoleAuthorization_AdminRole(t *testing.T) {
	// Setup
	handler := httpPres.NewHandler(nil, nil, nil, nil, nil, nil, nil, nil, nil)

	apiKeyConfig := &config.APIKeyConfig{
		Enabled: true,
		KeysMap: []config.APIKeyInfo{
			{Key: "admin-key", Role: config.RoleAdmin},
		},
	}

	routerConfig := httpPres.RouterConfig{
		APIKeyConfig: apiKeyConfig,
	}

	router := httpPres.NewRouter(handler, routerConfig)

	// Test: Admin role can access GET endpoints
	t.Run("admin role can access GET /api/contacts/check", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/contacts/check?phone=+1234567890&session_id=test", nil)
		req.Header.Set("X-API-Key", "admin-key")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should not get 403 Forbidden
		if w.Code == http.StatusForbidden {
			t.Errorf("Expected admin role to access GET endpoint, got 403 Forbidden")
		}
	})

	// Test: Admin role can access POST endpoints
	t.Run("admin role can access POST /api/messages", func(t *testing.T) {
		text := "test"
		reqBody := dto.SendMessageRequest{
			SessionID: "test",
			To:        "+1234567890",
			Type:      "text",
			Content: dto.SendMessageContentInput{
				Text: &text,
			},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/messages", bytes.NewBuffer(body))
		req.Header.Set("X-API-Key", "admin-key")
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should not get 403 Forbidden
		if w.Code == http.StatusForbidden {
			t.Errorf("Expected admin role to access POST endpoint, got 403 Forbidden")
		}
	})

	// Test: Admin role can access admin endpoints
	t.Run("admin role can access POST /api/internal/sessions/register", func(t *testing.T) {
		reqBody := map[string]string{
			"id":   "test",
			"name": "test",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/internal/sessions/register", bytes.NewBuffer(body))
		req.Header.Set("X-API-Key", "admin-key")
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should not get 403 Forbidden (might get other errors due to missing use cases, but not 403)
		if w.Code == http.StatusForbidden {
			t.Errorf("Expected admin role to access admin endpoint, got 403 Forbidden")
		}
	})
}

func TestRoleAuthorization_DefaultRole(t *testing.T) {
	// Setup
	handler := httpPres.NewHandler(nil, nil, nil, nil, nil, nil, nil, nil, nil)

	// Test with legacy keys (should default to write role)
	apiKeyConfig := &config.APIKeyConfig{
		Enabled: true,
		Keys:    []string{"legacy-key"},
	}

	routerConfig := httpPres.RouterConfig{
		APIKeyConfig: apiKeyConfig,
	}

	router := httpPres.NewRouter(handler, routerConfig)

	// Test: Legacy key defaults to write role and can access POST endpoints
	t.Run("legacy key defaults to write role", func(t *testing.T) {
		text := "test"
		reqBody := dto.SendMessageRequest{
			SessionID: "test",
			To:        "+1234567890",
			Type:      "text",
			Content: dto.SendMessageContentInput{
				Text: &text,
			},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/messages", bytes.NewBuffer(body))
		req.Header.Set("X-API-Key", "legacy-key")
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should not get 403 Forbidden (default write role should allow POST)
		if w.Code == http.StatusForbidden {
			t.Errorf("Expected legacy key with default write role to access POST endpoint, got 403 Forbidden")
		}
	})

	// Test: Legacy key cannot access admin endpoints
	t.Run("legacy key cannot access admin endpoints", func(t *testing.T) {
		reqBody := map[string]string{
			"id":   "test",
			"name": "test",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/internal/sessions/register", bytes.NewBuffer(body))
		req.Header.Set("X-API-Key", "legacy-key")
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected 403 Forbidden for legacy key on admin endpoint, got %d", w.Code)
		}
	})
}
