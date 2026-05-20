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

	"origadmin/application/origstudio/internal/conf"
	"origadmin/application/origstudio/internal/dal/enums"
	"origadmin/application/origstudio/internal/features/media/ffmpeg"
	"origadmin/application/origstudio/internal/infra/pubsub"
	"origadmin/application/origstudio/internal/features/media/dto"
	systembiz "origadmin/application/origstudio/internal/features/system/biz"
)

// Upload status constants
const (
	StatusPending   = enums.UploadStatusPending
	StatusUploading = enums.UploadStatusUploading
	StatusCompleted = enums.UploadStatusCompleted
	StatusAborted   = enums.UploadStatusAborted
)

// UploadSession represents an upload session for multipart uploads.
type UploadSession = dto.UploadSession

// UploadRepo defines the storage operations for upload sessions.
type UploadRepo = dto.UploadRepo

type UploadUseCase struct {
	repo           UploadRepo
	mediaRepo      MediaRepo
	profileRepo    dto.EncodeProfileRepo
	encodingRepo   dto.EncodingTaskRepo
	mediaUseCase   *MediaUseCase
	storage        Storage
	publisher      message.Publisher
	paths          *conf.StoragePaths
	chunkSize      int
	log            *log.Helper
	mu             sync.Mutex
	configProvider systembiz.ConfigProvider
}

func NewUploadUseCase(
	repo UploadRepo,
	mediaRepo MediaRepo,
	profileRepo dto.EncodeProfileRepo,
	encodingRepo dto.EncodingTaskRepo,
	mediaUseCase *MediaUseCase,
	storage Storage,
	paths *conf.StoragePaths,
	chunkSize int,
	logger log.Logger,
	configProvider systembiz.ConfigProvider,
) *UploadUseCase {
	return &UploadUseCase{
		repo:           repo,
		mediaRepo:      mediaRepo,
		profileRepo:    profileRepo,
		encodingRepo:   encodingRepo,
		mediaUseCase:   mediaUseCase,
		storage:        storage,
		paths:          paths,
		chunkSize:      chunkSize,
		log:            log.NewHelper(log.With(logger, "module", "upload.biz")),
		configProvider: configProvider,
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
	userID *string,
) (*UploadSession, error) {
	if !uc.isUploadAllowed(ctx) {
		return nil, fmt.Errorf("upload is disabled")
	}

	maxSize := uc.getMaxUploadSize(ctx, contentType)
	if maxSize > 0 && fileSize > maxSize {
		return nil, fmt.Errorf("file size %d exceeds maximum allowed size %d", fileSize, maxSize)
	}

	if !uc.isFormatAllowed(ctx, contentType) {
		return nil, fmt.Errorf("file format %s is not allowed", contentType)
	}

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

	// Determine userID for path generation; fallback to "_system" if unknown
	userID := "_system"
	if session.UserID != nil && *session.UserID != "" {
		userID = *session.UserID
	}

	// Inject userID into context so LocalStorage can use it for path generation
	ctx = ContextWithUserID(ctx, userID)

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

	// Define final path using StoragePaths with user isolation and date sharding
	ext := filepath.Ext(session.Filename)
	filename := fmt.Sprintf("%s%s", uploadID, ext)

	// Determine userID for path generation; fallback to "_system" if unknown
	userID := "_system"
	if session.UserID != nil && *session.UserID != "" {
		userID = *session.UserID
	}

	// Generate relative path for DB storage: temp/{userID}/{yyyy}/{MM}/{filename}
	// File is first merged to temp, then promoted to originals after transcoding
	finalPath := uc.paths.RelativeTemp(userID, filename)

	// Inject userID into context so LocalStorage can use it for path generation
	ctx = ContextWithUserID(ctx, userID)

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

	// Extract full media info via ffprobe for video/audio files
	fullPath := uc.paths.FullPath(finalPath)
	var mediaInfo *ffmpeg.MediaInfo
	var duration time.Duration
	if strings.Contains(session.ContentType, "video") || strings.Contains(session.ContentType, "audio") {
		if info, err := ffmpeg.GetMediaInfo(ctx, fullPath); err != nil {
			uc.log.Errorf("failed to extract media info for %s: %v", fullPath, err)
		} else {
			mediaInfo = info
			if info.Duration > 0 {
				duration = time.Duration(info.Duration * float64(time.Second))
			}
			uc.log.Infof("media info: codec=%s/%s bitrate=%d/%d fps=%.2f res=%dx%d sample=%d channels=%d",
				info.VideoCodec, info.AudioCodec, info.VideoBitRate, info.AudioBitRate,
				info.FPS, info.Width, info.Height, info.SampleRate, info.Channels)
		}
	}

	// Check max duration for video files
	if strings.Contains(session.ContentType, "video") {
		if maxDuration := uc.getMaxVideoDuration(ctx); maxDuration > 0 && duration > maxDuration {
			_ = uc.storage.Delete(ctx, finalPath)
			return nil, fmt.Errorf("video duration %s exceeds maximum allowed duration %s", duration, maxDuration)
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
		Extension:      filepath.Ext(session.Filename),
		EncodingStatus: string(enums.MediaEncodingStatusPending),
	}
	if finalCategoryID != nil {
		media.CategoryId = *finalCategoryID
	}
	if session.UserID != nil {
		media.UserId = *session.UserID
	}
	if mediaInfo != nil {
		media.Width = int32(mediaInfo.Width)
		media.Height = int32(mediaInfo.Height)
	}

	media.Type = "file"
	if strings.Contains(session.ContentType, "video") {
		media.Type = "video"
	} else if strings.Contains(session.ContentType, "image") {
		media.Type = "image"
	} else if strings.Contains(session.ContentType, "audio") {
		media.Type = "audio"
	}

	entityMedia, createdMedia, err := uc.mediaRepo.CreateWithEntity(ctx, media)
	if err != nil {
		return nil, err
	}

	// Promote temp file to originals immediately after media record creation
	promoteFilename := filepath.Base(finalPath)
	promotedPath, err := uc.paths.PromoteToOriginal(userID, promoteFilename)
	if err != nil {
		_ = uc.storage.Delete(ctx, finalPath)
		_ = uc.mediaRepo.Delete(ctx, entityMedia.ID)
		return nil, fmt.Errorf("promote to originals: %w", err)
	}
	createdMedia.Url = promotedPath
	media.Url = promotedPath
	if _, err := uc.mediaRepo.Update(ctx, media); err != nil {
		uc.log.Errorf("failed to update media URL after promotion: %v", err)
	}

	session.Status = StatusCompleted
	session.StoragePath = promotedPath
	session.Sha256 = sha256
	_ = uc.repo.UpdateSession(ctx, session)

	// Clean up temporary parts
	_ = uc.storage.DeleteParts(ctx, uploadID)

	// Background media processing (Thumbnail + HLS Transcoding)
	if strings.HasPrefix(session.ContentType, "video/") {
		payload, _ := json.Marshal(MediaEncodeRequest{
			MediaID:     entityMedia.ID,
			MediaPath:   media.Url,
			ContentType: session.ContentType,
		})
		msg := pubsub.NewMessage(payload)
		if uc.publisher != nil {
			if err := uc.publisher.Publish(pubsub.MediaEncodeRequestTopic, msg); err != nil {
				uc.log.Errorf("failed to publish encode request for media %s: %v", entityMedia.ID, err)
			}
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
	userID string,
	status enums.UploadStatus,
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

// CleanupExpiredTemp removes temp directories for failed/expired transcodes.
// It queries media records whose URL starts with "temp/" and whose create_time
// is older than the configured TTL, then deletes the temp files and marks the
// media encoding status as failed. Called periodically by a cron job.
func (uc *UploadUseCase) CleanupExpiredTemp(ctx context.Context, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = 48 * time.Hour
	}
	cutoff := time.Now().Add(-ttl)

	uc.log.Infof("running cleanup of expired temp files (TTL=%s, cutoff=%s)", ttl, cutoff.Format(time.RFC3339))

	medias, err := uc.mediaRepo.ListTempMediaBefore(ctx, cutoff)
	if err != nil {
		return fmt.Errorf("list temp media: %w", err)
	}

	cleaned := 0
	for _, m := range medias {
		if !strings.HasPrefix(m.Url, "temp/") {
			continue
		}
		if err := uc.storage.Delete(ctx, m.Url); err != nil {
			uc.log.Warnf("failed to delete temp file %s for media %s: %v", m.Url, m.Id, err)
			continue
		}

		// Mark media as failed since it was never successfully transcoded
		m.EncodingStatus = string(enums.MediaEncodingStatusFailed)
		if _, err := uc.mediaRepo.Update(ctx, m); err != nil {
			uc.log.Warnf("failed to update media %s status after temp cleanup: %v", m.Id, err)
		}
		cleaned++
		uc.log.Infof("cleaned up expired temp file for media %s: %s", m.Id, m.Url)
	}

	uc.log.Infof("temp cleanup completed: %d files cleaned", cleaned)
	return nil
}

// RetryTranscode re-triggers transcoding for a failed media item.
// It validates the media state, cleans up old encoding tasks, resets the status,
// and publishes a new encode request to the transcode pipeline.
// Uses mutex to prevent concurrent retry of the same media.
func (uc *UploadUseCase) RetryTranscode(ctx context.Context, mediaID string) error {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	media, err := uc.mediaRepo.Get(ctx, mediaID)
	if err != nil {
		return fmt.Errorf("media not found: %w", err)
	}

	// Only allow retry for failed media — do not interrupt in-progress tasks
	if media.EncodingStatus != string(enums.MediaEncodingStatusFailed) {
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
		uc.log.Warnf("failed to delete old encoding tasks for media %s: %v", mediaID, err)
	}

	// Reset media status to pending
	media.EncodingStatus = string(enums.MediaEncodingStatusPending)
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

	uc.log.Infof("retry transcoding requested for media %s", mediaID)
	return nil
}

func (uc *UploadUseCase) isUploadAllowed(ctx context.Context) bool {
	if uc.configProvider == nil {
		return true
	}
	val := uc.configProvider.Get(ctx, "allow_upload")
	return val != "false" && val != "0"
}

func (uc *UploadUseCase) getMaxUploadSize(ctx context.Context, contentType string) int64 {
	if uc.configProvider == nil {
		return 0
	}
	switch {
	case strings.HasPrefix(contentType, "video/"):
		if val := uc.configProvider.Get(ctx, "max_upload_size_video"); val != "" {
			if size, err := parseSize(val); err == nil {
				return size
			}
		}
	case strings.HasPrefix(contentType, "image/"):
		if val := uc.configProvider.Get(ctx, "max_upload_size_image"); val != "" {
			if size, err := parseSize(val); err == nil {
				return size
			}
		}
	}
	return 0
}

func (uc *UploadUseCase) isFormatAllowed(ctx context.Context, contentType string) bool {
	if uc.configProvider == nil {
		return true
	}
	var allowedFormats string
	switch {
	case strings.HasPrefix(contentType, "video/"):
		allowedFormats = uc.configProvider.Get(ctx, "allowed_video_formats")
	case strings.HasPrefix(contentType, "image/"):
		allowedFormats = uc.configProvider.Get(ctx, "allowed_image_formats")
	default:
		return true
	}
	if allowedFormats == "" {
		return true
	}
	ext := strings.TrimPrefix(contentType, "video/")
	ext = strings.TrimPrefix(ext, "image/")
	allowedList := strings.Split(strings.ToLower(allowedFormats), ",")
	for _, f := range allowedList {
		if strings.TrimSpace(f) == ext {
			return true
		}
	}
	return false
}

func (uc *UploadUseCase) getMaxVideoDuration(ctx context.Context) time.Duration {
	if uc.configProvider == nil {
		return 0
	}
	val := uc.configProvider.Get(ctx, "max_video_duration")
	if val == "" {
		return 0
	}
	seconds, err := parseSizeInt(val)
	if err != nil {
		return 0
	}
	return time.Duration(seconds) * time.Second
}

func parseSize(s string) (int64, error) {
	var size int64
	_, err := fmt.Sscanf(s, "%d", &size)
	return size, err
}

func parseSizeInt(s string) (int, error) {
	var size int
	_, err := fmt.Sscanf(s, "%d", &size)
	return size, err
}
