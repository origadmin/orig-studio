/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package service implements the HTTP handlers for the auth feature module.
package service

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"origadmin/application/origcms/api/gen/v1/types"
	http2 "origadmin/application/origcms/internal/helpers/http"
	ginadapter "origadmin/application/origcms/internal/helpers/http/gin"
	"origadmin/application/origcms/internal/infra/auth"
	"origadmin/application/origcms/internal/features/user/biz"
	"origadmin/application/origcms/internal/features/user/dto"
	"origadmin/application/origcms/internal/server"
)

// AuthHandler handles /api/v1/auth/* routes.
type AuthHandler struct {
	uc  *biz.UserUseCase
	jwt *auth.Manager
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(uc *biz.UserUseCase, jwt *auth.Manager) *AuthHandler {
	return &AuthHandler{uc: uc, jwt: jwt}
}

// RegisterRoutes registers the handler's routes.
func (h *AuthHandler) RegisterRoutes(r http2.Router) {
	authGroup := r.Group("/auth")
	{
		// Public auth routes
		authGroup.POST("/signin", h.login())
		authGroup.POST("/signup", h.registerUser())
		authGroup.POST("/refresh", h.refreshToken())
		authGroup.POST("/signout", h.logout())
	}
}

// LoginRequest is the request body for POST /api/v1/auth/login.
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest is the request body for POST /api/v1/auth/register.
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=64"`
	Password string `json:"password" binding:"required,min=6,max=128"`
	Email    string `json:"email"    binding:"omitempty,email"`
	Nickname string `json:"nickname"`
}

// TokenResponse is the response body for successful auth.
// Fields match the frontend Token interface in request.ts.
type TokenResponse struct {
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"refresh_token"`
	TokenType    string     `json:"token_type"`
	ExpiresIn    int64      `json:"expires_in"` // seconds, matches JWT TTL
	User         *LoginUser `json:"user"`
}

// LoginUser is the user info returned in login response, including the is_staff field needed by frontend
type LoginUser struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname,omitempty"`
	Email    string `json:"email,omitempty"`
	IsStaff  bool   `json:"is_staff"`
}

// Login godoc: POST /api/v1/auth/signin
func (h *AuthHandler) login() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)

		var req LoginRequest
		if err := gc.ShouldBindJSON(&req); err != nil {
			http2.Fail(ctx, server.ErrBadRequest, err.Error())
			return nil
		}

		// Look up user by username (entity for role field)
		u, err := h.uc.GetUserByUsername(ctx.Request().Context(), req.Username)
		if err != nil {
			http2.Fail(ctx, server.ErrUnauthorized, "invalid credentials")
			return nil
		}

		// Get role from entity (types.User doesn't have role field)
		userRole := "user"
		if entUser, entErr := h.uc.GetUserEntity(ctx.Request().Context(), u.Id); entErr == nil &&
			entUser.Role != "" {
			userRole = string(entUser.Role)
		}

		// Verify password
		if err := h.uc.VerifyPassword(ctx.Request().Context(), u.Id, req.Password); err != nil {
			http2.Fail(ctx, server.ErrUnauthorized, "invalid credentials")
			return nil
		}

		token, err := h.jwt.Generate(u.Id, u.Username, u.IsStaff, userRole)
		if err != nil {
			slog.Error("failed to generate token", "err", err)
			http2.Fail(ctx, server.ErrInternal, "token generation failed")
			return nil
		}

		refreshToken, err := h.jwt.GenerateRefreshToken(u.Id, u.Username, u.IsStaff, userRole)
		if err != nil {
			slog.Error("failed to generate refresh token", "err", err)
			http2.Fail(ctx, server.ErrInternal, "refresh token generation failed")
			return nil
		}

		// Return simplified user info, ensure it includes the is_staff field
		loginUser := &LoginUser{
			Id:       u.Id,
			Username: u.Username,
			Nickname: u.Nickname,
			Email:    u.Email,
			IsStaff:  u.IsStaff,
		}
		// Use http2.OK() to return unified response format {code:0, message:"ok", data:{...}}
		// per C016 unified API response convention.
		http2.OK(ctx, TokenResponse{AccessToken: token, RefreshToken: refreshToken, TokenType: "Bearer", ExpiresIn: int64(h.jwt.TTL().Seconds()), User: loginUser})
		return nil
	}
}

func (h *AuthHandler) registerUser() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)

		var req RegisterRequest
		if err := gc.ShouldBindJSON(&req); err != nil {
			http2.Fail(ctx, server.ErrBadRequest, err.Error())
			return nil
		}

		count, _ := h.uc.CountUsers(ctx.Request().Context())
		isFirstUser := count == 0

		newUser := &types.User{
			Username: req.Username,
			Nickname: req.Nickname,
			Email:    req.Email,
			Status:   1,
			IsStaff:  isFirstUser,
		}

		created, err := func() (*types.User, error) {
			hashed, herr := h.uc.HashPassword(req.Password)
			if herr != nil {
				return nil, herr
			}
			return h.uc.CreateUser(ctx.Request().Context(), newUser, hashed)
		}()
		if err != nil {
			slog.Error("register failed", "err", err)
			http2.Fail(ctx, server.ErrConflict, "registration failed: "+err.Error())
			return nil
		}

		userRole := "user"
		if isFirstUser {
			userRole = "admin"
			_ = h.uc.SetUserRole(ctx.Request().Context(), created.Id, "admin")
		}

		token, err := h.jwt.Generate(created.Id, created.Username, created.IsStaff, userRole)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, "token generation failed")
			return nil
		}

		// Generate refresh token
		refreshToken, err := h.jwt.GenerateRefreshToken(created.Id, created.Username, created.IsStaff, userRole)
		if err != nil {
			slog.Error("failed to generate refresh token", "err", err)
			http2.Fail(ctx, server.ErrInternal, "refresh token generation failed")
			return nil
		}

		loginUser := &LoginUser{
			Id:       created.Id,
			Username: created.Username,
			Nickname: created.Nickname,
			Email:    created.Email,
			IsStaff:  created.IsStaff,
		}
		// Use http2.Created() to return unified response format {code:0, message:"ok", data:{...}}
		// per C016 unified API response convention.
		http2.Created(ctx, TokenResponse{AccessToken: token, RefreshToken: refreshToken, TokenType: "Bearer", ExpiresIn: int64(h.jwt.TTL().Seconds()), User: loginUser})
		return nil
	}
}

// RefreshToken godoc: POST /api/v1/auth/refresh
func (h *AuthHandler) refreshToken() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)

		var req struct {
			RefreshToken string `json:"refresh_token" binding:"required"`
		}
		if err := gc.ShouldBindJSON(&req); err != nil {
			http2.Fail(ctx, server.ErrBadRequest, err.Error())
			return nil
		}

		// Parse the refresh token to get claims
		claims, err := h.jwt.Parse(req.RefreshToken)
		if err != nil {
			http2.Fail(ctx, server.ErrUnauthorized, "invalid refresh token")
			return nil
		}

		// Get user information
		u, err := h.uc.GetUser(ctx.Request().Context(), claims.GetUserID(), nil)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, "user not found")
			return nil
		}

		token, err := h.jwt.Generate(claims.GetUserID(), claims.Username, claims.IsStaff, claims.Role)
		if err != nil {
			slog.Error("failed to generate token", "err", err)
			http2.Fail(ctx, server.ErrInternal, "token generation failed")
			return nil
		}

		// Generate new refresh token
		refreshToken, err := h.jwt.GenerateRefreshToken(claims.GetUserID(), claims.Username, claims.IsStaff, claims.Role)
		if err != nil {
			slog.Error("failed to generate refresh token", "err", err)
			http2.Fail(ctx, server.ErrInternal, "refresh token generation failed")
			return nil
		}

		loginUser := &LoginUser{
			Id:       u.Id,
			Username: u.Username,
			Nickname: u.Nickname,
			Email:    u.Email,
			IsStaff:  u.IsStaff,
		}

		// Use http2.OK() to return unified response format {code:0, message:"ok", data:{...}}
		// per C016 unified API response convention.
		http2.OK(ctx, TokenResponse{AccessToken: token, RefreshToken: refreshToken, TokenType: "Bearer", ExpiresIn: int64(h.jwt.TTL().Seconds()), User: loginUser})
		return nil
	}
}

// Logout godoc: POST /api/v1/auth/logout (stateless: client discards token)
func (h *AuthHandler) logout() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		http2.OK(ctx, gin.H{"message": "logged out"})
		return nil
	}
}

// Me godoc: GET /api/v1/auth/me  (requires JWT)
func (h *AuthHandler) Me() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		claims, ok := server.GetClaimsCtx(ctx)
		if !ok {
			http2.Fail(ctx, server.ErrUnauthorized, "unauthorized")
			return nil
		}

		u, err := h.uc.GetUser(ctx.Request().Context(), claims.GetUserID(), &dto.UserQueryOption{
			WithProfile: true,
		})
		if err != nil {
			http2.Fail(ctx, server.ErrNotFound, "user not found")
			return nil
		}
		http2.OK(ctx, u)
		return nil
	}
}
