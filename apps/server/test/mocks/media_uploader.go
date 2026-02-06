package mocks

import (
	"context"

	"whatspire/internal/domain/entity"
	"whatspire/internal/domain/valueobject"
)

// MediaUploaderMock is a mock implementation of MediaUploader
type MediaUploaderMock struct {
	UploadImageFn    func(ctx context.Context, sessionID string, url string) (*entity.MediaUploadResult, error)
	UploadDocumentFn func(ctx context.Context, sessionID string, url string, filename string) (*entity.MediaUploadResult, error)
	UploadAudioFn    func(ctx context.Context, sessionID string, url string) (*entity.MediaUploadResult, error)
	UploadVideoFn    func(ctx context.Context, sessionID string, url string) (*entity.MediaUploadResult, error)
	UploadFn         func(ctx context.Context, sessionID string, info *entity.MediaDownloadInfo) (*entity.MediaUploadResult, error)
	Constraints      *valueobject.MediaConstraints
}

// NewMediaUploaderMock creates a new MediaUploaderMock
func NewMediaUploaderMock() *MediaUploaderMock {
	return &MediaUploaderMock{
		Constraints: valueobject.DefaultMediaConstraints(),
	}
}

// UploadImage mocks image upload
func (m *MediaUploaderMock) UploadImage(ctx context.Context, sessionID string, url string) (*entity.MediaUploadResult, error) {
	if m.UploadImageFn != nil {
		return m.UploadImageFn(ctx, sessionID, url)
	}
	return &entity.MediaUploadResult{
		URL:        "https://whatsapp.net/media/image123",
		MimeType:   "image/jpeg",
		FileLength: 1024,
	}, nil
}

// UploadDocument mocks document upload
func (m *MediaUploaderMock) UploadDocument(ctx context.Context, sessionID string, url string, filename string) (*entity.MediaUploadResult, error) {
	if m.UploadDocumentFn != nil {
		return m.UploadDocumentFn(ctx, sessionID, url, filename)
	}
	return &entity.MediaUploadResult{
		URL:        "https://whatsapp.net/media/doc123",
		MimeType:   "application/pdf",
		FileLength: 2048,
	}, nil
}

// UploadAudio mocks audio upload
func (m *MediaUploaderMock) UploadAudio(ctx context.Context, sessionID string, url string) (*entity.MediaUploadResult, error) {
	if m.UploadAudioFn != nil {
		return m.UploadAudioFn(ctx, sessionID, url)
	}
	return &entity.MediaUploadResult{
		URL:        "https://whatsapp.net/media/audio123",
		MimeType:   "audio/mpeg",
		FileLength: 4096,
	}, nil
}

// UploadVideo mocks video upload
func (m *MediaUploaderMock) UploadVideo(ctx context.Context, sessionID string, url string) (*entity.MediaUploadResult, error) {
	if m.UploadVideoFn != nil {
		return m.UploadVideoFn(ctx, sessionID, url)
	}
	return &entity.MediaUploadResult{
		URL:        "https://whatsapp.net/media/video123",
		MimeType:   "video/mp4",
		FileLength: 8192,
	}, nil
}

// Upload mocks generic upload
func (m *MediaUploaderMock) Upload(ctx context.Context, sessionID string, info *entity.MediaDownloadInfo) (*entity.MediaUploadResult, error) {
	if m.UploadFn != nil {
		return m.UploadFn(ctx, sessionID, info)
	}
	return &entity.MediaUploadResult{
		URL:        "https://whatsapp.net/media/generic123",
		MimeType:   "application/octet-stream",
		FileLength: 1024,
	}, nil
}

// GetConstraints returns media constraints
func (m *MediaUploaderMock) GetConstraints() *valueobject.MediaConstraints {
	return m.Constraints
}
