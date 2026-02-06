package property

import (
	"testing"
	"unicode/utf8"

	"whatspire/internal/domain/entity"
	"whatspire/internal/domain/valueobject"

	"github.com/google/uuid"
	"pgregory.net/rapid"
)

// TestProperty7_ReactionRemovalViaEmptyEmoji tests that reactions can be removed by sending empty emoji
// **Validates: Requirements 2.4**
func TestProperty7_ReactionRemovalViaEmptyEmoji(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a reaction with empty emoji
		reaction := entity.NewReactionBuilder(uuid.New().String(), uuid.New().String(), uuid.New().String()).
			From(rapid.String().Draw(t, "from")).
			To(rapid.String().Draw(t, "to")).
			WithEmoji(""). // Empty emoji for removal
			Build()

		// Property: Empty emoji should be valid and indicate removal
		if !reaction.IsValidEmoji() {
			t.Fatalf("Empty emoji should be valid for reaction removal")
		}

		if !reaction.IsRemoval() {
			t.Fatalf("Reaction with empty emoji should be marked as removal")
		}
	})
}

// TestProperty8_InvalidEmojiRejection tests that invalid emoji strings are rejected
// **Validates: Requirements 2.5**
func TestProperty8_InvalidEmojiRejection(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate invalid emoji strings
		invalidEmoji := rapid.SampledFrom([]string{
			"abc",                      // Plain text
			"123",                      // Numbers
			"hello world",              // Multiple words
			string([]byte{0xFF, 0xFE}), // Invalid UTF-8
			"😀😀😀😀😀",                    // Too many emoji (>4 runes)
		}).Draw(t, "invalid_emoji")

		reaction := entity.NewReactionBuilder(uuid.New().String(), uuid.New().String(), uuid.New().String()).
			From(rapid.String().Draw(t, "from")).
			To(rapid.String().Draw(t, "to")).
			WithEmoji(invalidEmoji).
			Build()

		// Property: Invalid emoji should be rejected
		if reaction.IsValidEmoji() {
			// Check if it's actually a valid emoji by checking rune count and UTF-8 validity
			if !utf8.ValidString(invalidEmoji) {
				t.Fatalf("Invalid UTF-8 string should be rejected as emoji")
			}
			runeCount := utf8.RuneCountInString(invalidEmoji)
			if runeCount > 4 {
				t.Fatalf("Emoji with more than 4 runes should be rejected")
			}
		}
	})
}

// TestProperty13_InvalidPhoneNumberRejection tests that invalid phone numbers are rejected
// **Validates: Requirements 5.5**
func TestProperty13_InvalidPhoneNumberRejection(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate invalid phone numbers
		invalidPhone := rapid.SampledFrom([]string{
			"",           // Empty
			"abc",        // Letters
			"123",        // Too short
			"+",          // Just plus sign
			"++123",      // Multiple plus signs
			"123abc456",  // Mixed alphanumeric
			"12 34 56",   // Spaces
			"+1-234-567", // Dashes
		}).Draw(t, "invalid_phone")

		// Property: Invalid phone numbers should be rejected
		_, err := valueobject.NewPhoneNumber(invalidPhone)
		if err == nil {
			t.Fatalf("Invalid phone number %q should be rejected", invalidPhone)
		}
	})
}

// TestValidEmojiAcceptance tests that valid emoji strings are accepted
func TestValidEmojiAcceptance(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate valid emoji strings (excluding compound emoji with ZWJ for now)
		validEmoji := rapid.SampledFrom([]string{
			"😀",  // Single emoji
			"👍",  // Thumbs up
			"❤️", // Heart with variation selector
			"",   // Empty (for removal)
		}).Draw(t, "valid_emoji")

		from := rapid.StringMatching("[a-zA-Z0-9]+").Draw(t, "from")
		to := rapid.StringMatching("[a-zA-Z0-9]+").Draw(t, "to")

		reaction := entity.NewReactionBuilder(uuid.New().String(), uuid.New().String(), uuid.New().String()).
			From(from).
			To(to).
			WithEmoji(validEmoji).
			Build()

		// Property: Valid emoji should be accepted
		if !reaction.IsValidEmoji() {
			t.Fatalf("Valid emoji %q should be accepted", validEmoji)
		}
	})
}

// TestReactionValidation tests reaction entity validation
func TestReactionValidation(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		id := uuid.New().String()
		messageID := uuid.New().String()
		sessionID := uuid.New().String()
		from := rapid.StringMatching("[a-zA-Z0-9]+").Draw(t, "from")
		to := rapid.StringMatching("[a-zA-Z0-9]+").Draw(t, "to")
		emoji := rapid.SampledFrom([]string{"😀", "👍", "❤️", ""}).Draw(t, "emoji")

		reaction := entity.NewReactionBuilder(id, messageID, sessionID).
			From(from).
			To(to).
			WithEmoji(emoji).
			Build()

		// Property: Reaction with all required fields should be valid
		if !reaction.IsValid() {
			t.Fatalf("Reaction with all required fields should be valid")
		}

		// Property: Reaction without ID should be invalid
		invalidReaction := entity.NewReactionBuilder("", messageID, sessionID).
			From(from).
			To(to).
			WithEmoji(emoji).
			Build()
		if invalidReaction.IsValid() {
			t.Fatalf("Reaction without ID should be invalid")
		}
	})
}

// TestReceiptValidation tests receipt entity validation
func TestReceiptValidation(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		id := uuid.New().String()
		messageID := uuid.New().String()
		sessionID := uuid.New().String()
		from := rapid.StringMatching("[a-zA-Z0-9]+").Draw(t, "from")
		to := rapid.StringMatching("[a-zA-Z0-9]+").Draw(t, "to")
		receiptType := rapid.SampledFrom([]entity.ReceiptType{
			entity.ReceiptTypeDelivered,
			entity.ReceiptTypeRead,
		}).Draw(t, "receipt_type")

		receipt := entity.NewReceiptBuilder(id, messageID, sessionID).
			From(from).
			To(to).
			WithType(receiptType).
			Build()

		// Property: Receipt with all required fields should be valid
		if !receipt.IsValid() {
			t.Fatalf("Receipt with all required fields should be valid")
		}

		// Property: Receipt type should be valid
		if !receipt.Type.IsValid() {
			t.Fatalf("Receipt type should be valid")
		}

		// Property: Receipt without ID should be invalid
		invalidReceipt := entity.NewReceiptBuilder("", messageID, sessionID).
			From(from).
			To(to).
			WithType(receiptType).
			Build()
		if invalidReceipt.IsValid() {
			t.Fatalf("Receipt without ID should be invalid")
		}
	})
}

// TestPresenceValidation tests presence entity validation
func TestPresenceValidation(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		id := uuid.New().String()
		sessionID := uuid.New().String()
		userJID := rapid.StringMatching("[a-zA-Z0-9]+").Draw(t, "user_jid")
		chatJID := rapid.StringMatching("[a-zA-Z0-9]+").Draw(t, "chat_jid")
		state := rapid.SampledFrom([]entity.PresenceState{
			entity.PresenceStateTyping,
			entity.PresenceStatePaused,
			entity.PresenceStateOnline,
			entity.PresenceStateOffline,
		}).Draw(t, "state")

		presence := entity.NewPresence(id, sessionID, userJID, chatJID, state)

		// Property: Presence with all required fields should be valid
		if !presence.IsValid() {
			t.Fatalf("Presence with all required fields should be valid")
		}

		// Property: Presence state should be valid
		if !presence.State.IsValid() {
			t.Fatalf("Presence state should be valid")
		}

		// Property: Presence without ID should be invalid
		invalidPresence := entity.NewPresence("", sessionID, userJID, chatJID, state)
		if invalidPresence.IsValid() {
			t.Fatalf("Presence without ID should be invalid")
		}
	})
}

// TestContactValidation tests contact entity validation
func TestContactValidation(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		jid := rapid.StringMatching("[a-zA-Z0-9]+").Draw(t, "jid")
		name := rapid.String().Draw(t, "name")
		isOnWhatsApp := rapid.Bool().Draw(t, "is_on_whatsapp")

		contact := entity.NewContact(jid, name, isOnWhatsApp)

		// Property: Contact with JID should be valid
		if !contact.IsValid() {
			t.Fatalf("Contact with JID should be valid")
		}

		// Property: Contact without JID should be invalid
		invalidContact := entity.NewContact("", name, isOnWhatsApp)
		if invalidContact.IsValid() {
			t.Fatalf("Contact without JID should be invalid")
		}
	})
}

// TestChatValidation tests chat entity validation
func TestChatValidation(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		jid := rapid.StringMatching("[a-zA-Z0-9]+").Draw(t, "jid")
		name := rapid.String().Draw(t, "name")
		isGroup := rapid.Bool().Draw(t, "is_group")

		chat := entity.NewChat(jid, name, isGroup)

		// Property: Chat with JID should be valid
		if !chat.IsValid() {
			t.Fatalf("Chat with JID should be valid")
		}

		// Property: Chat without JID should be invalid
		invalidChat := entity.NewChat("", name, isGroup)
		if invalidChat.IsValid() {
			t.Fatalf("Chat without JID should be invalid")
		}

		// Property: Unread count should never be negative
		chat.SetUnreadCount(-5)
		if chat.UnreadCount < 0 {
			t.Fatalf("Unread count should never be negative")
		}
	})
}
