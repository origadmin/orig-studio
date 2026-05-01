package service

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	types "origadmin/application/origcms/api/gen/v1/types"
	pb "origadmin/application/origcms/api/gen/v1/media"
	"origadmin/application/origcms/internal/handler"
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

func (h *CategoryHandler) RegisterRoutes(rg *gin.RouterGroup) {
	r := handler.NewGinRouterAdapter(rg)
	categories := r.Group("/categories")
	{
		// ================================
		// 1. STATIC ROUTES (NO PARAMETERS) - MUST BE FIRST
		// ================================
		categories.GET("", h.listCategories())
		categories.POST("", server.WithJWT(h.jwt, func(w http.ResponseWriter, r *http.Request) {
			c := handler.NewGinContextAdapterFromHTTP(w, r)
			var input struct {
				Name        string `json:"name"`
				Description string `json:"description"`
				Slug        string `json:"slug"`
			}
			if err := c.Bind(&input); err != nil {
				server.Fail(c.GinContext(), server.ErrBadRequest, err.Error())
				return
			}

			cat, err := h.uc.CreateCategory(r.Context(), &biz.Category{
				Name:        input.Name,
				Description: input.Description,
				Slug:        input.Slug,
			})
			if err != nil {
				server.Fail(c.GinContext(), server.ErrInternal, err.Error())
				return
			}

			server.Created(c.GinContext(), &pb.CreateCategoryResponse{
				Category: bizCategoryToProto(cat),
			})
		}))

		// ================================
		// 2. PARAMETER ROUTES (WITH :id) - MUST BE LAST
		// ================================
		categories.GET("/:id", func(w http.ResponseWriter, r *http.Request) {
			c := handler.NewGinContextAdapterFromHTTP(w, r)
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				server.Fail(c.GinContext(), server.ErrBadRequest, "invalid category id")
				return
			}
			cat, err := h.uc.GetCategory(r.Context(), id)
			if err != nil {
				server.Fail(c.GinContext(), server.ErrNotFound, "category not found")
				return
			}
			server.OK(c.GinContext(), &pb.GetCategoryResponse{
				Category: bizCategoryToProto(cat),
			})
		})

		categories.PUT("/:id", server.WithJWT(h.jwt, func(w http.ResponseWriter, r *http.Request) {
			c := handler.NewGinContextAdapterFromHTTP(w, r)
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				server.Fail(c.GinContext(), server.ErrBadRequest, "invalid category id")
				return
			}

			var input struct {
				Name        string `json:"name"`
				Description string `json:"description"`
				Slug        string `json:"slug"`
			}
			if err := c.Bind(&input); err != nil {
				server.Fail(c.GinContext(), server.ErrBadRequest, err.Error())
				return
			}

			cat, err := h.uc.UpdateCategory(r.Context(), &biz.Category{
				ID:          id,
				Name:        input.Name,
				Description: input.Description,
				Slug:        input.Slug,
			})
			if err != nil {
				server.Fail(c.GinContext(), server.ErrInternal, err.Error())
				return
			}

			server.OK(c.GinContext(), &pb.UpdateCategoryResponse{
				Category: bizCategoryToProto(cat),
			})
		}))

		categories.PATCH("/:id", server.WithJWT(h.jwt, func(w http.ResponseWriter, r *http.Request) {
			c := handler.NewGinContextAdapterFromHTTP(w, r)
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				server.Fail(c.GinContext(), server.ErrBadRequest, "invalid category id")
				return
			}

			var input biz.UpdateCategoryInput
			if err := c.Bind(&input); err != nil {
				server.Fail(c.GinContext(), server.ErrBadRequest, err.Error())
				return
			}

			cat, err := h.uc.UpdateCategoryPartial(r.Context(), id, &input)
			if err != nil {
				server.Fail(c.GinContext(), server.ErrInternal, err.Error())
				return
			}

			server.OK(c.GinContext(), &pb.UpdateCategoryResponse{
				Category: bizCategoryToProto(cat),
			})
		}))

		categories.DELETE("/:id", server.WithJWT(h.jwt, func(w http.ResponseWriter, r *http.Request) {
			c := handler.NewGinContextAdapterFromHTTP(w, r)
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				server.Fail(c.GinContext(), server.ErrBadRequest, "invalid category id")
				return
			}
			err = h.uc.DeleteCategory(r.Context(), id)
			if err != nil {
				server.Fail(c.GinContext(), server.ErrInternal, err.Error())
				return
			}
			server.OK(c.GinContext(), &pb.DeleteCategoryResponse{
				Empty: &emptypb.Empty{},
			})
		}))
	}
}

func (h *CategoryHandler) listCategories() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := handler.NewGinContextAdapterFromHTTP(w, r)
		page, _ := strconv.Atoi(c.Query("page"))
		if page == 0 {
			page = 1
		}
		limit, _ := strconv.Atoi(c.Query("page_size"))
		if limit == 0 {
			limit = 100
		}
		// Normalize pagination parameters
		page, limit = repo.NormalizeHTTPPagination(page, limit)
		items, err := h.uc.ListCategories(r.Context())
		if err != nil {
			server.Fail(c.GinContext(), server.ErrInternal, err.Error())
			return
		}

		// Convert biz categories to proto categories
		pbCategories := make([]*types.Category, len(items))
		for i, item := range items {
			pbCategories[i] = bizCategoryToProto(item)
		}

		server.OK(c.GinContext(), &pb.ListCategoriesResponse{
			Items:     pbCategories,
			Total:      int32(len(items)),
			Page:       int32(page),
			PageSize:   int32(limit),
		})
	}
}

// bizCategoryToProto converts a biz.Category to a proto types.Category.
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
