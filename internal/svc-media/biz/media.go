/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package biz

import (
	"context"

	"github.com/origadmin/runtime/log"
	"origadmin/application/origcms/api/gen/v1/types"
	"origadmin/application/origcms/internal/svc-media/dto"
)

type MediaUseCase struct {
	repo dto.MediaRepo
	log  *log.Helper
}

func NewMediaUseCase(repo dto.MediaRepo, logger log.Logger) *MediaUseCase {
	return &MediaUseCase{
		repo: repo,
		log:  log.NewHelper(log.With(logger, "module", "media.biz")),
	}
}

func (uc *MediaUseCase) GetMedia(ctx context.Context, id int64) (*types.Media, error) {
	return uc.repo.Get(ctx, id)
}

func (uc *MediaUseCase) ListMedias(ctx context.Context, opts ...*dto.MediaQueryOption) ([]*types.Media, int32, error) {
	return uc.repo.List(ctx, opts...)
}

func (uc *MediaUseCase) CreateMedia(ctx context.Context, media *types.Media) (*types.Media, error) {
	return uc.repo.Create(ctx, media)
}

func (uc *MediaUseCase) UpdateMedia(ctx context.Context, media *types.Media) (*types.Media, error) {
	return uc.repo.Update(ctx, media)
}

func (uc *MediaUseCase) DeleteMedia(ctx context.Context, id int64) error {
	return uc.repo.Delete(ctx, id)
}

func (uc *MediaUseCase) IncrementViewCount(ctx context.Context, id int64) (int64, error) {
	return uc.repo.IncrementViewCount(ctx, id)
}
