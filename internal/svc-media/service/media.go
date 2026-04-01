/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package service

import (
	"context"

	"origadmin/application/origcms/api/gen/v1/media"

	"github.com/origadmin/runtime/errors"
	"github.com/origadmin/runtime/log"
	"origadmin/application/origcms/internal/svc-media/dto"
)

type MediaService struct {
	media.UnimplementedMediaServiceServer
	repo dto.MediaRepo
	log  *log.Helper
}

func NewMediaService(repo dto.MediaRepo, logger log.Logger) *MediaService {
	return &MediaService{
		repo: repo,
		log:  log.NewHelper(log.With(logger, "module", "media.service")),
	}
}

func (s *MediaService) ListMedias(ctx context.Context, req *media.ListMediasRequest) (*media.ListMediasResponse, error) {
	opts := dto.ListMediasRequestToQueryOption(req)
	items, total, err := s.repo.List(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &media.ListMediasResponse{
		Medias:   items,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (s *MediaService) GetMedia(ctx context.Context, req *media.GetMediaRequest) (*media.GetMediaResponse, error) {
	item, err := s.repo.Get(ctx, req.Id)
	if err != nil {
		return nil, errors.NotFound("MEDIA_NOT_FOUND", "Media not found")
	}
	return &media.GetMediaResponse{Media: item}, nil
}

func (s *MediaService) CreateMedia(ctx context.Context, req *media.CreateMediaRequest) (*media.CreateMediaResponse, error) {
	item, err := s.repo.Create(ctx, req.Media)
	if err != nil {
		return nil, err
	}
	return &media.CreateMediaResponse{Media: item}, nil
}

func (s *MediaService) UpdateMedia(ctx context.Context, req *media.UpdateMediaRequest) (*media.UpdateMediaResponse, error) {
	item, err := s.repo.Update(ctx, req.Media)
	if err != nil {
		return nil, err
	}
	return &media.UpdateMediaResponse{Media: item}, nil
}

func (s *MediaService) DeleteMedia(ctx context.Context, req *media.DeleteMediaRequest) (*media.DeleteMediaResponse, error) {
	err := s.repo.Delete(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &media.DeleteMediaResponse{}, nil
}

func (s *MediaService) IncrementViewCount(ctx context.Context, req *media.IncrementViewCountRequest) (*media.IncrementViewCountResponse, error) {
	count, err := s.repo.IncrementViewCount(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &media.IncrementViewCountResponse{ViewCount: count}, nil
}
