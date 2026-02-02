package whatsapp

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"

	"whatspire/internal/domain/errors"
	"whatspire/internal/domain/repository"

	"go.mau.fi/whatsmeow"
	waE2E "go.mau.fi/whatsmeow/proto/waE2E"
)

// MediaDownloadHelper helps download and store media from WhatsApp messages
type MediaDownloadHelper struct {
	storage repository.MediaStorage
}

// NewMediaDownloadHelper creates a new media download helper
func NewMediaDownloadHelper(storage repository.MediaStorage) *MediaDownloadHelper {
	return &MediaDownloadHelper{
		storage: storage,
	}
}

// DownloadAndStoreImage downloads an image from WhatsApp and stores it
func (h *MediaDownloadHelper) DownloadAndStoreImage(
	ctx context.Context,
	client *whatsmeow.Client,
	sessionID, messageID string,
	imageMsg *waE2E.ImageMessage,
) (string, string, error) {
	// Check file size before downloading
	fileSize := int64(imageMsg.GetFileLength())
	if err := h.storage.ValidateSize(fileSize); err != nil {
		return "", "", err
	}

	// Download image data from WhatsApp
	data, err := client.Download(ctx, imageMsg)
	if err != nil {
		return "", "", errors.ErrMediaDownloadFailed.WithCause(err)
	}

	// Determine MIME type and extension
	mimeType := imageMsg.GetMimetype()
	if mimeType == "" {
		mimeType = "image/jpeg" // Default
	}
	extension := getExtensionFromMimeType(mimeType)

	// Store the media
	reader := bytes.NewReader(data)
	filePath, publicURL, err := h.storage.DownloadAndStore(ctx, sessionID, messageID, reader, mimeType, extension)
	if err != nil {
		return "", "", err
	}

	return filePath, publicURL, nil
}

// DownloadAndStoreVideo downloads a video from WhatsApp and stores it
func (h *MediaDownloadHelper) DownloadAndStoreVideo(
	ctx context.Context,
	client *whatsmeow.Client,
	sessionID, messageID string,
	videoMsg *waE2E.VideoMessage,
) (string, string, error) {
	// Check file size before downloading
	fileSize := int64(videoMsg.GetFileLength())
	if err := h.storage.ValidateSize(fileSize); err != nil {
		return "", "", err
	}

	// Download video data from WhatsApp
	data, err := client.Download(ctx, videoMsg)
	if err != nil {
		return "", "", errors.ErrMediaDownloadFailed.WithCause(err)
	}

	// Determine MIME type and extension
	mimeType := videoMsg.GetMimetype()
	if mimeType == "" {
		mimeType = "video/mp4" // Default
	}
	extension := getExtensionFromMimeType(mimeType)

	// Store the media
	reader := bytes.NewReader(data)
	filePath, publicURL, err := h.storage.DownloadAndStore(ctx, sessionID, messageID, reader, mimeType, extension)
	if err != nil {
		return "", "", err
	}

	return filePath, publicURL, nil
}

// DownloadAndStoreAudio downloads an audio file from WhatsApp and stores it
func (h *MediaDownloadHelper) DownloadAndStoreAudio(
	ctx context.Context,
	client *whatsmeow.Client,
	sessionID, messageID string,
	audioMsg *waE2E.AudioMessage,
) (string, string, error) {
	// Check file size before downloading
	fileSize := int64(audioMsg.GetFileLength())
	if err := h.storage.ValidateSize(fileSize); err != nil {
		return "", "", err
	}

	// Download audio data from WhatsApp
	data, err := client.Download(ctx, audioMsg)
	if err != nil {
		return "", "", errors.ErrMediaDownloadFailed.WithCause(err)
	}

	// Determine MIME type and extension
	mimeType := audioMsg.GetMimetype()
	if mimeType == "" {
		mimeType = "audio/ogg" // Default for WhatsApp voice messages
	}
	extension := getExtensionFromMimeType(mimeType)

	// Store the media
	reader := bytes.NewReader(data)
	filePath, publicURL, err := h.storage.DownloadAndStore(ctx, sessionID, messageID, reader, mimeType, extension)
	if err != nil {
		return "", "", err
	}

	return filePath, publicURL, nil
}

// DownloadAndStoreDocument downloads a document from WhatsApp and stores it
func (h *MediaDownloadHelper) DownloadAndStoreDocument(
	ctx context.Context,
	client *whatsmeow.Client,
	sessionID, messageID string,
	docMsg *waE2E.DocumentMessage,
) (string, string, error) {
	// Check file size before downloading
	fileSize := int64(docMsg.GetFileLength())
	if err := h.storage.ValidateSize(fileSize); err != nil {
		return "", "", err
	}

	// Download document data from WhatsApp
	data, err := client.Download(ctx, docMsg)
	if err != nil {
		return "", "", errors.ErrMediaDownloadFailed.WithCause(err)
	}

	// Determine MIME type and extension
	mimeType := docMsg.GetMimetype()
	if mimeType == "" {
		mimeType = "application/octet-stream" // Default
	}

	// Try to get extension from filename first
	extension := ""
	if docMsg.FileName != nil && *docMsg.FileName != "" {
		extension = filepath.Ext(*docMsg.FileName)
	}
	if extension == "" {
		extension = getExtensionFromMimeType(mimeType)
	}

	// Store the media
	reader := bytes.NewReader(data)
	filePath, publicURL, err := h.storage.DownloadAndStore(ctx, sessionID, messageID, reader, mimeType, extension)
	if err != nil {
		return "", "", err
	}

	return filePath, publicURL, nil
}

// DownloadAndStoreSticker downloads a sticker from WhatsApp and stores it
func (h *MediaDownloadHelper) DownloadAndStoreSticker(
	ctx context.Context,
	client *whatsmeow.Client,
	sessionID, messageID string,
	stickerMsg *waE2E.StickerMessage,
) (string, string, error) {
	// Check file size before downloading
	fileSize := int64(stickerMsg.GetFileLength())
	if err := h.storage.ValidateSize(fileSize); err != nil {
		return "", "", err
	}

	// Download sticker data from WhatsApp
	data, err := client.Download(ctx, stickerMsg)
	if err != nil {
		return "", "", errors.ErrMediaDownloadFailed.WithCause(err)
	}

	// Determine MIME type and extension
	mimeType := stickerMsg.GetMimetype()
	if mimeType == "" {
		mimeType = "image/webp" // Default for WhatsApp stickers
	}
	extension := getExtensionFromMimeType(mimeType)

	// Store the media
	reader := bytes.NewReader(data)
	filePath, publicURL, err := h.storage.DownloadAndStore(ctx, sessionID, messageID, reader, mimeType, extension)
	if err != nil {
		return "", "", err
	}

	return filePath, publicURL, nil
}

// getExtensionFromMimeType returns a file extension for a given MIME type
func getExtensionFromMimeType(mimeType string) string {
	// Normalize MIME type
	mimeType = strings.ToLower(strings.TrimSpace(mimeType))

	// Common MIME type to extension mappings
	extensions := map[string]string{
		// Images
		"image/jpeg":    ".jpg",
		"image/jpg":     ".jpg",
		"image/png":     ".png",
		"image/gif":     ".gif",
		"image/webp":    ".webp",
		"image/bmp":     ".bmp",
		"image/svg+xml": ".svg",

		// Videos
		"video/mp4":       ".mp4",
		"video/mpeg":      ".mpeg",
		"video/quicktime": ".mov",
		"video/x-msvideo": ".avi",
		"video/webm":      ".webm",
		"video/3gpp":      ".3gp",

		// Audio
		"audio/mpeg":  ".mp3",
		"audio/ogg":   ".ogg",
		"audio/wav":   ".wav",
		"audio/webm":  ".weba",
		"audio/aac":   ".aac",
		"audio/x-m4a": ".m4a",
		"audio/mp4":   ".m4a",
		"audio/opus":  ".opus",

		// Documents
		"application/pdf":    ".pdf",
		"application/msword": ".doc",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": ".docx",
		"application/vnd.ms-excel": ".xls",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         ".xlsx",
		"application/vnd.ms-powerpoint":                                             ".ppt",
		"application/vnd.openxmlformats-officedocument.presentationml.presentation": ".pptx",
		"application/zip":              ".zip",
		"application/x-rar-compressed": ".rar",
		"application/x-7z-compressed":  ".7z",
		"text/plain":                   ".txt",
		"text/csv":                     ".csv",
		"application/json":             ".json",
		"application/xml":              ".xml",
		"text/html":                    ".html",
	}

	if ext, ok := extensions[mimeType]; ok {
		return ext
	}

	// Try to extract extension from MIME type (e.g., "image/jpeg" -> ".jpeg")
	parts := strings.Split(mimeType, "/")
	if len(parts) == 2 {
		return "." + parts[1]
	}

	return ".bin" // Default fallback
}

// GetMediaMetadata extracts metadata from a media message
func GetMediaMetadata(msg interface{}) (mimeType string, size uint64, filename string) {
	switch m := msg.(type) {
	case *waE2E.ImageMessage:
		mimeType = m.GetMimetype()
		size = m.GetFileLength()
		filename = ""
	case *waE2E.VideoMessage:
		mimeType = m.GetMimetype()
		size = m.GetFileLength()
		filename = ""
	case *waE2E.AudioMessage:
		mimeType = m.GetMimetype()
		size = m.GetFileLength()
		filename = ""
	case *waE2E.DocumentMessage:
		mimeType = m.GetMimetype()
		size = m.GetFileLength()
		if m.FileName != nil {
			filename = *m.FileName
		}
	case *waE2E.StickerMessage:
		mimeType = m.GetMimetype()
		size = m.GetFileLength()
		filename = ""
	}
	return
}
