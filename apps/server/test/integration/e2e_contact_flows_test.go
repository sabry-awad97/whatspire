package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"whatspire/internal/application/dto"
	"whatspire/internal/application/usecase"
	"whatspire/internal/domain/entity"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== E2E Test: Contact Operations ====================

func TestE2E_ContactOperations_CheckPhoneNumber(t *testing.T) {
	// Setup mocks
	waClient := NewWhatsAppClientMock()
	sessionRepo := NewSessionRepositoryMock()

	// Create session
	session := entity.NewSession(testSessionID, "Test Session")
	session.SetStatus(entity.StatusConnected)
	sessionRepo.Sessions[testSessionID] = session
	waClient.Connected[testSessionID] = true

	// Setup use cases
	contactUC := usecase.NewContactUseCase(waClient)

	// Setup router
	router := setupE2ERouter(nil, nil, nil, nil, nil, contactUC)

	// Check phone number
	req := httptest.NewRequest(http.MethodGet, "/api/contacts/check?phone=+1234567890&session_id=550e8400-e29b-41d4-a716-446655440000", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Verify HTTP response
	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.APIResponse[dto.ContactResponse]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)
	assert.True(t, response.Data.IsOnWhatsApp)
	assert.NotEmpty(t, response.Data.JID)
}

func TestE2E_ContactOperations_GetUserProfile(t *testing.T) {
	// Setup mocks
	waClient := NewWhatsAppClientMock()
	sessionRepo := NewSessionRepositoryMock()

	// Create session
	session := entity.NewSession(testSessionID, "Test Session")
	session.SetStatus(entity.StatusConnected)
	sessionRepo.Sessions[testSessionID] = session
	waClient.Connected[testSessionID] = true

	// Setup use cases
	contactUC := usecase.NewContactUseCase(waClient)

	// Setup router
	router := setupE2ERouter(nil, nil, nil, nil, nil, contactUC)

	// Get user profile
	req := httptest.NewRequest(http.MethodGet, "/api/contacts/1234567890@s.whatsapp.net/profile?session_id=550e8400-e29b-41d4-a716-446655440000", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Verify HTTP response
	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.APIResponse[dto.ContactResponse]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "1234567890@s.whatsapp.net", response.Data.JID)
	assert.Equal(t, "Test User", response.Data.Name)
	assert.NotNil(t, response.Data.AvatarURL)
	assert.NotNil(t, response.Data.Status)
}

func TestE2E_ContactOperations_ListContacts(t *testing.T) {
	// Setup mocks
	waClient := NewWhatsAppClientMock()
	sessionRepo := NewSessionRepositoryMock()

	// Create session
	session := entity.NewSession(testSessionID, "Test Session")
	session.SetStatus(entity.StatusConnected)
	sessionRepo.Sessions[testSessionID] = session
	waClient.Connected[testSessionID] = true

	// Setup use cases
	contactUC := usecase.NewContactUseCase(waClient)

	// Setup router
	router := setupE2ERouter(nil, nil, nil, nil, nil, contactUC)

	// List contacts
	req := httptest.NewRequest(http.MethodGet, "/api/sessions/550e8400-e29b-41d4-a716-446655440000/contacts", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Verify HTTP response
	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.APIResponse[dto.ContactListResponse]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)
	assert.Len(t, response.Data.Contacts, 2)
	assert.Equal(t, "1234567890@s.whatsapp.net", response.Data.Contacts[0].JID)
	assert.Equal(t, "Contact 1", response.Data.Contacts[0].Name)
}

func TestE2E_ContactOperations_ListChats(t *testing.T) {
	// Setup mocks
	waClient := NewWhatsAppClientMock()
	sessionRepo := NewSessionRepositoryMock()

	// Create session
	session := entity.NewSession(testSessionID, "Test Session")
	session.SetStatus(entity.StatusConnected)
	sessionRepo.Sessions[testSessionID] = session
	waClient.Connected[testSessionID] = true

	// Setup use cases
	contactUC := usecase.NewContactUseCase(waClient)

	// Setup router
	router := setupE2ERouter(nil, nil, nil, nil, nil, contactUC)

	// List chats
	req := httptest.NewRequest(http.MethodGet, "/api/sessions/550e8400-e29b-41d4-a716-446655440000/chats", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Verify HTTP response
	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.APIResponse[dto.ChatListResponse]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)
	assert.Len(t, response.Data.Chats, 2)
	assert.Equal(t, "1234567890@s.whatsapp.net", response.Data.Chats[0].JID)
	assert.Equal(t, "Chat 1", response.Data.Chats[0].Name)
	assert.False(t, response.Data.Chats[0].IsGroup)
	assert.Equal(t, "group123@g.us", response.Data.Chats[1].JID)
	assert.True(t, response.Data.Chats[1].IsGroup)
}

func TestE2E_DisconnectedSession_ContactOperationsFail(t *testing.T) {
	// Setup mocks
	waClient := NewWhatsAppClientMock()

	// Create disconnected session
	waClient.Connected[testSessionID] = false

	// Setup use cases
	contactUC := usecase.NewContactUseCase(waClient)

	// Setup router
	router := setupE2ERouter(nil, nil, nil, nil, nil, contactUC)

	// Test check phone number
	req := httptest.NewRequest(http.MethodGet, "/api/contacts/check?phone=+1234567890&session_id=550e8400-e29b-41d4-a716-446655440000", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response dto.APIResponse[interface{}]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "DISCONNECTED", response.Error.Code)
}
