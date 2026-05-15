package biz

import (
	"context"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"origadmin/application/origstudio/api/gen/v1/types"
	"origadmin/application/origstudio/internal/data/entity"
	"origadmin/application/origstudio/internal/data/enums"
	"origadmin/application/origstudio/internal/features/media/dto"
)

// mockMediaRepo is a mock of MediaRepo
type mockMediaRepo struct {
	mock.Mock
}

func (m *mockMediaRepo) Get(ctx context.Context, id string, opts ...*dto.MediaQueryOption) (*types.Media, error) {
	args := m.Called(ctx, id, opts)
	return args.Get(0).(*types.Media), args.Error(1)
}

func (m *mockMediaRepo) GetWithEntity(ctx context.Context, id string, opts ...*dto.MediaQueryOption) (*entity.Media, *types.Media, error) {
	args := m.Called(ctx, id, opts)
	return nil, args.Get(0).(*types.Media), args.Error(1)
}

func (m *mockMediaRepo) List(ctx context.Context, opts ...*dto.MediaQueryOption) ([]*types.Media, int32, error) {
	return nil, 0, nil
}

func (m *mockMediaRepo) ListWithEntities(ctx context.Context, opts ...*dto.MediaQueryOption) ([]*entity.Media, []*types.Media, int32, error) {
	return nil, nil, 0, nil
}

func (m *mockMediaRepo) Create(ctx context.Context, media *types.Media, opts ...*dto.MediaCreateOption) (*types.Media, error) {
	return nil, nil
}

func (m *mockMediaRepo) CreateWithEntity(ctx context.Context, media *types.Media) (*entity.Media, *types.Media, error) {
	return nil, nil, nil
}

func (m *mockMediaRepo) Update(ctx context.Context, media *types.Media, opts ...*dto.MediaUpdateOption) (*types.Media, error) {
	return nil, nil
}

func (m *mockMediaRepo) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockMediaRepo) ListCategories(ctx context.Context, opts ...*dto.CategoryQueryOption) ([]*types.Category, int32, error) {
	return nil, 0, nil
}

func (m *mockMediaRepo) GetCategory(ctx context.Context, id string) (*types.Category, error) {
	return nil, nil
}

func (m *mockMediaRepo) IncrementViewCount(ctx context.Context, id string) (int64, error) {
	return 0, nil
}

func (m *mockMediaRepo) UpdateCommentCount(ctx context.Context, id string, delta int) error {
	return nil
}

func (m *mockMediaRepo) UpdateLikeCount(ctx context.Context, id string, delta int) error {
	return nil
}

func (m *mockMediaRepo) UpdateDislikeCount(ctx context.Context, id string, delta int) error {
	return nil
}

func (m *mockMediaRepo) UpdateFavoriteCount(ctx context.Context, id string, delta int) error {
	return nil
}

func (m *mockMediaRepo) GetEntityByID(ctx context.Context, id string) (*entity.Media, error) {
	return nil, nil
}

func (m *mockMediaRepo) GetEntityByShortToken(ctx context.Context, shortToken string) (*entity.Media, error) {
	return nil, nil
}

func (m *mockMediaRepo) ResetStaleProcessing(ctx context.Context) (int, error) {
	return 0, nil
}

func (m *mockMediaRepo) CountByEncodingStatus(ctx context.Context) (*dto.StatusCounts, error) {
	return nil, nil
}

func (m *mockMediaRepo) ListFilteredByEncodingStatus(ctx context.Context, statuses []string, page, pageSize int) ([]*types.Media, int, error) {
	return nil, 0, nil
}

func (m *mockMediaRepo) GetByShortToken(ctx context.Context, shortToken string) (*types.Media, error) {
	return nil, nil
}

func (m *mockMediaRepo) GetByID(ctx context.Context, id string) (*types.Media, error) {
	return m.Get(ctx, id)
}

func (m *mockMediaRepo) ResolveToID(ctx context.Context, shortToken string) (string, error) {
	return "", nil
}

func (m *mockMediaRepo) UpdateSpriteFields(ctx context.Context, mediaID string, spriteStatus string, spritePath string, vttPath string) error {
	return nil
}

func (m *mockMediaRepo) UpdateThumbnailFields(ctx context.Context, mediaID string, thumbnail string, thumbnailTime float64) error {
	return nil
}

func (m *mockMediaRepo) UpdatePreviewFilePath(ctx context.Context, mediaID string, previewFilePath string) error {
	return nil
}

func (m *mockMediaRepo) UpdateDimensions(ctx context.Context, mediaID string, width, height int) error {
	return nil
}

func (m *mockMediaRepo) ListTempMediaBefore(ctx context.Context, cutoff time.Time) ([]*types.Media, error) {
	return nil, nil
}

// mockEncodingTaskRepo is a mock of EncodingTaskRepo
type mockEncodingTaskRepo struct {
	mock.Mock
}

func (m *mockEncodingTaskRepo) Create(ctx context.Context, task *dto.EncodingTask) (*dto.EncodingTask, error) {
	return nil, nil
}

func (m *mockEncodingTaskRepo) Update(ctx context.Context, task *dto.EncodingTask) (*dto.EncodingTask, error) {
	args := m.Called(ctx, task)
	return args.Get(0).(*dto.EncodingTask), args.Error(1)
}

func (m *mockEncodingTaskRepo) Get(ctx context.Context, id string) (*dto.EncodingTask, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*dto.EncodingTask), args.Error(1)
}

func (m *mockEncodingTaskRepo) ListByMedia(ctx context.Context, mediaId string) ([]*dto.EncodingTask, error) {
	return nil, nil
}

func (m *mockEncodingTaskRepo) DeleteByMedia(ctx context.Context, mediaID string) error {
	return nil
}

func (m *mockEncodingTaskRepo) ListFlat(ctx context.Context, status string, mediaId *string, profileFilter string, profileID int, chunkFilter string, searchQuery string, offset, limit int) ([]*dto.EncodingTask, int, error) {
	return nil, 0, nil
}

func (m *mockEncodingTaskRepo) CountByStatus(ctx context.Context) (*dto.StatusCounts, error) {
	return nil, nil
}

func (m *mockEncodingTaskRepo) CountByStatusWithFilter(ctx context.Context, status string, mediaId *string, profileFilter string, profileID int, chunkFilter string, searchQuery string) (*dto.StatusCounts, error) {
	return nil, nil
}

func TestMediaUseCase_RetryTask_StatusUpdate(t *testing.T) {
	ctx := context.Background()
	mediaRepo := new(mockMediaRepo)
	taskRepo := new(mockEncodingTaskRepo)

	uc := &MediaUseCase{
		repo:         mediaRepo,
		encodingRepo: taskRepo,
		log:          log.NewHelper(log.DefaultLogger),
	}

	taskID := "123"
	mediaID := "456"

	originalTask := &dto.EncodingTask{
		Id:           taskID,
		MediaId:      mediaID,
		Status:       enums.EncodingTaskStatusFailed,
		ErrorMessage: "something went wrong",
	}

	media := &types.Media{
		Id:       mediaID,
		Url:      "uploads/test.mp4",
		MimeType: "video/mp4",
	}

	// 1. Setup expectations
	taskRepo.On("Get", ctx, taskID).Return(originalTask, nil)

	// Check that Update is called with status="pending" and progress=0
	taskRepo.On("Update", ctx, mock.MatchedBy(func(t *dto.EncodingTask) bool {
		return t.Id == taskID && t.Status == enums.EncodingTaskStatusPending && t.ErrorMessage == ""
	})).Return(originalTask, nil)

	mediaRepo.On("Get", ctx, mediaID, mock.Anything).Return(media, nil)

	// 2. Execute
	updatedTask, err := uc.RetryTask(ctx, taskID)

	// 3. Verify
	assert.NoError(t, err)
	assert.NotNil(t, updatedTask)
	assert.Equal(t, enums.EncodingTaskStatusPending, originalTask.Status)
	assert.Equal(t, "", originalTask.ErrorMessage)

	taskRepo.AssertExpectations(t)
	mediaRepo.AssertExpectations(t)
}
