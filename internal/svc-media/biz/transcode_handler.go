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

	"origadmin/application/origcms/internal/helpers/ffmpeg"
	"origadmin/application/origcms/internal/pubsub"
)

// MediaEncodeRequest is the payload for media.encode.request messages.
type MediaEncodeRequest struct {
	MediaID     int64  `json:"media_id"`
	MediaPath   string `json:"media_path"`
	ContentType string `json:"content_type"`
}

// MediaEncodeEvent is the payload for progress/completion messages.
type MediaEncodeEvent struct {
	MediaID  int64         `json:"media_id"`
	Task     *EncodingTask `json:"task,omitempty"`
	Status   string        `json:"status"` // processing, success, failed
	Progress int           `json:"progress"`
	Error    string        `json:"error,omitempty"`
}

// TranscodeHandler handles incoming media.encode.request messages.
// It orchestrates the full transcoding pipeline:
//
//	thumbnail → parallel profile transcodes (direct HLS or GIF preview) → master playlist → status determination
type TranscodeHandler struct {
	mediaUC      *MediaUseCase
	profileRepo  EncodeProfileRepo
	encodingRepo EncodingTaskRepo
	mediaRepo    MediaRepo
	worker       TranscodeWorker
	publisher    message.Publisher
	logger       *log.Helper
	baseDir      string // e.g., "./data/uploads"
}

// NewTranscodeHandler creates a new TranscodeHandler.
func NewTranscodeHandler(
	mediaUC *MediaUseCase,
	profileRepo EncodeProfileRepo,
	encodingRepo EncodingTaskRepo,
	mediaRepo MediaRepo,
	worker TranscodeWorker,
	publisher message.Publisher,
	logger log.Logger,
	baseDir string,
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

	h.logger.Infof("received encode request: media=%d path=%s", req.MediaID, req.MediaPath)

	if err := h.processMedia(ctx, &req); err != nil {
		h.logger.Errorf("media processing failed: media=%d err=%v", req.MediaID, err)
		return err
	}

	return nil
}

// processMedia runs the full transcoding pipeline for a single media.
//
// Pipeline:
//
//	1. Ensure UUID exists on media record
//	2. Generate thumbnail (if not already set)
//	3. Create encoding tasks for all active profiles
//	4. Submit video profile jobs → direct HLS output to hls/{uuid}/{profile_name}/
//	5. Submit preview job → GIF output to previews/{uuid}.gif
//	6. Collect results, generate master.m3u8 from successful variants
//	7. Determine final encoding_status:
//	       - all video tasks success → "success"
//	       - some video success + some failed → "partial"
//	       - all video tasks failed → "failed"
//	       - preview task outcome does NOT affect overall status
func (h *TranscodeHandler) processMedia(ctx context.Context, req *MediaEncodeRequest) error {
	mediaID := req.MediaID
	fullPath := filepath.Join(h.baseDir, req.MediaPath)

	procCtx, cancel := context.WithTimeout(context.Background(), 4*time.Hour)
	defer cancel()

	// --- Step 1: Load media and ensure UUID ---
	media, err := h.mediaRepo.Get(procCtx, mediaID)
	if err != nil {
		return fmt.Errorf("get media %d: %w", mediaID, err)
	}

	// Generate UUID if not present (for secure public paths)
	if media.Uuid == "" {
		media.Uuid = GenerateUUID()
		if _, err := h.mediaRepo.Update(procCtx, media); err != nil {
			h.logger.Warnf("failed to save UUID for media %d: %v", mediaID, err)
		}
	}
	mediaUUID := media.Uuid

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
				h.logger.Warnf("failed to save thumbnail for media %d: %v", mediaID, err)
			}
		} else {
			h.logger.Warnf("thumbnail generation failed for media %d: %v", mediaID, err)
		}
	}

	// --- Step 3: Fetch active profiles and create encoding tasks ---
	profiles, err := h.profileRepo.ListActive(procCtx)
	if err != nil {
		return fmt.Errorf("list profiles: %w", err)
	}

	// Separate video and preview profiles
	var videoProfiles, previewProfiles []*EncodeProfile
	for _, p := range profiles {
		if IsPreviewProfile(p) {
			previewProfiles = append(previewProfiles, p)
		} else if IsVideoProfile(p) {
			videoProfiles = append(videoProfiles, p)
		} else {
			h.logger.Infof("skipping unknown profile type: name=%s ext=%s", p.Name, p.Extension)
		}
	}

	allProfiles := append(videoProfiles, previewProfiles...)

	// Create tasks in DB for all applicable profiles
	var tasks []*EncodingTask
	for _, p := range allProfiles {
		task := &EncodingTask{
			MediaId:   mediaID,
			ProfileId: p.Id,
			Status:    "pending",
			Progress:  0,
		}
		t, err := h.encodingRepo.Create(procCtx, task)
		if err != nil {
			h.logger.Warnf("failed to create encoding task for profile %s: %v", p.Name, err)
			continue
		}
		tasks = append(tasks, t)
	}

	if len(tasks) == 0 {
		h.logger.Warnf("no encoding tasks created for media %d", mediaID)
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

		// Mark task as processing
		t.Status = "processing"
		t.Progress = 10
		if _, err := h.encodingRepo.Update(procCtx, t); err != nil {
			h.logger.Warnf("failed to update task %d status: %v", t.Id, err)
		}
		h.mediaUC.Publish(mediaID, &EncodingEvent{MediaId: mediaID, Task: t})

		// Determine output directory based on profile type
		var outputDir string
		if IsVideoProfile(profile) {
			outputDir = filepath.Join(hlsBaseDir, profile.Name)
		} else {
			// Preview: output goes to previews/ dir; we pass hlsBaseDir as anchor
			outputDir = hlsBaseDir // executePreviewJob will navigate up to previews/
		}

		job := TranscodeJob{
			MediaID:   mediaID,
			TaskID:    int64(t.Id),
			Profile:   profile,
			InputPath: fullPath,
			OutputDir: outputDir,
			UUID:      mediaUUID,
		}

		wg.Add(1)
		go func(task *EncodingTask) {
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
		}(t)
	}

	// Close channel after all goroutines finish
	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	// --- Step 6: Collect results and determine final state ---
	type taskOutcome struct {
		task        *EncodingTask
		variantInfo *ffmpeg.VariantInfo // only set for successful video profiles
		isPreview   bool
	}

	var videoOutcomes []taskOutcome
	var previewOutcomes []taskOutcome

	for result := range resultsCh {
		t, _ := h.encodingRepo.Get(procCtx, result.taskID)
		if t == nil {
			continue
		}

		// Get profile info for this task
		profile, _ := h.profileRepo.Get(procCtx, t.ProfileId)
		outcome := taskOutcome{task: t, isPreview: IsPreviewProfile(profile)}

		if result.err != nil {
			t.Status = "failed"
			t.ErrorMessage = result.err.Error()
			t.Progress = 0
		} else {
			t.Status = "success"
			t.Progress = 100

			if IsVideoProfile(profile) {
				// Store HLS path: hls/{uuid}/{profile_name}/index.m3u8
				relativePlaylist := fmt.Sprintf("%s/%s/index.m3u8", mediaUUID, profile.Name)
				t.OutputPath = relativePlaylist

				outcome.variantInfo = &ffmpeg.VariantInfo{
					Path:       fmt.Sprintf("%s/index.m3u8", profile.Name), // relative within hlsBaseDir
					Bandwidth:  estimateBandwidth(profile),
					Resolution: ffmpeg.ResolutionToSize(profile.Resolution),
					Name:       profile.Name,
				}
			} else if IsPreviewProfile(profile) {
				// Store GIF path: previews/{uuid}.gif
				t.OutputPath = fmt.Sprintf("previews/%s.gif", mediaUUID)
			}
		}

		if _, err := h.encodingRepo.Update(procCtx, t); err != nil {
			h.logger.Warnf("failed to update task %d: %v", t.Id, err)
		}
		h.mediaUC.Publish(mediaID, &EncodingEvent{MediaId: mediaID, Task: t})

		// Publish progress event
		h.publishEvent(ctx, &MediaEncodeEvent{
			MediaID:  mediaID,
			Task:     t,
			Status:   t.Status,
			Progress: t.Progress,
			Error:    t.ErrorMessage,
		})

		if outcome.isPreview {
			previewOutcomes = append(previewOutcomes, outcome)
		} else {
			videoOutcomes = append(videoOutcomes, outcome)
		}
	}

	// --- Determine final status based on VIDEO tasks only ---
	videoSuccessCount := 0
	videoFailedCount := 0
	var variantInfos []ffmpeg.VariantInfo

	for _, o := range videoOutcomes {
		if o.task.Status == "success" {
			videoSuccessCount++
			if o.variantInfo != nil {
				variantInfos = append(variantInfos, *o.variantInfo)
			}
		} else {
			videoFailedCount++
		}
	}

	switch {
	case len(videoOutcomes) == 0:
		// No video profiles at all (only preview?) — treat as success
		media.EncodingStatus = "success"
	case videoSuccessCount == len(videoOutcomes):
		// All video tasks succeeded
		media.EncodingStatus = "success"
	case videoSuccessCount > 0 && videoFailedCount > 0:
		// Partial: some succeeded, some failed — content is partially playable
		media.EncodingStatus = "partial"
	default:
		// All video tasks failed
		media.EncodingStatus = "failed"
	}

	// --- Generate master playlist if we have successful video variants ---
	if len(variantInfos) > 0 {
		masterRelPath := fmt.Sprintf("hls/%s/master.m3u8", mediaUUID)
		if _, err := GenerateMasterPlaylist(hlsBaseDir, variantInfos); err != nil {
			h.logger.Errorf("master playlist generation failed for media %d: %v", mediaID, err)
			// Don't change status — variants are still individually playable
		} else {
			media.HlsFile = masterRelPath
			h.logger.Infof("HLS master ready: media=%d → %s", mediaID, masterRelPath)
		}
	} else if videoFailedCount > 0 && videoSuccessCount == 0 {
		// Clean up empty HLS directory
		os.RemoveAll(hlsBaseDir)
	}

	// Final media update
	if _, err := h.mediaRepo.Update(procCtx, media); err != nil {
		h.logger.Errorf("failed to update media %d final status: %v", mediaID, err)
	}

	// Publish completion event
	h.publishEvent(ctx, &MediaEncodeEvent{
		MediaID: mediaID,
		Status:  media.EncodingStatus,
	})

	h.logger.Infof("media processing complete: media=%d uuid=%s status=%s (video: %d ok / %d fail)",
		mediaID, mediaUUID, media.EncodingStatus, videoSuccessCount, videoFailedCount)

	return nil
}

// waitForOutput polls for the expected output file after worker submission.
// For video profiles: checks for index.m3u8; for preview: checks for .gif file.
func (h *TranscodeHandler) waitForOutput(job TranscodeJob, task *EncodingTask, result *transcodeResult) error {
	var expectedFile string
	maxAttempts := 120 // max 10 min wait per task (5s interval)

	if IsVideoProfile(job.Profile) {
		expectedFile = filepath.Join(job.OutputDir, "index.m3u8")
	} else if IsPreviewProfile(job.Profile) {
		expectedFile = filepath.Join(filepath.Dir(job.OutputDir), "..", "previews", fmt.Sprintf("%s.gif", job.UUID))
	} else {
		return nil // unknown type, nothing to wait for
	}

	for i := 0; i < maxAttempts; i++ {
		time.Sleep(5 * time.Second)
		if _, err := os.Stat(expectedFile); err == nil {
			return nil // file exists
		}
	}

	return fmt.Errorf("timeout waiting for output: %s", expectedFile)
}

// transcodeResult holds the result of a single transcode job execution.
type transcodeResult struct {
	taskID int
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
	if event.Status == "success" || event.Status == "failed" || event.Status == "partial" {
		topic = pubsub.MediaEncodeCompletedTopic
	}

	if err := h.publisher.Publish(topic, msg); err != nil {
		h.logger.Warnf("failed to publish event to %s: %v", topic, err)
	}
}

// estimateBandwidth extracts bandwidth from profile's BentoParameters (e.g., "--video-bitrate 400k" → 400000).
// Falls back to resolution-based estimation if not parseable.
func estimateBandwidth(p *EncodeProfile) int {
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
