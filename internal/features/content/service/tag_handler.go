package service

import (
	"strconv"
	"time"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	types "origadmin/application/origcms/api/gen/v1/types"
	pb "origadmin/application/origcms/api/gen/v1/media"
	http2 "origadmin/application/origcms/internal/helpers/http"
	ginadapter "origadmin/application/origcms/internal/helpers/http/gin"
	"origadmin/application/origcms/internal/infra/auth"
	"origadmin/application/origcms/internal/helpers/repo"
	"origadmin/application/origcms/internal/server"
	"origadmin/application/origcms/internal/features/content/biz"
)

type TagHandler struct {
	uc  *biz.CategoryTagUseCase
	jwt *auth.Manager
}

func NewTagHandler(uc *biz.CategoryTagUseCase, jwt *auth.Manager) *TagHandler {
	return &TagHandler{uc: uc, jwt: jwt}
}

func (h *TagHandler) RegisterRoutes(r http2.Router) {
	tags := r.Group("/tags")
	{
		// ================================
		// 1. STATIC ROUTES (NO PARAMETERS) - MUST BE FIRST
		// ================================
		tags.GET("", h.listTags())
		tags.POST("", server.WithJWTCtx(h.jwt, h.createTag()))

		// ================================
		// 2. NESTED RESOURCE ROUTES (WITH :id) - MUST BE BEFORE MAIN :id ROUTES
		// ================================
		// GET /tags/:tag_id/media — list media by tag
		tags.GET("/:id/media", h.getMediaByTag())

		// ================================
		// 3. MAIN RESOURCE PARAMETER ROUTES (WITH :id) - MUST BE LAST
		// ================================
		tags.GET("/:id", h.getTag())
		tags.PUT("/:id", server.WithJWTCtx(h.jwt, h.updateTag()))
		tags.DELETE("/:id", server.WithJWTCtx(h.jwt, h.deleteTag()))
	}
}

func (h *TagHandler) createTag() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)
		var input struct {
			Title string `json:"title"`
		}
		if err := gc.ShouldBindJSON(&input); err != nil {
			server.FailCtx(ctx, server.ErrBadRequest, err.Error())
			return nil
		}

		t, err := h.uc.CreateTag(ctx.Request().Context(), &biz.Tag{
			Title: input.Title,
		})
		if err != nil {
			server.FailCtx(ctx, server.ErrInternal, err.Error())
			return nil
		}

		server.CreatedCtx(ctx, &pb.CreateTagResponse{
			Tag: bizTagToProto(t),
		})
		return nil
	}
}

func (h *TagHandler) getTag() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)
		id, err := strconv.Atoi(gc.Param("id"))
		if err != nil {
			server.FailCtx(ctx, server.ErrBadRequest, "invalid tag id")
			return nil
		}
		t, err := h.uc.GetTag(ctx.Request().Context(), id)
		if err != nil {
			server.FailCtx(ctx, server.ErrInternal, err.Error())
			return nil
		}
		server.OKCtx(ctx, &pb.GetTagResponse{
			Tag: bizTagToProto(t),
		})
		return nil
	}
}

func (h *TagHandler) updateTag() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)
		id, err := strconv.Atoi(gc.Param("id"))
		if err != nil {
			server.FailCtx(ctx, server.ErrBadRequest, "invalid tag id")
			return nil
		}
		var input struct {
			Title string `json:"title"`
		}
		if err := gc.ShouldBindJSON(&input); err != nil {
			server.FailCtx(ctx, server.ErrBadRequest, err.Error())
			return nil
		}

		t, err := h.uc.UpdateTag(ctx.Request().Context(), &biz.Tag{
			ID:    id,
			Title: input.Title,
		})
		if err != nil {
			server.FailCtx(ctx, server.ErrInternal, err.Error())
			return nil
		}

		server.OKCtx(ctx, &pb.UpdateTagResponse{
			Tag: bizTagToProto(t),
		})
		return nil
	}
}

func (h *TagHandler) deleteTag() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)
		id, _ := strconv.Atoi(gc.Param("id"))
		err := h.uc.DeleteTag(ctx.Request().Context(), id)
		if err != nil {
			server.FailCtx(ctx, server.ErrInternal, err.Error())
			return nil
		}
		server.OKCtx(ctx, &pb.DeleteTagResponse{
			Empty: &emptypb.Empty{},
		})
		return nil
	}
}

func (h *TagHandler) listTags() http2.HandlerFunc {
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
		// Normalize pagination parameters
		page, limit = repo.NormalizeHTTPPagination(page, limit)
		items, total, err := h.uc.ListTags(ctx.Request().Context(), page, limit)
		if err != nil {
			server.FailCtx(ctx, server.ErrInternal, err.Error())
			return nil
		}

		pbTags := make([]*types.Tag, len(items))
		for i, item := range items {
			pbTags[i] = bizTagToProto(item)
		}

		totalPages := int32(0)
		if limit > 0 {
			totalPages = (int32(total) + int32(limit) - 1) / int32(limit)
		}

		server.OKCtx(ctx, &pb.ListTagsResponse{
			Total:      int32(total),
			Items:      pbTags,
			Page:       int32(page),
			PageSize:   int32(limit),
			TotalPages: totalPages,
		})
		return nil
	}
}

// getMediaByTag returns all media associated with a specific tag.
// GET /api/v1/tags/:id/media
func (h *TagHandler) getMediaByTag() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		server.FailCtx(ctx, server.ErrNotFound, "not implemented in UseCase")
		return nil
	}
}

// bizTagToProto converts a biz.Tag to a proto types.Tag.
func bizTagToProto(t *biz.Tag) *types.Tag {
	if t == nil {
		return nil
	}
	return &types.Tag{
		Id:         int64(t.ID),
		Name:       t.Title,
		Slug:       t.Slug,
		MediaCount: int64(t.MediaCount),
		CreateTime: timestamppb.New(time.Now()), // biz.Tag lacks timestamp fields
	}
}
