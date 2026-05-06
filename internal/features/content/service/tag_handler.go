package service

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	types "origadmin/application/origcms/api/gen/v1/types"
	pb "origadmin/application/origcms/api/gen/v1/media"
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

func (h *TagHandler) RegisterRoutes(rg *gin.RouterGroup) {
	r := ginadapter.NewStdRouterAdapter(rg)
	tags := r.Group("/tags")
	{
		// ================================
		// 1. STATIC ROUTES (NO PARAMETERS) - MUST BE FIRST
		// ================================
		tags.GET("", h.listTags())
		tags.POST("", server.WithJWT(h.jwt, h.createTag()))

		// ================================
		// 2. NESTED RESOURCE ROUTES (WITH :id) - MUST BE BEFORE MAIN :id ROUTES
		// ================================
		// GET /tags/:tag_id/media — list media by tag
		tags.GET("/:id/media", h.getMediaByTag())

		// ================================
		// 3. MAIN RESOURCE PARAMETER ROUTES (WITH :id) - MUST BE LAST
		// ================================
		tags.GET("/:id", h.getTag())
		tags.PUT("/:id", server.WithJWT(h.jwt, h.updateTag()))
		tags.DELETE("/:id", server.WithJWT(h.jwt, h.deleteTag()))
	}
}

func (h *TagHandler) createTag() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		var input struct {
			Title string `json:"title"`
		}
		if err := gc.ShouldBindJSON(&input); err != nil {
			server.Fail(gc, server.ErrBadRequest, err.Error())
			return
		}

		t, err := h.uc.CreateTag(r.Context(), &biz.Tag{
			Title: input.Title,
		})
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.Created(gc, &pb.CreateTagResponse{
			Tag: bizTagToProto(t),
		})
	}
}

func (h *TagHandler) getTag() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		id, err := strconv.Atoi(gc.Param("id"))
		if err != nil {
			server.Fail(gc, server.ErrBadRequest, "invalid tag id")
			return
		}
		t, err := h.uc.GetTag(r.Context(), id)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}
		server.OK(gc, &pb.GetTagResponse{
			Tag: bizTagToProto(t),
		})
	}
}

func (h *TagHandler) updateTag() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		id, err := strconv.Atoi(gc.Param("id"))
		if err != nil {
			server.Fail(gc, server.ErrBadRequest, "invalid tag id")
			return
		}
		var input struct {
			Title string `json:"title"`
		}
		if err := gc.ShouldBindJSON(&input); err != nil {
			server.Fail(gc, server.ErrBadRequest, err.Error())
			return
		}

		t, err := h.uc.UpdateTag(r.Context(), &biz.Tag{
			ID:    id,
			Title: input.Title,
		})
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		server.OK(gc, &pb.UpdateTagResponse{
			Tag: bizTagToProto(t),
		})
	}
}

func (h *TagHandler) deleteTag() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		id, _ := strconv.Atoi(gc.Param("id"))
		err := h.uc.DeleteTag(r.Context(), id)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}
		server.OK(gc, &pb.DeleteTagResponse{
			Empty: &emptypb.Empty{},
		})
	}
}

func (h *TagHandler) listTags() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
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
		items, total, err := h.uc.ListTags(r.Context(), page, limit)
		if err != nil {
			server.Fail(gc, server.ErrInternal, err.Error())
			return
		}

		pbTags := make([]*types.Tag, len(items))
		for i, item := range items {
			pbTags[i] = bizTagToProto(item)
		}

		totalPages := int32(0)
		if limit > 0 {
			totalPages = (int32(total) + int32(limit) - 1) / int32(limit)
		}

		server.OK(gc, &pb.ListTagsResponse{
			Total:      int32(total),
			Items:      pbTags,
			Page:       int32(page),
			PageSize:   int32(limit),
			TotalPages: totalPages,
		})
	}
}

// getMediaByTag returns all media associated with a specific tag.
// GET /api/v1/tags/:id/media
func (h *TagHandler) getMediaByTag() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gc := ginadapter.GetGinContext(r)
		server.Fail(gc, server.ErrNotFound, "not implemented in UseCase")
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
