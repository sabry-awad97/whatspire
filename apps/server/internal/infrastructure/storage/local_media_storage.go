package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"whatspire/internal/domain/errors"
	"whatspire/internal/domain/repository"
)

// LocalMediaStorage implements MediaStorage using local filesystem
type LocalMediaStorage struct {
	config repository.MediaStorageConfig
}

// NewLocalMediaStorage creates a new local media storage instance
func NewLocalMediaStorage(config repository.MediaStorageConfig) (*LocalMediaStorage, error) {
	// Ensure base path exists
	if err := os.MkdirAll(config.BasePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create media directory: %w", err)
	}

	return &LocalMediaStorage{
		config: config,
	}, nil
}

// DownloadAndStore downloads media from WhatsApp and stores it locally
func (s *LocalMediaStorage) DownloadAndStore(ctx context.Context, sessionID, messageID string, mediaData io.Reader, mimeType, extension string) (string, string, error) {
	// Generate unique filename
	filename := s.generateFilename(sessionID, messageID, extension)

	// Create session directory if it doesn't exist
	sessionDir := filepath.Join(s.config.BasePath, sessionID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create session directory: %w", err)
	}

	// Full file path
	filePath := filepath.Join(sessionDir, filename)

	// Create file
	file, err := os.Create(filePath)
	if err != nil {
		return "", "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy media data to file with size tracking
	written, err := io.Copy(file, mediaData)
	if err != nil {
		// Clean up partial file
		_ = os.Remove(filePath)
		return "", "", fmt.Errorf("failed to write media data: %w", err)
	}

	// Validate size
	if err := s.ValidateSize(written); err != nil {
		// Clean up file that exceeds size limit
		_ = os.Remove(filePath)
		return "", "", err
	}

	// Generate relative path for storage
	relativePath := filepath.Join(sessionID, filename)

	// Generate public URL
	publicURL := s.GetMediaURL(relativePath)

	return relativePath, publicURL, nil
}

// GetMediaURL returns the public URL for accessing stored media
func (s *LocalMediaStorage) GetMediaURL(filePath string) string {
	// Normalize path separators for URLs
	urlPath := strings.ReplaceAll(filePath, "\\", "/")
	return fmt.Sprintf("%s/%s", strings.TrimSuffix(s.config.BaseURL, "/"), urlPath)
}

// DeleteMedia removes a media file from storage
func (s *LocalMediaStorage) DeleteMedia(ctx context.Context, filePath string) error {
	fullPath := s.GetMediaPath(filePath)

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return errors.ErrNotFound
		}
		return fmt.Errorf("failed to delete media file: %w", err)
	}

	return nil
}

// GetMediaPath returns the full local path for a media file
func (s *LocalMediaStorage) GetMediaPath(filePath string) string {
	return filepath.Join(s.config.BasePath, filePath)
}

// ValidateSize checks if the media size is within allowed limits
func (s *LocalMediaStorage) ValidateSize(size int64) error {
	if size > s.config.MaxFileSize {
		return errors.ErrMediaTooLarge.WithMessage(
			fmt.Sprintf("media size %d bytes exceeds maximum allowed size %d bytes", size, s.config.MaxFileSize),
		)
	}
	return nil
}

// generateFilename generates a unique filename for media
func (s *LocalMediaStorage) generateFilename(sessionID, messageID, extension string) string {
	// Ensure extension starts with a dot
	if extension != "" && !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}

	// Use message ID as filename for uniqueness
	return fmt.Sprintf("%s%s", messageID, extension)
}

// GetConfig returns the storage configuration
func (s *LocalMediaStorage) GetConfig() repository.MediaStorageConfig {
	return s.config
}
