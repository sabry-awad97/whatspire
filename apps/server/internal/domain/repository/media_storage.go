package repository

import (
	"context"
	"io"
)

// MediaStorage defines operations for storing and retrieving media files
type MediaStorage interface {
	// DownloadAndStore downloads media from WhatsApp and stores it locally
	// Returns the local file path and public URL
	DownloadAndStore(ctx context.Context, sessionID, messageID string, mediaData io.Reader, mimeType, extension string) (string, string, error)

	// GetMediaURL returns the public URL for accessing stored media
	GetMediaURL(filePath string) string

	// DeleteMedia removes a media file from storage
	DeleteMedia(ctx context.Context, filePath string) error

	// GetMediaPath returns the full local path for a media file
	GetMediaPath(filePath string) string

	// ValidateSize checks if the media size is within allowed limits
	ValidateSize(size int64) error
}

// MediaStorageConfig holds configuration for media storage
type MediaStorageConfig struct {
	// BasePath is the local directory where media files are stored
	BasePath string

	// BaseURL is the public URL prefix for accessing media files
	BaseURL string

	// MaxFileSize is the maximum allowed file size in bytes
	MaxFileSize int64
}
