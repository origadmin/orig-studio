/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package biz

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"

	"origadmin/application/origstudio/internal/features/media/ffmpeg"
	"origadmin/application/origstudio/internal/features/media/dto"
)

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// TranscodeJob represents a single transcoding job for one profile of one media.
type TranscodeJob struct {
	MediaID      string
	TaskID       string
	Profile      *dto.EncodeProfile
	InputPath    string // source video file path
	OutputDir    string // final output directory (e.g., hls/{id}/{profile_name}/)
	UUID         string // media ID for path construction (deprecated, use MediaID instead)
	EncodingRepo dto.EncodingTaskRepo
	MediaUC      *MediaUseCase
	Logger       *log.Helper
}

// WorkerPoolStatus reports the current state of the worker pool.
type WorkerPoolStatus struct {
	MaxWorkers    int32
	ActiveWorkers int32
	PendingJobs   int32
}

// TranscodeWorker is the interface for submitting and managing transcode jobs.
// CE uses goroutineWorker; EE can provide an asynqWorker.
type TranscodeWorker interface {
	Submit(ctx context.Context, job TranscodeJob) error
	Status() WorkerPoolStatus
	Shutdown(ctx context.Context) error
}

// goroutineWorker implements TranscodeWorker using goroutines + semaphore.
type goroutineWorker struct {
	sem          sync.Locker
	maxWorkers   int32
	activeCount  atomic.Int32
	pendingCount atomic.Int32
	logger       *log.Helper
	jobTimeout   time.Duration
}

// NewGoroutineWorker creates a new worker pool that limits concurrent ffmpeg processes.
// maxWorkers controls how many ffmpeg processes can run simultaneously.
func NewGoroutineWorker(maxWorkers int32, logger *log.Helper) *goroutineWorker {
	return &goroutineWorker{
		sem:        newCountingSemaphore(int64(maxWorkers)),
		maxWorkers: maxWorkers,
		logger:     logger,
		jobTimeout: 2 * time.Hour,
	}
}

func (w *goroutineWorker) Submit(ctx context.Context, job TranscodeJob) error {
	w.pendingCount.Add(1)

	// Acquire semaphore — blocks if maxWorkers are already running
	w.sem.Lock()

	w.pendingCount.Add(-1)
	w.activeCount.Add(1)

	// Use job's logger if available, otherwise use worker's logger
	logger := w.logger
	if job.Logger != nil {
		logger = job.Logger
	}

	go func() {
		defer func() {
			w.activeCount.Add(-1)
			w.sem.Unlock()
		}()

		jobCtx, cancel := context.WithTimeout(ctx, w.jobTimeout)
		defer cancel()

		if err := executeTranscodeJob(jobCtx, job, logger); err != nil {
			logger.Errorf("transcode job failed: media=%s profile=%s err=%v",
				job.MediaID, job.Profile.Name, err)
		}
	}()

	return nil
}

func (w *goroutineWorker) Status() WorkerPoolStatus {
	return WorkerPoolStatus{
		MaxWorkers:    w.maxWorkers,
		ActiveWorkers: w.activeCount.Load(),
		PendingJobs:   w.pendingCount.Load(),
	}
}

func (w *goroutineWorker) Shutdown(ctx context.Context) error {
	// Drain: wait for all active workers to finish
	for w.activeCount.Load() > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}
	return nil
}

// executeTranscodeJob runs a single profile transcode job to completion.
//
// For video profiles (mp4/webm): directly outputs HLS segments via ffmpeg's HLS muxer
// into hls/{id}/{profile_name}/ with index.m3u8 + segment_XXX.ts files.
//
// For preview/GIF profile: generates animated GIF preview into previews/{id}.gif.
func executeTranscodeJob(ctx context.Context, job TranscodeJob, logger *log.Helper) error {
	profile := job.Profile

	// Update task status to processing when we actually start executing
	if job.EncodingRepo != nil {
		task, err := job.EncodingRepo.Get(ctx, job.TaskID)
		if err == nil && task != nil {
			task.Status = "processing"
			if _, err := job.EncodingRepo.Update(ctx, task); err != nil {
				logger.Warnf("failed to update task %s status: %v", job.TaskID, err)
			}
			if job.MediaUC != nil {
				job.MediaUC.Publish(job.MediaID, &EncodingEvent{MediaId: job.MediaID, Task: task})
			}
		}
	}

	var execErr error
	switch {
	case IsPreviewProfile(profile):
		// GIF preview generation
		execErr = executePreviewJob(ctx, job, logger)

	case IsVideoProfile(profile):
		// Direct HLS transcoding (no intermediate MP4)
		execErr = executeVideoHLSJob(ctx, job, logger)

	default:
		logger.Infof(
			"skipping unsupported profile %s (extension=%s)",
			profile.Name,
			profile.Extension,
		)
		return nil
	}

	// If execution failed, update task status to failed
	if execErr != nil && job.EncodingRepo != nil {
		task, err := job.EncodingRepo.Get(ctx, job.TaskID)
		if err == nil && task != nil {
			task.Status = "failed"
			task.ErrorMessage = execErr.Error()
			if _, err := job.EncodingRepo.Update(ctx, task); err != nil {
				logger.Warnf("failed to update task %s status to failed: %v", job.TaskID, err)
			}
			if job.MediaUC != nil {
				job.MediaUC.Publish(job.MediaID, &EncodingEvent{MediaId: job.MediaID, Task: task})
			}
		}
	}

	return execErr
}

// IsVideoProfile returns true if this is a standard video transcoding profile.
func IsVideoProfile(p *dto.EncodeProfile) bool {
	return p.Extension == "mp4" || p.Extension == "webm"
}

// IsPreviewProfile returns true if this is the GIF preview profile.
func IsPreviewProfile(p *dto.EncodeProfile) bool {
	return p.Extension == "gif" || strings.EqualFold(p.Name, "preview")
}

// executeVideoHLSJob runs direct HLS transcoding from source video to HLS segments.
// Output: {outputDir}/index.m3u8 + segment_001.ts, segment_002.ts, ...
func executeVideoHLSJob(ctx context.Context, job TranscodeJob, logger *log.Helper) error {
	profile := job.Profile

	if ffmpeg.IsSkipResolution(profile.Resolution) {
		logger.Infof(
			"skipping profile %s (resolution=%q is not transcodable)",
			profile.Name,
			profile.Resolution,
		)
		return fmt.Errorf("non-transcodable resolution: %s", profile.Resolution)
	}

	logger.Infof("[HLS] transcoding media=%s profile=%s res=%s codec=%s → %s",
		job.MediaID, profile.Name, profile.Resolution, profile.VideoCodec, job.OutputDir)

	duration := 0.0
	if dur, err := ffmpeg.GetVideoDuration(ctx, job.InputPath); err == nil {
		duration = dur.Seconds()
	}

	progressCb := func(progress int, frame int64, fps float64, speed string, currentTime float64) {
		if job.EncodingRepo == nil {
			return
		}

		task, err := job.EncodingRepo.Get(ctx, job.TaskID)
		if err != nil || task == nil {
			return
		}

		if progress >= 0 && progress <= 100 {
			if job.MediaUC != nil {
				// Create a copy of the task to avoid modifying the original
				taskCopy := *task
				// Publish with complete progress data
				job.MediaUC.Publish(job.MediaID, &EncodingEvent{
					MediaId:  job.MediaID,
					Task:     &taskCopy,
					Progress: progress,
					Speed:    speed,
					Fps:      fps,
					Time:     currentTime,
				})
			}
		}

		logger.Infof("[HLS] progress: media=%s profile=%s progress=%d%% frame=%d fps=%.1f speed=%s time=%.1fs",
			job.MediaID, profile.Name, progress, frame, fps, speed, currentTime)
	}

	err := ffmpeg.TranscodeToHLSWithProgress(
		ctx,
		job.InputPath,
		job.OutputDir,
		profile.Name,
		profile.Resolution,
		profile.VideoCodec,
		profile.AudioCodec,
		profile.VideoBitrate,
		profile.AudioBitrate,
		duration,
		progressCb,
	)
	if err != nil {
		return fmt.Errorf("direct HLS transcode failed for profile %s: %w", profile.Name, err)
	}

	logger.Infof(
		"[HLS] complete: media=%s profile=%s → %s/index.m3u8",
		job.MediaID,
		profile.Name,
		job.OutputDir,
	)
	return nil
}

// executePreviewJob generates an animated GIF preview from the source video.
// Output: {baseDir}/previews/{id}.gif
func executePreviewJob(ctx context.Context, job TranscodeJob, logger *log.Helper) error {
	previewDir := filepath.Join(job.OutputDir, "..", "previews") // up from hls/{id} to base dir
	if err := os.MkdirAll(previewDir, 0o755); err != nil {
		return fmt.Errorf("failed to create preview directory: %w", err)
	}

	gifPath := filepath.Join(previewDir, fmt.Sprintf("%s.gif", job.MediaID))

	scale := extractScaleParam(job.Profile.BentoParameters)

	logger.Infof("[GIF] generating preview: media=%s → %s", job.MediaID, gifPath)

	// Update task status to processing
	if job.EncodingRepo != nil {
		task, err := job.EncodingRepo.Get(ctx, job.TaskID)
		if err == nil && task != nil {
			task.Status = "processing"
			if _, err := job.EncodingRepo.Update(ctx, task); err != nil {
				logger.Warnf("failed to update task %s status: %v", job.TaskID, err)
			}
			if job.MediaUC != nil {
				job.MediaUC.Publish(job.MediaID, &EncodingEvent{MediaId: job.MediaID, Task: task})
			}
		}
	}

	err := ffmpeg.GenerateGIFPreview(ctx, job.InputPath, gifPath, scale)
	if err != nil {
		// Update task status to failed
		if job.EncodingRepo != nil {
			task, getErr := job.EncodingRepo.Get(ctx, job.TaskID)
			if getErr == nil && task != nil {
				task.Status = "failed"
				task.ErrorMessage = err.Error()
				if _, updateErr := job.EncodingRepo.Update(ctx, task); updateErr != nil {
					logger.Warnf("failed to update task %s status to failed: %v", job.TaskID, updateErr)
				}
				if job.MediaUC != nil {
					job.MediaUC.Publish(job.MediaID, &EncodingEvent{MediaId: job.MediaID, Task: task})
				}
			}
		}
		return fmt.Errorf("GIF preview failed for profile %s: %w", job.Profile.Name, err)
	}

	// Update task status to success
	if job.EncodingRepo != nil {
		task, err := job.EncodingRepo.Get(ctx, job.TaskID)
		if err == nil && task != nil {
			task.Status = "success"
			task.OutputPath = fmt.Sprintf("previews/%s.gif", job.MediaID)
			if _, err := job.EncodingRepo.Update(ctx, task); err != nil {
				logger.Warnf("failed to update task %s status to success: %v", job.TaskID, err)
			}
			if job.MediaUC != nil {
				job.MediaUC.Publish(job.MediaID, &EncodingEvent{MediaId: job.MediaID, Task: task})
			}
		}
	}

	logger.Infof("[GIF] complete: media=%s → %s", job.MediaID, gifPath)
	return nil
}

// GenerateMasterPlaylist creates the master.m3u8 at hlsBaseDir referencing all successful variants.
// Returns the relative path to master.m3u8 on success.
func GenerateMasterPlaylist(hlsBaseDir string, variants []ffmpeg.VariantInfo) (string, error) {
	if len(variants) == 0 {
		return "", fmt.Errorf("no variants provided")
	}

	if err := ffmpeg.GenerateMasterPlaylist(hlsBaseDir, variants); err != nil {
		return "", err
	}

	masterPath := filepath.Join(hlsBaseDir, "master.m3u8")
	if _, err := os.Stat(masterPath); err != nil {
		return "", fmt.Errorf("master.m3u8 not found after generation: %w", err)
	}

	return masterPath, nil
}

// GenerateThumbnail extracts a thumbnail frame from a video file.
func GenerateThumbnail(ctx context.Context, inputPath, outputDir, filename string) (string, error) {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create thumbnail directory: %w", err)
	}

	thumbPath := filepath.Join(outputDir, filename)

	// Try at 5 seconds first, fallback to 0
	err := ffmpeg.ExtractThumbnail(ctx, inputPath, thumbPath, "00:00:05")
	if err != nil {
		if err2 := ffmpeg.ExtractThumbnail(ctx, inputPath, thumbPath, "00:00:00"); err2 != nil {
			return "", fmt.Errorf(
				"thumbnail extraction failed: %w (fallback also failed: %v)",
				err,
				err2,
			)
		}
	}

	return thumbPath, nil
}

// GenerateUUID creates a new UUID string. Extracted as a function for testability.
func GenerateUUID() string {
	return uuid.New().String()
}

// extractScaleParam extracts the scale value from BentoParameters string.
// e.g., "--fps 10 --scale 320" → "320:-1" (ffmpeg scale filter format)
// Returns "320:-1" as default if not found.
func extractScaleParam(bentoParams string) string {
	fields := strings.Fields(bentoParams)
	for i, f := range fields {
		if f == "-scale" && i+1 < len(fields) {
			width := fields[i+1]
			return fmt.Sprintf("%s:-1", width)
		}
		if f == "--scale" && i+1 < len(fields) {
			width := fields[i+1]
			return fmt.Sprintf("%s:-1", width)
		}
	}
	return "320:-1"
}

// countingSemaphore is a simple sync.Locker-based semaphore.
type countingSemaphore struct {
	ch chan struct{}
}

func newCountingSemaphore(n int64) *countingSemaphore {
	return &countingSemaphore{ch: make(chan struct{}, n)}
}

func (s *countingSemaphore) Lock() {
	s.ch <- struct{}{}
}

func (s *countingSemaphore) Unlock() {
	<-s.ch
}
