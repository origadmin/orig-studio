/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package server is the M1 monolith entry point for origcms.
// It wires up ent, svc-user biz/data, and a Gin HTTP server in a single process.
// Microservice split starts at M2.
package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"origadmin/application/origcms/internal/auth"
	"origadmin/application/origcms/internal/svc-user/biz"
	"origadmin/application/origcms/internal/svc-user/dto"

	"origadmin/application/origcms/api/gen/v1/types"
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

func (h *AuthHandler) Register(group *gin.RouterGroup) {
	authGroup := group.Group("/auth")
	{
		authGroup.POST("/signin", h.Login)
		authGroup.POST("/signup", h.SignUp)
		authGroup.POST("/signout", h.Logout)

		// Protected auth routes
		protected := authGroup.Group("")
		protected.Use(JWTMiddleware(h.jwt))
		{
			protected.GET("/me", h.Me)
		}
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
	AccessToken string      `json:"access_token"`
	TokenType   string      `json:"token_type"`
	ExpiresIn   int64       `json:"expires_in"` // seconds, matches JWT TTL
	User        *LoginUser  `json:"user"`
}

// LoginUser 是登录响应中返回的用户信息，包含前端需要的 is_staff 字段
type LoginUser struct {
	Id       int64  `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname,omitempty"`
	Email    string `json:"email,omitempty"`
	IsStaff  bool   `json:"is_staff"`
}

// Login godoc: POST /api/v1/auth/signin
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Look up user by username
	u, err := h.uc.GetUserByUsername(c.Request.Context(), req.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// DEBUG: 打印用户信息
	slog.Info("user from GetUserByUsername", "id", u.Id, "username", u.Username, "is_staff", u.IsStaff)

	// Verify password
	if err := h.uc.VerifyPassword(c.Request.Context(), u.Id, req.Password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token, err := h.jwt.Generate(u.Id, u.Username, u.IsStaff)
	if err != nil {
		slog.Error("failed to generate token", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
		return
	}

	// 返回简化版用户信息，确保包含 is_staff 字段
	loginUser := &LoginUser{
		Id:       u.Id,
		Username: u.Username,
		Nickname: u.Nickname,
		Email:    u.Email,
		IsStaff:  u.IsStaff,
	}
	c.JSON(http.StatusOK, TokenResponse{AccessToken: token, TokenType: "Bearer", ExpiresIn: 86400, User: loginUser})
}

// SignUp godoc: POST /api/v1/auth/signup
func (h *AuthHandler) SignUp(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newUser := &types.User{
		Username:   req.Username,
		Nickname:  req.Nickname,
		Email:     req.Email,
		Status:    1,
		IsStaff:   true, // 第一个注册的用户自动成为管理员
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
		c.JSON(http.StatusConflict, gin.H{"error": "registration failed: " + err.Error()})
		return
	}

	token, err := h.jwt.Generate(created.Id, created.Username, created.IsStaff)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
		return
	}

	// 返回简化版用户信息，确保包含 is_staff 字段
	loginUser := &LoginUser{
		Id:       created.Id,
		Username: created.Username,
		Nickname: created.Nickname,
		Email:    created.Email,
		IsStaff:  created.IsStaff,
	}
	c.JSON(http.StatusCreated, TokenResponse{AccessToken: token, TokenType: "Bearer", ExpiresIn: 86400, User: loginUser})
}

// Me godoc: GET /api/v1/auth/me  (requires JWT)
func (h *AuthHandler) Me(c *gin.Context) {
	claims, ok := c.MustGet("claims").(*auth.Claims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	u, err := h.uc.GetUser(c.Request.Context(), claims.UserID, &dto.UserQueryOption{
		WithProfile: true,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, u)
}

// Logout godoc: POST /api/v1/auth/logout (stateless: client discards token)
func (h *AuthHandler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

// JWTMiddleware validates Bearer token and injects claims into context.
func JWTMiddleware(jwtMgr *auth.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if len(header) < 8 || header[:7] != "Bearer " {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid Authorization header"})
			return
		}
		claims, err := jwtMgr.Parse(header[7:])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token: " + err.Error()})
			return
		}
		c.Set("claims", claims)
		c.Next()
	}
}

// HealthHandler returns a simple health check response.
func HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "origcms"})
}

// contextKey is a typed context key to avoid collisions.
type contextKey string

const claimsKey contextKey = "claims"

// claimsFromContext extracts JWT claims from a standard context.Context.
func claimsFromContext(ctx context.Context) (*auth.Claims, bool) {
	v := ctx.Value(claimsKey)
	c, ok := v.(*auth.Claims)
	return c, ok
}

// writeJSON is a helper for net/http handlers (used by gateway package).
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
