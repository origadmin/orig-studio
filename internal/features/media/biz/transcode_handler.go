/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package biz

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origstudio/internal/conf"
	"origadmin/application/origstudio/internal/data/enums"
	contentbiz "origadmin/application/origstudio/internal/features/content/biz"
	"origadmin/application/origstudio/internal/features/media/ffmpeg"
	"origadmin/application/origstudio/internal/infra/pubsub"
	"origadmin/application/origstudio/internal/features/media/dto"
)

// MediaEncodeRequest is the payload for media.encode.request messages.
type MediaEncodeRequest struct {
	MediaID     string  `json:"media_id"`
	MediaPath   string  `json:"media_path"`
	ContentType string  `json:"content_type"`
	TaskID      *string `json:"task_id,omitempty"` // 可选：只重试特定任务
}

// MediaEncodeEvent is the payload for progress/completion messages.
type MediaEncodeEvent struct {
	MediaID string        `json:"media_id"`
	Task    *EncodingTask `json:"task,omitempty"`
	Status  string        `json:"status"` // processing, success, failed
	Error   string        `json:"error,omitempty"`
}

// TranscodeHandler handles incoming media.encode.request messages.
// It orchestrates the full transcoding pipeline:
//
//	thumbnail → parallel profile transcodes (direct HLS or GIF preview) → master playlist → status determination
type TranscodeHandler struct {
	mediaUC        *MediaUseCase
	profileRepo    dto.EncodeProfileRepo
	encodingRepo   dto.EncodingTaskRepo
	mediaRepo      MediaRepo
	worker         TranscodeWorker
	publisher      message.Publisher
	logger         *log.Helper
	paths          *conf.StoragePaths
	storage        Storage
	taskTimeout    time.Duration
	spriteUC       *SpriteUseCase
	notificationUC *contentbiz.NotificationUseCase
}

// NewTranscodeHandler creates a new TranscodeHandler.
func NewTranscodeHandler(
	mediaUC *MediaUseCase,
	profileRepo dto.EncodeProfileRepo,
	encodingRepo dto.EncodingTaskRepo,
	mediaRepo MediaRepo,
	worker TranscodeWorker,
	publisher message.Publisher,
	logger log.Logger,
	paths *conf.StoragePaths,
	storage Storage,
	taskTimeout time.Duration,
	spriteUC *SpriteUseCase,
	notificationUC *contentbiz.NotificationUseCase,
) *TranscodeHandler {
	return &TranscodeHandler{
		mediaUC:        mediaUC,
		profileRepo:    profileRepo,
		encodingRepo:   encodingRepo,
		mediaRepo:      mediaRepo,
		worker:         worker,
		publisher:      publisher,
		logger:         log.NewHelper(log.With(logger, "module", "transcode.handler")),
		paths:          paths,
		storage:        storage,
		taskTimeout:    taskTimeout,
		spriteUC:       spriteUC,
		notificationUC: notificationUC,
	}
}

// Handle processes a media.encode.request message.
// This is the entry point called by the Watermill router.
func (h *TranscodeHandler) Handle(msg *message.Message) error {
	ctx := msg.Context()

	var req MediaEncodeRequest
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		h.logger.Errorf("failed to unmarshal encode request: %v", err)
		return fmt.Errorf("unmarshal: %w", err)
	}

	h.logger.Infof("received encode request: media=%s path=%s", req.MediaID, req.MediaPath)

	if err := h.processMedia(ctx, &req); err != nil {
		h.logger.Errorf("media processing failed: media=%s err=%v", req.MediaID, err)
		return err
	}

	return nil
}

// processMedia runs the full transcoding pipeline for a single media.
//
// Pipeline:
//
//  1. Ensure UUID exists on media record
//  2. Generate thumbnail (if not already set)
//  3. Create encoding tasks for all active profiles
//  4. Submit video profile jobs → direct HLS output to hls/{id}/{profile_name}/
//  5. Submit preview job → GIF output to previews/{id}.gif
//  6. Collect results, generate master.m3u8 from successful variants
//  7. Determine final encoding_status:
//     - all video tasks success → "success"
//     - some video success + some failed → "partial"
//     - all video tasks failed → "failed"
//     - preview task outcome does NOT affect overall status
func (h *TranscodeHandler) processMedia(ctx context.Context, req *MediaEncodeRequest) error {
	mediaID := req.MediaID

	procCtx, cancel := context.WithTimeout(context.Background(), h.taskTimeout)
	defer cancel()

	// --- Step 1: Load media ---
	media, err := h.mediaRepo.Get(procCtx, mediaID)
	if err != nil {
		return fmt.Errorf("get media %s: %w", mediaID, err)
	}

	// Use media.Url from DB (may have been updated by preprocessing) rather than
	// the request payload, which could reference the original pre-remux path.
	sourcePath := media.Url
	if sourcePath == "" {
		sourcePath = req.MediaPath
	}

	var workDir *LocalWorkDir
	fullPath := h.paths.FullPath(sourcePath)
	if _, err := os.Stat(fullPath); err != nil {
		workDir, err = DownloadToLocalWorkDir(procCtx, h.storage, sourcePath, h.paths.FullPath)
		if err != nil {
			return fmt.Errorf("download source for S3: %w", err)
		}
		if workDir != nil {
			fullPath = workDir.LocalPath
		}
	}
	defer workDir.Cleanup()

	// Use media ID (which is already a UUID) for secure public paths
	mediaUUID := media.Id

	// Idempotency: if media is already successfully encoded, skip reprocessing
	if media.EncodingStatus == "success" && media.HlsFile != "" {
		h.logger.Infof("media %s already encoded successfully, skipping (redelivered message)", mediaID)
		// Ensure already-encoded media is published (backward compat for videos encoded before auto-publish)
		if media.State != "active" || media.ReviewStatus != "reviewed" || !media.Listable {
			media.State = "active"
			media.ReviewStatus = "reviewed"
			media.Listable = h.mediaUC.ShouldBeListable(media)
			if _, err := h.mediaRepo.Update(procCtx, media); err != nil {
				h.logger.Warnf("failed to auto-publish already-encoded media %s: %v", mediaID, err)
			}
		}
		return nil
	}

	// Update media status to processing
	media.EncodingStatus = "processing"
	if _, err := h.mediaRepo.Update(procCtx, media); err != nil {
		h.logger.Warnf("failed to update media status to processing: %v", err)
	}

	// --- Step 2: Generate thumbnail ---
	if media.Thumbnail == "" {
		thumbDir := h.paths.ThumbnailsDir()
		thumbFilename := fmt.Sprintf("%s.jpg", mediaUUID)
		thumbGeneratedPath, err := GenerateThumbnail(procCtx, fullPath, thumbDir, thumbFilename)
		if err == nil {
			media.Thumbnail = h.paths.RelativeThumbnail(mediaUUID)

			// Verify thumbnail file exists and log size
			if fi, statErr := os.Stat(thumbGeneratedPath); statErr == nil {
				h.logger.Infof("thumbnail generated for media %s: %s (%d bytes)", mediaID, thumbGeneratedPath, fi.Size())
			} else {
				h.logger.Warnf("thumbnail file missing after generation for media %s: %v", mediaID, statErr)
			}

			// Upload thumbnail to storage for non-local backends (S3/Hybrid)
			if data, readErr := os.ReadFile(thumbGeneratedPath); readErr == nil {
				thumbRelPath := h.paths.RelativeThumbnail(mediaUUID)
				if _, upErr := h.storage.Upload(procCtx, thumbRelPath, bytes.NewReader(data), int64(len(data)), "image/jpeg"); upErr != nil {
					h.logger.Warnf("failed to upload thumbnail to storage for media %s: %v", mediaID, upErr)
				}
			} else {
				h.logger.Warnf("failed to read thumbnail for media %s: %v (path=%s)", mediaID, readErr, thumbGeneratedPath)
			}

			if _, err := h.mediaRepo.Update(procCtx, media); err != nil {
				h.logger.Warnf("failed to save thumbnail for media %s: %v", mediaID, err)
			}
		} else {
			h.logger.Warnf("thumbnail generation failed for media %s: %v", mediaID, err)
		}
	}

	// --- Step 2.5: Get video resolution and save ---
	sourceHeight := 0
	if strings.HasPrefix(media.MimeType, "video/") {
		if srcW, srcH, err := ffmpeg.GetVideoResolution(procCtx, fullPath); err != nil {
			h.logger.Warnf("failed to get video resolution for media %s: %v", mediaID, err)
		} else {
			sourceHeight = srcH
			if err := h.mediaRepo.UpdateDimensions(procCtx, mediaID, srcW, srcH); err != nil {
				h.logger.Warnf("failed to save video dimensions for media %s: %v", mediaID, err)
			}
			h.logger.Infof("media %s resolution: %dx%d", mediaID, srcW, srcH)
		}
	}

	// --- Step 3: Fetch active profiles and get or create encoding tasks ---
	profiles, err := h.profileRepo.ListActive(procCtx)
	if err != nil {
		return fmt.Errorf("list profiles: %w", err)
	}

	// Separate video and preview profiles
	var videoProfiles, previewProfiles []*dto.EncodeProfile
	for _, p := range profiles {
		if IsPreviewProfile(p) {
			previewProfiles = append(previewProfiles, p)
		} else if IsVideoProfile(p) {
			profileHeight := resolutionToHeight(p.Resolution)
			if sourceHeight > 0 && profileHeight > sourceHeight {
				h.logger.Infof("skipping profile %s (height=%d) for media %s (source height=%d): would upscale",
					p.Name, profileHeight, mediaID, sourceHeight)
				continue
			}
			videoProfiles = append(videoProfiles, p)
		} else {
			h.logger.Infof("skipping unknown profile type: name=%s ext=%s", p.Name, p.Extension)
		}
	}

	allProfiles := append(videoProfiles, previewProfiles...)

	// Get existing tasks for this media
	existingTasks, err := h.encodingRepo.ListByMedia(procCtx, mediaID)
	if err != nil {
		h.logger.Warnf("failed to get existing tasks for media %s: %v", mediaID, err)
		existingTasks = nil
	}

	// Create a map of profile ID to existing task for quick lookup
	existingTaskMap := make(map[int]*EncodingTask)
	for _, t := range existingTasks {
		existingTaskMap[t.ProfileId] = t
	}

	// Get or create tasks in DB for all applicable profiles
	var tasks []*EncodingTask

	// 如果指定了 TaskID，说明是特定任务的重试
	if req.TaskID != nil {
		// 直接获取该任务，而不依赖 allProfiles (可能该 profile 已被设置为不活跃，但任务仍需处理)
		existingTask, err := h.encodingRepo.Get(procCtx, *req.TaskID)
		if err != nil {
			return fmt.Errorf("get encoding task %s: %w", *req.TaskID, err)
		}

		if existingTask.MediaId != mediaID {
			return fmt.Errorf("task %s does not belong to media %s", *req.TaskID, mediaID)
		}

		// 任务状态已经在 RetryTask 中被重置为 pending，这里不需要再次重置
		// 只需要将任务加入处理队列
		tasks = append(tasks, existingTask)
		h.logger.Infof("processing specific task %s (media=%s)", existingTask.Id, mediaID)
	} else {
		// 没有指定 TaskID，处理所有需要处理的任务（初始上传或重试所有失败）
		for _, p := range allProfiles {
			// Check if there's an existing task for this profile
			if existingTask, exists := existingTaskMap[p.Id]; exists {
				// If task is already successful, skip it
				if existingTask.Status == enums.EncodingTaskStatusSuccess {
					h.logger.Infof("skipping already successful task %s for profile %s (media=%s)", existingTask.Id, p.Name, mediaID)
					continue
				}
				// Only process tasks that are already in pending state (for retries)
				// Do NOT reset other states - let them remain as is
				if existingTask.Status == enums.EncodingTaskStatusPending {
					tasks = append(tasks, existingTask)
					h.logger.Infof("processing pending task %s for profile %s (media=%s)", existingTask.Id, p.Name, mediaID)
					continue
				}
				// For non-pending, non-success tasks, skip them
				h.logger.Infof("skipping task %s with status %s for profile %s (media=%s)", existingTask.Id, existingTask.Status, p.Name, mediaID)
				continue
			}

			// Create a new task if no existing task found (this is for initial upload)
			task := &EncodingTask{
				MediaId:   mediaID,
				ProfileId: p.Id,
				Status:    enums.EncodingTaskStatusPending,
			}
			t, err := h.encodingRepo.Create(procCtx, task)
			if err != nil {
				h.logger.Warnf("failed to create encoding task for profile %s: %v", p.Name, err)
				continue
			}
			tasks = append(tasks, t)
		}
	}

	if len(tasks) == 0 {
		// Check if all tasks are already in a terminal state (success/failed) — this is a redelivered message
		allExistingTasks, _ := h.encodingRepo.ListByMedia(procCtx, mediaID)
		allTerminal := len(allExistingTasks) > 0
		videoSuccess := 0
		videoTotal := 0
		for _, et := range allExistingTasks {
			prof, pErr := h.profileRepo.Get(procCtx, et.ProfileId)
			if pErr != nil || prof == nil {
				continue
			}
			if IsVideoProfile(prof) {
				videoTotal++
				if et.Status == enums.EncodingTaskStatusSuccess {
					videoSuccess++
				}
			}
			if et.Status != enums.EncodingTaskStatusSuccess && et.Status != enums.EncodingTaskStatusFailed {
				allTerminal = false
			}
		}
		if allTerminal {
			h.logger.Infof("media %s already processed (redelivered message), skipping", mediaID)
			// Determine final status from existing tasks
			switch {
			case videoTotal == 0:
				media.EncodingStatus = "success"
			case videoSuccess == videoTotal:
				media.EncodingStatus = "success"
			case videoSuccess > 0:
				media.EncodingStatus = "partial"
			default:
				media.EncodingStatus = "failed"
			}
			// Auto-publish on redelivery too
			if media.EncodingStatus == "success" {
				media.State = "active"
				media.ReviewStatus = "reviewed"
				media.Listable = h.mediaUC.ShouldBeListable(media)
			}
			// Update HLS file reference if needed
		if media.HlsFile == "" && media.EncodingStatus == "success" {
			media.HlsFile = h.paths.RelativeHLSMaster(mediaUUID)
		}
			h.mediaRepo.Update(procCtx, media)
			h.publishEvent(ctx, &MediaEncodeEvent{
				MediaID: mediaID,
				Status:  media.EncodingStatus,
			})
			return nil
		}
		h.logger.Warnf("no encoding tasks created for media %s", mediaID)
		media.EncodingStatus = "failed"
		h.mediaRepo.Update(procCtx, media)
		return nil
	}

	// --- Step 4+5: Submit all jobs in parallel ---
	// Directory layout:
	//   hls/{uuid}/{profile_name}/index.m3u8 + segment_XXX.ts  (video profiles)
	//   previews/{uuid}.gif                                    (preview)

	hlsBaseDir := h.paths.HLSDirForMedia(mediaUUID)

	var wg sync.WaitGroup
	resultsCh := make(chan transcodeResult, len(tasks))

	for _, t := range tasks {
		profile, err := h.profileRepo.Get(procCtx, t.ProfileId)
		if err != nil {
			h.logger.Warnf("profile %d not found: %v", t.ProfileId, err)
			continue
		}

		// Determine output directory based on profile type
		var outputDir string
		if IsVideoProfile(profile) {
			outputDir = filepath.Join(hlsBaseDir, profile.Name)
		} else {
			// Preview: output goes directly to previews/ directory
			outputDir = h.paths.PreviewsDir()
		}

		// 创建局部变量，避免闭包问题
		task := t
		job := TranscodeJob{
			MediaID:      mediaID,
			TaskID:       task.Id,
			Profile:      profile,
			InputPath:    fullPath,
			OutputDir:    outputDir,
			EncodingRepo: h.encodingRepo,
			MediaUC:      h.mediaUC,
			Logger:       h.logger,
		}

		wg.Add(1)
		go func() {
			defer wg.Done()

			result := transcodeResult{taskID: task.Id}

			if err := h.worker.Submit(procCtx, job); err != nil {
				result.err = fmt.Errorf("worker submit: %w", err)
				resultsCh <- result
				return
			}

			// Wait for the output file(s) to appear
			if err := h.waitForOutput(job, task, &result); err != nil {
				result.err = err
			}

			resultsCh <- result
		}()
	}

	// Close channel after all goroutines finish
	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	// --- Step 6: Collect and update processing results ---
	for result := range resultsCh {
		t, _ := h.encodingRepo.Get(procCtx, result.taskID)
		if t == nil {
			continue
		}

		profile, _ := h.profileRepo.Get(procCtx, t.ProfileId)
		if result.err != nil {
			t.Status = enums.EncodingTaskStatusFailed
			t.ErrorMessage = result.err.Error()
		} else {
			t.Status = enums.EncodingTaskStatusSuccess
			t.ErrorMessage = ""

			if profile != nil {
				if IsVideoProfile(profile) {
					t.OutputPath = h.paths.RelativeHLSProfile(mediaUUID, profile.Name)
				} else if IsPreviewProfile(profile) {
					t.OutputPath = h.paths.RelativePreview(mediaUUID)
				}
			}
		}

		if _, err := h.encodingRepo.Update(procCtx, t); err != nil {
			h.logger.Warnf("failed to update task %s: %v", t.Id, err)
		}

		// Notify frontend of task change
		if t.Status == enums.EncodingTaskStatusSuccess {
			h.mediaUC.Publish(mediaID, &EncodingEvent{
				MediaId:  mediaID,
				Task:     t,
				Progress: 100,
				Speed:    "",
				Fps:      0,
				Time:     0,
			})
		} else {
			h.mediaUC.Publish(mediaID, &EncodingEvent{MediaId: mediaID, Task: t})
		}
		h.publishEvent(ctx, &MediaEncodeEvent{
			MediaID: mediaID,
			Task:    t,
			Status:  string(t.Status),
			Error:   t.ErrorMessage,
		})
	}

	// --- Step 7: Consolidate overall media status based on ALL tasks ---
	allTasks, err := h.encodingRepo.ListByMedia(procCtx, mediaID)
	if err != nil {
		return fmt.Errorf("failed to fetch all tasks for status update: %w", err)
	}

	videoSuccessCount := 0
	videoFailedCount := 0
	videoTotalCount := 0
	var variantInfos []ffmpeg.VariantInfo

	for _, t := range allTasks {
		profile, _ := h.profileRepo.Get(procCtx, t.ProfileId)
		if profile == nil {
			h.logger.Warnf("task %s has no profile, skipping", t.Id)
			continue
		}

		h.logger.Infof("processing task %s: profile=%s status=%s isPreview=%v", t.Id, profile.Name, t.Status, IsPreviewProfile(profile))

		if IsPreviewProfile(profile) {
			// Update preview file path if preview task succeeded
		if t.Status == enums.EncodingTaskStatusSuccess {
			media.PreviewFilePath = h.paths.RelativePreview(mediaUUID)
			h.logger.Infof("set preview file path to: %s", media.PreviewFilePath)
		}
			continue
		}

		videoTotalCount++
		if t.Status == enums.EncodingTaskStatusSuccess {
			videoSuccessCount++
			variantInfos = append(variantInfos, ffmpeg.VariantInfo{
				Path:       fmt.Sprintf("%s/index.m3u8", profile.Name),
				Bandwidth:  estimateBandwidth(profile),
				Resolution: ffmpeg.ResolutionToSize(profile.Resolution),
				Name:       profile.Name,
			})
		} else if t.Status == enums.EncodingTaskStatusFailed {
			videoFailedCount++
		}
	}

	// Update overall media encoding status
	switch {
	case videoTotalCount == 0:
		media.EncodingStatus = "success" // No video to process
	case videoSuccessCount == videoTotalCount:
		media.EncodingStatus = "success"
	case videoSuccessCount > 0:
		media.EncodingStatus = "partial"
	default:
		media.EncodingStatus = "failed"
	}

	// Auto-publish: when video transcoding succeeds, automatically set state to active
	// and review_status to reviewed so the video becomes publicly visible
	if media.EncodingStatus == "success" {
		media.State = "active"
		media.ReviewStatus = "reviewed"
	}

	// Recalculate listable now that encoding_status and state may have changed
	media.Listable = h.mediaUC.ShouldBeListable(media)

	// Regenerate master playlist if we have any successful variants
	if len(variantInfos) > 0 {
		masterRelPath := h.paths.RelativeHLSMaster(mediaUUID)
		if _, err := GenerateMasterPlaylist(hlsBaseDir, variantInfos); err != nil {
			h.logger.Errorf("master playlist generation failed: %v", err)
		} else {
			media.HlsFile = masterRelPath
		}
	} else if videoFailedCount > 0 && videoSuccessCount == 0 {
		os.RemoveAll(hlsBaseDir)
		media.HlsFile = ""
	}

	// Final media update
	if _, err := h.mediaRepo.Update(procCtx, media); err != nil {
		h.logger.Errorf("failed to update media final status: %v", err)
	}

	// BUG-264: 转码终态通知上传者（设计锚点 12-NOTIFICATION_TYPES_AND_PREFS.md §4）。
	// 仅正常完成路径触发（redelivery 分支已提前 return，不重复通知）。
	if h.notificationUC != nil {
		if err := h.notificationUC.NotifyTranscodeStatus(procCtx, media.UserId, media.Title, media.EncodingStatus, ""); err != nil {
			h.logger.Warnf("failed to notify transcode status for media %s: %v", mediaID, err)
		}
	}

	// Final completion notification
	h.publishEvent(ctx, &MediaEncodeEvent{
		MediaID: mediaID,
		Status:  media.EncodingStatus,
	})

	h.logger.Infof("media processing complete: media=%s uuid=%s status=%s (video: %d ok / %d fail)",
		mediaID, mediaUUID, media.EncodingStatus, videoSuccessCount, videoFailedCount)

	if workDir != nil {
		if err := h.uploadOutputDirs(procCtx, mediaUUID); err != nil {
			h.logger.Warnf("failed to upload output dirs to S3 for media %s: %v", mediaID, err)
		}
	}

	if media.Type == "video" && h.spriteUC != nil {
		go func() {
			if err := h.spriteUC.ProcessPostTranscode(context.Background(), mediaID); err != nil {
				h.logger.Warnf("post-transcode processing failed for media %s: %v", mediaID, err)
			}
		}()
	}

	return nil
}

// waitForOutput polls for the expected output file after worker submission.
// For video profiles: checks for index.m3u8; for preview: checks for .gif file.
// Progress updates are handled by the transcoding worker itself for video profiles.
// For preview profiles, we provide basic progress updates here.
func (h *TranscodeHandler) waitForOutput(
	job TranscodeJob,
	task *EncodingTask,
	result *transcodeResult,
) error {
	var expectedFile string
	maxAttempts := 600 // max 20 min wait per task (2s interval)

	if IsVideoProfile(job.Profile) {
		expectedFile = filepath.Join(job.OutputDir, "index.m3u8")
	} else if IsPreviewProfile(job.Profile) {
		expectedFile = filepath.Join(job.OutputDir, fmt.Sprintf("%s.gif", job.MediaID))
	} else {
		return nil // unknown type, nothing to wait for
	}

	// 等待任务开始执行（状态变为 processing、success 或 failed）
	maxWaitForStart := 120 // 最多等待 10 分钟（5s interval）
	for i := 0; i < maxWaitForStart; i++ {
		// 检查任务状态
		currentTask, err := h.encodingRepo.Get(context.Background(), task.Id)
		if err == nil && currentTask != nil {
			switch currentTask.Status {
			case enums.EncodingTaskStatusSuccess:
				// Task already completed successfully (fast job), skip to file check
				h.logger.Infof("task %s already completed successfully, checking output", task.Id)
				if _, err := os.Stat(expectedFile); err == nil {
					return nil
				}
				// File not found yet despite success status; continue to file-wait loop
				goto fileWait
			case enums.EncodingTaskStatusProcessing:
				h.logger.Infof("task %s has started processing, beginning file wait", task.Id)
				goto fileWait
			case enums.EncodingTaskStatusFailed:
				h.logger.Warnf("task %s failed during execution: %s", task.Id, currentTask.ErrorMessage)
				return fmt.Errorf("task %s failed: %s", task.Id, currentTask.ErrorMessage)
			}
		}
		time.Sleep(5 * time.Second)
		if i == maxWaitForStart-1 {
			return fmt.Errorf("task %s did not start processing within timeout", task.Id)
		}
	}

fileWait:
	// 初始延迟：给转码任务时间开始生成文件
	time.Sleep(2 * time.Second)

	// Poll for output file
	// For video profiles, progress is updated by the transcoding worker
	// For preview profiles, we provide basic progress updates here
	for i := 0; i < maxAttempts; i++ {
		// Check task status first
		currentTask, err := h.encodingRepo.Get(context.Background(), task.Id)
		if err == nil && currentTask != nil {
			if currentTask.Status == enums.EncodingTaskStatusFailed {
				h.logger.Warnf("task %s failed during file wait: %s", task.Id, currentTask.ErrorMessage)
				return fmt.Errorf("task %s failed: %s", task.Id, currentTask.ErrorMessage)
			}
			if currentTask.Status == enums.EncodingTaskStatusSuccess {
				if _, err := os.Stat(expectedFile); err == nil {
					return nil
				}
			}
		}

		// For preview profiles, provide basic progress updates
		// Video profiles have their own progress updates from ffmpeg
		if IsPreviewProfile(job.Profile) && currentTask != nil {
			// Basic progress: 20% to 90%
			progress := 20 + (i * 70 / maxAttempts)
			if progress > 90 {
				progress = 90
			}

			// Publish progress update via SSE
			if job.MediaUC != nil {
				taskCopy := *currentTask
				job.MediaUC.Publish(job.MediaID, &EncodingEvent{
					MediaId:  job.MediaID,
					Task:     &taskCopy,
					Progress: progress,
					Speed:    "",
					Fps:      0,
					Time:     0,
				})
			}
		}

		time.Sleep(2 * time.Second)

		// Check if expected file/directory exists
		if _, err := os.Stat(expectedFile); err == nil {
			return nil // file/directory exists
		}
	}

	return fmt.Errorf("timeout waiting for output: %s", expectedFile)
}

// transcodeResult holds the result of a single transcode job execution.
type transcodeResult struct {
	taskID string
	err    error
}

// publishEvent sends an event to the progress/completion topic.
func (h *TranscodeHandler) publishEvent(ctx context.Context, event *MediaEncodeEvent) {
	if h.publisher == nil {
		return
	}

	payload, err := json.Marshal(event)
	if err != nil {
		h.logger.Warnf("failed to marshal encode event: %v", err)
		return
	}

	msg := pubsub.NewMessage(payload)

	topic := pubsub.MediaEncodeProgressTopic
	if event.Status == string(enums.MediaEncodingStatusSuccess) || event.Status == string(enums.MediaEncodingStatusFailed) || event.Status == string(enums.MediaEncodingStatusPartial) {
		topic = pubsub.MediaEncodeCompletedTopic
	}

	if err := h.publisher.Publish(topic, msg); err != nil {
		h.logger.Warnf("failed to publish event to %s: %v", topic, err)
	}
}

func (h *TranscodeHandler) uploadOutputDirs(ctx context.Context, mediaUUID string) error {
	hlsDir := h.paths.HLSDirForMedia(mediaUUID)
	if _, err := os.Stat(hlsDir); err == nil {
		if err := h.storage.UploadDir(ctx, hlsDir, h.paths.Relative("hls", mediaUUID)); err != nil {
			h.logger.Warnf("failed to upload HLS dir to S3: %v", err)
		}
	}

	thumbDir := h.paths.ThumbnailsDir()
	if _, err := os.Stat(thumbDir); err == nil {
		thumbPattern := filepath.Join(thumbDir, mediaUUID+".*")
		matches, _ := filepath.Glob(thumbPattern)
		for _, m := range matches {
			relKey := h.paths.Relative("thumbnails", filepath.Base(m))
			// Read into memory first to avoid source/destination path conflict
			// where LocalStorage.Upload's os.Create truncates the file before io.Copy.
			data, err := os.ReadFile(m)
			if err != nil {
				continue
			}
			if _, err := h.storage.Upload(ctx, relKey, bytes.NewReader(data), int64(len(data)), ""); err != nil {
				h.logger.Warnf("failed to upload thumbnail %s to storage: %v", relKey, err)
			}
		}
	}

	previewPath := h.paths.PreviewAbsPath(mediaUUID)
	if _, err := os.Stat(previewPath); err == nil {
		relKey := h.paths.RelativePreview(mediaUUID)
		// Read into memory first to avoid source/destination path conflict.
		data, err := os.ReadFile(previewPath)
		if err == nil {
			if _, err := h.storage.Upload(ctx, relKey, bytes.NewReader(data), int64(len(data)), ""); err != nil {
				h.logger.Warnf("failed to upload preview %s to storage: %v", relKey, err)
			}
		}
	}

	spriteDir := h.paths.SpriteDirAbs(mediaUUID)
	if _, err := os.Stat(spriteDir); err == nil {
		if err := h.storage.UploadDir(ctx, spriteDir, h.paths.Relative("sprites", mediaUUID)); err != nil {
			h.logger.Warnf("failed to upload sprite dir to S3: %v", err)
		}
	}

	return nil
}

func estimateBandwidth(p *dto.EncodeProfile) int {
	// Try parsing from BentoParameters
	if p.BentoParameters != "" {
		fields := strings.Fields(p.BentoParameters)
		for i, f := range fields {
			if f == "--video-bitrate" && i+1 < len(fields) {
				bps := parseBitrateToBps(fields[i+1])
				if bps > 0 {
					return bps
				}
			}
		}
	}

	// Fallback: estimate by resolution height
	switch p.Resolution {
	case "2160":
		return 20_000_000
	case "1440":
		return 12_000_000
	case "1080":
		return 8_000_000
	case "720":
		return 4_000_000
	case "480":
		return 2_000_000
	case "360":
		return 1_000_000
	case "240":
		return 500_000
	default:
		return 1_000_000
	}
}

// resolutionToHeight converts a resolution string to its height value.
// Accepts formats: "720", "1280x720", returns 0 for non-numeric values.
func resolutionToHeight(resolution string) int {
	if resolution == "" || resolution == "-" {
		return 0
	}
	if strings.Contains(resolution, "x") {
		parts := strings.SplitN(resolution, "x", 2)
		h, err := strconv.Atoi(parts[1])
		if err != nil {
			return 0
		}
		return h
	}
	h, err := strconv.Atoi(resolution)
	if err != nil {
		return 0
	}
	return h
}
