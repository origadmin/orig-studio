package service

import (
	"strconv"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "origadmin/application/origstudio/api/gen/v1/media"
	types "origadmin/application/origstudio/api/gen/v1/types"
	repotypes "origadmin/application/origstudio/internal/domain/types"
	"origadmin/application/origstudio/internal/features/content/biz"
	"origadmin/application/origstudio/internal/infra/auth"
	http2 "origadmin/application/origstudio/internal/pkg/http"
	"origadmin/application/origstudio/internal/server"
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
		tags.GET("", h.listTags())
		tags.POST("", server.WithJWTCtx(h.jwt, h.createTag()))
		tags.GET("/:slug/media", h.getMediaByTag())
		tags.GET("/:slug", h.getTag())
		tags.PUT("/:slug", server.WithJWTCtx(h.jwt, h.updateTag()))
		tags.DELETE("/:slug", server.WithJWTCtx(h.jwt, h.deleteTag()))
	}
}

func parseTagID(idStr string) (int, error) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (h *TagHandler) createTag() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		var input struct {
			Title string `json:"title"`
		}
		if err := ctx.Bind(&input); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		title := strings.TrimSpace(input.Title)
		if title == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "tag title is required")
		}

		t, err := h.uc.CreateTag(ctx.Request().Context(), &biz.Tag{
			Title: title,
		})
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.CreatedCtx(ctx, &pb.CreateTagResponse{
			Tag: bizTagToProto(t),
		})
	}
}

func (h *TagHandler) getTag() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		slug := ctx.Var("slug")
		if slug == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "tag slug is required")
		}
		t, err := h.uc.GetTagBySlug(ctx.Request().Context(), slug)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "tag not found")
		}
		return server.OKCtx(ctx, &pb.GetTagResponse{
			Tag: bizTagToProto(t),
		})
	}
}

func (h *TagHandler) updateTag() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		slug := ctx.Var("slug")
		if slug == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "tag slug is required")
		}

		existing, err := h.uc.GetTagBySlug(ctx.Request().Context(), slug)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "tag not found")
		}

		var input struct {
			Title string `json:"title"`
		}
		if err := ctx.Bind(&input); err != nil {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}

		t, err := h.uc.UpdateTag(ctx.Request().Context(), &biz.Tag{
			ID:    existing.ID,
			Title: strings.TrimSpace(input.Title),
		})
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		return server.OKCtx(ctx, &pb.UpdateTagResponse{
			Tag: bizTagToProto(t),
		})
	}
}

func (h *TagHandler) deleteTag() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		slug := ctx.Var("slug")
		if slug == "" {
			return server.FailCtx(ctx, server.ErrBadRequest, "tag slug is required")
		}

		existing, err := h.uc.GetTagBySlug(ctx.Request().Context(), slug)
		if err != nil {
			return server.FailCtx(ctx, server.ErrNotFound, "tag not found")
		}

		err = h.uc.DeleteTag(ctx.Request().Context(), existing.ID)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}
		return server.OKCtx(ctx, &pb.DeleteTagResponse{
			Empty: &emptypb.Empty{},
		})
	}
}

func (h *TagHandler) listTags() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		page, _ := strconv.Atoi(ctx.QueryVar("page"))
		if page == 0 {
			page = 1
		}
		limit, _ := strconv.Atoi(ctx.QueryVar("page_size"))
		if limit == 0 {
			limit = 100
		}
		page, limit = repotypes.NormalizeHTTPPagination(page, limit)
		items, total, err := h.uc.ListTags(ctx.Request().Context(), page, limit)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}

		pbTags := make([]*types.Tag, len(items))
		for i, item := range items {
			pbTags[i] = bizTagToProto(item)
		}

		totalPages := int32(0)
		if limit > 0 {
			totalPages = (int32(total) + int32(limit) - 1) / int32(limit)
		}

		return server.OKCtx(ctx, &pb.ListTagsResponse{
			Total:      int32(total),
			Items:      pbTags,
			Page:       int32(page),
			PageSize:   int32(limit),
			TotalPages: totalPages,
		})
	}
}

func (h *TagHandler) getMediaByTag() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		return server.FailCtx(ctx, server.ErrNotFound, "not implemented")
	}
}

func bizTagToProto(t *biz.Tag) *types.Tag {
	if t == nil {
		return nil
	}
	// BUG-180: the create time used to be stamped with time.Now() on every
	// response, so every tag looked "just created" and any client-side ordering
	// by create_time was meaningless. Emit the stored value; only fall back to
	// now when the row genuinely carries no timestamp.
	createTime := t.CreateTime
	if createTime.IsZero() {
		createTime = time.Now()
	}
	return &types.Tag{
		Id:          int64(t.ID),
		Title:       t.Title,
		Slug:        t.Slug,
		Description: t.Description,
		Color:       t.Color,
		Status:      types.TagStatus(t.Status),
		MediaCount:  int64(t.MediaCount),
		CreateTime:  timestamppb.New(createTime),
	}
}
