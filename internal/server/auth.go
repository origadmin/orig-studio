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

	"origadmin/application/origcms/api/gen/v1/types"
	"origadmin/application/origcms/internal/auth"
	"origadmin/application/origcms/internal/svc-user/biz"
	"origadmin/application/origcms/internal/svc-user/dto"
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

// LoginUser 是登录响应中返回的用户信息，包含前端需要的 is_staff 字段
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Look up user by username (entity for role field)
	u, err := h.uc.GetUserByUsername(c.Request.Context(), req.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Get role from entity (types.User doesn't have role field)
	userRole := "user"
	if entUser, entErr := h.uc.GetUserEntity(c.Request.Context(), u.Uuid); entErr == nil &&
		entUser.Role != "" {
		userRole = string(entUser.Role)
	}

	// Verify password
	if err := h.uc.VerifyPassword(c.Request.Context(), u.Uuid, req.Password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token, err := h.jwt.Generate(u.Uuid, u.Username, u.IsStaff, userRole)
	if err != nil {
		slog.Error("failed to generate token", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
		return
	}

	// Generate refresh token
	refreshToken, err := h.jwt.GenerateRefreshToken(u.Uuid, u.Username, u.IsStaff, userRole)
	if err != nil {
		slog.Error("failed to generate refresh token", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "refresh token generation failed"})
		return
	}

	// 返回简化版用户信息，确保包含 is_staff 字段
	loginUser := &LoginUser{
		Id:       u.Uuid,
		Username: u.Username,
		Nickname: u.Nickname,
		Email:    u.Email,
		IsStaff:  u.IsStaff,
	}
	c.JSON(
		http.StatusOK,
		TokenResponse{AccessToken: token, RefreshToken: refreshToken, TokenType: "Bearer", ExpiresIn: int64(h.jwt.TTL().Seconds()), User: loginUser},
	)
}

// RegisterUser godoc: POST /api/v1/auth/register
func (h *AuthHandler) RegisterUser(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusConflict, gin.H{"error": "registration failed: " + err.Error()})
		return
	}

	userRole := "user"
	if isFirstUser {
		userRole = "admin"
		_ = h.uc.SetUserRole(c.Request.Context(), created.Uuid, "admin")
	}

	token, err := h.jwt.Generate(created.Uuid, created.Username, created.IsStaff, userRole)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
		return
	}

	// Generate refresh token
	refreshToken, err := h.jwt.GenerateRefreshToken(created.Uuid, created.Username, created.IsStaff, userRole)
	if err != nil {
		slog.Error("failed to generate refresh token", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "refresh token generation failed"})
		return
	}

	// 返回简化版用户信息，确保包含 is_staff 字段
	loginUser := &LoginUser{
		Id:       created.Uuid,
		Username: created.Username,
		Nickname: created.Nickname,
		Email:    created.Email,
		IsStaff:  created.IsStaff,
	}
	c.JSON(
		http.StatusCreated,
		TokenResponse{AccessToken: token, RefreshToken: refreshToken, TokenType: "Bearer", ExpiresIn: int64(h.jwt.TTL().Seconds()), User: loginUser},
	)
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

// RefreshToken godoc: POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse the refresh token to get claims
	claims, err := h.jwt.Parse(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	// Generate new access token
	token, err := h.jwt.Generate(claims.UserID, claims.Username, claims.IsStaff, claims.Role)
	if err != nil {
		slog.Error("failed to generate token", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
		return
	}

	// Generate new refresh token
	refreshToken, err := h.jwt.GenerateRefreshToken(claims.UserID, claims.Username, claims.IsStaff, claims.Role)
	if err != nil {
		slog.Error("failed to generate refresh token", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "refresh token generation failed"})
		return
	}

	c.JSON(http.StatusOK, TokenResponse{
		AccessToken:  token,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(h.jwt.TTL().Seconds()),
	})
}

// Logout godoc: POST /api/v1/auth/logout (stateless: client discards token)
func (h *AuthHandler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
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
