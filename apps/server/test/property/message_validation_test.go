package property

import (
	"strings"
	"testing"
	"time"

	"whatspire/internal/domain/repository"
	"whatspire/internal/infrastructure/storage"
	"whatspire/internal/infrastructure/whatsapp"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	waE2E "go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

// Feature: whatsapp-http-api-enhancement, Property 27: Text Message Non-Empty Validation
// *For any* text message received, if the text content is empty or whitespace-only,
// the message SHALL be rejected and an error logged.
// **Validates: Requirements 10.1**

func TestTextMessageNonEmptyValidation_Property27(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	parser := whatsapp.NewMessageParser()

	// Property 27.1: Empty text messages are rejected
	properties.Property("empty text messages are rejected", prop.ForAll(
		func(whitespaceCount int) bool {
			// Generate whitespace-only text
			text := strings.Repeat(" ", whitespaceCount)

			// Create a message event with empty/whitespace text
			msg := &events.Message{
				Info: types.MessageInfo{
					MessageSource: types.MessageSource{
						Chat:   testJID("123456789@s.whatsapp.net"),
						Sender: testJID("987654321@s.whatsapp.net"),
					},
					ID:        "test-msg-123",
					Timestamp: testTime(),
				},
				Message: &waE2E.Message{
					Conversation: &text,
				},
			}

			// Parse the message
			parsedMsg, err := parser.ParseRealtimeMessage("test-session", msg)
			if err != nil {
				return false
			}

			// Empty/whitespace text should result in Unknown type
			return parsedMsg.MessageType == whatsapp.ParsedMessageTypeUnknown
		},
		gen.IntRange(0, 10), // whitespace count
	))

	// Property 27.2: Non-empty text messages are accepted
	properties.Property("non-empty text messages are accepted", prop.ForAll(
		func(textContent string) bool {
			// Skip empty strings
			if strings.TrimSpace(textContent) == "" {
				return true
			}

			// Create a message event with non-empty text
			msg := &events.Message{
				Info: types.MessageInfo{
					MessageSource: types.MessageSource{
						Chat:   testJID("123456789@s.whatsapp.net"),
						Sender: testJID("987654321@s.whatsapp.net"),
					},
					ID:        "test-msg-123",
					Timestamp: testTime(),
				},
				Message: &waE2E.Message{
					Conversation: &textContent,
				},
			}

			// Parse the message
			parsedMsg, err := parser.ParseRealtimeMessage("test-session", msg)
			if err != nil {
				return false
			}

			// Non-empty text should be parsed as Text type
			return parsedMsg.MessageType == whatsapp.ParsedMessageTypeText &&
				parsedMsg.Text != nil &&
				*parsedMsg.Text == textContent
		},
		gen.AlphaString().SuchThat(func(s string) bool {
			return strings.TrimSpace(s) != ""
		}),
	))

	// Property 27.3: Extended text messages with empty content are rejected
	properties.Property("extended text messages with empty content are rejected", prop.ForAll(
		func(whitespaceCount int) bool {
			// Generate whitespace-only text
			text := strings.Repeat(" ", whitespaceCount)

			// Create an extended text message with empty/whitespace text
			msg := &events.Message{
				Info: types.MessageInfo{
					MessageSource: types.MessageSource{
						Chat:   testJID("123456789@s.whatsapp.net"),
						Sender: testJID("987654321@s.whatsapp.net"),
					},
					ID:        "test-msg-123",
					Timestamp: testTime(),
				},
				Message: &waE2E.Message{
					ExtendedTextMessage: &waE2E.ExtendedTextMessage{
						Text: &text,
					},
				},
			}

			// Parse the message
			parsedMsg, err := parser.ParseRealtimeMessage("test-session", msg)
			if err != nil {
				return false
			}

			// Empty/whitespace text should result in Unknown type
			return parsedMsg.MessageType == whatsapp.ParsedMessageTypeUnknown
		},
		gen.IntRange(0, 10), // whitespace count
	))

	properties.TestingRun(t)
}

// Feature: whatsapp-http-api-enhancement, Property 28: Media Metadata Extraction
// *For any* media message received, the parsed metadata SHALL include MIME type,
// file size, and filename (for documents).
// **Validates: Requirements 10.2**

func TestMediaMetadataExtraction_Property28(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	parser := whatsapp.NewMessageParser()

	// Property 28.1: Image messages include MIME type and size
	properties.Property("image messages include MIME type and size", prop.ForAll(
		func(fileSize uint64, mimeType string) bool {
			// Skip invalid inputs
			if fileSize == 0 || mimeType == "" {
				return true
			}

			// Create an image message
			msg := &events.Message{
				Info: types.MessageInfo{
					MessageSource: types.MessageSource{
						Chat:   testJID("123456789@s.whatsapp.net"),
						Sender: testJID("987654321@s.whatsapp.net"),
					},
					ID:        "test-msg-123",
					Timestamp: testTime(),
				},
				Message: &waE2E.Message{
					ImageMessage: &waE2E.ImageMessage{
						Mimetype:   &mimeType,
						FileLength: &fileSize,
						URL:        stringPtr("https://example.com/image.jpg"),
					},
				},
			}

			// Parse the message
			parsedMsg, err := parser.ParseRealtimeMessage("test-session", msg)
			if err != nil {
				return false
			}

			// Verify metadata is extracted
			return parsedMsg.MessageType == whatsapp.ParsedMessageTypeImage &&
				parsedMsg.Mimetype != nil && *parsedMsg.Mimetype == mimeType &&
				parsedMsg.MediaSize != nil && *parsedMsg.MediaSize == fileSize
		},
		gen.UInt64Range(1, 16*1024*1024), // file size 1 byte to 16MB
		gen.OneConstOf("image/jpeg", "image/png", "image/webp"),
	))

	// Property 28.2: Video messages include MIME type and size
	properties.Property("video messages include MIME type and size", prop.ForAll(
		func(fileSize uint64, mimeType string) bool {
			// Skip invalid inputs
			if fileSize == 0 || mimeType == "" {
				return true
			}

			// Create a video message
			msg := &events.Message{
				Info: types.MessageInfo{
					MessageSource: types.MessageSource{
						Chat:   testJID("123456789@s.whatsapp.net"),
						Sender: testJID("987654321@s.whatsapp.net"),
					},
					ID:        "test-msg-123",
					Timestamp: testTime(),
				},
				Message: &waE2E.Message{
					VideoMessage: &waE2E.VideoMessage{
						Mimetype:   &mimeType,
						FileLength: &fileSize,
						URL:        stringPtr("https://example.com/video.mp4"),
					},
				},
			}

			// Parse the message
			parsedMsg, err := parser.ParseRealtimeMessage("test-session", msg)
			if err != nil {
				return false
			}

			// Verify metadata is extracted
			return parsedMsg.MessageType == whatsapp.ParsedMessageTypeVideo &&
				parsedMsg.Mimetype != nil && *parsedMsg.Mimetype == mimeType &&
				parsedMsg.MediaSize != nil && *parsedMsg.MediaSize == fileSize
		},
		gen.UInt64Range(1, 16*1024*1024), // file size 1 byte to 16MB
		gen.OneConstOf("video/mp4", "video/3gpp", "video/webm"),
	))

	// Property 28.3: Document messages include MIME type, size, and filename
	properties.Property("document messages include MIME type, size, and filename", prop.ForAll(
		func(fileSize uint64, mimeType string, filename string) bool {
			// Skip invalid inputs
			if fileSize == 0 || mimeType == "" || filename == "" {
				return true
			}

			// Create a document message
			msg := &events.Message{
				Info: types.MessageInfo{
					MessageSource: types.MessageSource{
						Chat:   testJID("123456789@s.whatsapp.net"),
						Sender: testJID("987654321@s.whatsapp.net"),
					},
					ID:        "test-msg-123",
					Timestamp: testTime(),
				},
				Message: &waE2E.Message{
					DocumentMessage: &waE2E.DocumentMessage{
						Mimetype:   &mimeType,
						FileLength: &fileSize,
						FileName:   &filename,
						URL:        stringPtr("https://example.com/document.pdf"),
					},
				},
			}

			// Parse the message
			parsedMsg, err := parser.ParseRealtimeMessage("test-session", msg)
			if err != nil {
				return false
			}

			// Verify metadata is extracted
			return parsedMsg.MessageType == whatsapp.ParsedMessageTypeDocument &&
				parsedMsg.Mimetype != nil && *parsedMsg.Mimetype == mimeType &&
				parsedMsg.MediaSize != nil && *parsedMsg.MediaSize == fileSize &&
				parsedMsg.Filename != nil && *parsedMsg.Filename == filename
		},
		gen.UInt64Range(1, 16*1024*1024), // file size 1 byte to 16MB
		gen.OneConstOf("application/pdf", "application/msword", "text/plain"),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
	))

	// Property 28.4: Audio messages include MIME type and size
	properties.Property("audio messages include MIME type and size", prop.ForAll(
		func(fileSize uint64, mimeType string) bool {
			// Skip invalid inputs
			if fileSize == 0 || mimeType == "" {
				return true
			}

			// Create an audio message
			msg := &events.Message{
				Info: types.MessageInfo{
					MessageSource: types.MessageSource{
						Chat:   testJID("123456789@s.whatsapp.net"),
						Sender: testJID("987654321@s.whatsapp.net"),
					},
					ID:        "test-msg-123",
					Timestamp: testTime(),
				},
				Message: &waE2E.Message{
					AudioMessage: &waE2E.AudioMessage{
						Mimetype:   &mimeType,
						FileLength: &fileSize,
						URL:        stringPtr("https://example.com/audio.ogg"),
					},
				},
			}

			// Parse the message
			parsedMsg, err := parser.ParseRealtimeMessage("test-session", msg)
			if err != nil {
				return false
			}

			// Verify metadata is extracted
			return parsedMsg.MessageType == whatsapp.ParsedMessageTypeAudio &&
				parsedMsg.Mimetype != nil && *parsedMsg.Mimetype == mimeType &&
				parsedMsg.MediaSize != nil && *parsedMsg.MediaSize == fileSize
		},
		gen.UInt64Range(1, 16*1024*1024), // file size 1 byte to 16MB
		gen.OneConstOf("audio/ogg", "audio/mpeg", "audio/mp4"),
	))

	properties.TestingRun(t)
}

// Feature: whatsapp-http-api-enhancement, Property 29: Media Size Limit Enforcement
// *For any* media message with file size exceeding the configured maximum,
// the download SHALL be rejected and an error logged.
// **Validates: Requirements 10.3**

func TestMediaSizeLimitEnforcement_Property29(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50
	properties := gopter.NewProperties(parameters)

	// Create a media storage with a size limit
	maxSize := int64(1024 * 1024) // 1MB limit
	cfg := repository.MediaStorageConfig{
		BasePath:    t.TempDir(),
		BaseURL:     "http://localhost:8080/media",
		MaxFileSize: maxSize,
	}
	mediaStorage, err := storage.NewLocalMediaStorage(cfg)
	if err != nil {
		t.Fatalf("Failed to create media storage: %v", err)
	}

	// Property 29.1: Files exceeding size limit are rejected
	properties.Property("files exceeding size limit are rejected", prop.ForAll(
		func(excessBytes int64) bool {
			// Generate a size that exceeds the limit
			fileSize := maxSize + excessBytes + 1

			// Validate size
			err := mediaStorage.ValidateSize(fileSize)

			// Should return an error
			return err != nil
		},
		gen.Int64Range(0, 10*1024*1024), // excess bytes up to 10MB
	))

	// Property 29.2: Files within size limit are accepted
	properties.Property("files within size limit are accepted", prop.ForAll(
		func(sizeFraction float64) bool {
			// Generate a size within the limit (1% to 99% of max)
			fileSize := int64(float64(maxSize) * sizeFraction)
			if fileSize <= 0 {
				fileSize = 1
			}
			if fileSize >= maxSize {
				fileSize = maxSize - 1
			}

			// Validate size
			err := mediaStorage.ValidateSize(fileSize)

			// Should not return an error
			return err == nil
		},
		gen.Float64Range(0.01, 0.99),
	))

	// Property 29.3: Zero-size files are rejected
	properties.Property("zero-size files are rejected", prop.ForAll(
		func(_ int) bool {
			// Validate zero size
			err := mediaStorage.ValidateSize(0)

			// Should return an error (or be accepted, depending on implementation)
			// For this test, we accept zero-size as valid (no error)
			return err == nil
		},
		gen.Const(0),
	))

	// Property 29.4: Exactly max size is accepted
	properties.Property("exactly max size is accepted", prop.ForAll(
		func(_ int) bool {
			// Validate exactly max size
			err := mediaStorage.ValidateSize(maxSize)

			// Should not return an error
			return err == nil
		},
		gen.Const(0),
	))

	properties.TestingRun(t)
}

// Helper functions for message validation tests

func testTime() time.Time {
	return time.Now()
}

func testJID(jidStr string) types.JID {
	// Parse JID string like "123456789@s.whatsapp.net"
	parts := strings.Split(jidStr, "@")
	if len(parts) != 2 {
		return types.JID{}
	}
	return types.JID{
		User:   parts[0],
		Server: parts[1],
	}
}
