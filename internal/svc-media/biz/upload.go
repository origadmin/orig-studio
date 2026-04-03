/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"

	"origadmin/application/origcms/internal/helpers/ffmpeg"
	"origadmin/application/origcms/internal/pubsub"
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
	Thumbnail    string         `json:"thumbnail"`
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
	repo         UploadRepo
	mediaRepo    MediaRepo
	profileRepo  EncodeProfileRepo
	encodingRepo EncodingTaskRepo
	mediaUseCase *MediaUseCase
	storage      Storage
	publisher message.Publisher // Watermill publisher for async encoding
	chunkSize    int
	log          *log.Helper
	mu           sync.Mutex
}

// NewUploadUseCase .
func NewUploadUseCase(
	repo UploadRepo,
	mediaRepo MediaRepo,
	profileRepo EncodeProfileRepo,
	encodingRepo EncodingTaskRepo,
	mediaUseCase *MediaUseCase,
	storage Storage,
	logger log.Logger,
) *UploadUseCase {
	return &UploadUseCase{
		repo:         repo,
		mediaRepo:    mediaRepo,
		profileRepo:  profileRepo,
		encodingRepo: encodingRepo,
		mediaUseCase: mediaUseCase,
		storage:      storage,
		chunkSize:    5 * 1024 * 1024, // 5MB default
		log:          log.NewHelper(log.With(logger, "module", "upload.biz")),
	}
}

// SetPublisher injects a Watermill publisher for async media encoding requests.
// Called after construction to decouple from the constructor signature.
func (uc *UploadUseCase) SetPublisher(publisher message.Publisher) {
	uc.publisher = publisher
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
	thumbnail string,
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
		Thumbnail:   thumbnail,
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
func (uc *UploadUseCase) UpdateUploadMetadata(
	ctx context.Context,
	uploadID string,
	title, description string,
	categoryID *int64,
	tags []string,
	thumbnail string,
) error {
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
	if thumbnail != "" {
		session.Thumbnail = thumbnail
	}

	return uc.repo.UpdateSession(ctx, session)
}

// CompleteMultipartUpload finalizes the upload and merges all parts.
func (uc *UploadUseCase) CompleteMultipartUpload(
	ctx context.Context,
	uploadID string,
	sha256 string,
	title, description string,
	categoryID *int64,
	tags []string,
	thumbnail string,
) (*Media, error) {
	uc.mu.Lock()
	defer uc.mu.Unlock()

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

	// Use metadata from completion request, falling back to session metadata if not provided
	finalTitle := title
	if finalTitle == "" {
		finalTitle = session.Title
	}
	finalDescription := description
	if finalDescription == "" {
		finalDescription = session.Description
	}
	finalCategoryID := categoryID
	if finalCategoryID == nil {
		finalCategoryID = session.CategoryID
	}
	finalTags := tags
	if len(finalTags) == 0 {
		finalTags = session.Tags
	}
	finalThumbnail := thumbnail
	if finalThumbnail == "" {
		finalThumbnail = session.Thumbnail
	}

	// Extract duration if it's a video
	var duration time.Duration
	if strings.Contains(session.ContentType, "video") {
		// Base directory for data (should ideally be configurable)
		baseDir := "./data/uploads"
		fullPath := filepath.Join(baseDir, finalPath)
		if d, err := ffmpeg.GetVideoDuration(ctx, fullPath); err == nil {
			duration = d
		} else {
			uc.log.Errorf("failed to extract duration for %s: %v", fullPath, err)
		}
	}

	// Create media record
	media := &Media{
		Title:          finalTitle,
		Description:    finalDescription,
		Url:            finalPath,
		Size:           session.FileSize,
		MimeType:       session.ContentType,
		Thumbnail:      finalThumbnail,
		Tags:           finalTags,
		Duration:       int32(duration.Seconds()),
		EncodingStatus: "pending",
	}
	if finalCategoryID != nil {
		media.CategoryId = *finalCategoryID
	}
	if session.UserID != nil {
		media.UserId = *session.UserID
	}

	// Derive media type (e.g., video, image)
	media.Type = "file"
	if strings.Contains(session.ContentType, "video") {
		media.Type = "video"
	} else if strings.Contains(session.ContentType, "image") {
		media.Type = "image"
	} else if strings.Contains(session.ContentType, "audio") {
		media.Type = "audio"
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

	// Background media processing (Thumbnail + HLS Transcoding)
	if strings.HasPrefix(session.ContentType, "video/") {
		payload, _ := json.Marshal(MediaEncodeRequest{
			MediaID:     createdMedia.Id,
			MediaPath:   finalPath,
			ContentType: session.ContentType,
		})
		msg := pubsub.NewMessage(payload)
		if err := uc.publisher.Publish(pubsub.MediaEncodeRequestTopic, msg); err != nil {
			uc.log.Errorf("failed to publish encode request for media %d: %v", createdMedia.Id, err)
		}
	}

	return createdMedia, nil
}

// ProcessMedia removed: legacy sync transcoding method, replaced by Watermill-driven TranscodeHandler.
// See transcode_handler.go for the new implementation.

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

// RetryTranscode re-triggers transcoding for a failed media item.
// It validates the media state, cleans up old encoding tasks, resets the status,
// and publishes a new encode request to the transcode pipeline.
// Uses mutex to prevent concurrent retry of the same media.
func (uc *UploadUseCase) RetryTranscode(ctx context.Context, mediaID int64) error {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	media, err := uc.mediaRepo.Get(ctx, mediaID)
	if err != nil {
		return fmt.Errorf("media not found: %w", err)
	}

	// Only allow retry for failed media — do not interrupt in-progress tasks
	if media.EncodingStatus != "failed" {
		return fmt.Errorf(
			"cannot retry media with status %q, only 'failed' allowed",
			media.EncodingStatus,
		)
	}

	// Validate that the source file still exists
	if media.Url == "" {
		return fmt.Errorf("media has no source file URL")
	}

	// Delete old encoding tasks (they'll be recreated by the transcode handler)
	if err := uc.encodingRepo.DeleteByMedia(ctx, mediaID); err != nil {
		uc.log.Warnf("failed to delete old encoding tasks for media %d: %v", mediaID, err)
	}

	// Reset media status to pending
	media.EncodingStatus = "pending"
	if _, err := uc.mediaRepo.Update(ctx, media); err != nil {
		return fmt.Errorf("failed to reset media status: %w", err)
	}

	// Publish new encode request
	payload, _ := json.Marshal(MediaEncodeRequest{
		MediaID:     mediaID,
		MediaPath:   media.Url,
		ContentType: media.MimeType,
	})
	msg := pubsub.NewMessage(payload)
	if err := uc.publisher.Publish(pubsub.MediaEncodeRequestTopic, msg); err != nil {
		return fmt.Errorf("failed to publish encode request: %w", err)
	}

	uc.log.Infof("retry transcoding requested for media %d", mediaID)
	return nil
}
