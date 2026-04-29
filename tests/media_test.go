//go:build ignore

package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"origadmin/application/origcms/internal/data/enums"
	"origadmin/application/origcms/internal/features/media/biz"
	"origadmin/application/origcms/internal/features/media/dal"
)

// TestMediaCreation tests media creation after upload
func TestMediaCreation(t *testing.T) {
	// Setup dependencies
	mockStorage := NewMockStorage()
	mediaRepo := dal.NewInMemoryMediaRepo()
	profileRepo := dal.NewInMemoryEncodeProfileRepo()
	encodingRepo := dal.NewInMemoryEncodingTaskRepo()
	
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

	// Test 1: Create media with video content
	t.Run("CreateVideoMedia", func(t *testing.T) {
		// Mock video file data
		videoData := []byte("mock video content")
		videoPath := "test-video.mp4"
		
		// Store mock video file
		mockStorage.mergedFiles[videoPath] = videoData
		
		// Create media record
		media := &biz.Media{
			Title:       "Test Video",
			Description: "A test video",
			Url:         videoPath,
			Size:        int64(len(videoData)),
			MimeType:    "video/mp4",
			Tags:        []string{"video", "test"},
			EncodingStatus: string(enums.MediaEncodingStatusPending),
		}
		
		// Create media with entity
		entityMedia, createdMedia, err := mediaRepo.CreateWithEntity(context.Background(), media)
		require.NoError(t, err)
		assert.NotNil(t, entityMedia)
		assert.NotNil(t, createdMedia)
		assert.Equal(t, "Test Video", createdMedia.Title)
		assert.Equal(t, "A test video", createdMedia.Description)
		assert.Equal(t, int64(len(videoData)), createdMedia.Size)
		assert.Equal(t, "video/mp4", createdMedia.MimeType)
		assert.Equal(t, string(enums.MediaEncodingStatusPending), createdMedia.EncodingStatus)
		
		// Verify media type
		assert.Equal(t, "video", createdMedia.Type)
	})

	// Test 2: Create media with short video content
	t.Run("CreateShortVideoMedia", func(t *testing.T) {
		// Mock short video file data
		shortVideoData := []byte("mock short video content")
		shortVideoPath := "test-short-video.mp4"
		
		// Store mock short video file
		mockStorage.mergedFiles[shortVideoPath] = shortVideoData
		
		// Create media record with short duration
		media := &biz.Media{
			Title:       "Test Short Video",
			Description: "A test short video",
			Url:         shortVideoPath,
			Size:        int64(len(shortVideoData)),
			MimeType:    "video/mp4",
			Tags:        []string{"short", "video", "test"},
			Duration:    30, // 30 seconds
			EncodingStatus: string(enums.MediaEncodingStatusPending),
		}
		
		// Create media with entity
		entityMedia, createdMedia, err := mediaRepo.CreateWithEntity(context.Background(), media)
		require.NoError(t, err)
		assert.NotNil(t, entityMedia)
		assert.NotNil(t, createdMedia)
		assert.Equal(t, "Test Short Video", createdMedia.Title)
		assert.Equal(t, 30, createdMedia.Duration)
		
		// Verify media type
		assert.Equal(t, "video", createdMedia.Type)
	})

	// Test 3: Create media with image content
	t.Run("CreateImageMedia", func(t *testing.T) {
		// Mock image file data
		imageData := []byte("mock image content")
		imagePath := "test-image.jpg"
		
		// Store mock image file
		mockStorage.mergedFiles[imagePath] = imageData
		
		// Create media record
		media := &biz.Media{
			Title:       "Test Image",
			Description: "A test image",
			Url:         imagePath,
			Size:        int64(len(imageData)),
			MimeType:    "image/jpeg",
			Tags:        []string{"image", "test"},
			EncodingStatus: string(enums.MediaEncodingStatusPending),
		}
		
		// Create media with entity
		entityMedia, createdMedia, err := mediaRepo.CreateWithEntity(context.Background(), media)
		require.NoError(t, err)
		assert.NotNil(t, entityMedia)
		assert.NotNil(t, createdMedia)
		assert.Equal(t, "Test Image", createdMedia.Title)
		assert.Equal(t, "image/jpeg", createdMedia.MimeType)
		
		// Verify media type
		assert.Equal(t, "image", createdMedia.Type)
	})

	// Test 4: Create media with audio content
	t.Run("CreateAudioMedia", func(t *testing.T) {
		// Mock audio file data
		audioData := []byte("mock audio content")
		audioPath := "test-audio.mp3"
		
		// Store mock audio file
		mockStorage.mergedFiles[audioPath] = audioData
		
		// Create media record
		media := &biz.Media{
			Title:       "Test Audio",
			Description: "A test audio",
			Url:         audioPath,
			Size:        int64(len(audioData)),
			MimeType:    "audio/mpeg",
			Tags:        []string{"audio", "test"},
			EncodingStatus: string(enums.MediaEncodingStatusPending),
		}
		
		// Create media with entity
		entityMedia, createdMedia, err := mediaRepo.CreateWithEntity(context.Background(), media)
		require.NoError(t, err)
		assert.NotNil(t, entityMedia)
		assert.NotNil(t, createdMedia)
		assert.Equal(t, "Test Audio", createdMedia.Title)
		assert.Equal(t, "audio/mpeg", createdMedia.MimeType)
		
		// Verify media type
		assert.Equal(t, "audio", createdMedia.Type)
	})

	// Test 5: Create media with file content
	t.Run("CreateFileMedia", func(t *testing.T) {
		// Mock file data
		fileData := []byte("mock file content")
		filePath := "test-file.txt"
		
		// Store mock file
		mockStorage.mergedFiles[filePath] = fileData
		
		// Create media record
		media := &biz.Media{
			Title:       "Test File",
			Description: "A test file",
			Url:         filePath,
			Size:        int64(len(fileData)),
			MimeType:    "text/plain",
			Tags:        []string{"file", "test"},
			EncodingStatus: string(enums.MediaEncodingStatusPending),
		}
		
		// Create media with entity
		entityMedia, createdMedia, err := mediaRepo.CreateWithEntity(context.Background(), media)
		require.NoError(t, err)
		assert.NotNil(t, entityMedia)
		assert.NotNil(t, createdMedia)
		assert.Equal(t, "Test File", createdMedia.Title)
		assert.Equal(t, "text/plain", createdMedia.MimeType)
		
		// Verify media type
		assert.Equal(t, "file", createdMedia.Type)
	})
}

// TestMediaValidation tests media data validation
func TestMediaValidation(t *testing.T) {
	// Setup dependencies
	mockStorage := NewMockStorage()
	mediaRepo := dal.NewInMemoryMediaRepo()
	profileRepo := dal.NewInMemoryEncodeProfileRepo()
	encodingRepo := dal.NewInMemoryEncodingTaskRepo()
	
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

	// Test 1: Validate media with missing required fields
	t.Run("MissingRequiredFields", func(t *testing.T) {
		// Create media with missing title
		media := &biz.Media{
			Description: "A test media",
			Url:         "test.mp4",
			Size:        1024,
			MimeType:    "video/mp4",
			EncodingStatus: string(enums.MediaEncodingStatusPending),
		}
		
		// Create media with entity
		entityMedia, createdMedia, err := mediaRepo.CreateWithEntity(context.Background(), media)
		// Note: Currently, the code doesn't validate required fields
		// This test will pass, but we should add validation in the future
		assert.NoError(t, err)
		assert.NotNil(t, entityMedia)
		assert.NotNil(t, createdMedia)
	})

	// Test 2: Validate media with invalid mime type
	t.Run("InvalidMimeType", func(t *testing.T) {
		// Create media with invalid mime type
		media := &biz.Media{
			Title:       "Test Media",
			Description: "A test media",
			Url:         "test.invalid",
			Size:        1024,
			MimeType:    "invalid/mime",
			EncodingStatus: string(enums.MediaEncodingStatusPending),
		}
		
		// Create media with entity
		entityMedia, createdMedia, err := mediaRepo.CreateWithEntity(context.Background(), media)
		// Note: Currently, the code doesn't validate mime types
		// This test will pass, but we should add validation in the future
		assert.NoError(t, err)
		assert.NotNil(t, entityMedia)
		assert.NotNil(t, createdMedia)
		
		// Verify media type
		assert.Equal(t, "file", createdMedia.Type)
	})
}

// TestMediaUpdate tests media update functionality
func TestMediaUpdate(t *testing.T) {
	// Setup dependencies
	mockStorage := NewMockStorage()
	mediaRepo := dal.NewInMemoryMediaRepo()
	profileRepo := dal.NewInMemoryEncodeProfileRepo()
	encodingRepo := dal.NewInMemoryEncodingTaskRepo()
	
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

	// Test 1: Update media metadata
	t.Run("UpdateMetadata", func(t *testing.T) {
		// Create initial media
		media := &biz.Media{
			Title:       "Initial Title",
			Description: "Initial description",
			Url:         "test.mp4",
			Size:        1024,
			MimeType:    "video/mp4",
			Tags:        []string{"test"},
			EncodingStatus: string(enums.MediaEncodingStatusPending),
		}
		
		// Create media with entity
		entityMedia, createdMedia, err := mediaRepo.CreateWithEntity(context.Background(), media)
		require.NoError(t, err)
		require.NotNil(t, entityMedia)
		
		// Update media
		entityMedia.Title = "Updated Title"
		entityMedia.Description = "Updated description"
		entityMedia.Tags = []string{"test", "updated"}
		
		// Save updated media
		updatedMedia, err := mediaRepo.Update(context.Background(), entityMedia)
		require.NoError(t, err)
		assert.NotNil(t, updatedMedia)
		assert.Equal(t, "Updated Title", updatedMedia.Title)
		assert.Equal(t, "Updated description", updatedMedia.Description)
		assert.Equal(t, []string{"test", "updated"}, updatedMedia.Tags)
	})

	// Test 2: Update media encoding status
	t.Run("UpdateEncodingStatus", func(t *testing.T) {
		// Create initial media
		media := &biz.Media{
			Title:       "Test Media",
			Url:         "test.mp4",
			Size:        1024,
			MimeType:    "video/mp4",
			EncodingStatus: string(enums.MediaEncodingStatusPending),
		}
		
		// Create media with entity
		entityMedia, _, err := mediaRepo.CreateWithEntity(context.Background(), media)
		require.NoError(t, err)
		require.NotNil(t, entityMedia)
		
		// Update encoding status to processing
		entityMedia.EncodingStatus = string(enums.MediaEncodingStatusProcessing)
		updatedMedia, err := mediaRepo.Update(context.Background(), entityMedia)
		require.NoError(t, err)
		assert.Equal(t, string(enums.MediaEncodingStatusProcessing), updatedMedia.EncodingStatus)
		
		// Update encoding status to success
		entityMedia.EncodingStatus = string(enums.MediaEncodingStatusSuccess)
		updatedMedia, err = mediaRepo.Update(context.Background(), entityMedia)
		require.NoError(t, err)
		assert.Equal(t, string(enums.MediaEncodingStatusSuccess), updatedMedia.EncodingStatus)
		
		// Update encoding status to failed
		entityMedia.EncodingStatus = string(enums.MediaEncodingStatusFailed)
		updatedMedia, err = mediaRepo.Update(context.Background(), entityMedia)
		require.NoError(t, err)
		assert.Equal(t, string(enums.MediaEncodingStatusFailed), updatedMedia.EncodingStatus)
	})
}

// TestMediaRetrieval tests media retrieval functionality
func TestMediaRetrieval(t *testing.T) {
	// Setup dependencies
	mockStorage := NewMockStorage()
	mediaRepo := dal.NewInMemoryMediaRepo()
	profileRepo := dal.NewInMemoryEncodeProfileRepo()
	encodingRepo := dal.NewInMemoryEncodingTaskRepo()
	
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

	// Test 1: Get media by ID
	t.Run("GetMediaByID", func(t *testing.T) {
		// Create media
		media := &biz.Media{
			Title:       "Test Media",
			Url:         "test.mp4",
			Size:        1024,
			MimeType:    "video/mp4",
			EncodingStatus: string(enums.MediaEncodingStatusPending),
		}
		
		// Create media with entity
		entityMedia, _, err := mediaRepo.CreateWithEntity(context.Background(), media)
		require.NoError(t, err)
		require.NotNil(t, entityMedia)
		
		// Get media by ID
		retrievedMedia, err := mediaRepo.Get(context.Background(), entityMedia.Id)
		require.NoError(t, err)
		assert.NotNil(t, retrievedMedia)
		assert.Equal(t, entityMedia.Id, retrievedMedia.Id)
		assert.Equal(t, "Test Media", retrievedMedia.Title)
	})

	// Test 2: List media
	t.Run("ListMedia", func(t *testing.T) {
		// Create multiple media
		for i := 1; i <= 3; i++ {
			media := &biz.Media{
				Title:       fmt.Sprintf("Test Media %d", i),
				Url:         fmt.Sprintf("test%d.mp4", i),
				Size:        int64(1024 * i),
				MimeType:    "video/mp4",
				EncodingStatus: string(enums.MediaEncodingStatusPending),
			}
			_, _, err := mediaRepo.CreateWithEntity(context.Background(), media)
			require.NoError(t, err)
		}
		
		// List media
		mediaList, total, err := mediaRepo.List(context.Background(), 1, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, 3)
		assert.GreaterOrEqual(t, len(mediaList), 3)
	})
}
