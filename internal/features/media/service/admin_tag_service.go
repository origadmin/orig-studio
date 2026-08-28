package service

import (
	"context"
	"strconv"

	"github.com/origadmin/runtime/errors"
	"github.com/origadmin/runtime/log"

	media "origadmin/application/origstudio/api/gen/v1/media"
	"origadmin/application/origstudio/api/gen/v1/types"
	repotypes "origadmin/application/origstudio/internal/domain/types"
	"origadmin/application/origstudio/internal/data/convpb"
	"origadmin/application/origstudio/internal/data/entity/tag"
	"origadmin/application/origstudio/internal/features/media/biz"
	"origadmin/application/origstudio/internal/features/media/dto"
)

type AdminTagService struct {
	media.UnimplementedAdminTagServiceServer
	uc  *biz.MediaUseCase
	log *log.Helper
}

func NewAdminTagService(uc *biz.MediaUseCase, logger log.Logger) *AdminTagService {
	return &AdminTagService{
		uc:  uc,
		log: log.NewHelper(log.With(logger, "module", "media.service.admin-tag")),
	}
}

func (s *AdminTagService) ListAdminTags(ctx context.Context, req *media.ListAdminTagsRequest) (*media.ListAdminTagsResponse, error) {
	page, pageSize := repotypes.NormalizePagination(int(req.Page), int(req.PageSize))

	opts := &dto.TagQueryOption{
		QueryOption: repotypes.QueryOption{
			Page:     int32(page),
			PageSize: int32(pageSize),
			Keyword:  req.Keyword,
		},
	}

	items, total, err := s.uc.ListTags(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &media.ListAdminTagsResponse{
		Total:    total,
		Items:    items,
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}

func (s *AdminTagService) GetAdminTag(ctx context.Context, req *media.GetAdminTagRequest) (*media.GetAdminTagResponse, error) {
	item, err := s.uc.GetTag(ctx, req.Id)
	if err != nil {
		return nil, errors.NotFound("TAG_NOT_FOUND", "tag not found")
	}
	return &media.GetAdminTagResponse{Tag: item}, nil
}

func (s *AdminTagService) CreateAdminTag(ctx context.Context, req *media.CreateAdminTagRequest) (*media.CreateAdminTagResponse, error) {
	t := &types.Tag{
		Title:       req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Color:       req.Color,
	}
	if req.Status != "" {
		t.Status = convpb.ConvertTagStatusToTagStatusPB(tag.Status(req.Status))
	}
	created, err := s.uc.CreateTag(ctx, t)
	if err != nil {
		return nil, err
	}
	return &media.CreateAdminTagResponse{Tag: created}, nil
}

func (s *AdminTagService) UpdateAdminTag(ctx context.Context, req *media.UpdateAdminTagRequest) (*media.UpdateAdminTagResponse, error) {
	t := &types.Tag{
		Title:       req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Color:       req.Color,
	}
	if req.Id != "" {
		id, err := strconv.ParseInt(req.Id, 10, 64)
		if err == nil {
			t.Id = id
		}
	}
	if req.Status != "" {
		t.Status = convpb.ConvertTagStatusToTagStatusPB(tag.Status(req.Status))
	}
	updated, err := s.uc.UpdateTag(ctx, t)
	if err != nil {
		return nil, err
	}
	return &media.UpdateAdminTagResponse{Tag: updated}, nil
}

// PatchAdminTag applies a partial update: fetch the current tag, merge only the
// non-empty request fields, then persist via UpdateTag.
func (s *AdminTagService) PatchAdminTag(ctx context.Context, req *media.PatchAdminTagRequest) (*media.PatchAdminTagResponse, error) {
	existing, err := s.uc.GetTag(ctx, req.Id)
	if err != nil {
		return nil, errors.NotFound("TAG_NOT_FOUND", "tag not found")
	}
	if req.Name != "" {
		existing.Title = req.Name
	}
	if req.Slug != "" {
		existing.Slug = req.Slug
	}
	if req.Description != "" {
		existing.Description = req.Description
	}
	if req.Color != "" {
		existing.Color = req.Color
	}
	if req.Status != "" {
		existing.Status = convpb.ConvertTagStatusToTagStatusPB(tag.Status(req.Status))
	}
	updated, err := s.uc.UpdateTag(ctx, existing)
	if err != nil {
		return nil, err
	}
	return &media.PatchAdminTagResponse{Tag: updated}, nil
}

func (s *AdminTagService) DeleteAdminTag(ctx context.Context, req *media.DeleteAdminTagRequest) (*media.DeleteAdminTagResponse, error) {
	err := s.uc.DeleteTag(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &media.DeleteAdminTagResponse{}, nil
}

func (s *AdminTagService) BulkUpdateTags(ctx context.Context, req *media.BulkUpdateTagsRequest) (*media.BulkUpdateTagsResponse, error) {
	count := int32(0)
	for _, idStr := range req.Ids {
		t, err := s.uc.GetTag(ctx, idStr)
		if err != nil {
			continue
		}
		_, err = s.uc.UpdateTag(ctx, t)
		if err == nil {
			count++
		}
	}
	return &media.BulkUpdateTagsResponse{UpdatedCount: count}, nil
}

func (s *AdminTagService) ImportTags(ctx context.Context, req *media.ImportTagsRequest) (*media.ImportTagsResponse, error) {
	imported := int32(0)
	skipped := int32(0)
	for _, name := range req.Names {
		t := &types.Tag{
			Title: name,
			Slug:  name,
		}
		_, err := s.uc.CreateTag(ctx, t)
		if err != nil {
			skipped++
		} else {
			imported++
		}
	}
	return &media.ImportTagsResponse{
		ImportedCount: imported,
		SkippedCount:  skipped,
	}, nil
}
