//go:build ignore

package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"origadmin/application/origstudio/internal/data/enums"
	"origadmin/application/origstudio/internal/conf"
	"origadmin/application/origstudio/internal/infra/pubsub"
	"origadmin/application/origstudio/internal/features/media/biz"
	"origadmin/application/origstudio/internal/features/media/dal"
)

// MockPublisher implements message.Publisher interface for testing
type MockPublisher struct {
	messages []*message.Message
	topics   []string
}

func NewMockPublisher() *MockPublisher {
	return &MockPublisher{
		messages: make([]*message.Message, 0),
		topics:   make([]string, 0),
	}
}

func (p *MockPublisher) Publish(topic string, msg *message.Message) error {
	p.messages = append(p.messages, msg)
	p.topics = append(p.topics, topic)
	return nil
}

func (p *MockPublisher) Close() error {
	return nil
}

// MockMediaUseCaseWithPublish implements MediaUseCase with publish tracking
type MockMediaUseCaseWithPublish struct {
	*biz.MediaUseCase
	publishEvents []*biz.EncodingEvent
}

func NewMockMediaUseCaseWithPublish(mediaUC *biz.MediaUseCase) *MockMediaUseCaseWithPublish {
	return &MockMediaUseCaseWithPublish{
		MediaUseCase:   mediaUC,
		publishEvents: make([]*biz.EncodingEvent, 0),
	}
}

func (m *MockMediaUseCaseWithPublish) Publish(mediaID string, event *biz.EncodingEvent) {
	m.publishEvents = append(m.publishEvents, event)
	m.MediaUseCase.Publish(mediaID, event)
}

// TestTranscodeCompleteCallback tests transcode completion callbacks and notifications
func TestTranscodeCompleteCallback(t *testing.T) {
	// Setup dependencies
	mockStorage := NewMockStorage()
	mediaRepo := dal.NewInMemoryMediaRepo()
	profileRepo := dal.NewInMemoryEncodeProfileRepo()
	encodingRepo := dal.NewInMemoryEncodingTaskRepo()
	mockWorker := NewMockTranscodeWorker()
	mockPublisher := NewMockPublisher()
	testPaths := conf.NewStoragePaths(t.TempDir())
	
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

	mockMediaUC := NewMockMediaUseCaseWithPublish(mediaUC)

	transcodeHandler := biz.NewTranscodeHandler(
		mockMediaUC,
		profileRepo,
		encodingRepo,
		mediaRepo,
		mockWorker,
		mockPublisher,
		log.NewStdLogger(),
		testPaths,
		30*time.Minute,
		nil,
	)

	// Test 1: Transcode complete callback with success
	t.Run("SuccessCallback", func(t *testing.T) {
		// Create test media
		media := &biz.Media{
			Id:           "test-media-callback-1",
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
		
		// Check published messages
		assert.Greater(t, len(mockPublisher.messages), 0)
		assert.Greater(t, len(mockPublisher.topics), 0)
		
		// Verify at least one message was published to completed topic
		hasCompletedMessage := false
		for i, topic := range mockPublisher.topics {
			if topic == pubsub.MediaEncodeCompletedTopic {
				hasCompletedMessage = true
				// Verify message content
				var event biz.MediaEncodeEvent
				err := json.Unmarshal(mockPublisher.messages[i].Payload, &event)
				require.NoError(t, err)
				assert.Equal(t, entityMedia.Id, event.MediaID)
				assert.Equal(t, string(enums.MediaEncodingStatusSuccess), event.Status)
				break
			}
		}
		assert.True(t, hasCompletedMessage)
		
		// Check SSE publish events
		assert.Greater(t, len(mockMediaUC.publishEvents), 0)
		
		// Verify at least one event with progress 100
		hasProgress100 := false
		for _, event := range mockMediaUC.publishEvents {
			if event.Progress == 100 {
				hasProgress100 = true
				break
			}
		}
		assert.True(t, hasProgress100)
	})

	// Test 2: Transcode complete callback with failure
	t.Run("FailureCallback", func(t *testing.T) {
		// Clear previous messages and events
		mockPublisher.messages = make([]*message.Message, 0)
		mockPublisher.topics = make([]string, 0)
		mockMediaUC.publishEvents = make([]*biz.EncodingEvent, 0)
		
		// Create mock worker with failure
		mockWorkerWithFailure := NewMockTranscodeWorkerWithFailure("all")
		
		// Create new transcode handler with failure worker
		transcodeHandlerWithFailure := biz.NewTranscodeHandler(
			mockMediaUC,
			profileRepo,
			encodingRepo,
			mediaRepo,
			mockWorkerWithFailure,
			mockPublisher,
			log.NewStdLogger(),
			testPaths,
			30*time.Minute,
			nil,
		)
		
		// Create test media
		media := &biz.Media{
			Id:           "test-media-callback-2",
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
		err = transcodeHandlerWithFailure.processMedia(context.Background(), req)
		require.NoError(t, err)
		
		// Wait for transcode to complete
		time.Sleep(500 * time.Millisecond)
		
		// Check media status
		updatedMedia, err := mediaRepo.Get(context.Background(), entityMedia.Id)
		require.NoError(t, err)
		assert.Equal(t, string(enums.MediaEncodingStatusFailed), updatedMedia.EncodingStatus)
		
		// Check published messages
		assert.Greater(t, len(mockPublisher.messages), 0)
		assert.Greater(t, len(mockPublisher.topics), 0)
		
		// Verify at least one message was published to completed topic
		hasCompletedMessage := false
		for i, topic := range mockPublisher.topics {
			if topic == pubsub.MediaEncodeCompletedTopic {
				hasCompletedMessage = true
				// Verify message content
				var event biz.MediaEncodeEvent
				err := json.Unmarshal(mockPublisher.messages[i].Payload, &event)
				require.NoError(t, err)
				assert.Equal(t, entityMedia.Id, event.MediaID)
				assert.Equal(t, string(enums.MediaEncodingStatusFailed), event.Status)
				break
			}
		}
		assert.True(t, hasCompletedMessage)
		
		// Check SSE publish events
		assert.Greater(t, len(mockMediaUC.publishEvents), 0)
	})

	// Test 3: Transcode complete callback with partial success
	t.Run("PartialSuccessCallback", func(t *testing.T) {
		// Clear previous messages and events
		mockPublisher.messages = make([]*message.Message, 0)
		mockPublisher.topics = make([]string, 0)
		mockMediaUC.publishEvents = make([]*biz.EncodingEvent, 0)
		
		// Create mock worker with partial failure
		mockWorkerWithPartialFailure := NewMockTranscodeWorkerWithFailure("720p")
		
		// Create new transcode handler with partial failure worker
		transcodeHandlerWithPartialFailure := biz.NewTranscodeHandler(
			mockMediaUC,
			profileRepo,
			encodingRepo,
			mediaRepo,
			mockWorkerWithPartialFailure,
			mockPublisher,
			log.NewStdLogger(),
			testPaths,
			30*time.Minute,
			nil,
		)
		
		// Create test media
		media := &biz.Media{
			Id:           "test-media-callback-3",
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
		err = transcodeHandlerWithPartialFailure.processMedia(context.Background(), req)
		require.NoError(t, err)
		
		// Wait for transcode to complete
		time.Sleep(500 * time.Millisecond)
		
		// Check media status
		updatedMedia, err := mediaRepo.Get(context.Background(), entityMedia.Id)
		require.NoError(t, err)
		assert.Equal(t, "partial", updatedMedia.EncodingStatus)
		
		// Check published messages
		assert.Greater(t, len(mockPublisher.messages), 0)
		assert.Greater(t, len(mockPublisher.topics), 0)
		
		// Verify at least one message was published to completed topic
		hasCompletedMessage := false
		for i, topic := range mockPublisher.topics {
			if topic == pubsub.MediaEncodeCompletedTopic {
				hasCompletedMessage = true
				// Verify message content
				var event biz.MediaEncodeEvent
				err := json.Unmarshal(mockPublisher.messages[i].Payload, &event)
				require.NoError(t, err)
				assert.Equal(t, entityMedia.Id, event.MediaID)
				assert.Equal(t, "partial", event.Status)
				break
			}
		}
		assert.True(t, hasCompletedMessage)
		
		// Check SSE publish events
		assert.Greater(t, len(mockMediaUC.publishEvents), 0)
	})
}

// TestTranscodeProgressCallback tests transcode progress callbacks
func TestTranscodeProgressCallback(t *testing.T) {
	// Setup dependencies
	mockStorage := NewMockStorage()
	mediaRepo := dal.NewInMemoryMediaRepo()
	profileRepo := dal.NewInMemoryEncodeProfileRepo()
	encodingRepo := dal.NewInMemoryEncodingTaskRepo()
	mockWorker := NewMockTranscodeWorker()
	mockPublisher := NewMockPublisher()
	testPaths := conf.NewStoragePaths(t.TempDir())
	
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

	mockMediaUC := NewMockMediaUseCaseWithPublish(mediaUC)

	transcodeHandler := biz.NewTranscodeHandler(
		mockMediaUC,
		profileRepo,
		encodingRepo,
		mediaRepo,
		mockWorker,
		mockPublisher,
		log.NewStdLogger(),
		testPaths,
		30*time.Minute,
		nil,
	)

	// Test: Transcode progress callbacks
	t.Run("ProgressCallback", func(t *testing.T) {
		// Create test media
		media := &biz.Media{
			Id:           "test-media-progress-1",
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
		
		// Check SSE publish events
		assert.Greater(t, len(mockMediaUC.publishEvents), 0)
		
		// Verify events have proper structure
		for _, event := range mockMediaUC.publishEvents {
			assert.Equal(t, entityMedia.Id, event.MediaId)
			assert.NotNil(t, event.Task)
			// Progress should be between 0 and 100
			assert.GreaterOrEqual(t, event.Progress, 0)
			assert.LessOrEqual(t, event.Progress, 100)
		}
		
		// Verify at least one event with progress 100 (completion)
		hasCompletionEvent := false
		for _, event := range mockMediaUC.publishEvents {
			if event.Progress == 100 {
				hasCompletionEvent = true
				break
			}
		}
		assert.True(t, hasCompletionEvent)
	})
}

// TestTranscodeTaskStatusCallback tests transcode task status callbacks
func TestTranscodeTaskStatusCallback(t *testing.T) {
	// Setup dependencies
	mockStorage := NewMockStorage()
	mediaRepo := dal.NewInMemoryMediaRepo()
	profileRepo := dal.NewInMemoryEncodeProfileRepo()
	encodingRepo := dal.NewInMemoryEncodingTaskRepo()
	mockWorker := NewMockTranscodeWorker()
	mockPublisher := NewMockPublisher()
	testPaths := conf.NewStoragePaths(t.TempDir())
	
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

	mockMediaUC := NewMockMediaUseCaseWithPublish(mediaUC)

	transcodeHandler := biz.NewTranscodeHandler(
		mockMediaUC,
		profileRepo,
		encodingRepo,
		mediaRepo,
		mockWorker,
		mockPublisher,
		log.NewStdLogger(),
		testPaths,
		30*time.Minute,
		nil,
	)

	// Test: Transcode task status callbacks
	t.Run("TaskStatusCallback", func(t *testing.T) {
		// Create test media
		media := &biz.Media{
			Id:           "test-media-task-status-1",
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
		
		// Check encoding tasks
		tasks, err := encodingRepo.ListByMedia(context.Background(), entityMedia.Id)
		require.NoError(t, err)
		assert.Greater(t, len(tasks), 0)
		
		// Verify all tasks have status success
		for _, task := range tasks {
			assert.Equal(t, enums.EncodingTaskStatusSuccess, task.Status)
			assert.NotEmpty(t, task.OutputPath)
		}
		
		// Check published messages for task status updates
		assert.Greater(t, len(mockPublisher.messages), 0)
		
		// Verify at least one message per task
		taskIDs := make(map[string]bool)
		for _, msg := range mockPublisher.messages {
			var event biz.MediaEncodeEvent
			err := json.Unmarshal(msg.Payload, &event)
			if err == nil && event.Task != nil {
				taskIDs[event.Task.Id] = true
			}
		}
		assert.Equal(t, len(tasks), len(taskIDs))
	})
}
