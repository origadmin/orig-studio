/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origcms/internal/data/enums"
	"origadmin/application/origcms/internal/helpers/ffmpeg"
	"origadmin/application/origcms/internal/pubsub"
	"origadmin/application/origcms/internal/svc-media/dto"
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
	mediaUC      *MediaUseCase
	profileRepo  dto.EncodeProfileRepo
	encodingRepo dto.EncodingTaskRepo
	mediaRepo    MediaRepo
	worker       TranscodeWorker
	publisher    message.Publisher
	logger       *log.Helper
	baseDir      string
	taskTimeout  time.Duration
	spriteUC     *SpriteUseCase
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
	baseDir string,
	taskTimeout time.Duration,
	spriteUC *SpriteUseCase,
) *TranscodeHandler {
	return &TranscodeHandler{
		mediaUC:      mediaUC,
		profileRepo:  profileRepo,
		encodingRepo: encodingRepo,
		mediaRepo:    mediaRepo,
		worker:       worker,
		publisher:    publisher,
		logger:       log.NewHelper(log.With(logger, "module", "transcode.handler")),
		baseDir:      baseDir,
		taskTimeout:  taskTimeout,
		spriteUC:     spriteUC,
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
	fullPath := filepath.Join(h.baseDir, req.MediaPath)

	procCtx, cancel := context.WithTimeout(context.Background(), h.taskTimeout)
	defer cancel()

	// --- Step 1: Load media ---
	media, err := h.mediaRepo.Get(procCtx, mediaID)
	if err != nil {
		return fmt.Errorf("get media %s: %w", mediaID, err)
	}

	// Use media ID (which is already a UUID) for secure public paths
	mediaUUID := media.Id

	// Update media status to processing
	media.EncodingStatus = "processing"
	if _, err := h.mediaRepo.Update(procCtx, media); err != nil {
		h.logger.Warnf("failed to update media status to processing: %v", err)
	}

	// --- Step 2: Generate thumbnail ---
	if media.Thumbnail == "" {
		thumbDir := filepath.Join(h.baseDir, "thumbnails")
		thumbFilename := fmt.Sprintf("%s.jpg", mediaUUID)
		if _, err := GenerateThumbnail(procCtx, fullPath, thumbDir, thumbFilename); err == nil {
			media.Thumbnail = fmt.Sprintf("thumbnails/%s.jpg", mediaUUID)
			if _, err := h.mediaRepo.Update(procCtx, media); err != nil {
				h.logger.Warnf("failed to save thumbnail for media %s: %v", mediaID, err)
			}
		} else {
			h.logger.Warnf("thumbnail generation failed for media %s: %v", mediaID, err)
		}
	}

	// --- Step 3: Fetch active profiles and get or create encoding tasks ---
	profiles, err := h.profileRepo.ListActive(procCtx)
	if err != nil {
		return fmt.Errorf("list profiles: %w", err)
	}

	// Separate video, preview, and frames profiles
	var videoProfiles, previewProfiles, framesProfiles []*dto.EncodeProfile
	for _, p := range profiles {
		if IsPreviewProfile(p) {
			previewProfiles = append(previewProfiles, p)
		} else if IsFramesProfile(p) {
			framesProfiles = append(framesProfiles, p)
		} else if IsVideoProfile(p) {
			// Use all active video profiles regardless of video type
			videoProfiles = append(videoProfiles, p)
		} else {
			h.logger.Infof("skipping unknown profile type: name=%s ext=%s", p.Name, p.Extension)
		}
	}

	allProfiles := append(append(videoProfiles, previewProfiles...), framesProfiles...)

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
		h.logger.Warnf("no encoding tasks created for media %s", mediaID)
		media.EncodingStatus = "failed"
		h.mediaRepo.Update(procCtx, media)
		return nil
	}

	// --- Step 4+5: Submit all jobs in parallel ---
	// Directory layout:
	//   hls/{uuid}/{profile_name}/index.m3u8 + segment_XXX.ts  (video profiles)
	//   previews/{uuid}.gif                                    (preview)

	hlsBaseDir := filepath.Join(h.baseDir, "hls", mediaUUID)

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
		} else if IsFramesProfile(profile) {
			// Frames: output goes to frames/ dir; we pass hlsBaseDir as anchor
			outputDir = hlsBaseDir // executeFramesJob will navigate up to frames/
		} else {
			// Preview: output goes to previews/ dir; we pass hlsBaseDir as anchor
			outputDir = hlsBaseDir // executePreviewJob will navigate up to previews/
		}

		// 创建局部变量，避免闭包问题
		task := t
		job := TranscodeJob{
			MediaID:      mediaID,
			TaskID:       task.Id,
			Profile:      profile,
			InputPath:    fullPath,
			OutputDir:    outputDir,
			UUID:         mediaUUID,
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
					t.OutputPath = fmt.Sprintf("%s/%s/index.m3u8", mediaUUID, profile.Name)
				} else if IsPreviewProfile(profile) {
					t.OutputPath = fmt.Sprintf("previews/%s.gif", mediaUUID)
				} else if IsFramesProfile(profile) {
					t.OutputPath = fmt.Sprintf("frames/%s/", mediaUUID)
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
				media.PreviewFilePath = fmt.Sprintf("hls/previews/%s.gif", mediaUUID)
				h.logger.Infof("set preview file path to: %s", media.PreviewFilePath)
			}
			continue
		}

		if IsFramesProfile(profile) {
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

	// Regenerate master playlist if we have any successful variants
	if len(variantInfos) > 0 {
		masterRelPath := fmt.Sprintf("hls/%s/master.m3u8", mediaUUID)
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

	// Final completion notification
	h.publishEvent(ctx, &MediaEncodeEvent{
		MediaID: mediaID,
		Status:  media.EncodingStatus,
	})

	h.logger.Infof("media processing complete: media=%s uuid=%s status=%s (video: %d ok / %d fail)",
		mediaID, mediaUUID, media.EncodingStatus, videoSuccessCount, videoFailedCount)

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
		expectedFile = filepath.Join(job.OutputDir, "..", "previews", fmt.Sprintf("%s.gif", job.MediaID))
	} else if IsFramesProfile(job.Profile) {
		// For frames, we just wait for the directory to be created
		expectedFile = filepath.Join(job.OutputDir, "..", "frames")
	} else {
		return nil // unknown type, nothing to wait for
	}

	// 等待任务开始执行（状态变为 processing 或 failed）
	maxWaitForStart := 120 // 最多等待 10 分钟（5s interval）
	for i := 0; i < maxWaitForStart; i++ {
		// 检查任务状态
		currentTask, err := h.encodingRepo.Get(context.Background(), task.Id)
		if err == nil && currentTask != nil {
			if currentTask.Status == enums.EncodingTaskStatusProcessing {
				h.logger.Infof("task %s has started processing, beginning file wait", task.Id)
				break
			} else if currentTask.Status == enums.EncodingTaskStatusFailed {
				h.logger.Warnf("task %s failed during execution: %s", task.Id, currentTask.ErrorMessage)
				return fmt.Errorf("task %s failed: %s", task.Id, currentTask.ErrorMessage)
			}
		}
		time.Sleep(5 * time.Second)
		if i == maxWaitForStart-1 {
			return fmt.Errorf("task %s did not start processing within timeout", task.Id)
		}
	}

	// 初始延迟：给转码任务时间开始生成文件
	time.Sleep(2 * time.Second)

	// Poll for output file
	// For video profiles, progress is updated by the transcoding worker
	// For preview profiles, we provide basic progress updates here
	for i := 0; i < maxAttempts; i++ {
		// Check task status first
		currentTask, err := h.encodingRepo.Get(context.Background(), task.Id)
		if err == nil && currentTask != nil && currentTask.Status == enums.EncodingTaskStatusFailed {
			h.logger.Warnf("task %s failed during file wait: %s", task.Id, currentTask.ErrorMessage)
			return fmt.Errorf("task %s failed: %s", task.Id, currentTask.ErrorMessage)
		}

		// For preview and frames profiles, provide basic progress updates
		// Video profiles have their own progress updates from ffmpeg
		if (IsPreviewProfile(job.Profile) || IsFramesProfile(job.Profile)) && currentTask != nil {
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
			// For frames profile, we also check if at least one frame exists
			if IsFramesProfile(job.Profile) {
				framesDir := expectedFile
				frameFile := filepath.Join(framesDir, "frame_001.jpg")
				if _, err := os.Stat(frameFile); err == nil {
					return nil // directory and first frame exist
				}
				continue // wait more for frames to be generated
			}
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

// estimateBandwidth extracts bandwidth from profile's BentoParameters (e.g., "--video-bitrate 400k" → 400000).
// Falls back to resolution-based estimation if not parseable.
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
