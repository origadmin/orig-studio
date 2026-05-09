package service

import (
	"strconv"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	types "origadmin/application/origcms/api/gen/v1/types"
	pb "origadmin/application/origcms/api/gen/v1/media"
	http2 "origadmin/application/origcms/internal/helpers/http"
	ginadapter "origadmin/application/origcms/internal/helpers/http/gin"
	"origadmin/application/origcms/internal/infra/auth"
	"origadmin/application/origcms/internal/features/content/biz"
	"origadmin/application/origcms/internal/helpers/repo"
	"origadmin/application/origcms/internal/server"
)

type CategoryHandler struct {
	uc  *biz.CategoryTagUseCase
	jwt *auth.Manager
}

func NewCategoryHandler(uc *biz.CategoryTagUseCase, jwt *auth.Manager) *CategoryHandler {
	return &CategoryHandler{uc: uc, jwt: jwt}
}

func (h *CategoryHandler) RegisterRoutes(r http2.Router) {
	categories := r.Group("/categories")
	{
		categories.GET("", h.listCategories())
		categories.POST("", server.WithJWTCtx(h.jwt, h.createCategory()))

		categories.GET("/:id", h.getCategory())
		categories.PUT("/:id", server.WithJWTCtx(h.jwt, h.updateCategory()))
		categories.PATCH("/:id", server.WithJWTCtx(h.jwt, h.updateCategoryPartial()))
		categories.DELETE("/:id", server.WithJWTCtx(h.jwt, h.deleteCategory()))
	}
}

func (h *CategoryHandler) listCategories() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)
		page, _ := strconv.Atoi(gc.Query("page"))
		if page == 0 {
			page = 1
		}
		limit, _ := strconv.Atoi(gc.Query("page_size"))
		if limit == 0 {
			limit = 100
		}
		page, limit = repo.NormalizeHTTPPagination(page, limit)
		items, err := h.uc.ListCategories(ctx.Request().Context())
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}

		pbCategories := make([]*types.Category, len(items))
		for i, item := range items {
			pbCategories[i] = bizCategoryToProto(item)
		}

		http2.OK(ctx, &pb.ListCategoriesResponse{
			Items:     pbCategories,
			Total:     int32(len(items)),
			Page:      int32(page),
			PageSize:  int32(limit),
		})
		return nil
	}
}

func (h *CategoryHandler) createCategory() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)
		var input struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Slug        string `json:"slug"`
		}
		if err := gc.Bind(&input); err != nil {
			http2.Fail(ctx, server.ErrBadRequest, err.Error())
			return nil
		}

		cat, err := h.uc.CreateCategory(ctx.Request().Context(), &biz.Category{
			Name:        input.Name,
			Description: input.Description,
			Slug:        input.Slug,
		})
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}

		http2.Created(ctx, &pb.CreateCategoryResponse{
			Category: bizCategoryToProto(cat),
		})
		return nil
	}
}

func (h *CategoryHandler) getCategory() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)
		id, err := strconv.Atoi(gc.Param("id"))
		if err != nil {
			http2.Fail(ctx, server.ErrBadRequest, "invalid category id")
			return nil
		}
		cat, err := h.uc.GetCategory(ctx.Request().Context(), id)
		if err != nil {
			http2.Fail(ctx, server.ErrNotFound, "category not found")
			return nil
		}
		http2.OK(ctx, &pb.GetCategoryResponse{
			Category: bizCategoryToProto(cat),
		})
		return nil
	}
}

func (h *CategoryHandler) updateCategory() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)
		id, err := strconv.Atoi(gc.Param("id"))
		if err != nil {
			http2.Fail(ctx, server.ErrBadRequest, "invalid category id")
			return nil
		}

		var input struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Slug        string `json:"slug"`
		}
		if err := gc.Bind(&input); err != nil {
			http2.Fail(ctx, server.ErrBadRequest, err.Error())
			return nil
		}

		cat, err := h.uc.UpdateCategory(ctx.Request().Context(), &biz.Category{
			ID:          id,
			Name:        input.Name,
			Description: input.Description,
			Slug:        input.Slug,
		})
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}

		http2.OK(ctx, &pb.UpdateCategoryResponse{
			Category: bizCategoryToProto(cat),
		})
		return nil
	}
}

func (h *CategoryHandler) updateCategoryPartial() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)
		id, err := strconv.Atoi(gc.Param("id"))
		if err != nil {
			http2.Fail(ctx, server.ErrBadRequest, "invalid category id")
			return nil
		}

		var input biz.UpdateCategoryInput
		if err := gc.Bind(&input); err != nil {
			http2.Fail(ctx, server.ErrBadRequest, err.Error())
			return nil
		}

		cat, err := h.uc.UpdateCategoryPartial(ctx.Request().Context(), id, &input)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}

		http2.OK(ctx, &pb.UpdateCategoryResponse{
			Category: bizCategoryToProto(cat),
		})
		return nil
	}
}

func (h *CategoryHandler) deleteCategory() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)
		id, err := strconv.Atoi(gc.Param("id"))
		if err != nil {
			http2.Fail(ctx, server.ErrBadRequest, "invalid category id")
			return nil
		}
		err = h.uc.DeleteCategory(ctx.Request().Context(), id)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, &pb.DeleteCategoryResponse{
			Empty: &emptypb.Empty{},
		})
		return nil
	}
}

func bizCategoryToProto(c *biz.Category) *types.Category {
	pb := &types.Category{
		Id:          int64(c.ID),
		Name:        c.Name,
		Slug:        c.Slug,
		Description: c.Description,
		Status:      int32(c.Status),
		ParentId:    c.ParentID,
		Sequence:    int32(c.Sequence),
		MediaCount:  int64(c.MediaCount),
	}
	if !c.CreateTime.IsZero() {
		pb.CreateTime = timestamppb.New(c.CreateTime)
	}
	if !c.UpdateTime.IsZero() {
		pb.UpdateTime = timestamppb.New(c.UpdateTime)
	}
	return pb
}
