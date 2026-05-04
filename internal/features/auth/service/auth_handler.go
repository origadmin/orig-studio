/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package service implements the HTTP handlers for the auth feature module.
package service

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"origadmin/application/origcms/api/gen/v1/types"
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
func (h *AuthHandler) RegisterRoutes(rg *gin.RouterGroup) {
	authGroup := rg.Group("/auth")
	{
		// Public auth routes
		authGroup.POST("/signin", h.Login)
		authGroup.POST("/signup", h.RegisterUser)
		authGroup.POST("/refresh", h.RefreshToken)
		authGroup.POST("/signout", h.Logout)
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
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		server.Fail(c, server.ErrBadRequest, err.Error())
		return
	}

	// Look up user by username (entity for role field)
	u, err := h.uc.GetUserByUsername(c.Request.Context(), req.Username)
	if err != nil {
		server.Fail(c, server.ErrUnauthorized, "invalid credentials")
		return
	}

	// Get role from entity (types.User doesn't have role field)
	userRole := "user"
	if entUser, entErr := h.uc.GetUserEntity(c.Request.Context(), u.Id); entErr == nil &&
		entUser.Role != "" {
		userRole = string(entUser.Role)
	}

	// Verify password
	if err := h.uc.VerifyPassword(c.Request.Context(), u.Id, req.Password); err != nil {
		server.Fail(c, server.ErrUnauthorized, "invalid credentials")
		return
	}

	token, err := h.jwt.Generate(u.Id, u.Username, u.IsStaff, userRole)
	if err != nil {
		slog.Error("failed to generate token", "err", err)
		server.Fail(c, server.ErrInternal, "token generation failed")
		return
	}

	refreshToken, err := h.jwt.GenerateRefreshToken(u.Id, u.Username, u.IsStaff, userRole)
	if err != nil {
		slog.Error("failed to generate refresh token", "err", err)
		server.Fail(c, server.ErrInternal, "refresh token generation failed")
		return
	}

	// Return simplified user info, ensure it includes the is_staff field
	loginUser := &LoginUser{
		Id:       u.Id,
		Username: u.Username,
		Nickname: u.Nickname,
		Email:    u.Email,
		IsStaff:  u.IsStaff,
	}
	// Use server.OK() to return unified response format {code:0, message:"ok", data:{...}}
	// per C016 unified API response convention.
	server.OK(c, TokenResponse{AccessToken: token, RefreshToken: refreshToken, TokenType: "Bearer", ExpiresIn: int64(h.jwt.TTL().Seconds()), User: loginUser})
}

func (h *AuthHandler) RegisterUser(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		server.Fail(c, server.ErrBadRequest, err.Error())
		return
	}

	count, _ := h.uc.CountUsers(c.Request.Context())
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
		return h.uc.CreateUser(c.Request.Context(), newUser, hashed)
	}()
	if err != nil {
		slog.Error("register failed", "err", err)
		server.Fail(c, server.ErrConflict, "registration failed: " + err.Error())
		return
	}

	userRole := "user"
	if isFirstUser {
		userRole = "admin"
		_ = h.uc.SetUserRole(c.Request.Context(), created.Id, "admin")
	}

	token, err := h.jwt.Generate(created.Id, created.Username, created.IsStaff, userRole)
	if err != nil {
		server.Fail(c, server.ErrInternal, "token generation failed")
		return
	}

	// Generate refresh token
	refreshToken, err := h.jwt.GenerateRefreshToken(created.Id, created.Username, created.IsStaff, userRole)
	if err != nil {
		slog.Error("failed to generate refresh token", "err", err)
		server.Fail(c, server.ErrInternal, "refresh token generation failed")
		return
	}

	loginUser := &LoginUser{
		Id:       created.Id,
		Username: created.Username,
		Nickname: created.Nickname,
		Email:    created.Email,
		IsStaff:  created.IsStaff,
	}
	// Use server.Created() to return unified response format {code:0, message:"ok", data:{...}}
	// per C016 unified API response convention.
	server.Created(c, TokenResponse{AccessToken: token, RefreshToken: refreshToken, TokenType: "Bearer", ExpiresIn: int64(h.jwt.TTL().Seconds()), User: loginUser})
}

// RefreshToken godoc: POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		server.Fail(c, server.ErrBadRequest, err.Error())
		return
	}

	// Parse the refresh token to get claims
	claims, err := h.jwt.Parse(req.RefreshToken)
	if err != nil {
		server.Fail(c, server.ErrUnauthorized, "invalid refresh token")
		return
	}

	// Get user information
	u, err := h.uc.GetUser(c.Request.Context(), claims.GetUserID(), nil)
	if err != nil {
		server.Fail(c, server.ErrInternal, "user not found")
		return
	}

	token, err := h.jwt.Generate(claims.GetUserID(), claims.Username, claims.IsStaff, claims.Role)
	if err != nil {
		slog.Error("failed to generate token", "err", err)
		server.Fail(c, server.ErrInternal, "token generation failed")
		return
	}

	// Generate new refresh token
	refreshToken, err := h.jwt.GenerateRefreshToken(claims.GetUserID(), claims.Username, claims.IsStaff, claims.Role)
	if err != nil {
		slog.Error("failed to generate refresh token", "err", err)
		server.Fail(c, server.ErrInternal, "refresh token generation failed")
		return
	}

	loginUser := &LoginUser{
		Id:       u.Id,
		Username: u.Username,
		Nickname: u.Nickname,
		Email:    u.Email,
		IsStaff:  u.IsStaff,
	}

	// Use server.OK() to return unified response format {code:0, message:"ok", data:{...}}
	// per C016 unified API response convention.
	server.OK(c, TokenResponse{AccessToken: token, RefreshToken: refreshToken, TokenType: "Bearer", ExpiresIn: int64(h.jwt.TTL().Seconds()), User: loginUser})
}

// Logout godoc: POST /api/v1/auth/logout (stateless: client discards token)
func (h *AuthHandler) Logout(c *gin.Context) {
	server.OK(c, gin.H{"message": "logged out"})
}

// Me godoc: GET /api/v1/auth/me  (requires JWT)
func (h *AuthHandler) Me(c *gin.Context) {
	claims, ok := server.GetClaims(c)
	if !ok {
		server.Fail(c, server.ErrUnauthorized, "unauthorized")
		return
	}

	u, err := h.uc.GetUser(c.Request.Context(), claims.GetUserID(), &dto.UserQueryOption{
		WithProfile: true,
	})
	if err != nil {
		server.Fail(c, server.ErrNotFound, "user not found")
		return
	}
	server.OK(c, u)
}
