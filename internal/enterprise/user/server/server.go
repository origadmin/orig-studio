package server

import (
	"crypto/rand"
	"math/big"
	"net/http"
	"strconv"

	"github.com/google/wire"

	pb "origadmin/application/origstudio/api/gen/v1/user"
	types "origadmin/application/origstudio/api/gen/v1/types"
	tenantservice "origadmin/application/origstudio/internal/enterprise/user/tenant/service"
	userbiz "origadmin/application/origstudio/internal/features/user/biz"
	userdto "origadmin/application/origstudio/internal/features/user/dto"
	"origadmin/application/origstudio/internal/infra/auth"
	repotypes "origadmin/application/origstudio/internal/domain/types"
	http2 "origadmin/application/origstudio/internal/pkg/http"
	std "origadmin/application/origstudio/internal/pkg/http/std"
	"origadmin/application/origstudio/internal/server"
)

var ProviderSet = wire.NewSet(NewEnterpriseUserServer)

type EnterpriseUserServer struct {
	tenantHandler *tenantservice.Handler
	userUC        *userbiz.UserUseCase
	jwt           *auth.Manager
}

func NewEnterpriseUserServer(
	tenantHandler *tenantservice.Handler,
	userUC *userbiz.UserUseCase,
	jwt *auth.Manager,
) *EnterpriseUserServer {
	return &EnterpriseUserServer{
		tenantHandler: tenantHandler,
		userUC:        userUC,
		jwt:           jwt,
	}
}

func (s *EnterpriseUserServer) RegisterRoutes(r http2.Router) {
	s.tenantHandler.RegisterRoutes(r)
	s.registerAdminUserRoutes(r)
}

func (s *EnterpriseUserServer) registerAdminUserRoutes(r http2.Router) {
	adminUsers := r.Group("/admin/users")
	adminUsers.Use(server.AdminMiddlewareCtx(s.jwt))
	{
		adminUsers.GET("", s.adminListUsers())
		adminUsers.POST("", s.adminCreateUser())
		adminUsers.GET("/:id", s.adminGetUser())
		adminUsers.PUT("/:id", s.adminUpdateUser())
		adminUsers.DELETE("/:id", s.adminDeleteUser())
		adminUsers.PATCH("/:id/status", s.adminUpdateUserStatus())
		adminUsers.PATCH("/:id/role", s.adminUpdateUserRole())
	}
}

func (s *EnterpriseUserServer) HTTPHandler() http.Handler {
	router := std.NewRouter()
	apiV1 := router.Group("/api/v1")
	s.RegisterRoutes(apiV1)
	return router
}

func (s *EnterpriseUserServer) adminListUsers() http2.HandlerFunc {
	return func(ctx http2.Context) error {
			page, _ := strconv.Atoi(ctx.QueryVarDefault("page", "1"))
		pageSize, _ := strconv.Atoi(ctx.QueryVarDefault("page_size", "20"))
		page, pageSize = repotypes.NormalizeHTTPPagination(page, pageSize)

		keyword := ctx.QueryVar("keyword")

		opts := &userdto.UserQueryOption{
			QueryOption: repotypes.QueryOption{
				Page:     int32(page),
				PageSize: int32(pageSize),
				Keyword:  keyword,
			},
		}

		if role := ctx.QueryVar("role"); role != "" && role != "all" {
			opts.Role = role
		}

		if statusStr := ctx.QueryVar("status"); statusStr != "" && statusStr != "all" {
			statusMap := map[string]int32{
				"pending":   1,
				"active":    2,
				"inactive":  3,
				"suspended": 4,
				"rejected":  5,
			}
			if s, ok := statusMap[statusStr]; ok {
				opts.Status = &s
			}
		}

		users, total, err := s.userUC.ListUsers(ctx.Request().Context(), opts)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}

		resp := &pb.ListUsersResponse{
			Items:    users,
			Total:    total,
			Page:     int32(page),
			PageSize: int32(pageSize),
		}
		http2.OK(ctx, resp)
		return nil
	}
}

func (s *EnterpriseUserServer) adminGetUser() http2.HandlerFunc {
	return func(ctx http2.Context) error {
			id := ctx.Var("id")
		if id == "" {
			http2.Fail(ctx, server.ErrBadRequest, "user id is required")
			return nil
		}

		u, err := s.userUC.GetUser(ctx.Request().Context(), id)
		if err != nil {
			http2.Fail(ctx, server.ErrNotFound, "user not found")
			return nil
		}
		http2.OK(ctx, &pb.GetUserResponse{User: u})
		return nil
	}
}

func (s *EnterpriseUserServer) adminCreateUser() http2.HandlerFunc {
	return func(ctx http2.Context) error {
			var input struct {
			Username string `json:"username" binding:"required"`
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password"`
			Nickname string `json:"nickname"`
			Name     string `json:"name"`
			Role     string `json:"role"`
		}
		if err := ctx.BindJSON(&input); err != nil {
			http2.Fail(ctx, server.ErrBadRequest, err.Error())
			return nil
		}

		if input.Password != "" && len(input.Password) < 6 {
			http2.Fail(ctx, server.ErrBadRequest, "password must be at least 6 characters")
			return nil
		}

		var hashedPassword string
		var err error
		if input.Password != "" {
			hashedPassword, err = s.userUC.HashPassword(input.Password)
			if err != nil {
				http2.Fail(ctx, server.ErrInternal, "failed to hash password")
				return nil
			}
		} else {
			randomPwd := generateRandomPassword(12)
			hashedPassword, err = s.userUC.HashPassword(randomPwd)
			if err != nil {
				http2.Fail(ctx, server.ErrInternal, "failed to hash password")
				return nil
			}
		}

		user := &types.User{
			Username: input.Username,
			Email:    input.Email,
			Nickname: input.Nickname,
			Name:     input.Name,
		}

		created, err := s.userUC.CreateUser(ctx.Request().Context(), user, hashedPassword)
		if err != nil {
			http2.Fail(ctx, server.ErrConflict, err.Error())
			return nil
		}

		role := input.Role
		if role == "" {
			role = "user"
		}
		if role != "user" {
			if err := s.userUC.SetUserRole(ctx.Request().Context(), created.Id, role); err != nil {
				http2.Fail(ctx, server.ErrInternal, "failed to set role: "+err.Error())
				return nil
			}
		}

		http2.Created(ctx, &pb.CreateUserResponse{User: created})
		return nil
	}
}

func (s *EnterpriseUserServer) adminUpdateUser() http2.HandlerFunc {
	return func(ctx http2.Context) error {
			id := ctx.Var("id")
		if id == "" {
			http2.Fail(ctx, server.ErrBadRequest, "user id is required")
			return nil
		}

		existing, err := s.userUC.GetUser(ctx.Request().Context(), id)
		if err != nil {
			http2.Fail(ctx, server.ErrNotFound, "user not found")
			return nil
		}

		var input struct {
			Username *string `json:"username"`
			Nickname *string `json:"nickname"`
			Email    *string `json:"email"`
			Name     *string `json:"name"`
			Title    *string `json:"title"`
			Bio      *string `json:"description"`
			Location *string `json:"location"`
			Avatar   *string `json:"avatar"`
			Phone    *string `json:"phone"`
			Role     *string `json:"role"`
			Status   *string `json:"status"`
		}
		if err := ctx.BindJSON(&input); err != nil {
			http2.Fail(ctx, server.ErrBadRequest, err.Error())
			return nil
		}

		if input.Nickname != nil {
			existing.Nickname = *input.Nickname
		}
		if input.Email != nil {
			existing.Email = *input.Email
		}
		if input.Avatar != nil {
			existing.Avatar = *input.Avatar
		}
		if input.Name != nil {
			existing.Name = *input.Name
		}
		if input.Title != nil {
			existing.Title = *input.Title
		}
		if input.Bio != nil {
			existing.Description = *input.Bio
		}
		if input.Location != nil {
			existing.Location = *input.Location
		}
		if input.Phone != nil {
			existing.Phone = *input.Phone
		}

		if input.Role != nil && *input.Role != "" {
			if err := s.userUC.SetUserRole(ctx.Request().Context(), id, *input.Role); err != nil {
				http2.Fail(ctx, server.ErrInternal, "failed to update role: "+err.Error())
				return nil
			}
		}

		if input.Status != nil && *input.Status != "" {
			statusMap := map[string]int32{
				"pending":   1,
				"active":    2,
				"inactive":  3,
				"suspended": 4,
				"rejected":  5,
			}
			if statusCode, ok := statusMap[*input.Status]; ok {
				if err := s.userUC.UpdateUserStatus(ctx.Request().Context(), id, int8(statusCode)); err != nil {
					http2.Fail(ctx, server.ErrInternal, "failed to update status: "+err.Error())
					return nil
				}
			}
		}

		u, err := s.userUC.UpdateUser(ctx.Request().Context(), existing)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}

		http2.OK(ctx, &pb.UpdateUserResponse{User: u})
		return nil
	}
}

func (s *EnterpriseUserServer) adminDeleteUser() http2.HandlerFunc {
	return func(ctx http2.Context) error {
			id := ctx.Var("id")
		if id == "" {
			http2.Fail(ctx, server.ErrBadRequest, "user id is required")
			return nil
		}

		if err := s.userUC.DeleteUser(ctx.Request().Context(), id); err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, &pb.DeleteUserResponse{})
		return nil
	}
}

func (s *EnterpriseUserServer) adminUpdateUserStatus() http2.HandlerFunc {
	return func(ctx http2.Context) error {
			id := ctx.Var("id")
		if id == "" {
			http2.Fail(ctx, server.ErrBadRequest, "user id is required")
			return nil
		}

		var input struct {
			Status int32 `json:"status" binding:"required"`
		}
		if err := ctx.BindJSON(&input); err != nil {
			http2.Fail(ctx, server.ErrBadRequest, err.Error())
			return nil
		}

		if err := s.userUC.UpdateUserStatus(ctx.Request().Context(), id, int8(input.Status)); err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, map[string]interface{}{"success": true})
		return nil
	}
}

func (s *EnterpriseUserServer) adminUpdateUserRole() http2.HandlerFunc {
	return func(ctx http2.Context) error {
			id := ctx.Var("id")
		if id == "" {
			http2.Fail(ctx, server.ErrBadRequest, "user id is required")
			return nil
		}

		var req struct {
			Role string `json:"role" binding:"required"`
		}
		if err := ctx.BindJSON(&req); err != nil {
			http2.Fail(ctx, server.ErrBadRequest, err.Error())
			return nil
		}

		if err := s.userUC.SetUserRole(ctx.Request().Context(), id, req.Role); err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, map[string]interface{}{"success": true})
		return nil
	}
}

func generateRandomPassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			result[i] = charset[i%len(charset)]
			continue
		}
		result[i] = charset[n.Int64()]
	}
	return string(result)
}
