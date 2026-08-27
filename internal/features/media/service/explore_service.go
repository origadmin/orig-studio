package service

import (
	"context"

	"github.com/origadmin/runtime/log"

	media "origadmin/application/origstudio/api/gen/v1/media"
	repotypes "origadmin/application/origstudio/internal/domain/types"
	"origadmin/application/origstudio/internal/features/media/biz"
	"origadmin/application/origstudio/internal/features/media/dto"
)

type ExploreService struct {
	media.UnimplementedExploreServiceServer
	uc  *biz.MediaUseCase
	log *log.Helper
}

func NewExploreService(uc *biz.MediaUseCase, logger log.Logger) *ExploreService {
	return &ExploreService{
		uc:  uc,
		log: log.NewHelper(log.With(logger, "module", "media.service.explore")),
	}
}

func (s *ExploreService) GetTrending(ctx context.Context, req *media.GetTrendingRequest) (*media.GetTrendingResponse, error) {
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	opts := &dto.MediaQueryOption{
		QueryOption: repotypes.QueryOption{
			Page:     1,
			PageSize: int32(limit),
		},
		OrderBy:    "view_count",
		Descending: true,
		Listable:   boolPtr(true),
	}

	items, _, err := s.uc.ListMedias(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &media.GetTrendingResponse{Items: items}, nil
}
