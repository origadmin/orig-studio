/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package service

import (
	"image"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	apitypes "origadmin/application/origstudio/api/gen/v1/types"
	"origadmin/application/origstudio/internal/conf"
	http2 "origadmin/application/origstudio/internal/pkg/http"
	"origadmin/application/origstudio/internal/infra/auth"
	"origadmin/application/origstudio/internal/domain/types"
	contentbiz "origadmin/application/origstudio/internal/features/content/biz"
	userbiz "origadmin/application/origstudio/internal/features/user/biz"
	"origadmin/application/origstudio/internal/server"
)

// MeHandler handles /api/v1/me routes.
type MeHandler struct {
	userUC         *userbiz.UserUseCase
	likeFavoriteUC *contentbiz.LikeFavoriteUseCase
	playlistUC     *contentbiz.PlaylistChannelUseCase
	historyUC      *contentbiz.HistoryUseCase
	jwt            *auth.Manager
	paths          *conf.StoragePaths
}

// NewMeHandler creates a new MeHandler.
func NewMeHandler(
	userUC *userbiz.UserUseCase,
	likeFavoriteUC *contentbiz.LikeFavoriteUseCase,
	playlistUC *contentbiz.PlaylistChannelUseCase,
	historyUC *contentbiz.HistoryUseCase,
	jwt *auth.Manager,
	paths *conf.StoragePaths,
) *MeHandler {
	return &MeHandler{
		userUC:         userUC,
		likeFavoriteUC: likeFavoriteUC,
		playlistUC:     playlistUC,
		historyUC:      historyUC,
		jwt:            jwt,
		paths:          paths,
	}
}

// RegisterRoutes registers the handler's routes.
func (h *MeHandler) RegisterRoutes(r http2.Router) {
	me := r.Group("/me")
	me.Use(server.JWTMiddlewareCtx(h.jwt))
	{
		// ================================
		// 1. CURRENT USER PROFILE
		// ================================
		me.GET("", h.GetMe)
		me.PUT("", h.UpdateMe)
		me.PUT("/password", h.UpdatePassword)

		me.GET("/profile", h.GetProfile)
		me.PUT("/profile", h.UpdateProfile)
		me.POST("/avatar", h.UploadAvatar)
		me.GET("/setting", h.GetSetting)
		me.PUT("/setting", h.UpdateSetting)

		// ================================
		// 2. CURRENT USER RESOURCES
		// ================================
		me.GET("/playlists", h.GetPlaylists)
		me.POST("/playlists", h.CreatePlaylist)
		me.PATCH("/playlists/:id", h.UpdatePlaylist)
		me.DELETE("/playlists/:id", h.DeletePlaylist)
		me.POST("/playlists/:id/media", h.AddMediaToPlaylist)
		me.DELETE("/playlists/:id/media/:mediaId", h.RemoveMediaFromPlaylist)
		me.GET("/favorites", h.GetFavorites)
		me.DELETE("/favorites/:id", h.RemoveFavorite)
		me.GET("/likes", h.GetLikes)
		me.GET("/subscriptions", h.GetSubscriptions)
		me.GET("/followers", h.GetFollowers)

		// ================================
		// 3. WATCH HISTORY (independent from favorites/likes)
		// ================================
		me.GET("/history", h.GetHistory)
		me.POST("/history", h.UpsertHistory)
		me.POST("/history/sync", h.SyncHistory)
		me.DELETE("/history", h.ClearHistory)
		me.DELETE("/history/:id", h.RemoveHistoryItem)

		me.GET("/stats", h.GetStats)
		me.PUT("/slug", h.UpdateSlug)
		me.GET("/channels", h.GetChannels)
	}
}

// GetMe returns the current user's information.
func (h *MeHandler) GetMe(ctx http2.Context) error {
	claims, ok := server.GetClaimsCtx(ctx)
	if !ok {
		return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
	}

	user, err := h.userUC.GetUser(ctx, claims.GetUserID())
	if err != nil {
		return server.FailCtx(ctx, server.ErrUserNotFound, "User not found")
	}

	return server.OKCtx(ctx, user)
}

// UpdateMe updates the current user's information.
func (h *MeHandler) UpdateMe(ctx http2.Context) error {
	claims, ok := server.GetClaimsCtx(ctx)
	if !ok {
		return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
	}

	var input struct {
		Nickname string `json:"nickname"`
		Email    string `json:"email" binding:"omitempty,email"`
		Phone    string `json:"phone"`
		Bio      string `json:"bio"`
		Location string `json:"location"`
		Avatar   string `json:"avatar"`
	}

	if err := ctx.BindJSON(&input); err != nil {
		return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
	}

	user, err := h.userUC.GetUser(ctx, claims.GetUserID())
	if err != nil {
		return server.FailCtx(ctx, server.ErrUserNotFound, "User not found")
	}

	userChanged := false
	if input.Nickname != "" {
		user.Nickname = input.Nickname
		userChanged = true
	}
	if input.Email != "" {
		user.Email = input.Email
		userChanged = true
	}
	if input.Phone != "" {
		user.Phone = input.Phone
		userChanged = true
	}
	if input.Avatar != "" {
		user.Avatar = input.Avatar
		userChanged = true
	}

	if userChanged {
		user, err = h.userUC.UpdateUser(ctx, user)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}
	}

	profileChanged := input.Bio != "" || input.Location != ""
	if profileChanged {
		profile := &apitypes.UserProfile{
			Name:     user.Nickname,
			Bio:      input.Bio,
			Location: input.Location,
			Avatar:   user.Avatar,
		}
		if err := h.userUC.UpdateUserProfile(ctx, claims.GetUserID(), profile); err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}
	}

	return server.OKCtx(ctx, user)
}

// UpdatePassword updates the current user's password.
func (h *MeHandler) UpdatePassword(ctx http2.Context) error {
	claims, ok := server.GetClaimsCtx(ctx)
	if !ok {
		return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
	}

	var input struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}

	if err := ctx.BindJSON(&input); err != nil {
		return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
	}

	// Verify old password
	if err := h.userUC.VerifyPassword(ctx, claims.GetUserID(), input.OldPassword); err != nil {
		return server.FailCtx(ctx, server.ErrPasswordWrong, "Invalid old password")
	}

	hashedPassword, err := h.userUC.HashPassword(input.NewPassword)
	if err != nil {
		return server.FailCtx(ctx, server.ErrInternal, "password update failed")
	}

	if err := h.userUC.UpdateUserPassword(ctx, claims.GetUserID(), hashedPassword); err != nil {
		return server.FailCtx(ctx, server.ErrInternal, "password update failed")
	}

	return server.OKCtx(ctx, map[string]any{"message": "Password updated"})
}

func (h *MeHandler) GetProfile(ctx http2.Context) error {
	claims, ok := server.GetClaimsCtx(ctx)
	if !ok {
		return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
	}

	profile, err := h.userUC.GetUserProfile(ctx, claims.GetUserID())
	if err != nil {
		return server.FailCtx(ctx, server.ErrInternal, err.Error())
	}

	return server.OKCtx(ctx, profile)
}

func (h *MeHandler) UpdateProfile(ctx http2.Context) error {
	claims, ok := server.GetClaimsCtx(ctx)
	if !ok {
		return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
	}

	var input struct {
		Nickname string `json:"nickname"`
		Email    string `json:"email" binding:"omitempty,email"`
		Phone    string `json:"phone"`
		Bio      string `json:"bio"`
		Location string `json:"location"`
	}

	if err := ctx.BindJSON(&input); err != nil {
		return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
	}

	user, err := h.userUC.GetUser(ctx, claims.GetUserID())
	if err != nil {
		return server.FailCtx(ctx, server.ErrUserNotFound, "User not found")
	}

	userChanged := false
	if input.Nickname != "" {
		user.Nickname = input.Nickname
		userChanged = true
	}
	if input.Email != "" {
		user.Email = input.Email
		userChanged = true
	}
	if input.Phone != "" {
		user.Phone = input.Phone
		userChanged = true
	}

	if userChanged {
		user, err = h.userUC.UpdateUser(ctx, user)
		if err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}
	}

	profileChanged := input.Bio != "" || input.Location != ""
	if profileChanged {
		profile := &apitypes.UserProfile{
			Name:     user.Nickname,
			Bio:      input.Bio,
			Location: input.Location,
			Avatar:   user.Avatar,
		}
		if err := h.userUC.UpdateUserProfile(ctx, claims.GetUserID(), profile); err != nil {
			return server.FailCtx(ctx, server.ErrInternal, err.Error())
		}
	}

	return server.OKCtx(ctx, user)
}

func (h *MeHandler) UploadAvatar(ctx http2.Context) error {
	claims, ok := server.GetClaimsCtx(ctx)
	if !ok {
		return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
	}

	file, header, err := ctx.FormFile("avatar")
	if err != nil {
		return server.FailCtx(ctx, server.ErrBadRequest, "avatar file is required")
	}
	defer file.Close()

	const maxAvatarSize = 5 * 1024 * 1024 // 5MB
	if header.Size > maxAvatarSize {
		return server.FailCtx(ctx, server.ErrPayloadTooLarge, "avatar file too large, max 5MB")
	}

	contentType := header.Header.Get("Content-Type")
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/webp" {
		return server.FailCtx(ctx, server.ErrUnsupportedMediaType, "unsupported media type, use jpg/png/webp")
	}

	ext := filepath.Ext(header.Filename)
	if ext == "" {
		switch contentType {
		case "image/jpeg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "image/webp":
			ext = ".webp"
		}
	}

	// Use StoragePaths registry for avatar storage path
	avatarDir := h.paths.Dir("assets/avatars")
	if err := os.MkdirAll(avatarDir, 0755); err != nil {
		return server.FailCtx(ctx, server.ErrInternal, "failed to create avatar directory")
	}

	now := time.Now()
	yearMonth := strconv.Itoa(now.Year()) + strconv.Itoa(int(now.Month()))
	filename := claims.GetUserID() + "_" + yearMonth + ext
	dst := filepath.Join(avatarDir, filename)

	out, err := os.Create(dst)
	if err != nil {
		return server.FailCtx(ctx, server.ErrInternal, "failed to save avatar file")
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		return server.FailCtx(ctx, server.ErrInternal, "failed to write avatar file")
	}

	imgFile, err := os.Open(dst)
	if err != nil {
		os.Remove(dst)
		return server.FailCtx(ctx, server.ErrInternal, "failed to read avatar file")
	}
	defer imgFile.Close()

	imgConfig, _, err := image.DecodeConfig(imgFile)
	if err != nil {
		os.Remove(dst)
		return server.FailCtx(ctx, server.ErrBadRequest, "invalid image file")
	}

	const minDimension = 200
	const maxDimension = 4000
	if imgConfig.Width < minDimension || imgConfig.Height < minDimension {
		os.Remove(dst)
		return server.FailCtx(ctx, server.ErrBadRequest, "image too small, min 200x200")
	}
	if imgConfig.Width > maxDimension || imgConfig.Height > maxDimension {
		os.Remove(dst)
		return server.FailCtx(ctx, server.ErrBadRequest, "image too large, max 4000x4000")
	}

	// URL via gateway /files/ prefix
	avatarURL := "/files/assets/avatars/" + filename

	user, err := h.userUC.GetUser(ctx, claims.GetUserID())
	if err != nil {
		return server.FailCtx(ctx, server.ErrUserNotFound, "User not found")
	}
	user.Avatar = avatarURL

	updated, err := h.userUC.UpdateUser(ctx, user)
	if err != nil {
		return server.FailCtx(ctx, server.ErrInternal, err.Error())
	}

	return server.OKCtx(ctx, map[string]any{"avatar_url": updated.Avatar})
}

func (h *MeHandler) GetSetting(ctx http2.Context) error {
	claims, ok := server.GetClaimsCtx(ctx)
	if !ok {
		return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
	}

	setting, err := h.userUC.GetUserSetting(ctx, claims.GetUserID())
	if err != nil {
		return server.FailCtx(ctx, server.ErrInternal, err.Error())
	}

	return server.OKCtx(ctx, setting)
}

func (h *MeHandler) UpdateSetting(ctx http2.Context) error {
	claims, ok := server.GetClaimsCtx(ctx)
	if !ok {
		return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
	}

	var input struct {
		Theme       string            `json:"theme"`
		Language    string            `json:"language"`
		Timezone    string            `json:"timezone"`
		Preferences map[string]string `json:"preferences"`
	}

	if err := ctx.BindJSON(&input); err != nil {
		return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
	}

	setting := &apitypes.UserSetting{
		Theme:       input.Theme,
		Language:    input.Language,
		Timezone:    input.Timezone,
		Preferences: input.Preferences,
	}

	if err := h.userUC.UpdateUserSetting(ctx, claims.GetUserID(), setting); err != nil {
		return server.FailCtx(ctx, server.ErrInternal, err.Error())
	}

	setting, err := h.userUC.GetUserSetting(ctx, claims.GetUserID())
	if err != nil {
		return server.FailCtx(ctx, server.ErrInternal, err.Error())
	}

	return server.OKCtx(ctx, setting)
}

// GetPlaylists returns the current user's playlists.
func (h *MeHandler) GetPlaylists(ctx http2.Context) error {
	claims, ok := server.GetClaimsCtx(ctx)
	if !ok {
		return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
	}

	page, _ := strconv.Atoi(ctx.QueryVarDefault("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.QueryVarDefault("page_size", "20"))
	// Normalize pagination parameters
	page, pageSize = types.NormalizeHTTPPagination(page, pageSize)

	list, total, err := h.playlistUC.ListUserPlaylists(ctx, claims.GetUserID(), page, pageSize)
	if err != nil {
		return server.FailCtx(ctx, server.ErrInternal, err.Error())
	}

	return server.OKCtx(ctx, map[string]any{
		"items":     list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// CreatePlaylist creates a new playlist for the current user.
func (h *MeHandler) CreatePlaylist(ctx http2.Context) error {
	claims, ok := server.GetClaimsCtx(ctx)
	if !ok {
		return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
	}

	var input struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		IsPublic    *bool  `json:"is_public"`
	}
	if err := ctx.BindJSON(&input); err != nil {
		return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
	}

	if input.Title == "" {
		return server.FailCtx(ctx, server.ErrBadRequest, "title is required")
	}

	if input.Description == "" {
		input.Description = " "
	}

	// Default to public if not specified (fixes B099: playlists were created as
	// PRIVATE by default because Go bool zero-value is false, then the public
	// detail endpoint /playlists/:token couldn't verify ownership without JWT).
	isPublic := true
	if input.IsPublic != nil {
		isPublic = *input.IsPublic
	}

	p, err := h.playlistUC.CreatePlaylist(ctx, &contentbiz.Playlist{
		Title:       input.Title,
		Description: input.Description,
		UserID:      claims.GetUserID(),
		IsPublic:    isPublic,
	})
	if err != nil {
		return server.FailCtx(ctx, server.ErrInternal, err.Error())
	}

	return server.OKCtx(ctx, map[string]any{"playlist": p})
}

// AddMediaToPlaylist adds a media item to a playlist.
func (h *MeHandler) AddMediaToPlaylist(ctx http2.Context) error {
	claims, ok := server.GetClaimsCtx(ctx)
	if !ok {
		return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
	}

	playlistID := ctx.Var("id")
	if playlistID == "" {
		return server.FailCtx(ctx, server.ErrBadRequest, "playlist id is required")
	}

	var input struct {
		MediaID string `json:"media_id"`
	}
	if err := ctx.BindJSON(&input); err != nil {
		return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
	}

	if input.MediaID == "" {
		return server.FailCtx(ctx, server.ErrBadRequest, "media_id is required")
	}

	if err := h.playlistUC.AddMediaToPlaylist(ctx, playlistID, input.MediaID, claims.GetUserID(), false); err != nil {
		return server.FailCtx(ctx, server.ErrInternal, err.Error())
	}

	return server.OKCtx(ctx, map[string]any{"message": "media added to playlist"})
}

// UpdatePlaylist updates a playlist's title (and optionally description/is_public).
func (h *MeHandler) UpdatePlaylist(ctx http2.Context) error {
	claims, ok := server.GetClaimsCtx(ctx)
	if !ok {
		return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
	}

	playlistID := ctx.Var("id")
	if playlistID == "" {
		return server.FailCtx(ctx, server.ErrBadRequest, "playlist id is required")
	}

	var input struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		IsPublic    *bool  `json:"is_public"`
	}
	if err := ctx.BindJSON(&input); err != nil {
		return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
	}

	existing, err := h.playlistUC.GetPlaylist(ctx, playlistID)
	if err != nil {
		return server.FailCtx(ctx, server.ErrNotFound, "playlist not found")
	}

	if input.Title != "" {
		existing.Title = input.Title
	}
	if input.Description != "" {
		existing.Description = input.Description
	}
	if input.IsPublic != nil {
		existing.IsPublic = *input.IsPublic
	}

	updated, err := h.playlistUC.UpdatePlaylist(ctx, existing, claims.GetUserID(), false)
	if err != nil {
		return server.FailCtx(ctx, server.ErrInternal, err.Error())
	}

	return server.OKCtx(ctx, map[string]any{"playlist": updated})
}

// DeletePlaylist deletes a playlist owned by the current user.
func (h *MeHandler) DeletePlaylist(ctx http2.Context) error {
	claims, ok := server.GetClaimsCtx(ctx)
	if !ok {
		return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
	}

	playlistID := ctx.Var("id")
	if playlistID == "" {
		return server.FailCtx(ctx, server.ErrBadRequest, "playlist id is required")
	}

	if err := h.playlistUC.DeletePlaylist(ctx, playlistID, claims.GetUserID(), false); err != nil {
		return server.FailCtx(ctx, server.ErrInternal, err.Error())
	}

	return server.OKCtx(ctx, map[string]any{"message": "playlist deleted successfully"})
}

// RemoveMediaFromPlaylist removes a media item from a playlist.
func (h *MeHandler) RemoveMediaFromPlaylist(ctx http2.Context) error {
	claims, ok := server.GetClaimsCtx(ctx)
	if !ok {
		return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
	}

	playlistID := ctx.Var("id")
	mediaID := ctx.Var("mediaId")
	if playlistID == "" || mediaID == "" {
		return server.FailCtx(ctx, server.ErrBadRequest, "playlist id and media id are required")
	}

	if err := h.playlistUC.RemoveMediaFromPlaylist(ctx, playlistID, mediaID, claims.GetUserID(), false); err != nil {
		return server.FailCtx(ctx, server.ErrInternal, err.Error())
	}

	return server.OKCtx(ctx, map[string]any{"message": "media removed from playlist"})
}

func (h *MeHandler) GetFavorites(ctx http2.Context) error {
	claims, ok := server.GetClaimsCtx(ctx)
	if !ok {
		return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
	}

	page, _ := strconv.Atoi(ctx.QueryVarDefault("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.QueryVarDefault("page_size", "20"))
	// Normalize pagination parameters
	page, pageSize = types.NormalizeHTTPPagination(page, pageSize)

	favorites, total, err := h.likeFavoriteUC.ListUserFavoritesPaginated(ctx, claims.GetUserID(), page, pageSize)
	if err != nil {
		return server.FailCtx(ctx, server.ErrInternal, err.Error())
	}

	return server.OKCtx(ctx, map[string]any{
		"items":     favorites,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetLikes returns the current user's likes.
func (h *MeHandler) GetLikes(ctx http2.Context) error {
	claims, ok := server.GetClaimsCtx(ctx)
	if !ok {
		return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
	}

	likes, err := h.likeFavoriteUC.ListUserLikes(ctx, claims.GetUserID())
	if err != nil {
		return server.FailCtx(ctx, server.ErrInternal, err.Error())
	}

	return server.OKCtx(ctx, map[string]any{
		"items":     likes,
		"total":     len(likes),
		"page":      1,
		"page_size": len(likes),
	})
}

// GetSubscriptions returns the current user's subscriptions.
func (h *MeHandler) GetSubscriptions(ctx http2.Context) error {
	claims, ok := server.GetClaimsCtx(ctx)
	if !ok {
		return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
	}

	page, _ := strconv.Atoi(ctx.QueryVarDefault("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.QueryVarDefault("page_size", "20"))
	// Normalize pagination parameters
	page, pageSize = types.NormalizeHTTPPagination(page, pageSize)

	list, total, err := h.userUC.GetSubscriptions(
		ctx,
		claims.GetUserID(),
		page,
		pageSize,
	)
	if err != nil {
		return server.FailCtx(ctx, server.ErrInternal, err.Error())
	}

	return server.OKCtx(ctx, map[string]any{
		"items":     list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetHistory returns the current user's watch history from the user_history table.
// This replaces the old implementation that incorrectly used favorites+likes.
func (h *MeHandler) GetHistory(ctx http2.Context) error {
	claims, ok := server.GetClaimsCtx(ctx)
	if !ok {
		return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
	}

	page, _ := strconv.Atoi(ctx.QueryVarDefault("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.QueryVarDefault("page_size", "20"))
	page, pageSize = types.NormalizeHTTPPagination(page, pageSize)

	contentType := ctx.QueryVarDefault("content_type", "")

	items, total, err := h.historyUC.List(ctx, claims.GetUserID(), contentType, page, pageSize)
	if err != nil {
		return server.FailCtx(ctx, server.ErrInternal, err.Error())
	}

	return server.OKCtx(ctx, map[string]any{
		"items":     items,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *MeHandler) UpdateSlug(ctx http2.Context) error {
	claims, ok := server.GetClaimsCtx(ctx)
	if !ok {
		return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
	}

	var input struct {
		Slug string `json:"slug" binding:"required,min=3,max=64"`
	}
	if err := ctx.BindJSON(&input); err != nil {
		return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
	}

	if err := h.userUC.UpdateUserSlug(ctx, claims.GetUserID(), input.Slug); err != nil {
		return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
	}

	u, _ := h.userUC.GetUser(ctx, claims.GetUserID())
	return server.OKCtx(ctx, u)
}

func (h *MeHandler) GetChannels(ctx http2.Context) error {
	claims, ok := server.GetClaimsCtx(ctx)
	if !ok {
		return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
	}
	_ = claims
	page, _ := strconv.Atoi(ctx.QueryVarDefault("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.QueryVarDefault("page_size", "20"))
	page, pageSize = types.NormalizeHTTPPagination(page, pageSize)
	return server.OKCtx(ctx, map[string]any{
		"channels":  []interface{}{},
		"total":     0,
		"page":      page,
		"page_size": pageSize,
	})
}

// UpsertHistory creates or updates a history record (progress reporting).
func (h *MeHandler) UpsertHistory(ctx http2.Context) error {
	claims, ok := server.GetClaimsCtx(ctx)
	if !ok {
		return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
	}

	var input struct {
		ContentID       string `json:"content_id" binding:"required"`
		ContentType     string `json:"content_type" binding:"required"`
		ProgressSeconds int    `json:"progress_seconds"`
		DurationSeconds int    `json:"duration_seconds"`
		Title           string `json:"title"`
		Thumbnail       string `json:"thumbnail"`
		ShortToken      string `json:"short_token"`
	}

	if err := ctx.BindJSON(&input); err != nil {
		return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
	}

	if input.ContentID == "" {
		return server.FailCtx(ctx, server.ErrBadRequest, "content_id is required")
	}

	result, err := h.historyUC.Upsert(ctx, &contentbiz.History{
		UserID:          claims.GetUserID(),
		ContentID:       input.ContentID,
		ContentType:     input.ContentType,
		ProgressSeconds: input.ProgressSeconds,
		DurationSeconds: input.DurationSeconds,
		Title:           input.Title,
		Thumbnail:       input.Thumbnail,
		ShortToken:      input.ShortToken,
	})
	if err != nil {
		if err.Error() == "invalid content_type: "+input.ContentType {
			return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
		}
		return server.FailCtx(ctx, server.ErrInternal, err.Error())
	}

	return server.OKCtx(ctx, map[string]any{"item": result})
}

// SyncHistory batch-syncs history records (login merge).
func (h *MeHandler) SyncHistory(ctx http2.Context) error {
	claims, ok := server.GetClaimsCtx(ctx)
	if !ok {
		return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
	}

	var input struct {
		Items []struct {
			ContentID       string `json:"content_id"`
			ContentType     string `json:"content_type"`
			ProgressSeconds int    `json:"progress_seconds"`
			DurationSeconds int    `json:"duration_seconds"`
			IsFinished      bool   `json:"is_finished"`
			Title           string `json:"title"`
			Thumbnail       string `json:"thumbnail"`
			ShortToken      string `json:"short_token"`
		} `json:"items"`
	}

	if err := ctx.BindJSON(&input); err != nil {
		return server.FailCtx(ctx, server.ErrBadRequest, err.Error())
	}

	items := make([]*contentbiz.History, len(input.Items))
	for i, item := range input.Items {
		items[i] = &contentbiz.History{
			ContentID:       item.ContentID,
			ContentType:     item.ContentType,
			ProgressSeconds: item.ProgressSeconds,
			DurationSeconds: item.DurationSeconds,
			IsFinished:      item.IsFinished,
			Title:           item.Title,
			Thumbnail:       item.Thumbnail,
			ShortToken:      item.ShortToken,
		}
	}

	result, mergedCount, err := h.historyUC.Sync(ctx, claims.GetUserID(), items)
	if err != nil {
		return server.FailCtx(ctx, server.ErrInternal, err.Error())
	}

	return server.OKCtx(ctx, map[string]any{
		"items":        result,
		"merged_count": mergedCount,
	})
}

// ClearHistory clears all watch history for the current user.
// This only deletes from the user_history table and does NOT touch favorites or likes.
func (h *MeHandler) ClearHistory(ctx http2.Context) error {
	claims, ok := server.GetClaimsCtx(ctx)
	if !ok {
		return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
	}

	deletedCount, err := h.historyUC.ClearAll(ctx, claims.GetUserID())
	if err != nil {
		return server.FailCtx(ctx, server.ErrInternal, err.Error())
	}

	return server.OKCtx(ctx, map[string]any{"deleted_count": deletedCount})
}

// RemoveHistoryItem removes a single history item by its ID.
// This only deletes from the user_history table and does NOT touch favorites or likes.
func (h *MeHandler) RemoveHistoryItem(ctx http2.Context) error {
	_, ok := server.GetClaimsCtx(ctx)
	if !ok {
		return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
	}

	itemID := ctx.Var("id")
	if itemID == "" {
		return server.FailCtx(ctx, server.ErrBadRequest, "item id is required")
	}

	err := h.historyUC.Remove(ctx, itemID)
	if err != nil {
		return server.FailCtx(ctx, server.ErrNotFound, "History item not found")
	}

	return server.OKCtx(ctx, map[string]any{"message": "History item removed"})
}

// GetStats returns the current user's statistics.
func (h *MeHandler) GetStats(ctx http2.Context) error {
	claims, ok := server.GetClaimsCtx(ctx)
	if !ok {
		return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
	}

	// TODO: Implement user stats
	return server.OKCtx(ctx, map[string]any{
		"user_id": claims.GetUserID(),
		"stats":   map[string]any{},
	})
}

// RemoveFavorite removes a favorite by its ID.
func (h *MeHandler) RemoveFavorite(ctx http2.Context) error {
	claims, ok := server.GetClaimsCtx(ctx)
	if !ok {
		return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
	}

	favoriteID := ctx.Var("id")
	if favoriteID == "" {
		return server.FailCtx(ctx, server.ErrBadRequest, "favorite id is required")
	}

	if err := h.likeFavoriteUC.RemoveFavoriteByID(ctx, claims.GetUserID(), favoriteID); err != nil {
		if err.Error() == "favorite not found" {
			return server.FailCtx(ctx, server.ErrNotFound, "Favorite not found")
		}
		return server.FailCtx(ctx, server.ErrInternal, err.Error())
	}

	return server.OKCtx(ctx, map[string]any{"message": "Favorite removed"})
}

// GetFollowers returns users who follow the current user.
func (h *MeHandler) GetFollowers(ctx http2.Context) error {
	claims, ok := server.GetClaimsCtx(ctx)
	if !ok {
		return server.FailCtx(ctx, server.ErrUnauthorized, "unauthorized")
	}

	page, _ := strconv.Atoi(ctx.QueryVarDefault("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.QueryVarDefault("page_size", "20"))
	// Normalize pagination parameters
	page, pageSize = types.NormalizeHTTPPagination(page, pageSize)

	list, total, err := h.userUC.GetSubscribers(
		ctx,
		claims.GetUserID(),
		page,
		pageSize,
	)
	if err != nil {
		return server.FailCtx(ctx, server.ErrInternal, err.Error())
	}

	items := make([]interface{}, 0, len(list))
	for _, u := range list {
		var createdAt string
		if u.CreateTime != nil {
			createdAt = u.CreateTime.AsTime().Format("2006-01-02T15:04:05Z07:00")
		}
		items = append(items, map[string]any{
			"id":            u.Id,
			"user_id":       u.Id,
			"username":      u.Username,
			"nickname":      u.Nickname,
			"avatar":        u.Avatar,
			"subscribed_at": createdAt,
		})
	}

	return server.OKCtx(ctx, map[string]any{
		"items":     items,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}
