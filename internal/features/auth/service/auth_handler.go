/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package service implements the HTTP handlers for the auth feature module.
package service

import (
	"context"
	"fmt"
	"log/slog"

	"origadmin/application/origstudio/api/gen/v1/types"
	http2 "origadmin/application/origstudio/internal/pkg/http"
	"origadmin/application/origstudio/internal/infra/auth"
	authdto "origadmin/application/origstudio/internal/features/auth/dto"
	"origadmin/application/origstudio/internal/features/user/biz"
	"origadmin/application/origstudio/internal/features/user/dto"
	"origadmin/application/origstudio/internal/server"
	systembiz "origadmin/application/origstudio/internal/features/system/biz"
)

type AuditFunc func(ctx context.Context, userID, username, action, ip, userAgent, result string)

type AuthHandler struct {
	uc            *biz.UserUseCase
	jwt           *auth.Manager
	configProvider systembiz.ConfigProvider
	auditFn       AuditFunc
}

func NewAuthHandler(uc *biz.UserUseCase, jwt *auth.Manager, configProvider systembiz.ConfigProvider) *AuthHandler {
	return &AuthHandler{uc: uc, jwt: jwt, configProvider: configProvider}
}

func (h *AuthHandler) SetAuditFunc(fn AuditFunc) {
	h.auditFn = fn
}

func (h *AuthHandler) RegisterRoutes(r http2.Router) {
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/signin", h.login())
		authGroup.POST("/signup", h.registerUser())
		authGroup.POST("/refresh", h.refreshToken())
		authGroup.POST("/signout", h.logout())
		authGroup.PUT("/password", server.WithJWTCtx(h.jwt, h.changePassword()))
	}
}

func (h *AuthHandler) login() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		var req authdto.LoginRequest
		if err := ctx.BindJSON(&req); err != nil {
			http2.Fail(ctx, server.ErrBadRequest, err.Error())
			return nil
		}

		u, err := h.uc.GetUserByUsername(ctx.Request().Context(), req.Username)
		if err != nil {
			h.auditLogin(ctx, "", req.Username, "failure")
			http2.Fail(ctx, server.ErrUnauthorized, "invalid credentials")
			return nil
		}

		if u.Status != types.UserStatus_USER_STATUS_ACTIVE {
			h.auditLogin(ctx, u.Id, u.Username, "failure")
			http2.Fail(ctx, server.ErrForbidden, "account is not active")
			return nil
		}

		userRole := "user"
		if entUser, entErr := h.uc.GetUserEntity(ctx.Request().Context(), u.Id); entErr == nil && entUser.Role != "" {
			userRole = string(entUser.Role)
		}

		if err := h.uc.VerifyPassword(ctx.Request().Context(), u.Id, req.Password); err != nil {
			h.auditLogin(ctx, u.Id, u.Username, "failure")
			http2.Fail(ctx, server.ErrUnauthorized, "invalid credentials")
			return nil
		}

		token, err := h.jwt.Generate(u.Id, u.Username, userRole)
		if err != nil {
			slog.Error("failed to generate token", "err", err)
			http2.Fail(ctx, server.ErrInternal, "token generation failed")
			return nil
		}

		refreshToken, err := h.jwt.GenerateRefreshToken(u.Id, u.Username, userRole)
		if err != nil {
			slog.Error("failed to generate refresh token", "err", err)
			http2.Fail(ctx, server.ErrInternal, "refresh token generation failed")
			return nil
		}

		loginUser := &authdto.LoginUser{
			Id:       u.Id,
			Username: u.Username,
			Nickname: u.Nickname,
			Email:    u.Email,
			Role:     userRole,
		}

		h.auditLogin(ctx, u.Id, u.Username, "success")
		http2.OK(ctx, authdto.TokenResponse{AccessToken: token, RefreshToken: refreshToken, TokenType: "Bearer", ExpiresIn: int64(h.jwt.TTL().Seconds()), User: loginUser})
		return nil
	}
}

func (h *AuthHandler) registerUser() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		var req authdto.RegisterRequest
		if err := ctx.BindJSON(&req); err != nil {
			http2.Fail(ctx, server.ErrBadRequest, err.Error())
			return nil
		}

		if h.configProvider != nil && !h.configProvider.GetBool(ctx.Request().Context(), "allow_registration") {
			http2.Fail(ctx, server.ErrForbidden, "registration is disabled")
			return nil
		}

		minLen := 6
		if h.configProvider != nil {
			if v := h.configProvider.GetInt(ctx.Request().Context(), "min_password_length"); v > 0 {
				minLen = v
			}
		}
		if len(req.Password) < minLen {
			http2.Fail(ctx, server.ErrBadRequest, fmt.Sprintf("password must be at least %d characters", minLen))
			return nil
		}

		count, _ := h.uc.CountUsers(ctx.Request().Context())
		isFirstUser := count == 0

		newUser := &types.User{
			Username: req.Username,
			Nickname: req.Nickname,
			Email:    req.Email,
			Status:   types.UserStatus_USER_STATUS_ACTIVE,
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
			_ = h.uc.SetUserSuperuser(ctx.Request().Context(), created.Id, true)
		}

		token, err := h.jwt.Generate(created.Id, created.Username, userRole)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, "token generation failed")
			return nil
		}

		refreshToken, err := h.jwt.GenerateRefreshToken(created.Id, created.Username, userRole)
		if err != nil {
			slog.Error("failed to generate refresh token", "err", err)
			http2.Fail(ctx, server.ErrInternal, "refresh token generation failed")
			return nil
		}

		loginUser := &authdto.LoginUser{
			Id:       created.Id,
			Username: created.Username,
			Nickname: created.Nickname,
			Email:    created.Email,
			Role:     userRole,
		}

		http2.Created(ctx, authdto.TokenResponse{AccessToken: token, RefreshToken: refreshToken, TokenType: "Bearer", ExpiresIn: int64(h.jwt.TTL().Seconds()), User: loginUser})
		return nil
	}
}

func (h *AuthHandler) refreshToken() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		var req struct {
			RefreshToken string `json:"refresh_token" binding:"required"`
		}
		if err := ctx.BindJSON(&req); err != nil {
			http2.Fail(ctx, server.ErrBadRequest, err.Error())
			return nil
		}

		claims, err := h.jwt.Parse(req.RefreshToken)
		if err != nil {
			http2.Fail(ctx, server.ErrUnauthorized, "invalid refresh token")
			return nil
		}

		u, err := h.uc.GetUser(ctx.Request().Context(), claims.GetUserID(), nil)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, "user not found")
			return nil
		}

		if u.Status != types.UserStatus_USER_STATUS_ACTIVE {
			http2.Fail(ctx, server.ErrForbidden, "account is not active")
			return nil
		}

		token, err := h.jwt.Generate(claims.GetUserID(), claims.Username, claims.Role)
		if err != nil {
			slog.Error("failed to generate token", "err", err)
			http2.Fail(ctx, server.ErrInternal, "token generation failed")
			return nil
		}

		refreshToken, err := h.jwt.GenerateRefreshToken(claims.GetUserID(), claims.Username, claims.Role)
		if err != nil {
			slog.Error("failed to generate refresh token", "err", err)
			http2.Fail(ctx, server.ErrInternal, "refresh token generation failed")
			return nil
		}

		loginUser := &authdto.LoginUser{
			Id:       u.Id,
			Username: u.Username,
			Nickname: u.Nickname,
			Email:    u.Email,
			Role:     u.Role,
		}

		http2.OK(ctx, authdto.TokenResponse{AccessToken: token, RefreshToken: refreshToken, TokenType: "Bearer", ExpiresIn: int64(h.jwt.TTL().Seconds()), User: loginUser})
		return nil
	}
}

func (h *AuthHandler) logout() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		if claims, ok := server.GetClaimsCtx(ctx); ok {
			h.auditLogout(ctx, claims.GetUserID(), claims.Username)
		} else {
			header := ctx.GetHeader("Authorization")
			if len(header) >= 8 && header[:7] == "Bearer " {
				if claims, err := h.jwt.Parse(header[7:]); err == nil {
					h.auditLogout(ctx, claims.GetUserID(), claims.Username)
				}
			}
		}
		http2.OK(ctx, map[string]any{"message": "logged out"})
		return nil
	}
}

func (h *AuthHandler) changePassword() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		claims, ok := server.GetClaimsCtx(ctx)
		if !ok {
			http2.Fail(ctx, server.ErrUnauthorized, "unauthorized")
			return nil
		}

		var req authdto.ChangePasswordRequest
		if err := ctx.BindJSON(&req); err != nil {
			http2.Fail(ctx, server.ErrBadRequest, err.Error())
			return nil
		}

		if err := h.uc.VerifyPassword(ctx.Request().Context(), claims.GetUserID(), req.CurrentPassword); err != nil {
			http2.Fail(ctx, server.ErrUnauthorized, "invalid current password")
			return nil
		}

		hashedPassword, err := h.uc.HashPassword(req.NewPassword)
		if err != nil {
			slog.Error("failed to hash password", "err", err)
			http2.Fail(ctx, server.ErrInternal, "password update failed")
			return nil
		}

		if err := h.uc.UpdateUserPassword(ctx.Request().Context(), claims.GetUserID(), hashedPassword); err != nil {
			slog.Error("failed to update password", "err", err)
			http2.Fail(ctx, server.ErrInternal, "password update failed")
			return nil
		}

		http2.OK(ctx, map[string]any{"message": "password updated"})
		return nil
	}
}

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

func (h *AuthHandler) auditLogin(ctx http2.Context, userID, username, result string) {
	if h.auditFn == nil {
		return
	}
	ip := ctx.ClientIP()
	ua := ctx.GetHeader("User-Agent")
	h.auditFn(ctx.Request().Context(), userID, username, "login", ip, ua, result)
}

func (h *AuthHandler) auditLogout(ctx http2.Context, userID, username string) {
	if h.auditFn == nil {
		return
	}
	ip := ctx.ClientIP()
	ua := ctx.GetHeader("User-Agent")
	h.auditFn(ctx.Request().Context(), userID, username, "logout", ip, ua, "success")
}
