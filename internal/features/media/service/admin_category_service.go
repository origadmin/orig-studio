package service

import (
	"context"
	"strconv"

	"github.com/origadmin/runtime/errors"
	"github.com/origadmin/runtime/log"

	media "origadmin/application/origstudio/api/gen/v1/media"
	"origadmin/application/origstudio/api/gen/v1/types"
	repotypes "origadmin/application/origstudio/internal/domain/types"
	"origadmin/application/origstudio/internal/features/media/biz"
	"origadmin/application/origstudio/internal/features/media/dto"
)

type AdminCategoryService struct {
	media.UnimplementedAdminCategoryServiceServer
	uc  *biz.MediaUseCase
	log *log.Helper
}

func NewAdminCategoryService(uc *biz.MediaUseCase, logger log.Logger) *AdminCategoryService {
	return &AdminCategoryService{
		uc:  uc,
		log: log.NewHelper(log.With(logger, "module", "media.service.admin-category")),
	}
}

func (s *AdminCategoryService) ListAdminCategories(ctx context.Context, req *media.ListAdminCategoriesRequest) (*media.ListAdminCategoriesResponse, error) {
	page, pageSize := repotypes.NormalizePagination(int(req.Page), int(req.PageSize))

	opts := &dto.CategoryQueryOption{
		QueryOption: repotypes.QueryOption{
			Page:     int32(page),
			PageSize: int32(pageSize),
			Keyword:  req.Keyword,
		},
	}

	items, total, err := s.uc.ListCategories(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &media.ListAdminCategoriesResponse{
		Total:    total,
		Items:    items,
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}

func (s *AdminCategoryService) GetAdminCategory(ctx context.Context, req *media.GetAdminCategoryRequest) (*media.GetAdminCategoryResponse, error) {
	item, err := s.uc.GetCategory(ctx, req.Id)
	if err != nil {
		return nil, errors.NotFound("CATEGORY_NOT_FOUND", "category not found")
	}
	return &media.GetAdminCategoryResponse{Category: item}, nil
}

func (s *AdminCategoryService) CreateAdminCategory(ctx context.Context, req *media.CreateAdminCategoryRequest) (*media.CreateAdminCategoryResponse, error) {
	cat := &types.Category{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Thumbnail:   req.Thumbnail,
		Icon:        req.Icon,
		Color:       req.Color,
		Sequence:    req.Order,
	}
	if req.ParentId > 0 {
		cat.ParentId = req.ParentId
	}
	created, err := s.uc.CreateCategory(ctx, cat)
	if err != nil {
		return nil, err
	}
	return &media.CreateAdminCategoryResponse{Category: created}, nil
}

func (s *AdminCategoryService) UpdateAdminCategory(ctx context.Context, req *media.UpdateAdminCategoryRequest) (*media.UpdateAdminCategoryResponse, error) {
	cat := &types.Category{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Thumbnail:   req.Thumbnail,
		Icon:        req.Icon,
		Color:       req.Color,
		Sequence:    req.Order,
	}
	if req.ParentId > 0 {
		cat.ParentId = req.ParentId
	}
	if req.Id != "" {
		id, err := strconv.ParseInt(req.Id, 10, 64)
		if err == nil {
			cat.Id = id
		}
	}
	if req.Status > 0 {
		cat.Status = req.Status
	}
	updated, err := s.uc.UpdateCategory(ctx, cat)
	if err != nil {
		return nil, err
	}
	return &media.UpdateAdminCategoryResponse{Category: updated}, nil
}

// PatchAdminCategory applies a partial update: fetch the current category, merge
// only the non-zero request fields, then persist via UpdateCategory.
func (s *AdminCategoryService) PatchAdminCategory(ctx context.Context, req *media.PatchAdminCategoryRequest) (*media.PatchAdminCategoryResponse, error) {
	existing, err := s.uc.GetCategory(ctx, req.Id)
	if err != nil {
		return nil, errors.NotFound("CATEGORY_NOT_FOUND", "category not found")
	}
	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Slug != "" {
		existing.Slug = req.Slug
	}
	if req.Description != "" {
		existing.Description = req.Description
	}
	if req.ParentId > 0 {
		existing.ParentId = req.ParentId
	}
	if req.Status > 0 {
		existing.Status = req.Status
	}
	if req.Thumbnail != "" {
		existing.Thumbnail = req.Thumbnail
	}
	if req.Icon != "" {
		existing.Icon = req.Icon
	}
	if req.Color != "" {
		existing.Color = req.Color
	}
	if req.Order != 0 {
		existing.Sequence = req.Order
	}
	updated, err := s.uc.UpdateCategory(ctx, existing)
	if err != nil {
		return nil, err
	}
	return &media.PatchAdminCategoryResponse{Category: updated}, nil
}

func (s *AdminCategoryService) DeleteAdminCategory(ctx context.Context, req *media.DeleteAdminCategoryRequest) (*media.DeleteAdminCategoryResponse, error) {
	err := s.uc.DeleteCategory(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &media.DeleteAdminCategoryResponse{}, nil
}
