package biz

import (
	"context"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"origadmin/application/origcms/internal/data/enums"
	"origadmin/application/origcms/internal/svc-media/dto"
)

// mockMediaRepo is a mock of MediaRepo
type mockMediaRepo struct {
	mock.Mock
}

func (m *mockMediaRepo) Create(
	ctx context.Context,
	media *Media,
) (*Media, error) {
	return nil, nil
}
func (m *mockMediaRepo) Get(ctx context.Context, id int64) (*Media, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*Media), args.Error(1)
}

func (m *mockMediaRepo) List(
	ctx context.Context,
	opts ...*dto.MediaQueryOption,
) ([]*Media, int32, error) {
	return nil, 0, nil
}

func (m *mockMediaRepo) Update(
	ctx context.Context,
	media *Media,
) (*Media, error) {
	return nil, nil
}
func (m *mockMediaRepo) Delete(ctx context.Context, id int64) error { return nil }
func (m *mockMediaRepo) IncrementViewCount(ctx context.Context, id int64) (int64, error) {
	return 0, nil
}

func (m *mockMediaRepo) UpdateCommentCount(ctx context.Context, id int64, delta int) error {
	return nil
}

func (m *mockMediaRepo) UpdateLikeCount(
	ctx context.Context,
	id int64,
	delta int,
) error {
	return nil
}
func (m *mockMediaRepo) UpdateDislikeCount(ctx context.Context, id int64, delta int) error {
	return nil
}

func (m *mockMediaRepo) UpdateFavoriteCount(ctx context.Context, id int64, delta int) error {
	return nil
}
func (m *mockMediaRepo) ResetStaleProcessing(ctx context.Context) (int, error) { return 0, nil }
func (m *mockMediaRepo) CountByEncodingStatus(ctx context.Context) (*StatusCounts, error) {
	return nil, nil
}

func (m *mockMediaRepo) ListFilteredByEncodingStatus(
	ctx context.Context,
	statuses []string,
	page, pageSize int,
) ([]*Media, int, error) {
	return nil, 0, nil
}

// mockEncodingTaskRepo is a mock of EncodingTaskRepo
type mockEncodingTaskRepo struct {
	mock.Mock
}

func (m *mockEncodingTaskRepo) Create(
	ctx context.Context,
	task *EncodingTask,
) (*EncodingTask, error) {
	return nil, nil
}

func (m *mockEncodingTaskRepo) Update(
	ctx context.Context,
	task *EncodingTask,
) (*EncodingTask, error) {
	args := m.Called(ctx, task)
	return args.Get(0).(*EncodingTask), args.Error(1)
}

func (m *mockEncodingTaskRepo) Get(ctx context.Context, id int) (*EncodingTask, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*EncodingTask), args.Error(1)
}

func (m *mockEncodingTaskRepo) ListByMedia(
	ctx context.Context,
	mediaId int64,
) ([]*EncodingTask, error) {
	return nil, nil
}

func (m *mockEncodingTaskRepo) DeleteByMedia(
	ctx context.Context,
	mediaID int64,
) error {
	return nil
}

func (m *mockEncodingTaskRepo) ListFlat(
	ctx context.Context,
	status string,
	mediaId *int64,
	profileFilter string,
	chunkFilter string,
	searchQuery string,
	offset, limit int,
) ([]*EncodingTask, int, error) {
	return nil, 0, nil
}

func (m *mockEncodingTaskRepo) CountByStatus(ctx context.Context) (*StatusCounts, error) {
	return nil, nil
}

func (m *mockEncodingTaskRepo) CountByStatusWithFilter(ctx context.Context, status string, mediaId *int64, profileFilter string, chunkFilter string, searchQuery string) (*StatusCounts, error) {
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

	taskID := 123
	mediaID := int64(456)

	originalTask := &EncodingTask{
		Id:           taskID,
		MediaId:      mediaID,
		Status:       "failed",
		Progress:     45,
		ErrorMessage: "something went wrong",
	}

	media := &Media{
		Id:       mediaID,
		Url:      "uploads/test.mp4",
		MimeType: "video/mp4",
	}

	// 1. Setup expectations
	taskRepo.On("Get", ctx, taskID).Return(originalTask, nil)

	// Check that Update is called with status="pending" and progress=0
	taskRepo.On("Update", ctx, mock.MatchedBy(func(t *EncodingTask) bool {
		return t.Id == taskID && t.Status == "pending" && t.Progress == 0 && t.ErrorMessage == ""
	})).Return(originalTask, nil)

	mediaRepo.On("Get", ctx, mediaID).Return(media, nil)

	// 2. Execute
	updatedTask, err := uc.RetryTask(ctx, taskID)

	// 3. Verify
	assert.NoError(t, err)
	assert.NotNil(t, updatedTask)
	assert.Equal(t, enums.EncodingTaskStatusPending, originalTask.Status)
	assert.Equal(t, 0, originalTask.Progress)
	assert.Equal(t, "", originalTask.ErrorMessage)

	taskRepo.AssertExpectations(t)
	mediaRepo.AssertExpectations(t)
}
