package property

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"whatspire/internal/domain/repository"
	"whatspire/internal/infrastructure/storage"

	"github.com/google/uuid"
	"pgregory.net/rapid"
)

// TestProperty3_MediaDownloadAndStorageRoundTrip tests that media can be downloaded and stored, then retrieved
// **Validates: Requirements 1.4**
func TestProperty3_MediaDownloadAndStorageRoundTrip(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	rapid.Check(t, func(rt *rapid.T) {
		// Create media storage
		config := repository.MediaStorageConfig{
			BasePath:    tempDir,
			BaseURL:     "http://localhost:8080/media",
			MaxFileSize: 16 * 1024 * 1024, // 16MB
		}
		mediaStorage, err := storage.NewLocalMediaStorage(config)
		if err != nil {
			rt.Fatalf("Failed to create media storage: %v", err)
		}

		// Generate test data
		sessionID := uuid.New().String()
		messageID := uuid.New().String()
		mediaData := rapid.SliceOfN(rapid.Byte(), 100, 1000).Draw(rt, "media_data")
		mimeType := rapid.SampledFrom([]string{
			"image/jpeg",
			"image/png",
			"video/mp4",
			"audio/ogg",
			"application/pdf",
		}).Draw(rt, "mime_type")
		extension := rapid.SampledFrom([]string{".jpg", ".png", ".mp4", ".ogg", ".pdf"}).Draw(rt, "extension")

		ctx := context.Background()

		// Store media
		reader := bytes.NewReader(mediaData)
		filePath, publicURL, err := mediaStorage.DownloadAndStore(ctx, sessionID, messageID, reader, mimeType, extension)
		if err != nil {
			rt.Fatalf("Failed to store media: %v", err)
		}

		// Property: File path should not be empty
		if filePath == "" {
			rt.Fatalf("File path should not be empty")
		}

		// Property: Public URL should not be empty
		if publicURL == "" {
			rt.Fatalf("Public URL should not be empty")
		}

		// Property: File should exist on disk
		fullPath := mediaStorage.GetMediaPath(filePath)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			rt.Fatalf("File should exist on disk at %s", fullPath)
		}

		// Property: Stored file should have the same content
		storedData, err := os.ReadFile(fullPath)
		if err != nil {
			rt.Fatalf("Failed to read stored file: %v", err)
		}
		if !bytes.Equal(storedData, mediaData) {
			rt.Fatalf("Stored data does not match original data")
		}

		// Property: Public URL should contain the file path
		if publicURL != mediaStorage.GetMediaURL(filePath) {
			rt.Fatalf("Public URL mismatch")
		}

		// Cleanup: Delete media
		err = mediaStorage.DeleteMedia(ctx, filePath)
		if err != nil {
			rt.Fatalf("Failed to delete media: %v", err)
		}

		// Property: File should not exist after deletion
		if _, err := os.Stat(fullPath); !os.IsNotExist(err) {
			rt.Fatalf("File should not exist after deletion")
		}
	})
}

// TestProperty29_MediaSizeLimitEnforcement tests that media exceeding size limits is rejected
// **Validates: Requirements 10.3**
func TestProperty29_MediaSizeLimitEnforcement(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	rapid.Check(t, func(rt *rapid.T) {
		// Create media storage with small size limit
		maxSize := int64(1024) // 1KB limit
		config := repository.MediaStorageConfig{
			BasePath:    tempDir,
			BaseURL:     "http://localhost:8080/media",
			MaxFileSize: maxSize,
		}
		mediaStorage, err := storage.NewLocalMediaStorage(config)
		if err != nil {
			rt.Fatalf("Failed to create media storage: %v", err)
		}

		// Generate test data larger than limit
		oversizedData := rapid.SliceOfN(rapid.Byte(), int(maxSize)+100, int(maxSize)+1000).Draw(rt, "oversized_data")

		sessionID := uuid.New().String()
		messageID := uuid.New().String()
		mimeType := "image/jpeg"
		extension := ".jpg"

		ctx := context.Background()

		// Try to store oversized media
		reader := bytes.NewReader(oversizedData)
		_, _, err = mediaStorage.DownloadAndStore(ctx, sessionID, messageID, reader, mimeType, extension)

		// Property: Oversized media should be rejected
		if err == nil {
			rt.Fatalf("Oversized media should be rejected")
		}

		// Note: The implementation creates the file during write, then deletes it on validation failure
		// So we don't check for file existence here as it's implementation-dependent
	})
}

// TestMediaStorageValidation tests media storage validation
func TestMediaStorageValidation(t *testing.T) {
	tempDir := t.TempDir()

	rapid.Check(t, func(rt *rapid.T) {
		maxSize := rapid.Int64Range(1024, 16*1024*1024).Draw(rt, "max_size")
		config := repository.MediaStorageConfig{
			BasePath:    tempDir,
			BaseURL:     "http://localhost:8080/media",
			MaxFileSize: maxSize,
		}
		mediaStorage, err := storage.NewLocalMediaStorage(config)
		if err != nil {
			rt.Fatalf("Failed to create media storage: %v", err)
		}

		// Property: Size within limit should be valid
		validSize := rapid.Int64Range(1, maxSize).Draw(rt, "valid_size")
		if err := mediaStorage.ValidateSize(validSize); err != nil {
			rt.Fatalf("Size within limit should be valid: %v", err)
		}

		// Property: Size exceeding limit should be invalid
		invalidSize := maxSize + rapid.Int64Range(1, 1024*1024).Draw(rt, "excess")
		if err := mediaStorage.ValidateSize(invalidSize); err == nil {
			rt.Fatalf("Size exceeding limit should be invalid")
		}
	})
}

// TestMediaURLGeneration tests media URL generation
func TestMediaURLGeneration(t *testing.T) {
	tempDir := t.TempDir()

	rapid.Check(t, func(rt *rapid.T) {
		baseURL := rapid.SampledFrom([]string{
			"http://localhost:8080/media",
			"https://example.com/media",
			"http://localhost:8080/media/",
		}).Draw(rt, "base_url")

		config := repository.MediaStorageConfig{
			BasePath:    tempDir,
			BaseURL:     baseURL,
			MaxFileSize: 16 * 1024 * 1024,
		}
		mediaStorage, err := storage.NewLocalMediaStorage(config)
		if err != nil {
			rt.Fatalf("Failed to create media storage: %v", err)
		}

		sessionID := rapid.String().Draw(rt, "session_id")
		filename := rapid.String().Draw(rt, "filename")
		filePath := filepath.Join(sessionID, filename)

		// Generate URL
		url := mediaStorage.GetMediaURL(filePath)

		// Property: URL should not be empty
		if url == "" {
			rt.Fatalf("URL should not be empty")
		}

		// Property: URL should contain the file path
		// Normalize path separators for comparison
		normalizedPath := filepath.ToSlash(filePath)
		if url != "http://localhost:8080/media/"+normalizedPath &&
			url != "https://example.com/media/"+normalizedPath {
			rt.Fatalf("URL should contain the file path")
		}
	})
}
