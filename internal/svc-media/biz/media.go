/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"origadmin/application/origcms/api/gen/v1/types" // Import the generated Media type
	"origadmin/application/origcms/internal/svc-media/dto"
)

// Media is a wrapper for types.Media for biz layer.
type Media = types.Media

// MediaRepo defines the storage operations for media.
type MediaRepo interface {
	Create(ctx context.Context, media *Media) (*Media, error)
	Get(ctx context.Context, id int64) (*Media, error)
	List(ctx context.Context, opts ...*dto.MediaQueryOption) ([]*Media, int32, error)
	Update(ctx context.Context, media *Media) (*Media, error)
	Delete(ctx context.Context, id int64) error
	IncrementViewCount(ctx context.Context, id int64) (int64, error)
}

type MediaUseCase struct {
	repo MediaRepo
	log  *log.Helper
}

func NewMediaUseCase(repo MediaRepo, logger log.Logger) *MediaUseCase {
	return &MediaUseCase{
		repo: repo,
		log:  log.NewHelper(log.With(logger, "module", "media.biz")),
	}
}

func (uc *MediaUseCase) GetMedia(ctx context.Context, id int64) (*Media, error) {
	return uc.repo.Get(ctx, id)
}

func (uc *MediaUseCase) ListMedias(ctx context.Context, opts ...*dto.MediaQueryOption) ([]*Media, int32, error) {
	return uc.repo.List(ctx, opts...)
}

func (uc *MediaUseCase) CreateMedia(ctx context.Context, media *Media) (*Media, error) {
	return uc.repo.Create(ctx, media)
}

func (uc *MediaUseCase) UpdateMedia(ctx context.Context, media *Media) (*Media, error) {
	return uc.repo.Update(ctx, media)
}

func (uc *MediaUseCase) DeleteMedia(ctx context.Context, id int64) error {
	return uc.repo.Delete(ctx, id)
}

func (uc *MediaUseCase) IncrementViewCount(ctx context.Context, id int64) (int64, error) {
	return uc.repo.IncrementViewCount(ctx, id)
}
