/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package biz

import (
	"context"
	"os"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"

	"origadmin/application/origcms/api/gen/v1/types"
	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/svc-media/dto"
)

// MockReviewRepo 模拟审核相关的媒体仓库
type MockReviewRepo struct {
	media map[string]*Media
}

func NewMockReviewRepo() *MockReviewRepo {
	return &MockReviewRepo{
		media: make(map[string]*Media),
	}
}

func (r *MockReviewRepo) Create(ctx context.Context, media *Media, opts ...*dto.MediaCreateOption) (*Media, error) {
	r.media[media.Id] = media
	return media, nil
}

func (r *MockReviewRepo) Get(ctx context.Context, id string, opts ...*dto.MediaQueryOption) (*Media, error) {
	return r.media[id], nil
}

func (r *MockReviewRepo) List(ctx context.Context, opts ...*dto.MediaQueryOption) ([]*Media, int32, error) {
	var mediaList []*Media
	for _, media := range r.media {
		mediaList = append(mediaList, media)
	}
	return mediaList, int32(len(mediaList)), nil
}

func (r *MockReviewRepo) Update(ctx context.Context, media *Media, opts ...*dto.MediaUpdateOption) (*Media, error) {
	r.media[media.Id] = media
	return media, nil
}

func (r *MockReviewRepo) Delete(ctx context.Context, id string) error {
	delete(r.media, id)
	return nil
}

func (r *MockReviewRepo) CreateWithEntity(ctx context.Context, media *Media) (*entity.Media, *Media, error) {
	r.media[media.Id] = media
	return &entity.Media{ID: media.Id}, media, nil
}

func (r *MockReviewRepo) ListCategories(ctx context.Context, opts ...*dto.CategoryQueryOption) ([]*types.Category, int32, error) {
	return nil, 0, nil
}

func (r *MockReviewRepo) GetCategory(ctx context.Context, id string) (*types.Category, error) {
	return nil, nil
}

func (r *MockReviewRepo) IncrementViewCount(ctx context.Context, id string) (int64, error) {
	return 0, nil
}

func (r *MockReviewRepo) UpdateCommentCount(ctx context.Context, id string, delta int) error {
	return nil
}

func (r *MockReviewRepo) UpdateLikeCount(ctx context.Context, id string, delta int) error {
	return nil
}

func (r *MockReviewRepo) UpdateDislikeCount(ctx context.Context, id string, delta int) error {
	return nil
}

func (r *MockReviewRepo) UpdateFavoriteCount(ctx context.Context, id string, delta int) error {
	return nil
}

func (r *MockReviewRepo) ResetStaleProcessing(ctx context.Context) (int, error) {
	return 0, nil
}

func (r *MockReviewRepo) CountByEncodingStatus(ctx context.Context) (*dto.StatusCounts, error) {
	return &dto.StatusCounts{}, nil
}

func (r *MockReviewRepo) ListFilteredByEncodingStatus(ctx context.Context, statuses []string, page, pageSize int) ([]*Media, int, error) {
	var mediaList []*Media
	for _, media := range r.media {
		mediaList = append(mediaList, media)
	}
	return mediaList, len(mediaList), nil
}

// TestMediaUseCase_ReviewApprove 测试审核通过
func TestMediaUseCase_ReviewApprove(t *testing.T) {
	repo := NewMockReviewRepo()
	logger := log.NewStdLogger(os.Stdout)
	
	uc := NewMediaUseCase(repo, nil, nil, nil, nil, logger)
	
	ctx := context.Background()
	
	// 创建一个转码完成的媒体
	media := &Media{
		Id:             "media-123",
		Title:          "Test Video",
		EncodingStatus: "success",
		IsReviewed:     false,
		ReviewStatus:   "pending_review",
		Listable:       false,
		State:          "active",
	}
	
	// 保存媒体
	_, err := repo.Create(ctx, media)
	assert.NoError(t, err)
	
	// 审核通过
	updatedMedia, err := uc.ReviewMedia(ctx, "media-123", true, "审核通过", "user-456")
	assert.NoError(t, err)
	assert.NotNil(t, updatedMedia)
	assert.True(t, updatedMedia.IsReviewed)
	assert.Equal(t, "reviewed", updatedMedia.ReviewStatus)
	assert.True(t, updatedMedia.Listable)
}

// TestMediaUseCase_ReviewReject 测试审核驳回
func TestMediaUseCase_ReviewReject(t *testing.T) {
	repo := NewMockReviewRepo()
	logger := log.NewStdLogger(os.Stdout)
	
	uc := NewMediaUseCase(repo, nil, nil, nil, nil, logger)
	
	ctx := context.Background()
	
	// 创建一个转码完成的媒体
	media := &Media{
		Id:             "media-123",
		Title:          "Test Video",
		EncodingStatus: "success",
		IsReviewed:     false,
		ReviewStatus:   "pending_review",
		Listable:       false,
		State:          "active",
	}
	
	// 保存媒体
	_, err := repo.Create(ctx, media)
	assert.NoError(t, err)
	
	// 审核驳回
	updatedMedia, err := uc.ReviewMedia(ctx, "media-123", false, "内容不符合规范", "user-456")
	assert.NoError(t, err)
	assert.NotNil(t, updatedMedia)
	assert.False(t, updatedMedia.IsReviewed)
	assert.Equal(t, "rejected", updatedMedia.ReviewStatus)
	assert.False(t, updatedMedia.Listable)
}

// TestMediaUseCase_ShouldBeListable 测试可见性计算
func TestMediaUseCase_ShouldBeListable(t *testing.T) {
	repo := NewMockReviewRepo()
	logger := log.NewStdLogger(os.Stdout)
	
	uc := NewMediaUseCase(repo, nil, nil, nil, nil, logger)
	
	// 测试条件满足
	media1 := &Media{
		Id:             "media-123",
		Title:          "Test Video",
		EncodingStatus: "success",
		IsReviewed:     true,
		State:          "active",
	}
	
	listable1 := uc.ShouldBeListable(media1)
	assert.True(t, listable1)
	
	// 测试条件不满足 - 转码未完成
	media2 := &Media{
		Id:             "media-456",
		Title:          "Test Video 2",
		EncodingStatus: "processing",
		IsReviewed:     true,
		State:          "active",
	}
	
	listable2 := uc.ShouldBeListable(media2)
	assert.False(t, listable2)
	
	// 测试条件不满足 - 未审核
	media3 := &Media{
		Id:             "media-789",
		Title:          "Test Video 3",
		EncodingStatus: "success",
		IsReviewed:     false,
		State:          "active",
	}
	
	listable3 := uc.ShouldBeListable(media3)
	assert.False(t, listable3)
	
	// 测试条件不满足 - 不是公开状态
	media4 := &Media{
		Id:             "media-000",
		Title:          "Test Video 4",
		EncodingStatus: "success",
		IsReviewed:     true,
		State:          "draft",
	}
	
	listable4 := uc.ShouldBeListable(media4)
	assert.False(t, listable4)
}
