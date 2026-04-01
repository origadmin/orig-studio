/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package biz

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"origadmin/application/origcms/internal/helpers/ffmpeg"
)

const (
	StatusPending   = "pending"
	StatusUploading = "uploading"
	StatusCompleted = "completed"
	StatusAborted   = "aborted"
)

// UploadSession represents an upload session for multipart uploads.
type UploadSession struct {
	UploadID     string         `json:"upload_id"`
	Filename     string         `json:"filename"`
	FileSize     int64          `json:"file_size"`
	ContentType  string         `json:"content_type"`
	TotalParts   int            `json:"total_parts"`
	ChunkSize    int            `json:"chunk_size"`
	UploadedSize int64          `json:"uploaded_size"`
	Title        string         `json:"title"`
	Description  string         `json:"description"`
	CategoryID   *int64         `json:"category_id"`
	Tags         []string       `json:"tags"`
	UserID       *int64         `json:"user_id"`
	Status       string         `json:"status"`
	Parts        map[int]string `json:"parts"` // part_number -> etag
	Sha256       string         `json:"sha256"`
	StoragePath  string         `json:"storage_path"`
	TempDir      string         `json:"temp_dir"`
	ExpiresAt    time.Time      `json:"expires_at"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// UploadRepo defines the storage operations for upload sessions.
type UploadRepo interface {
	CreateSession(ctx context.Context, session *UploadSession) error
	GetSession(ctx context.Context, uploadID string) (*UploadSession, error)
	UpdateSession(ctx context.Context, session *UploadSession) error
	DeleteSession(ctx context.Context, uploadID string) error
	ListSessions(
		ctx context.Context,
		userID int64,
		status string,
		page, pageSize int,
	) ([]*UploadSession, int, error)
	// DeleteExpiredSessions finds and deletes sessions that have expired.
	// Returns the list of upload IDs deleted.
	DeleteExpiredSessions(ctx context.Context, now time.Time) ([]string, error)
}

// Storage defines the interface for storing and merging file parts.
type Storage interface {
	StorePart(ctx context.Context, uploadID string, partNumber int, data []byte) (string, error)
	MergeParts(ctx context.Context, uploadID string, totalParts int, finalPath string) error
	DeleteParts(ctx context.Context, uploadID string) error
}

type UploadUseCase struct {
	repo      UploadRepo
	mediaRepo MediaRepo
	storage   Storage
	chunkSize int
	log       *log.Helper
	mu        sync.Mutex
}

// NewUploadUseCase .
func NewUploadUseCase(
	repo UploadRepo,
	mediaRepo MediaRepo,
	storage Storage,
	logger log.Logger,
) *UploadUseCase {
	return &UploadUseCase{
		repo:      repo,
		mediaRepo: mediaRepo,
		storage:   storage,
		chunkSize: 2 * 1024 * 1024, // 2MB default
		log:       log.NewHelper(log.With(logger, "module", "upload.biz")),
	}
}

// InitiateMultipartUpload starts a new multipart upload.
func (uc *UploadUseCase) InitiateMultipartUpload(
	ctx context.Context,
	filename string,
	fileSize int64,
	contentType string,
	title, description string,
	categoryID *int64,
	tags []string,
	userID *int64,
) (*UploadSession, error) {
	uploadID := uuid.New().String()
	totalParts := int(math.Ceil(float64(fileSize) / float64(uc.chunkSize)))

	session := &UploadSession{
		UploadID:    uploadID,
		Filename:    filename,
		FileSize:    fileSize,
		ContentType: contentType,
		TotalParts:  totalParts,
		ChunkSize:   uc.chunkSize,
		Title:       title,
		Description: description,
		CategoryID:  categoryID,
		Tags:        tags,
		UserID:      userID,
		Status:      StatusPending,
		Parts:       make(map[int]string),
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}

	if err := uc.repo.CreateSession(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
}

// UploadPart handles a single part upload.
func (uc *UploadUseCase) UploadPart(
	ctx context.Context,
	uploadID string,
	partNumber int,
	data []byte,
) (string, error) {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	session, err := uc.repo.GetSession(ctx, uploadID)
	if err != nil {
		return "", err
	}

	if session.Status == StatusCompleted || session.Status == StatusAborted {
		return "", fmt.Errorf("upload session %s is already %s", uploadID, session.Status)
	}

	etag, err := uc.storage.StorePart(ctx, uploadID, partNumber, data)
	if err != nil {
		return "", err
	}

	session.Parts[partNumber] = etag
	session.UploadedSize += int64(len(data))
	session.Status = StatusUploading

	if err := uc.repo.UpdateSession(ctx, session); err != nil {
		return "", err
	}

	return etag, nil
}

// UpdateUploadMetadata updates the metadata of an ongoing upload session.
func (uc *UploadUseCase) UpdateUploadMetadata(ctx context.Context, uploadID string, title, description string, categoryID *int64, tags []string) error {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	session, err := uc.repo.GetSession(ctx, uploadID)
	if err != nil {
		return err
	}

	if session.Status == StatusCompleted || session.Status == StatusAborted {
		return fmt.Errorf("cannot update metadata for %s upload session", session.Status)
	}

	if title != "" {
		session.Title = title
	}
	session.Description = description
	session.CategoryID = categoryID
	session.Tags = tags

	return uc.repo.UpdateSession(ctx, session)
}

// CompleteMultipartUpload finalizes the upload and merges all parts.
func (uc *UploadUseCase) CompleteMultipartUpload(
	ctx context.Context,
	uploadID string,
	sha256 string,
) (*Media, error) {
	session, err := uc.repo.GetSession(ctx, uploadID)
	if err != nil {
		return nil, err
	}

	if len(session.Parts) < session.TotalParts {
		return nil, fmt.Errorf(
			"not all parts uploaded: %d/%d",
			len(session.Parts),
			session.TotalParts,
		)
	}

	// Define final path (should be configurable)
	ext := filepath.Ext(session.Filename)
	finalPath := fmt.Sprintf("uploads/%s%s", uploadID, ext)

	if err := uc.storage.MergeParts(ctx, uploadID, session.TotalParts, finalPath); err != nil {
		return nil, err
	}

	// Create media record
	media := &Media{
		Title:       session.Title,
		Description: session.Description,
		Url:         finalPath,
		Size:        session.FileSize,
		MimeType:    session.ContentType,
	}
	if session.UserID != nil {
		media.UserId = *session.UserID
	}

	createdMedia, err := uc.mediaRepo.Create(ctx, media)
	if err != nil {
		return nil, err
	}

	session.Status = StatusCompleted
	session.StoragePath = finalPath
	session.Sha256 = sha256
	_ = uc.repo.UpdateSession(ctx, session)

	// Clean up temporary parts
	_ = uc.storage.DeleteParts(ctx, uploadID)

	// Background thumbnail generation
	if strings.HasPrefix(session.ContentType, "video/") {
		go uc.generateThumbnail(context.Background(), createdMedia.Id, finalPath, session.ContentType)
	}

	return createdMedia, nil
}

// generateThumbnail extracts a frame from the video and updates the media record.
func (uc *UploadUseCase) generateThumbnail(ctx context.Context, mediaID int64, mediaPath, contentType string) {
	// Base directory for data (should ideally be configurable)
	baseDir := "./data/uploads"
	fullPath := filepath.Join(baseDir, mediaPath)
	thumbDir := filepath.Join(baseDir, "thumbnails")
	thumbFilename := fmt.Sprintf("%d.jpg", mediaID)
	thumbPath := filepath.Join(thumbDir, thumbFilename)

	if err := os.MkdirAll(thumbDir, 0755); err != nil {
		uc.log.Errorf("failed to create thumbnails directory: %v", err)
		return
	}

	uc.log.Infof("generating thumbnail for media %d at %s", mediaID, thumbPath)

	// Extract thumbnail at 5 seconds (or 0 if video is shorter)
	err := ffmpeg.ExtractThumbnail(ctx, fullPath, thumbPath, "00:00:05")
	if err != nil {
		uc.log.Errorf("failed to generate thumbnail for media %d: %v", mediaID, err)
		// Try at 0 seconds as fallback
		err = ffmpeg.ExtractThumbnail(ctx, fullPath, thumbPath, "00:00:00")
		if err != nil {
			uc.log.Errorf("fallback thumbnail generation also failed for media %d: %v", mediaID, err)
			return
		}
	}

	// Update media record with thumbnail URL
	// The URL should be relative for the static file server
	thumbUrl := fmt.Sprintf("thumbnails/%s", thumbFilename)

	// We need a fresh context because the original might have been cancelled
	updateCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	media, err := uc.mediaRepo.Get(updateCtx, mediaID)
	if err != nil {
		uc.log.Errorf("failed to get media %d for thumbnail update: %v", mediaID, err)
		return
	}

	media.Thumbnail = thumbUrl
	if _, err := uc.mediaRepo.Update(updateCtx, media); err != nil {
		uc.log.Errorf("failed to update media %d with thumbnail: %v", mediaID, err)
	} else {
		uc.log.Infof("successfully generated and updated thumbnail for media %d", mediaID)
	}
}

// AbortMultipartUpload cancels the upload and cleans up.
func (uc *UploadUseCase) AbortMultipartUpload(ctx context.Context, uploadID string) error {
	session, err := uc.repo.GetSession(ctx, uploadID)
	if err != nil {
		return err
	}

	session.Status = StatusAborted
	if err := uc.repo.UpdateSession(ctx, session); err != nil {
		return err
	}

	return uc.storage.DeleteParts(ctx, uploadID)
}

func (uc *UploadUseCase) GetSession(ctx context.Context, uploadID string) (*UploadSession, error) {
	return uc.repo.GetSession(ctx, uploadID)
}

func (uc *UploadUseCase) ListSessions(
	ctx context.Context,
	userID int64,
	status string,
	page, pageSize int,
) ([]*UploadSession, int, error) {
	return uc.repo.ListSessions(ctx, userID, status, page, pageSize)
}

// CleanupExpiredSessions removes sessions and temporary files that have expired.
func (uc *UploadUseCase) CleanupExpiredSessions(ctx context.Context) error {
	uc.log.Info("running cleanup of expired upload sessions")
	ids, err := uc.repo.DeleteExpiredSessions(ctx, time.Now())
	if err != nil {
		return err
	}

	for _, id := range ids {
		uc.log.Infof("cleaning up temporary parts for expired upload: %s", id)
		_ = uc.storage.DeleteParts(ctx, id)
	}

	return nil
}
