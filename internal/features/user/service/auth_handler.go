package service

import (
	"context"

	"google.golang.org/grpc/metadata"

	userv1 "origadmin/application/origstudio/api/gen/v1/user"
	http2 "origadmin/application/origstudio/internal/pkg/http"
	"origadmin/application/origstudio/internal/infra/auth"
	"origadmin/application/origstudio/internal/server"
)

type AuthHandler struct {
	svc *UserService
	jwt *auth.Manager
}

func NewAuthHandler(svc *UserService, jwt *auth.Manager) *AuthHandler {
	return &AuthHandler{
		svc: svc,
		jwt: jwt,
	}
}

func (h *AuthHandler) RegisterRoutes(r http2.Router) {
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/signin", h.Login)
		authGroup.POST("/signout", h.Logout)
		authGroup.POST("/refresh", h.RefreshToken)
		authGroup.POST("/signup", h.Register)
		authGroup.GET("/me", server.WithJWTCtx(h.jwt, h.GetCurrentUser))
		authGroup.PUT("/password", server.WithJWTCtx(h.jwt, h.ChangePassword))
	}
}

func (h *AuthHandler) withAuth(goCtx context.Context, ctx http2.Context) context.Context {
	authHeader := ctx.GetHeader("Authorization")
	if authHeader != "" {
		md := metadata.Pairs("authorization", authHeader)
		goCtx = metadata.NewIncomingContext(goCtx, md)
	}
	return goCtx
}

func (h *AuthHandler) Login(ctx http2.Context) error {
	var req userv1.LoginRequest
	if err := ctx.BindJSON(&req); err != nil {
		return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
	}

	resp, err := h.svc.Login(ctx, &req)
	if err != nil {
		return server.FailCtx(ctx, server.ErrInternal, err.Error())
	}

	return server.OKCtx(ctx, resp)
}

func (h *AuthHandler) Logout(ctx http2.Context) error {
	var req userv1.LogoutRequest
	_ = ctx.BindJSON(&req)

	resp, err := h.svc.Logout(ctx, &req)
	if err != nil {
		return server.FailCtx(ctx, server.ErrInternal, err.Error())
	}

	return server.OKCtx(ctx, resp)
}

func (h *AuthHandler) RefreshToken(ctx http2.Context) error {
	var req userv1.RefreshTokenRequest
	if err := ctx.BindJSON(&req); err != nil {
		return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
	}

	// Forward the Authorization header (if any) into the gRPC-style incoming
	// metadata context so UserService.RefreshToken can fall back to using the
	// Bearer token from the header when `refresh_token` is missing from the
	// request body. This matches the behaviour of `GetCurrentUser` /
	// `ChangePassword` via `withAuth` and enables common frontend auth
	// libraries that call /refresh with an empty body and the refresh token
	// carried in the Authorization header.
	resp, err := h.svc.RefreshToken(h.withAuth(ctx, ctx), &req)
	if err != nil {
		return server.FailCtx(ctx, server.ErrInternal, err.Error())
	}

	return server.OKCtx(ctx, resp)
}

func (h *AuthHandler) Register(ctx http2.Context) error {
	var req userv1.RegisterRequest
	if err := ctx.BindJSON(&req); err != nil {
		return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
	}

	resp, err := h.svc.Register(ctx, &req)
	if err != nil {
		return server.FailCtx(ctx, server.ErrInternal, err.Error())
	}

	return server.OKCtx(ctx, resp)
}

func (h *AuthHandler) GetCurrentUser(ctx http2.Context) error {
	authCtx := h.withAuth(ctx, ctx)
	resp, err := h.svc.GetCurrentUser(authCtx, &userv1.GetCurrentUserRequest{})
	if err != nil {
		return server.FailCtx(ctx, server.ErrInternal, err.Error())
	}

	return server.OKCtx(ctx, resp)
}

func (h *AuthHandler) ChangePassword(ctx http2.Context) error {
	var req userv1.ChangePasswordRequest
	if err := ctx.BindJSON(&req); err != nil {
		return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
	}

	authCtx := h.withAuth(ctx, ctx)
	resp, err := h.svc.ChangePassword(authCtx, &req)
	if err != nil {
		return server.FailCtx(ctx, server.ErrInternal, err.Error())
	}

	return server.OKCtx(ctx, resp)
}
