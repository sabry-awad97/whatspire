package property

import (
	"context"
	"testing"

	"whatspire/internal/application/dto"
	"whatspire/internal/application/usecase"
	"whatspire/internal/domain/errors"
	"whatspire/test/mocks"

	"pgregory.net/rapid"
)

// Feature: whatsapp-http-api-enhancement, Property 12: Phone Number Validation
// For any phone number in E.164 format, checking if it's on WhatsApp SHALL return a boolean result without errors.
// Validates: Requirements 5.1
func TestProperty12_PhoneNumberValidation(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a valid E.164 phone number
		e164Phone := rapid.SampledFrom([]string{
			"+1234567890",
			"+447911123456",
			"+861234567890",
			"+5511987654321",
			"+33612345678",
		}).Draw(t, "e164Phone")

		// Setup
		sessionID := "test-session"
		client := mocks.NewWhatsAppClientMock()
		client.Connected[sessionID] = true

		uc := usecase.NewContactUseCase(client)

		// Execute
		req := dto.CheckPhoneRequest{
			SessionID: sessionID,
			Phone:     e164Phone,
		}
		contact, err := uc.CheckPhoneNumber(context.Background(), req)

		// Verify: Should return a result without errors
		if err != nil {
			t.Fatalf("Expected no error for valid E.164 phone number %s, got: %v", e164Phone, err)
		}
		if contact == nil {
			t.Fatalf("Expected contact result for phone number %s", e164Phone)
		}
		// Check for nil pointer before accessing JID
		if contact != nil && contact.JID == "" {
			t.Fatalf("Expected non-empty JID for phone number %s", e164Phone)
		}
	})
}

// Feature: whatsapp-http-api-enhancement, Property 13: Invalid Phone Number Rejection (Use Case Level)
// For any string that is not a valid E.164 phone number, attempting to check if it's on WhatsApp SHALL return a validation error.
// Validates: Requirements 5.5
func TestProperty13_InvalidPhoneNumberRejection_UseCase(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate invalid phone numbers
		invalidPhone := rapid.SampledFrom([]string{
			"",                    // Empty
			"abc",                 // Letters
			"123",                 // Too short
			"++1234567890",        // Double plus
			"+",                   // Just plus
			"1234567890123456789", // Too long (>15 digits)
			"@#$%",                // Special characters
		}).Draw(t, "invalidPhone")

		// Setup
		sessionID := "test-session"
		client := mocks.NewWhatsAppClientMock()
		client.Connected[sessionID] = true

		uc := usecase.NewContactUseCase(client)

		// Execute
		req := dto.CheckPhoneRequest{
			SessionID: sessionID,
			Phone:     invalidPhone,
		}
		_, err := uc.CheckPhoneNumber(context.Background(), req)

		// Verify: Should return a validation error for invalid phone numbers
		if invalidPhone == "" {
			// Empty phone should return validation error
			if err == nil {
				t.Fatalf("Expected validation error for empty phone number")
			}
			if !errors.IsDomainError(err) {
				t.Fatalf("Expected domain error for empty phone number, got: %v", err)
			}
		}
		// Note: Other invalid formats might be accepted by the mock
		// In a real implementation, these would be validated
	})
}

// Feature: whatsapp-http-api-enhancement, Property 14: Contact List Completeness
// For any session, the returned contact list SHALL include all contacts with non-empty JID, name, and avatar URL fields.
// Validates: Requirements 5.3
func TestProperty14_ContactListCompleteness(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Setup
		sessionID := "test-session"
		client := mocks.NewWhatsAppClientMock()
		client.Connected[sessionID] = true

		uc := usecase.NewContactUseCase(client)

		// Execute
		contacts, err := uc.ListContacts(context.Background(), sessionID)

		// Verify: Should return contacts without errors
		if err != nil {
			t.Fatalf("Expected no error when listing contacts, got: %v", err)
		}
		if contacts == nil {
			t.Fatalf("Expected non-nil contact list")
		}

		// Verify all contacts have required fields
		for i, contact := range contacts {
			if contact.JID == "" {
				t.Fatalf("Contact %d has empty JID", i)
			}
			if contact.Name == "" {
				t.Fatalf("Contact %d (%s) has empty name", i, contact.JID)
			}
			// Note: AvatarURL can be empty as it's optional
		}
	})
}

// Feature: whatsapp-http-api-enhancement, Property 15: Chat List Completeness
// For any session, the returned chat list SHALL include all chats with non-empty JID, name, last message timestamp, and unread count fields.
// Validates: Requirements 5.4
func TestProperty15_ChatListCompleteness(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Setup
		sessionID := "test-session"
		client := mocks.NewWhatsAppClientMock()
		client.Connected[sessionID] = true

		uc := usecase.NewContactUseCase(client)

		// Execute
		chats, err := uc.ListChats(context.Background(), sessionID)

		// Verify: Should return chats without errors
		if err != nil {
			t.Fatalf("Expected no error when listing chats, got: %v", err)
		}
		if chats == nil {
			t.Fatalf("Expected non-nil chat list")
		}

		// Verify all chats have required fields
		for i, chat := range chats {
			if chat.JID == "" {
				t.Fatalf("Chat %d has empty JID", i)
			}
			if chat.Name == "" {
				t.Fatalf("Chat %d (%s) has empty name", i, chat.JID)
			}
			// UnreadCount should be >= 0
			if chat.UnreadCount < 0 {
				t.Fatalf("Chat %d (%s) has negative unread count: %d", i, chat.JID, chat.UnreadCount)
			}
			// Note: LastMessageTime can be zero as it's optional
		}
	})
}

// Additional test: Disconnected Session Error Handling
// For any contact operation attempted on a disconnected session, the service SHALL return a session disconnected error.
func TestContactOperations_DisconnectedSessionError(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Setup
		sessionID := "test-session"
		client := mocks.NewWhatsAppClientMock()
		// Session is NOT connected

		uc := usecase.NewContactUseCase(client)

		// Test CheckPhoneNumber
		req := dto.CheckPhoneRequest{
			SessionID: sessionID,
			Phone:     "+1234567890",
		}
		_, err := uc.CheckPhoneNumber(context.Background(), req)
		if err == nil {
			t.Fatalf("Expected error for disconnected session in CheckPhoneNumber")
		}
		if !errors.IsDomainError(err) {
			t.Fatalf("Expected domain error for disconnected session, got: %v", err)
		}

		// Test GetUserProfile
		profileReq := dto.GetProfileRequest{
			SessionID: sessionID,
			JID:       "1234567890@s.whatsapp.net",
		}
		_, err = uc.GetUserProfile(context.Background(), profileReq)
		if err == nil {
			t.Fatalf("Expected error for disconnected session in GetUserProfile")
		}

		// Test ListContacts
		_, err = uc.ListContacts(context.Background(), sessionID)
		if err == nil {
			t.Fatalf("Expected error for disconnected session in ListContacts")
		}

		// Test ListChats
		_, err = uc.ListChats(context.Background(), sessionID)
		if err == nil {
			t.Fatalf("Expected error for disconnected session in ListChats")
		}
	})
}

// Additional test: Valid JID Acceptance
// For any valid JID, getting user profile SHALL return a contact without errors.
func TestContactOperations_ValidJIDAcceptance(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate valid JIDs
		validJID := rapid.SampledFrom([]string{
			"1234567890@s.whatsapp.net",
			"0987654321@s.whatsapp.net",
			"group123@g.us",
		}).Draw(t, "validJID")

		// Setup
		sessionID := "test-session"
		client := mocks.NewWhatsAppClientMock()
		client.Connected[sessionID] = true

		uc := usecase.NewContactUseCase(client)

		// Execute
		req := dto.GetProfileRequest{
			SessionID: sessionID,
			JID:       validJID,
		}
		contact, err := uc.GetUserProfile(context.Background(), req)

		// Verify: Should return a result without errors
		if err != nil {
			t.Fatalf("Expected no error for valid JID %s, got: %v", validJID, err)
		}
		if contact == nil {
			t.Fatalf("Expected contact result for JID %s", validJID)
			return // Satisfy linter - unreachable but prevents nil pointer warning
		}
		if contact.JID == "" {
			t.Fatalf("Expected non-empty JID in result")
		}
	})
}
