/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package server implements the server handlers for the application.
package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"origadmin/application/origcms/api/gen/v1/types"
	"origadmin/application/origcms/internal/auth"
	"origadmin/application/origcms/internal/handler"
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

// Register registers the handler's routes.
func (h *AuthHandler) Register(r handler.Router) {
	authGroup := r.Group("/auth")
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
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Look up user by username (entity for role field)
	u, err := h.uc.GetUserByUsername(r.Context(), req.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Get role from entity (types.User doesn't have role field)
	userRole := "user"
	if entUser, entErr := h.uc.GetUserEntity(r.Context(), u.Id); entErr == nil &&
		entUser.Role != "" {
		userRole = string(entUser.Role)
	}

	// Verify password
	if err := h.uc.VerifyPassword(r.Context(), u.Id, req.Password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token, err := h.jwt.Generate(u.Id, u.Username, u.IsStaff, userRole)
	if err != nil {
		slog.Error("failed to generate token", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
		return
	}

	refreshToken, err := h.jwt.GenerateRefreshToken(u.Id, u.Username, u.IsStaff, userRole)
	if err != nil {
		slog.Error("failed to generate refresh token", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "refresh token generation failed"})
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
	c.JSON(
		http.StatusOK,
		TokenResponse{AccessToken: token, RefreshToken: refreshToken, TokenType: "Bearer", ExpiresIn: int64(h.jwt.TTL().Seconds()), User: loginUser},
	)
}

func (h *AuthHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	count, _ := h.uc.CountUsers(r.Context())
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
		return h.uc.CreateUser(r.Context(), newUser, hashed)
	}()
	if err != nil {
		slog.Error("register failed", "err", err)
		c.JSON(http.StatusConflict, gin.H{"error": "registration failed: " + err.Error()})
		return
	}

	userRole := "user"
	if isFirstUser {
		userRole = "admin"
		_ = h.uc.SetUserRole(r.Context(), created.Id, "admin")
	}

	token, err := h.jwt.Generate(created.Id, created.Username, created.IsStaff, userRole)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
		return
	}

	// Generate refresh token
	refreshToken, err := h.jwt.GenerateRefreshToken(created.Id, created.Username, created.IsStaff, userRole)
	if err != nil {
		slog.Error("failed to generate refresh token", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "refresh token generation failed"})
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
	c.JSON(
		http.StatusCreated,
		TokenResponse{AccessToken: token, RefreshToken: refreshToken, TokenType: "Bearer", ExpiresIn: int64(h.jwt.TTL().Seconds()), User: loginUser},
	)
}

// Me godoc: GET /api/v1/auth/me  (requires JWT)
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	claims, ok := c.Get("claims").(*auth.Claims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	u, err := h.uc.GetUser(r.Context(), claims.GetUserID(), &dto.UserQueryOption{
		WithProfile: true,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, u)
}

// RefreshToken godoc: POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.Bind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse the refresh token to get claims
	claims, err := h.jwt.Parse(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	// Get user information
	u, err := h.uc.GetUser(r.Context(), claims.GetUserID(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
		return
	}

	token, err := h.jwt.Generate(claims.GetUserID(), claims.Username, claims.IsStaff, claims.Role)
	if err != nil {
		slog.Error("failed to generate token", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
		return
	}

	// Generate new refresh token
	refreshToken, err := h.jwt.GenerateRefreshToken(claims.GetUserID(), claims.Username, claims.IsStaff, claims.Role)
	if err != nil {
		slog.Error("failed to generate refresh token", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "refresh token generation failed"})
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

	c.JSON(http.StatusOK, TokenResponse{
		AccessToken:  token,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(h.jwt.TTL().Seconds()),
		User:         loginUser,
	})
}

// Logout godoc: POST /api/v1/auth/logout (stateless: client discards token)
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	c := handler.NewGinContextAdapterFromHTTP(w, r)
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
