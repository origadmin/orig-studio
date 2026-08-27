/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * gRPC implementations for public CategoryService and TagService.
 */

package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "origadmin/application/origstudio/api/gen/v1/media"
	types "origadmin/application/origstudio/api/gen/v1/types"
	"origadmin/application/origstudio/internal/features/content/biz"
)

// CategoryServiceServer implements the gRPC CategoryServiceServer interface.
type CategoryServiceServer struct {
	pb.UnimplementedCategoryServiceServer
	uc  *biz.CategoryTagUseCase
	log *log.Helper
}

func NewCategoryServiceServer(uc *biz.CategoryTagUseCase, logger log.Logger) *CategoryServiceServer {
	return &CategoryServiceServer{
		uc:  uc,
		log: log.NewHelper(log.With(logger, "module", "service/category_grpc")),
	}
}

func (s *CategoryServiceServer) ListCategories(ctx context.Context, req *pb.ListCategoriesRequest) (*pb.ListCategoriesResponse, error) {
	items, err := s.uc.ListCategories(ctx)
	if err != nil {
		s.log.Errorf("ListCategories failed: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbItems := make([]*types.Category, len(items))
	for i, c := range items {
		pbItems[i] = bizCategoryToProto(c)
	}

	return &pb.ListCategoriesResponse{
		Items: pbItems,
		Total: int32(len(pbItems)),
	}, nil
}

func (s *CategoryServiceServer) GetCategory(ctx context.Context, req *pb.GetCategoryRequest) (*pb.GetCategoryResponse, error) {
	if req.Slug == "" {
		return nil, status.Error(codes.InvalidArgument, "slug is required")
	}

	item, err := s.uc.GetCategoryBySlug(ctx, req.Slug)
	if err != nil {
		return nil, status.Error(codes.NotFound, "category not found")
	}

	return &pb.GetCategoryResponse{
		Category: bizCategoryToProto(item),
	}, nil
}

func (s *CategoryServiceServer) CreateCategory(ctx context.Context, req *pb.CreateCategoryRequest) (*pb.CreateCategoryResponse, error) {
	if req.Category == nil {
		return nil, status.Error(codes.InvalidArgument, "category is required")
	}

	c := protoCategoryToBiz(req.Category)
	created, err := s.uc.CreateCategory(ctx, c)
	if err != nil {
		s.log.Errorf("CreateCategory failed: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateCategoryResponse{
		Category: bizCategoryToProto(created),
	}, nil
}

func (s *CategoryServiceServer) UpdateCategory(ctx context.Context, req *pb.UpdateCategoryRequest) (*pb.UpdateCategoryResponse, error) {
	if req.Category == nil {
		return nil, status.Error(codes.InvalidArgument, "category is required")
	}

	c := protoCategoryToBiz(req.Category)
	updated, err := s.uc.UpdateCategory(ctx, c)
	if err != nil {
		s.log.Errorf("UpdateCategory failed: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UpdateCategoryResponse{
		Category: bizCategoryToProto(updated),
	}, nil
}

func (s *CategoryServiceServer) DeleteCategory(ctx context.Context, req *pb.DeleteCategoryRequest) (*pb.DeleteCategoryResponse, error) {
	if req.Slug == "" {
		return nil, status.Error(codes.InvalidArgument, "slug is required")
	}

	item, err := s.uc.GetCategoryBySlug(ctx, req.Slug)
	if err != nil {
		return nil, status.Error(codes.NotFound, "category not found")
	}

	if err := s.uc.DeleteCategory(ctx, item.ID); err != nil {
		s.log.Errorf("DeleteCategory failed: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteCategoryResponse{}, nil
}

// TagServiceServer implements the gRPC TagServiceServer interface.
type TagServiceServer struct {
	pb.UnimplementedTagServiceServer
	uc  *biz.CategoryTagUseCase
	log *log.Helper
}

func NewTagServiceServer(uc *biz.CategoryTagUseCase, logger log.Logger) *TagServiceServer {
	return &TagServiceServer{
		uc:  uc,
		log: log.NewHelper(log.With(logger, "module", "service/tag_grpc")),
	}
}

func (s *TagServiceServer) ListTags(ctx context.Context, req *pb.ListTagsRequest) (*pb.ListTagsResponse, error) {
	page := int(req.Page)
	pageSize := int(req.PageSize)
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 20
	}

	items, total, err := s.uc.ListTags(ctx, page, pageSize)
	if err != nil {
		s.log.Errorf("ListTags failed: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbItems := make([]*types.Tag, len(items))
	for i, t := range items {
		pbItems[i] = bizTagToProto(t)
	}

	return &pb.ListTagsResponse{
		Items:    pbItems,
		Total:    int32(total),
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}

func (s *TagServiceServer) GetTag(ctx context.Context, req *pb.GetTagRequest) (*pb.GetTagResponse, error) {
	if req.Slug == "" {
		return nil, status.Error(codes.InvalidArgument, "slug is required")
	}

	item, err := s.uc.GetTagBySlug(ctx, req.Slug)
	if err != nil {
		return nil, status.Error(codes.NotFound, "tag not found")
	}

	return &pb.GetTagResponse{
		Tag: bizTagToProto(item),
	}, nil
}

func (s *TagServiceServer) CreateTag(ctx context.Context, req *pb.CreateTagRequest) (*pb.CreateTagResponse, error) {
	if req.Tag == nil {
		return nil, status.Error(codes.InvalidArgument, "tag is required")
	}

	t := protoTagToBiz(req.Tag)
	created, err := s.uc.CreateTag(ctx, t)
	if err != nil {
		s.log.Errorf("CreateTag failed: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateTagResponse{
		Tag: bizTagToProto(created),
	}, nil
}

func (s *TagServiceServer) UpdateTag(ctx context.Context, req *pb.UpdateTagRequest) (*pb.UpdateTagResponse, error) {
	if req.Tag == nil {
		return nil, status.Error(codes.InvalidArgument, "tag is required")
	}

	t := protoTagToBiz(req.Tag)
	updated, err := s.uc.UpdateTag(ctx, t)
	if err != nil {
		s.log.Errorf("UpdateTag failed: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UpdateTagResponse{
		Tag: bizTagToProto(updated),
	}, nil
}

func (s *TagServiceServer) DeleteTag(ctx context.Context, req *pb.DeleteTagRequest) (*pb.DeleteTagResponse, error) {
	if req.Slug == "" {
		return nil, status.Error(codes.InvalidArgument, "slug is required")
	}

	item, err := s.uc.GetTagBySlug(ctx, req.Slug)
	if err != nil {
		return nil, status.Error(codes.NotFound, "tag not found")
	}

	if err := s.uc.DeleteTag(ctx, item.ID); err != nil {
		s.log.Errorf("DeleteTag failed: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteTagResponse{}, nil
}

// protoCategoryToBiz converts a proto types.Category to a biz.Category.
func protoCategoryToBiz(c *types.Category) *biz.Category {
	return &biz.Category{
		ID:          int(c.Id),
		Name:        c.Name,
		Slug:        c.Slug,
		Description: c.Description,
		Status:      int(c.Status),
		ParentID:    c.ParentId,
		Sequence:    int(c.Sequence),
	}
}

// protoTagToBiz converts a proto types.Tag to a biz.Tag.
//
// BUG-180 (creation half): the public TagService previously only copied
// Title/Slug, so any tag created through POST /api/v1/tags lost its
// description and color. The admin panel posts those fields, and routing
// admin tag creation through the public endpoint (which is the only tag
// write path that survives the gateway) must preserve them.
func protoTagToBiz(t *types.Tag) *biz.Tag {
	return &biz.Tag{
		ID:          int(t.Id),
		Title:       t.Title,
		Slug:        t.Slug,
		Description: t.Description,
		Color:       t.Color,
		Status:      int(t.Status),
	}
}
