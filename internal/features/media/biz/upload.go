/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package biz

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"

	"origadmin/application/origstudio/api/gen/v1/types"
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

	maxPartSize = 100 * 1024 * 1024
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
//
// A类修复: tempPath 和 finalPath 在创建 session 时一次性计算并存入 session。
// CompleteMultipartUpload 使用 session 中记录的固定路径，不重新调用 time.Now()。
func (uc *UploadUseCase) InitiateMultipartUpload(
	ctx context.Context,
	filename string,
	fileSize int64,
	contentType string,
	title, description string,
	categoryID *int64,
	channelID *string,
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
	ext := filepath.Ext(filename)

	uid := "_system"
	if userID != nil && *userID != "" {
		uid = *userID
	}

	// Use a fixed reference time for this session so all paths are consistent
	// across Initiate → UploadPart → Complete, regardless of calendar boundaries.
	now := time.Now()
	tempFilename := fmt.Sprintf("%s%s", uploadID, ext)
	tempPath := uc.paths.RelativeTempAt(uid, tempFilename, now)
	finalPath := uc.paths.RelativeOriginalAt(uid, tempFilename, now)
	tempDir := uc.paths.TempUploadDirAt(uid, uploadID, now)

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
		ChannelID:   channelID,
		Tags:        tags,
		Thumbnail:   thumbnail,
		UserID:      userID,
		Status:      StatusPending,
		Parts:       make(map[int]string),
		TempPath:    tempPath,
		FinalPath:   finalPath,
		TempDir:     tempDir,
		ExpiresAt:   now.Add(24 * time.Hour),
		CreateTime:  now,
		UpdateTime:  now,
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
	r io.Reader,
	size int64,
) (string, error) {
	if size > maxPartSize {
		return "", fmt.Errorf("part size %d exceeds maximum allowed size %d", size, maxPartSize)
	}

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

	// Inject userID and session creation time into context for time-safe paths
	ctx = ContextWithUserID(ctx, userID)
	ctx = ContextWithSessionCreateTime(ctx, session.CreateTime)

	etag, err := uc.storage.StorePart(ctx, uploadID, partNumber, r, size)
	if err != nil {
		return "", err
	}

	session.Parts[partNumber] = etag
	session.UploadedSize += size
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
	channelID *string,
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
	session.ChannelID = channelID
	session.Tags = tags
	if thumbnail != "" {
		session.Thumbnail = thumbnail
	}

	return uc.repo.UpdateSession(ctx, session)
}

// sniffRealMediaType 以文件真实字节为权威源判定媒体类型，纠正客户端误标的 Content-Type（BUG-067）。
// 返回：规范化后的 MIME、媒体类型(file/image/video/audio)、规范扩展名、是否需要转码。
func sniffRealMediaType(fullPath, declared, filename string) (mime, mediaType, ext string, needsEncoding bool) {
	fallback := func() (string, string, string, bool) {
		mt := declared
		if mt == "" {
			mt = "application/octet-stream"
		}
		t := "file"
		switch {
		case strings.Contains(mt, "video"):
			t = "video"
		case strings.Contains(mt, "image"):
			t = "image"
		case strings.Contains(mt, "audio"):
			t = "audio"
		}
		return mt, t, filepath.Ext(filename), strings.HasPrefix(mt, "video/")
	}

	f, err := os.Open(fullPath)
	if err != nil {
		return fallback()
	}
	defer f.Close()
	buf := make([]byte, 512)
	n, _ := f.Read(buf)
	detected := http.DetectContentType(buf[:n])
	switch {
	case strings.HasPrefix(detected, "image/"):
		return detected, "image", canonicalExt(detected), false
	case strings.HasPrefix(detected, "video/"):
		return detected, "video", canonicalExt(detected), true
	case strings.HasPrefix(detected, "audio/"):
		return detected, "audio", canonicalExt(detected), false
	case detected == "application/octet-stream":
		// 嗅探失败（如 mkv/ts 等容器），用 ffprobe 兜底判定是否为真视频/音频
		if info, err := ffmpeg.GetMediaInfo(context.Background(), fullPath); err == nil {
			if info.VideoCodec != "" {
				return "video/mp4", "video", ".mp4", true
			}
			if info.AudioCodec != "" {
				return "audio/mpeg", "audio", ".mp3", false
			}
		}
		return fallback()
	default:
		return fallback()
	}
}

// canonicalExt 由 MIME 返回规范扩展名（含点）。未知类型返回空串。
func canonicalExt(mime string) string {
	switch {
	case mime == "image/jpeg":
		return ".jpg"
	case mime == "image/png":
		return ".png"
	case mime == "image/gif":
		return ".gif"
	case mime == "image/webp":
		return ".webp"
	case strings.HasPrefix(mime, "video/"):
		return ".mp4"
	case strings.HasPrefix(mime, "audio/"):
		return ".mp3"
	default:
		return ""
	}
}

// CompleteMultipartUpload finalizes the upload and merges all parts.
//
// A类修复:
//   1. 前置状态机校验：pending/aborted/completed 状态拒绝操作
//   2. 使用 session.TempPath（Initiate 时固定的路径），不重新调用 time.Now()
//   3. sha256 为可选，不强制计算；后端统一计算SHA256用于完整性校验
func (uc *UploadUseCase) CompleteMultipartUpload(
	ctx context.Context,
	uploadID string,
	expectedSha256 string,
	title, description string,
	categoryID *int64,
	channelID *string,
	tags []string,
	thumbnail string,
) (*Media, error) {
	uc.mu.Lock()
	session, err := uc.repo.GetSession(ctx, uploadID)
	if err != nil {
		uc.mu.Unlock()
		return nil, err
	}

	// A类修复: 状态机前置校验
	if session.Status == StatusAborted {
		uc.mu.Unlock()
		return nil, fmt.Errorf("upload session %s is already aborted", uploadID)
	}
	if session.Status == StatusCompleted {
		uc.mu.Unlock()
		return nil, fmt.Errorf("upload session %s is already completed", uploadID)
	}
	if len(session.Parts) == 0 {
		uc.mu.Unlock()
		return nil, fmt.Errorf("upload not started: no parts uploaded yet")
	}

	if len(session.Parts) < session.TotalParts {
		uc.mu.Unlock()
		return nil, fmt.Errorf(
			"not all parts uploaded: %d/%d",
			len(session.Parts),
			session.TotalParts,
		)
	}

	userID := "_system"
	if session.UserID != nil && *session.UserID != "" {
		userID = *session.UserID
	}

	// A类修复: 使用 session 中记录的固定 TempPath，不重新计算时间
	tempPath := session.TempPath
	sessionCreateTime := session.CreateTime
	if tempPath == "" {
		ext := filepath.Ext(session.Filename)
		filename := fmt.Sprintf("%s%s", session.UploadID, ext)
		if sessionCreateTime.IsZero() {
			sessionCreateTime = time.Now()
		}
		tempPath = uc.paths.RelativeTempAt(userID, filename, sessionCreateTime)
	}
	ctx = ContextWithUserID(ctx, userID)
	ctx = ContextWithSessionCreateTime(ctx, sessionCreateTime)

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
	finalChannelID := channelID
	if finalChannelID == nil {
		finalChannelID = session.ChannelID
	}
	finalTags := tags
	if len(finalTags) == 0 {
		finalTags = session.Tags
	}
	finalThumbnail := thumbnail
	if finalThumbnail == "" {
		finalThumbnail = session.Thumbnail
	}

	sessionTotalParts := session.TotalParts
	sessionFileSize := session.FileSize
	sessionContentType := session.ContentType
	sessionUserID := session.UserID

	uc.mu.Unlock()

	if err := uc.storage.MergeParts(ctx, uploadID, sessionTotalParts, tempPath); err != nil {
		return nil, err
	}

	fullPath := uc.paths.FullPath(tempPath)
	var mergeWorkDir *LocalWorkDir
	if _, err := os.Stat(fullPath); err != nil {
		mergeWorkDir, err = DownloadToLocalWorkDir(ctx, uc.storage, tempPath, uc.paths.FullPath)
		if err != nil {
			uc.log.Warnf("failed to download merged file for media info: %v", err)
		} else if mergeWorkDir != nil {
			fullPath = mergeWorkDir.LocalPath
		}
	}
	defer mergeWorkDir.Cleanup()

	var mediaInfo *ffmpeg.MediaInfo
	var duration time.Duration
	if strings.Contains(sessionContentType, "video") || strings.Contains(sessionContentType, "audio") {
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

	var imgWidth, imgHeight int
	if strings.Contains(sessionContentType, "image") {
		if f, err := os.Open(fullPath); err == nil {
			cfg, _, decErr := image.DecodeConfig(f)
			f.Close()
			if decErr == nil && cfg.Width > 0 && cfg.Height > 0 {
				imgWidth = cfg.Width
				imgHeight = cfg.Height
				uc.log.Infof("image info: %s %dx%d", sessionContentType, imgWidth, imgHeight)
			} else if decErr != nil {
				uc.log.Warnf("failed to decode image config for %s: %v", fullPath, decErr)
			}
		} else {
			uc.log.Warnf("failed to open merged file for image info %s: %v", fullPath, err)
		}
	}

	var fileSha256 string
	if f, err := os.Open(fullPath); err == nil {
		hash := sha256.New()
		if _, copyErr := io.Copy(hash, f); copyErr == nil {
			fileSha256 = hex.EncodeToString(hash.Sum(nil))
			uc.log.Infof("file sha256: %s (size=%d)", fileSha256, sessionFileSize)
		} else {
			uc.log.Warnf("failed to compute sha256 for %s: %v", fullPath, copyErr)
		}
		f.Close()
	} else {
		uc.log.Warnf("failed to open merged file for sha256 %s: %v", fullPath, err)
	}

	if expectedSha256 != "" && fileSha256 != "" && !strings.EqualFold(expectedSha256, fileSha256) {
		_ = uc.storage.Delete(ctx, tempPath)
		return nil, fmt.Errorf("sha256 mismatch: expected %s, got %s", expectedSha256, fileSha256)
	}

	if strings.Contains(sessionContentType, "video") {
		if maxDuration := uc.getMaxVideoDuration(ctx); maxDuration > 0 && duration > maxDuration {
			_ = uc.storage.Delete(ctx, tempPath)
			return nil, fmt.Errorf("video duration %s exceeds maximum allowed duration %s", duration, maxDuration)
		}
	}

	promotedPath, err := uc.storage.PromoteToOriginal(ctx, tempPath)
	if err != nil {
		_ = uc.storage.Delete(ctx, tempPath)
		return nil, fmt.Errorf("promote to originals: %w", err)
	}

	// BUG-067: 以真实文件字节为权威源纠正客户端误标的 Content-Type（如图片声明为 video/*）
	// 在加锁前完成文件嗅探，避免持锁 I/O。
	sessForName, _ := uc.repo.GetSession(ctx, uploadID)
	sessFilename := ""
	if sessForName != nil {
		sessFilename = sessForName.Filename
	}
	realMime, realType, realExt, realNeedsEncoding := sniffRealMediaType(fullPath, sessionContentType, sessFilename)

	uc.mu.Lock()

	session, err = uc.repo.GetSession(ctx, uploadID)
	if err != nil {
		uc.mu.Unlock()
		_ = uc.storage.Delete(ctx, promotedPath)
		return nil, err
	}
	if session.Status == StatusCompleted || session.Status == StatusAborted {
		uc.mu.Unlock()
		_ = uc.storage.Delete(ctx, promotedPath)
		return nil, fmt.Errorf("upload session %s is already %s", uploadID, session.Status)
	}

	uidStr := ""
	if sessionUserID != nil {
		uidStr = *sessionUserID
	}
	media := &Media{
		Title:          finalTitle,
		Description:    finalDescription,
		Url:            promotedPath,
		Size:           sessionFileSize,
		MimeType:       realMime,
		Thumbnail:      finalThumbnail,
		Tags:           finalTags,
		Duration:       int32(duration.Seconds()),
		Extension:      realExt,
		Sha256:         fileSha256,
		EncodingStatus: string(enums.MediaEncodingStatusPending),
		State:          "draft",
		Privacy:        types.Privacy_PRIVACY_PUBLIC,
		// BUG-139: platform feature modes (comments_mode / downloads_mode) drive
		// per-media defaults — opt_out → on, opt_in/disabled → off. Falls back to
		// comments on / download off when unconfigured (keeps existing behavior).
		AllowDownload:  uc.defaultFeatureBool(ctx, "downloads_mode", false),
		EnableComments: uc.defaultFeatureBool(ctx, "comments_mode", true),
		Listable:       false,
		Featured:       false,
	}
	if finalCategoryID != nil {
		media.CategoryId = *finalCategoryID
	}
	if finalChannelID != nil && *finalChannelID != "" {
		media.ChannelId = *finalChannelID
	}
	// BUG-091: fallback to user's default channel when no explicit channel provided
	if media.ChannelId == "" && uc.mediaUseCase != nil && userID != "" && userID != "_system" {
		if defChID, defErr := uc.mediaUseCase.GetDefaultChannelID(ctx, userID); defErr == nil && defChID != "" {
			media.ChannelId = defChID
		} else if defErr != nil {
			uc.log.Warnf("GetDefaultChannelID fallback skipped for user %s: %v", userID, defErr)
		}
	}
	if sessionUserID != nil {
		media.UserId = *sessionUserID
		media.CreateAuthor = uidStr
		media.UpdateAuthor = uidStr
	}
	if mediaInfo != nil {
		media.Width = int32(mediaInfo.Width)
		media.Height = int32(mediaInfo.Height)
	} else if imgWidth > 0 && imgHeight > 0 {
		media.Width = int32(imgWidth)
		media.Height = int32(imgHeight)
	}

	media.Type = realType

	// Non-video media (image, audio) does not need transcoding.
	// Set encoding_status=success and state=active immediately so the media
	// is visible and properly listed. Only video enters the transcode pipeline.
	if !realNeedsEncoding {
		media.EncodingStatus = string(enums.MediaEncodingStatusSuccess)
		media.State = "active"
		media.Listable = true
	}

	entityMedia, createdMedia, err := uc.mediaRepo.CreateWithEntity(ctx, media)
	if err != nil {
		uc.mu.Unlock()
		_ = uc.storage.Delete(ctx, promotedPath)
		return nil, err
	}

	session.Status = StatusCompleted
	session.StoragePath = promotedPath
	session.Sha256 = fileSha256
	_ = uc.repo.UpdateSession(ctx, session)

	_ = uc.storage.DeleteParts(ctx, uploadID)

	needsEncoding := realNeedsEncoding
	entityMediaIDCopy := entityMedia.ID
	userIDCopy := userID
	mediaForEncode := createdMedia

	uc.mu.Unlock()

	if needsEncoding {
		go func() {
			uc.preprocessAndEncode(context.Background(), mediaForEncode, entityMediaIDCopy, userIDCopy)
		}()
	}

	return createdMedia, nil
}

// ProcessMedia removed: legacy sync transcoding method, replaced by Watermill-driven TranscodeHandler.
// See transcode_handler.go for the new implementation.

// AbortMultipartUpload cancels the upload and cleans up.
//
// A类修复: 已 completed/aborted 的 session 不可重复操作
func (uc *UploadUseCase) AbortMultipartUpload(ctx context.Context, uploadID string) error {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	session, err := uc.repo.GetSession(ctx, uploadID)
	if err != nil {
		return err
	}

	if session.Status == StatusCompleted {
		return fmt.Errorf("upload session %s is already completed, cannot abort", uploadID)
	}
	if session.Status == StatusAborted {
		return fmt.Errorf("upload session %s is already aborted", uploadID)
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
	// The allow-list stores file EXTENSIONS (e.g. "jpg,png,gif,webp"), but
	// the incoming value is a MIME content-type (e.g. "image/jpeg"). Extract
	// the subtype and normalize known aliases so the comparison is consistent.
	// Without this, image/jpeg (subtype "jpeg") was rejected because the list
	// only contained "jpg" — causing every JPEG upload to fail with HTTP 500.
	parts := strings.SplitN(contentType, "/", 2)
	subtype := ""
	if len(parts) == 2 {
		subtype = strings.ToLower(parts[1])
	}
	alias := map[string]string{
		"jpeg": "jpg",
		"jpe":  "jpg",
		"jfif": "jpg",
		"tif":  "tiff",
		"mpeg": "mpg",
	}
	normalized := subtype
	if a, ok := alias[subtype]; ok {
		normalized = a
	}
	allowedList := strings.Split(strings.ToLower(allowedFormats), ",")
	for _, f := range allowedList {
		f = strings.TrimSpace(f)
		if f == subtype || f == normalized {
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

func (uc *UploadUseCase) preprocessAndEncode(ctx context.Context, media *Media, entityMediaID, userID string) {
	fullPath := uc.paths.FullPath(media.Url)
	originalUrl := media.Url

	var workDir *LocalWorkDir
	if _, err := os.Stat(fullPath); err != nil {
		workDir, err = DownloadToLocalWorkDir(ctx, uc.storage, media.Url, uc.paths.FullPath)
		if err != nil {
			uc.log.Errorf("preprocess: download source for media %s: %v", entityMediaID, err)
			uc.handlePreprocessFailure(ctx, media, entityMediaID, "download source failed: "+err.Error())
			return
		}
		if workDir != nil {
			fullPath = workDir.LocalPath
		}
	}
	defer func() {
		if workDir != nil {
			workDir.Cleanup()
		}
	}()

	info, err := ffmpeg.GetMediaInfo(ctx, fullPath)
	if err != nil {
		uc.log.Errorf("preprocess: ffprobe failed for media %s: %v", entityMediaID, err)
		uc.handlePreprocessFailure(ctx, media, entityMediaID, "ffprobe failed: "+err.Error())
		return
	}

	check := ffmpeg.CheckWebFriendly(info)
	uc.log.Infof("preprocess: media=%s codec=%s/%s format=%s webFriendly=%v needsRemux=%v needsTranscode=%v",
		entityMediaID, info.VideoCodec, info.AudioCodec, info.FormatName,
		check.IsWebFriendly, check.NeedsRemux, check.NeedsTranscode)

	// Generate thumbnail immediately if not set (don't wait for transcode service)
	if media.Thumbnail == "" && strings.HasPrefix(media.MimeType, "video/") {
		thumbFilename := fmt.Sprintf("%s.jpg", entityMediaID)
		thumbDir := uc.paths.ThumbnailsDir()
		thumbGeneratedPath, err := GenerateThumbnail(ctx, fullPath, thumbDir, thumbFilename)
		if err == nil {
			media.Thumbnail = uc.paths.RelativeThumbnail(entityMediaID)

			// Verify thumbnail file exists and log its size
			if fi, statErr := os.Stat(thumbGeneratedPath); statErr == nil {
				uc.log.Infof("preprocess: thumbnail file created for media %s: %s (%d bytes)", entityMediaID, thumbGeneratedPath, fi.Size())
			} else {
				uc.log.Warnf("preprocess: thumbnail file missing after generation for media %s: %v", entityMediaID, statErr)
			}

			// Upload thumbnail to storage for non-local backends (S3/Hybrid).
			// Read into memory first to avoid source/destination path conflict
			// where LocalStorage.Upload's os.Create truncates the file before
			// io.Copy finishes reading from the same path.
			thumbRelPath := media.Thumbnail
			if data, readErr := os.ReadFile(thumbGeneratedPath); readErr == nil {
				if _, upErr := uc.storage.Upload(ctx, thumbRelPath, bytes.NewReader(data), int64(len(data)), "image/jpeg"); upErr != nil {
					uc.log.Warnf("preprocess: failed to upload thumbnail to storage for media %s: %v", entityMediaID, upErr)
				} else {
					uc.log.Infof("preprocess: thumbnail uploaded to storage for media %s: %s (%d bytes)", entityMediaID, thumbRelPath, len(data))
				}
			} else {
				uc.log.Warnf("preprocess: failed to read thumbnail file for media %s: %v (path=%s)", entityMediaID, readErr, thumbGeneratedPath)
			}

			if _, err := uc.mediaRepo.Update(ctx, media); err != nil {
				uc.log.Warnf("preprocess: failed to save thumbnail for media %s: %v", entityMediaID, err)
			}
		} else {
			uc.log.Warnf("preprocess: thumbnail generation failed for media %s: %v", entityMediaID, err)
		}
	}

	preprocessed := false

	switch {
	case check.IsWebFriendly:
		uc.log.Infof("preprocess: media %s is web-friendly, no conversion needed", entityMediaID)

	case check.NeedsRemux:
		outputFilename := fmt.Sprintf("%s.mp4", strings.TrimSuffix(filepath.Base(media.Url), filepath.Ext(media.Url)))
		outputRelPath := uc.paths.RelativeOriginal(userID, outputFilename)
		outputFullPath := uc.paths.FullPath(outputRelPath)

		if err := ffmpeg.RemuxToMP4(ctx, fullPath, outputFullPath); err != nil {
			uc.log.Errorf("preprocess: remux failed for media %s: %v", entityMediaID, err)
			uc.handlePreprocessFailure(ctx, media, entityMediaID, "remux failed: "+err.Error())
			return
		}
		uc.log.Infof("preprocess: remux succeeded for media %s: %s -> %s", entityMediaID, media.Url, outputRelPath)
		media.Url = outputRelPath
		media.Extension = ".mp4"
		media.MimeType = "video/mp4"
		preprocessed = true

		// Upload remuxed file to storage for non-local backends (S3/Hybrid).
		// Read into memory first to avoid source/destination path conflict
		// where LocalStorage.Upload's os.Create truncates the file before io.Copy.
		if data, fErr := os.ReadFile(outputFullPath); fErr == nil {
			if _, err := uc.storage.Upload(ctx, outputRelPath, bytes.NewReader(data), int64(len(data)), "video/mp4"); err != nil {
				uc.log.Warnf("preprocess: failed to upload remuxed file to storage: %v", err)
			}
		} else {
			uc.log.Warnf("preprocess: failed to read remuxed file for upload: %v", fErr)
		}

	case check.NeedsTranscode:
		outputFilename := fmt.Sprintf("%s.mp4", strings.TrimSuffix(filepath.Base(media.Url), filepath.Ext(media.Url)))
		outputRelPath := uc.paths.RelativeOriginal(userID, outputFilename)
		outputFullPath := uc.paths.FullPath(outputRelPath)

		if err := ffmpeg.QuickTranscodeToMP4(ctx, fullPath, outputFullPath); err != nil {
			uc.log.Errorf("preprocess: quick transcode failed for media %s: %v", entityMediaID, err)
			uc.handlePreprocessFailure(ctx, media, entityMediaID, "quick transcode failed: "+err.Error())
			return
		}
		uc.log.Infof("preprocess: quick transcode succeeded for media %s: %s -> %s", entityMediaID, media.Url, outputRelPath)
		media.Url = outputRelPath
		media.Extension = ".mp4"
		media.MimeType = "video/mp4"
		preprocessed = true

		// Upload transcoded file to storage for non-local backends (S3/Hybrid).
		// Read into memory first to avoid source/destination path conflict.
		if data, fErr := os.ReadFile(outputFullPath); fErr == nil {
			if _, err := uc.storage.Upload(ctx, outputRelPath, bytes.NewReader(data), int64(len(data)), "video/mp4"); err != nil {
				uc.log.Warnf("preprocess: failed to upload transcoded file to storage: %v", err)
			}
		} else {
			uc.log.Warnf("preprocess: failed to read transcoded file for upload: %v", fErr)
		}
	}

	if preprocessed {
		if _, err := uc.mediaRepo.Update(ctx, media); err != nil {
			uc.log.Errorf("preprocess: failed to update media URL: %v", err)
		}

		keepOriginal := uc.configProvider != nil && uc.configProvider.GetBool(ctx, "keep_original_files")
		if !keepOriginal && originalUrl != media.Url {
			shouldDelete := false
			if check.NeedsRemux {
				shouldDelete = !uc.configProvider.GetBool(ctx, "keep_original_after_remux")
			} else if check.NeedsTranscode {
				shouldDelete = !uc.configProvider.GetBool(ctx, "keep_original_after_transcode")
			}
			if shouldDelete {
				if err := uc.storage.Delete(ctx, originalUrl); err != nil {
					uc.log.Warnf("preprocess: failed to delete original file %s: %v", originalUrl, err)
				} else {
					uc.log.Infof("preprocess: deleted original file %s for media %s", originalUrl, entityMediaID)
				}
			}
		}
	}

	uc.publishEncodeRequest(ctx, entityMediaID, media.Url, media.MimeType)
}

// handlePreprocessFailure updates media encoding_status to "failed" and pushes SSE event
// when preprocessing (download/ffprobe/remux/transcode) fails.
// It still attempts to publish an encode request as a fallback so the transcode service
// can try to process the original file.
func (uc *UploadUseCase) handlePreprocessFailure(ctx context.Context, media *Media, entityMediaID, reason string) {
	// Update encoding_status to failed
	media.EncodingStatus = string(enums.MediaEncodingStatusFailed)
	if _, err := uc.mediaRepo.Update(ctx, media); err != nil {
		uc.log.Errorf("preprocess: failed to update media status to failed: %v", err)
	}

	// Push SSE failed event so frontend can react
	if uc.mediaUseCase != nil {
		uc.mediaUseCase.Publish(entityMediaID, &EncodingEvent{
			MediaId: entityMediaID,
			Task: &EncodingTask{
				Status: enums.EncodingTaskStatusFailed,
			},
		})
	}

	// Still try to publish encode request as fallback
	uc.publishEncodeRequest(ctx, entityMediaID, media.Url, media.MimeType)
}

func (uc *UploadUseCase) publishEncodeRequest(ctx context.Context, mediaID, mediaPath, contentType string) {
	payload, _ := json.Marshal(MediaEncodeRequest{
		MediaID:     mediaID,
		MediaPath:   mediaPath,
		ContentType: contentType,
	})
	msg := pubsub.NewMessage(payload)
	if uc.publisher != nil {
		if err := uc.publisher.Publish(pubsub.MediaEncodeRequestTopic, msg); err != nil {
			uc.log.Errorf("failed to publish encode request for media %s: %v", mediaID, err)
		}
	}
}

// defaultFeatureBool returns the per-media default for a BUG-139 feature mode
// (comments_mode / downloads_mode): opt_out → true (on by default), opt_in /
// disabled → false (off by default, or forced off when disabled). `fallback` is
// used when the setting is absent — comments default on, download default off.
func (uc *UploadUseCase) defaultFeatureBool(ctx context.Context, key string, fallback bool) bool {
	if uc.configProvider == nil {
		return fallback
	}
	switch uc.configProvider.Get(ctx, key) {
	case "opt_out":
		return true
	case "opt_in", "disabled":
		return false
	default:
		return fallback
	}
}
