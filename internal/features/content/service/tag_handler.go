package service

import (
	"strconv"
	"time"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	types "origadmin/application/origstudio/api/gen/v1/types"
	pb "origadmin/application/origstudio/api/gen/v1/media"
	http2 "origadmin/application/origstudio/internal/pkg/http"
	ginadapter "origadmin/application/origstudio/internal/pkg/http/gin"
	"origadmin/application/origstudio/internal/infra/auth"
	repotypes "origadmin/application/origstudio/internal/domain/types"
	"origadmin/application/origstudio/internal/server"
	"origadmin/application/origstudio/internal/features/content/biz"
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
		tags.GET("/:id/media", h.getMediaByTag())
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
			http2.Fail(ctx, http2.ErrBadRequest, err.Error())
			return nil
		}

		t, err := h.uc.CreateTag(ctx.Request().Context(), &biz.Tag{
			Title: input.Title,
		})
		if err != nil {
			http2.Fail(ctx, http2.ErrInternal, err.Error())
			return nil
		}

		http2.Created(ctx, &pb.CreateTagResponse{
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
			http2.Fail(ctx, http2.ErrBadRequest, "invalid tag id")
			return nil
		}
		t, err := h.uc.GetTag(ctx.Request().Context(), id)
		if err != nil {
			http2.Fail(ctx, http2.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, &pb.GetTagResponse{
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
			http2.Fail(ctx, http2.ErrBadRequest, "invalid tag id")
			return nil
		}
		var input struct {
			Title string `json:"title"`
		}
		if err := gc.ShouldBindJSON(&input); err != nil {
			http2.Fail(ctx, http2.ErrBadRequest, err.Error())
			return nil
		}

		t, err := h.uc.UpdateTag(ctx.Request().Context(), &biz.Tag{
			ID:    id,
			Title: input.Title,
		})
		if err != nil {
			http2.Fail(ctx, http2.ErrInternal, err.Error())
			return nil
		}

		http2.OK(ctx, &pb.UpdateTagResponse{
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
			http2.Fail(ctx, http2.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, &pb.DeleteTagResponse{
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
		page, limit = repotypes.NormalizeHTTPPagination(page, limit)
		items, total, err := h.uc.ListTags(ctx.Request().Context(), page, limit)
		if err != nil {
			http2.Fail(ctx, http2.ErrInternal, err.Error())
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

		http2.OK(ctx, &pb.ListTagsResponse{
			Total:      int32(total),
			Items:      pbTags,
			Page:       int32(page),
			PageSize:   int32(limit),
			TotalPages: totalPages,
		})
		return nil
	}
}

func (h *TagHandler) getMediaByTag() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		http2.Fail(ctx, http2.ErrNotFound, "not implemented in UseCase")
		return nil
	}
}

func bizTagToProto(t *biz.Tag) *types.Tag {
	if t == nil {
		return nil
	}
	return &types.Tag{
		Id:         int64(t.ID),
		Name:       t.Title,
		Slug:       t.Slug,
		MediaCount: int64(t.MediaCount),
		CreateTime: timestamppb.New(time.Now()),
	}
}
