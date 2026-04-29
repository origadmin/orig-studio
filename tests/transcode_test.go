//go:build ignore

package tests

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"origadmin/application/origcms/internal/data/enums"
	"origadmin/application/origcms/internal/features/media/biz"
	"origadmin/application/origcms/internal/features/media/dal"
)

// MockTranscodeWorker implements TranscodeWorker interface for testing
type MockTranscodeWorker struct {
	jobs []biz.TranscodeJob
}

func NewMockTranscodeWorker() *MockTranscodeWorker {
	return &MockTranscodeWorker{
		jobs: make([]biz.TranscodeJob, 0),
	}
}

func (w *MockTranscodeWorker) Submit(ctx context.Context, job biz.TranscodeJob) error {
	w.jobs = append(w.jobs, job)
	// Simulate successful transcode
	go func() {
		// Update task status to processing
		task, err := job.EncodingRepo.Get(ctx, job.TaskID)
		if err == nil && task != nil {
			task.Status = enums.EncodingTaskStatusProcessing
			job.EncodingRepo.Update(ctx, task)
			if job.MediaUC != nil {
				job.MediaUC.Publish(job.MediaID, &biz.EncodingEvent{MediaId: job.MediaID, Task: task})
			}
		}

		// Simulate processing time
		time.Sleep(100 * time.Millisecond)

		// Update task status to success
		task, err = job.EncodingRepo.Get(ctx, job.TaskID)
		if err == nil && task != nil {
			task.Status = enums.EncodingTaskStatusSuccess
			if job.Profile != nil {
				if biz.IsVideoProfile(job.Profile) {
					task.OutputPath = fmt.Sprintf("hls/%s/%s/index.m3u8", job.MediaID, job.Profile.Name)
				} else if biz.IsPreviewProfile(job.Profile) {
					task.OutputPath = fmt.Sprintf("previews/%s.gif", job.MediaID)
				}
			}
			job.EncodingRepo.Update(ctx, task)
			if job.MediaUC != nil {
				job.MediaUC.Publish(job.MediaID, &biz.EncodingEvent{
					MediaId:  job.MediaID,
					Task:     task,
					Progress: 100,
					Speed:    "1.0x",
					Fps:      30,
					Time:     10,
				})
			}
		}
	}()
	return nil
}

func (w *MockTranscodeWorker) Status() biz.WorkerPoolStatus {
	return biz.WorkerPoolStatus{
		MaxWorkers:    4,
		ActiveWorkers: 0,
		PendingJobs:   0,
	}
}

func (w *MockTranscodeWorker) Shutdown(ctx context.Context) error {
	return nil
}

// MockTranscodeWorkerWithFailure implements TranscodeWorker interface for testing failures
type MockTranscodeWorkerWithFailure struct {
	failProfile string
	jobs []biz.TranscodeJob
}

func NewMockTranscodeWorkerWithFailure(failProfile string) *MockTranscodeWorkerWithFailure {
	return &MockTranscodeWorkerWithFailure{
		failProfile: failProfile,
		jobs: make([]biz.TranscodeJob, 0),
	}
}

func (w *MockTranscodeWorkerWithFailure) Submit(ctx context.Context, job biz.TranscodeJob) error {
	w.jobs = append(w.jobs, job)
	// Simulate transcode with failure for specific profile
	go func() {
		// Update task status to processing
		task, err := job.EncodingRepo.Get(ctx, job.TaskID)
		if err == nil && task != nil {
			task.Status = enums.EncodingTaskStatusProcessing
			job.EncodingRepo.Update(ctx, task)
			if job.MediaUC != nil {
				job.MediaUC.Publish(job.MediaID, &biz.EncodingEvent{MediaId: job.MediaID, Task: task})
			}
		}

		// Simulate processing time
		time.Sleep(100 * time.Millisecond)

		// Update task status
		task, err = job.EncodingRepo.Get(ctx, job.TaskID)
		if err == nil && task != nil {
			if job.Profile != nil && job.Profile.Name == w.failProfile {
				// Fail this profile
				task.Status = enums.EncodingTaskStatusFailed
				task.ErrorMessage = "Simulated transcode failure"
			} else {
				// Succeed other profiles
				task.Status = enums.EncodingTaskStatusSuccess
				if biz.IsVideoProfile(job.Profile) {
					task.OutputPath = fmt.Sprintf("hls/%s/%s/index.m3u8", job.MediaID, job.Profile.Name)
				} else if biz.IsPreviewProfile(job.Profile) {
					task.OutputPath = fmt.Sprintf("previews/%s.gif", job.MediaID)
				}
			}
			job.EncodingRepo.Update(ctx, task)
			if job.MediaUC != nil {
				job.MediaUC.Publish(job.MediaID, &biz.EncodingEvent{
					MediaId:  job.MediaID,
					Task:     task,
					Progress: 100,
					Speed:    "1.0x",
					Fps:      30,
					Time:     10,
				})
			}
		}
	}()
	return nil
}

func (w *MockTranscodeWorkerWithFailure) Status() biz.WorkerPoolStatus {
	return biz.WorkerPoolStatus{
		MaxWorkers:    4,
		ActiveWorkers: 0,
		PendingJobs:   0,
	}
}

func (w *MockTranscodeWorkerWithFailure) Shutdown(ctx context.Context) error {
	return nil
}

// TestTranscode_Success tests successful transcode scenarios
func TestTranscode_Success(t *testing.T) {
	// Setup dependencies
	mockStorage := NewMockStorage()
	mediaRepo := dal.NewInMemoryMediaRepo()
	profileRepo := dal.NewInMemoryEncodeProfileRepo()
	encodingRepo := dal.NewInMemoryEncodingTaskRepo()
	mockWorker := NewMockTranscodeWorker()
	
	// Create test profiles
	createTestProfiles(t, profileRepo)
	
	mediaUC := biz.NewMediaUseCase(
		mediaRepo,
		profileRepo,
		encodingRepo,
		nil,
		mockStorage,
		nil,
		log.NewStdLogger(),
		nil,
	)
	
	transcodeHandler := biz.NewTranscodeHandler(
		mediaUC,
		profileRepo,
		encodingRepo,
		mediaRepo,
		mockWorker,
		nil,
		log.NewStdLogger(),
		"./data/uploads",
		30*time.Minute,
		nil,
	)

	// Test 1: Successful transcode for all profiles
	t.Run("AllProfilesSuccess", func(t *testing.T) {
		// Create test media
		media := &biz.Media{
			Id:           "test-media-1",
			Title:        "Test Video",
			Url:          "test.mp4",
			Size:         1024 * 1024, // 1MB
			MimeType:     "video/mp4",
			Type:         "video",
			EncodingStatus: string(enums.MediaEncodingStatusPending),
		}
		
		// Create media with entity
		entityMedia, _, err := mediaRepo.CreateWithEntity(context.Background(), media)
		require.NoError(t, err)
		require.NotNil(t, entityMedia)
		
		// Create encode request
		req := &biz.MediaEncodeRequest{
			MediaID:     entityMedia.Id,
			MediaPath:   entityMedia.Url,
			ContentType: entityMedia.MimeType,
		}
		
		// Process media
		err = transcodeHandler.processMedia(context.Background(), req)
		require.NoError(t, err)
		
		// Wait for transcode to complete
		time.Sleep(500 * time.Millisecond)
		
		// Check media status
		updatedMedia, err := mediaRepo.Get(context.Background(), entityMedia.Id)
		require.NoError(t, err)
		assert.Equal(t, string(enums.MediaEncodingStatusSuccess), updatedMedia.EncodingStatus)
		
		// Check encoding tasks
		tasks, err := encodingRepo.ListByMedia(context.Background(), entityMedia.Id)
		require.NoError(t, err)
		assert.Greater(t, len(tasks), 0)
		
		for _, task := range tasks {
			assert.Equal(t, enums.EncodingTaskStatusSuccess, task.Status)
			assert.NotEmpty(t, task.OutputPath)
		}
	})
}

// TestTranscode_Failure tests failure scenarios
func TestTranscode_Failure(t *testing.T) {
	// Setup dependencies
	mockStorage := NewMockStorage()
	mediaRepo := dal.NewInMemoryMediaRepo()
	profileRepo := dal.NewInMemoryEncodeProfileRepo()
	encodingRepo := dal.NewInMemoryEncodingTaskRepo()
	
	// Create test profiles
	createTestProfiles(t, profileRepo)
	
	mediaUC := biz.NewMediaUseCase(
		mediaRepo,
		profileRepo,
		encodingRepo,
		nil,
		mockStorage,
		nil,
		log.NewStdLogger(),
		nil,
	)

	// Test 1: Failure in one profile
	t.Run("PartialFailure", func(t *testing.T) {
		mockWorker := NewMockTranscodeWorkerWithFailure("720p")
		
		transcodeHandler := biz.NewTranscodeHandler(
			mediaUC,
			profileRepo,
			encodingRepo,
			mediaRepo,
			mockWorker,
			nil,
			log.NewStdLogger(),
			"./data/uploads",
			30*time.Minute,
			nil,
		)
		
		// Create test media
		media := &biz.Media{
			Id:           "test-media-2",
			Title:        "Test Video",
			Url:          "test.mp4",
			Size:         1024 * 1024, // 1MB
			MimeType:     "video/mp4",
			Type:         "video",
			EncodingStatus: string(enums.MediaEncodingStatusPending),
		}
		
		// Create media with entity
		entityMedia, _, err := mediaRepo.CreateWithEntity(context.Background(), media)
		require.NoError(t, err)
		require.NotNil(t, entityMedia)
		
		// Create encode request
		req := &biz.MediaEncodeRequest{
			MediaID:     entityMedia.Id,
			MediaPath:   entityMedia.Url,
			ContentType: entityMedia.MimeType,
		}
		
		// Process media
		err = transcodeHandler.processMedia(context.Background(), req)
		require.NoError(t, err)
		
		// Wait for transcode to complete
		time.Sleep(500 * time.Millisecond)
		
		// Check media status (should be partial success)
		updatedMedia, err := mediaRepo.Get(context.Background(), entityMedia.Id)
		require.NoError(t, err)
		assert.Equal(t, "partial", updatedMedia.EncodingStatus)
		
		// Check encoding tasks
		tasks, err := encodingRepo.ListByMedia(context.Background(), entityMedia.Id)
		require.NoError(t, err)
		assert.Greater(t, len(tasks), 0)
		
		hasSuccess := false
		hasFailure := false
		
		for _, task := range tasks {
			if task.Status == enums.EncodingTaskStatusSuccess {
				hasSuccess = true
			} else if task.Status == enums.EncodingTaskStatusFailed {
				hasFailure = true
			}
		}
		
		assert.True(t, hasSuccess)
		assert.True(t, hasFailure)
	})

	// Test 2: Failure in all profiles
	t.Run("CompleteFailure", func(t *testing.T) {
		// Clear previous tasks
		encodingRepo.DeleteByMedia(context.Background(), "test-media-3")
		
		mockWorker := NewMockTranscodeWorkerWithFailure("all")
		
		transcodeHandler := biz.NewTranscodeHandler(
			mediaUC,
			profileRepo,
			encodingRepo,
			mediaRepo,
			mockWorker,
			nil,
			log.NewStdLogger(),
			"./data/uploads",
			30*time.Minute,
			nil,
		)
		
		// Create test media
		media := &biz.Media{
			Id:           "test-media-3",
			Title:        "Test Video",
			Url:          "test.mp4",
			Size:         1024 * 1024, // 1MB
			MimeType:     "video/mp4",
			Type:         "video",
			EncodingStatus: string(enums.MediaEncodingStatusPending),
		}
		
		// Create media with entity
		entityMedia, _, err := mediaRepo.CreateWithEntity(context.Background(), media)
		require.NoError(t, err)
		require.NotNil(t, entityMedia)
		
		// Create encode request
		req := &biz.MediaEncodeRequest{
			MediaID:     entityMedia.Id,
			MediaPath:   entityMedia.Url,
			ContentType: entityMedia.MimeType,
		}
		
		// Process media
		err = transcodeHandler.processMedia(context.Background(), req)
		require.NoError(t, err)
		
		// Wait for transcode to complete
		time.Sleep(500 * time.Millisecond)
		
		// Check media status (should be failed)
		updatedMedia, err := mediaRepo.Get(context.Background(), entityMedia.Id)
		require.NoError(t, err)
		assert.Equal(t, string(enums.MediaEncodingStatusFailed), updatedMedia.EncodingStatus)
		
		// Check encoding tasks
		tasks, err := encodingRepo.ListByMedia(context.Background(), entityMedia.Id)
		require.NoError(t, err)
		assert.Greater(t, len(tasks), 0)
		
		for _, task := range tasks {
			assert.Equal(t, enums.EncodingTaskStatusFailed, task.Status)
			assert.NotEmpty(t, task.ErrorMessage)
		}
	})
}

// TestTranscode_Retry tests retry scenarios
func TestTranscode_Retry(t *testing.T) {
	// Setup dependencies
	mockStorage := NewMockStorage()
	mediaRepo := dal.NewInMemoryMediaRepo()
	profileRepo := dal.NewInMemoryEncodeProfileRepo()
	encodingRepo := dal.NewInMemoryEncodingTaskRepo()
	
	// Create test profiles
	createTestProfiles(t, profileRepo)
	
	mediaUC := biz.NewMediaUseCase(
		mediaRepo,
		profileRepo,
		encodingRepo,
		nil,
		mockStorage,
		nil,
		log.NewStdLogger(),
		nil,
	)

	// Test 1: Retry failed transcode
	t.Run("RetryFailedTranscode", func(t *testing.T) {
		// First, create a media with failed encoding status
		media := &biz.Media{
			Id:           "test-media-4",
			Title:        "Test Video",
			Url:          "test.mp4",
			Size:         1024 * 1024, // 1MB
			MimeType:     "video/mp4",
			Type:         "video",
			EncodingStatus: string(enums.MediaEncodingStatusFailed),
		}
		
		// Create media with entity
		entityMedia, _, err := mediaRepo.CreateWithEntity(context.Background(), media)
		require.NoError(t, err)
		require.NotNil(t, entityMedia)
		
		// Create failed encoding task
		task := &biz.EncodingTask{
			Id:       "test-task-1",
			MediaId:  entityMedia.Id,
			ProfileId: 1, // Assuming profile with ID 1 exists
			Status:   enums.EncodingTaskStatusFailed,
			ErrorMessage: "Previous failure",
		}
		
		createdTask, err := encodingRepo.Create(context.Background(), task)
		require.NoError(t, err)
		require.NotNil(t, createdTask)
		
		// Create mock worker for retry
		mockWorker := NewMockTranscodeWorker()
		
		transcodeHandler := biz.NewTranscodeHandler(
			mediaUC,
			profileRepo,
			encodingRepo,
			mediaRepo,
			mockWorker,
			nil,
			log.NewStdLogger(),
			"./data/uploads",
			30*time.Minute,
			nil,
		)
		
		// Create encode request with specific task ID to retry
		taskID := createdTask.Id
		req := &biz.MediaEncodeRequest{
			MediaID:     entityMedia.Id,
			MediaPath:   entityMedia.Url,
			ContentType: entityMedia.MimeType,
			TaskID:      &taskID,
		}
		
		// Process media (retry)
		err = transcodeHandler.processMedia(context.Background(), req)
		require.NoError(t, err)
		
		// Wait for transcode to complete
		time.Sleep(500 * time.Millisecond)
		
		// Check task status
		updatedTask, err := encodingRepo.Get(context.Background(), taskID)
		require.NoError(t, err)
		assert.Equal(t, enums.EncodingTaskStatusSuccess, updatedTask.Status)
		assert.Empty(t, updatedTask.ErrorMessage)
	})
}

// TestTranscode_DifferentProfileTypes tests different profile types
func TestTranscode_DifferentProfileTypes(t *testing.T) {
	// Setup dependencies
	mockStorage := NewMockStorage()
	mediaRepo := dal.NewInMemoryMediaRepo()
	profileRepo := dal.NewInMemoryEncodeProfileRepo()
	encodingRepo := dal.NewInMemoryEncodingTaskRepo()
	mockWorker := NewMockTranscodeWorker()
	
	// Create test profiles including preview and frames
	createTestProfilesWithPreviewAndFrames(t, profileRepo)
	
	mediaUC := biz.NewMediaUseCase(
		mediaRepo,
		profileRepo,
		encodingRepo,
		nil,
		mockStorage,
		nil,
		log.NewStdLogger(),
		nil,
	)

	transcodeHandler := biz.NewTranscodeHandler(
		mediaUC,
		profileRepo,
		encodingRepo,
		mediaRepo,
		mockWorker,
		nil,
		log.NewStdLogger(),
		"./data/uploads",
		30*time.Minute,
		nil,
	)

	// Test: Transcode with different profile types
	t.Run("DifferentProfileTypes", func(t *testing.T) {
		// Create test media
		media := &biz.Media{
			Id:           "test-media-5",
			Title:        "Test Video",
			Url:          "test.mp4",
			Size:         1024 * 1024, // 1MB
			MimeType:     "video/mp4",
			Type:         "video",
			EncodingStatus: string(enums.MediaEncodingStatusPending),
		}
		
		// Create media with entity
		entityMedia, _, err := mediaRepo.CreateWithEntity(context.Background(), media)
		require.NoError(t, err)
		require.NotNil(t, entityMedia)
		
		// Create encode request
		req := &biz.MediaEncodeRequest{
			MediaID:     entityMedia.Id,
			MediaPath:   entityMedia.Url,
			ContentType: entityMedia.MimeType,
		}
		
		// Process media
		err = transcodeHandler.processMedia(context.Background(), req)
		require.NoError(t, err)
		
		// Wait for transcode to complete
		time.Sleep(500 * time.Millisecond)
		
		// Check media status
		updatedMedia, err := mediaRepo.Get(context.Background(), entityMedia.Id)
		require.NoError(t, err)
		assert.Equal(t, string(enums.MediaEncodingStatusSuccess), updatedMedia.EncodingStatus)
		
		// Check encoding tasks
		tasks, err := encodingRepo.ListByMedia(context.Background(), entityMedia.Id)
		require.NoError(t, err)
		assert.Greater(t, len(tasks), 0)
		
		hasVideoTask := false
		hasPreviewTask := false
		
		for _, task := range tasks {
			assert.Equal(t, enums.EncodingTaskStatusSuccess, task.Status)
			assert.NotEmpty(t, task.OutputPath)
			
			// Check task types
			if strings.Contains(task.OutputPath, "hls/") {
				hasVideoTask = true
			} else if strings.Contains(task.OutputPath, "previews/") {
				hasPreviewTask = true
			}
		}
		
		assert.True(t, hasVideoTask)
		assert.True(t, hasPreviewTask)
	})
}

// Helper functions
func createTestProfiles(t *testing.T, profileRepo biz.EncodeProfileRepo) {
	profiles := []*biz.EncodeProfile{
		{
			Id:           1,
			Name:         "720p",
			Resolution:   "720",
			Extension:    "mp4",
			VideoCodec:   "h264",
			AudioCodec:   "aac",
			VideoBitrate: "2000k",
			AudioBitrate: "128k",
			IsActive:     true,
		},
		{
			Id:           2,
			Name:         "480p",
			Resolution:   "480",
			Extension:    "mp4",
			VideoCodec:   "h264",
			AudioCodec:   "aac",
			VideoBitrate: "1000k",
			AudioBitrate: "128k",
			IsActive:     true,
		},
		{
			Id:           3,
			Name:         "360p",
			Resolution:   "360",
			Extension:    "mp4",
			VideoCodec:   "h264",
			AudioCodec:   "aac",
			VideoBitrate: "500k",
			AudioBitrate: "128k",
			IsActive:     true,
		},
	}
	
	for _, profile := range profiles {
		_, err := profileRepo.Create(context.Background(), profile)
		require.NoError(t, err)
	}
}

func createTestProfilesWithPreviewAndFrames(t *testing.T, profileRepo biz.EncodeProfileRepo) {
	profiles := []*biz.EncodeProfile{
		{
			Id:           1,
			Name:         "720p",
			Resolution:   "720",
			Extension:    "mp4",
			VideoCodec:   "h264",
			AudioCodec:   "aac",
			VideoBitrate: "2000k",
			AudioBitrate: "128k",
			IsActive:     true,
		},
		{
			Id:           2,
			Name:         "preview",
			Resolution:   "320",
			Extension:    "gif",
			IsActive:     true,
			BentoParameters: "--fps 10 --scale 320",
		},
	}
	
	for _, profile := range profiles {
		_, err := profileRepo.Create(context.Background(), profile)
		require.NoError(t, err)
	}
}
