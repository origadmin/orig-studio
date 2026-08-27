package service

import (
	"strconv"
	"strings"

	"google.golang.org/protobuf/types/known/timestamppb"

	pb "origadmin/application/origstudio/api/gen/v1/media"
	types "origadmin/application/origstudio/api/gen/v1/types"
	"origadmin/application/origstudio/internal/features/content/biz"
	"origadmin/application/origstudio/internal/infra/auth"
	http2 "origadmin/application/origstudio/internal/pkg/http"
	"origadmin/application/origstudio/internal/server"
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

func parseCategoryID(idStr string) (int, error) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (h *CategoryHandler) listCategories() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		items, err := h.uc.ListActiveCategories(ctx.Request().Context())
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		pbCategories := make([]*types.Category, len(items))
		for i, item := range items {
			pbCategories[i] = bizCategoryToProto(item)
		}

		return server.OKCtx(ctx, &pb.ListCategoriesResponse{
			Items:    pbCategories,
			Total:    int32(len(pbCategories)),
			Page:     1,
			PageSize: int32(len(pbCategories)),
		})
	}
}

func (h *CategoryHandler) createCategory() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		var input struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Slug        string `json:"slug"`
			ParentID    *int64 `json:"parent_id"`
			Order       *int   `json:"order"`
			Status      *int   `json:"status"`
		}
		if err := ctx.Bind(&input); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		name := strings.TrimSpace(input.Name)
		if name == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "category name is required")
		}

		c := &biz.Category{
			Name:        name,
			Description: strings.TrimSpace(input.Description),
			Slug:        strings.TrimSpace(input.Slug),
			Status:      1,
			Sequence:    0,
		}
		if input.Status != nil {
			c.Status = *input.Status
		}
		if input.ParentID != nil {
			c.ParentID = *input.ParentID
		}
		if input.Order != nil {
			c.Sequence = *input.Order
		}

		cat, err := h.uc.CreateCategory(ctx.Request().Context(), c)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.CreatedCtx(ctx, &pb.CreateCategoryResponse{
			Category: bizCategoryToProto(cat),
		})
	}
}

func (h *CategoryHandler) getCategory() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		idStr := ctx.Var("id")
		id, err := parseCategoryID(idStr)
		if err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, "invalid category id")
		}
		cat, err := h.uc.GetCategory(ctx.Request().Context(), id)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "category not found")
		}
		return server.OKCtx(ctx, &pb.GetCategoryResponse{
			Category: bizCategoryToProto(cat),
		})
	}
}

func (h *CategoryHandler) updateCategory() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		idStr := ctx.Var("id")
		id, err := parseCategoryID(idStr)
		if err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, "invalid category id")
		}

		existing, err := h.uc.GetCategory(ctx.Request().Context(), id)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "category not found")
		}

		var input struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Slug        string `json:"slug"`
			ParentID    *int64 `json:"parent_id"`
			Order       *int   `json:"order"`
			Status      *int   `json:"status"`
		}
		if err := ctx.Bind(&input); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		c := &biz.Category{
			ID:          existing.ID,
			Name:        strings.TrimSpace(input.Name),
			Description: strings.TrimSpace(input.Description),
			Slug:        strings.TrimSpace(input.Slug),
			Status:      existing.Status,
			ParentID:    existing.ParentID,
			Sequence:    existing.Sequence,
		}
		if c.Name == "" {
			c.Name = existing.Name
		}
		if input.Status != nil {
			c.Status = *input.Status
		}
		if input.ParentID != nil {
			c.ParentID = *input.ParentID
		}
		if input.Order != nil {
			c.Sequence = *input.Order
		}

		cat, err := h.uc.UpdateCategory(ctx.Request().Context(), c)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, &pb.UpdateCategoryResponse{
			Category: bizCategoryToProto(cat),
		})
	}
}

func (h *CategoryHandler) updateCategoryPartial() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		idStr := ctx.Var("id")
		id, err := parseCategoryID(idStr)
		if err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, "invalid category id")
		}

		var input biz.UpdateCategoryInput
		if err := ctx.Bind(&input); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		cat, err := h.uc.UpdateCategoryPartial(ctx.Request().Context(), id, &input)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, &pb.UpdateCategoryResponse{
			Category: bizCategoryToProto(cat),
		})
	}
}

func (h *CategoryHandler) deleteCategory() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		idStr := ctx.Var("id")
		id, err := parseCategoryID(idStr)
		if err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, "invalid category id")
		}

		err = h.uc.DeleteCategory(ctx.Request().Context(), id)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}
		return server.OKCtx(ctx, &pb.DeleteCategoryResponse{})
	}
}

func bizCategoryToProto(c *biz.Category) *types.Category {
	pbCat := &types.Category{
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
		pbCat.CreateTime = timestamppb.New(c.CreateTime)
	}
	if !c.UpdateTime.IsZero() {
		pbCat.UpdateTime = timestamppb.New(c.UpdateTime)
	}
	return pbCat
}
