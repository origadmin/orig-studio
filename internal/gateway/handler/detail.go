/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package handler

import (
	"context"

	"github.com/origadmin/runtime/log"
	mediav1 "origadmin/application/origcms/api/gen/v1/media"
	userv1 "origadmin/application/origcms/api/gen/v1/user"
	"origadmin/application/origcms/internal/gateway/client"
)

// DetailHandler handles media detail page aggregation.
// Migrated from: internal/svc-portal/biz/portal.go GetVideoDetail
type DetailHandler struct {
	clients *client.Clients
	log     *log.Helper
}

// NewDetailHandler creates a new DetailHandler.
func NewDetailHandler(clients *client.Clients, logger log.Logger) *DetailHandler {
	return &DetailHandler{
		clients: clients,
		log:     log.NewHelper(log.With(logger, "module", "gateway.handler.detail")),
	}
}

// VideoDetailRequest defines the request for a video detail page.
type VideoDetailRequest struct {
	ID int64 `json:"id"`
}

// VideoDetailResponse aggregates media + author + related videos.
type VideoDetailResponse struct {
	Video         *MediaItem   `json:"video"`
	Author        *AuthorInfo  `json:"author"`
	RelatedVideos []*MediaItem `json:"related_videos"`
	TotalViews    int64        `json:"total_views"`
}

// AuthorInfo is a simplified user profile for display.
type AuthorInfo struct {
	ID          int64  `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
}

// GetVideoDetail returns aggregated media detail with author and related videos.
func (h *DetailHandler) GetVideoDetail(ctx context.Context, req *VideoDetailRequest) (*VideoDetailResponse, error) {
	videoResp, err := h.clients.Media.GetMedia(ctx, &mediav1.GetMediaRequest{Id: req.ID})
	if err != nil {
		h.log.Errorf("failed to get media %d: %v", req.ID, err)
		return nil, err
	}

	video := videoResp.Media
	resp := &VideoDetailResponse{
		Video: &MediaItem{
			ID:           video.Id,
			Title:        video.Title,
			Description:  video.Description,
			ThumbnailURL: video.Thumbnail,
			ViewCount:    video.ViewCount,
			AuthorID:     video.UserId,
		},
		TotalViews: video.ViewCount,
	}

	// Fetch author profile (best-effort, non-fatal)
	userResp, err := h.clients.User.GetUser(ctx, &userv1.GetUserRequest{Id: video.UserId})
	if err != nil {
		h.log.Warnf("failed to fetch author for media %d: %v", req.ID, err)
	} else if userResp.User != nil {
		resp.Author = &AuthorInfo{
			ID:          userResp.User.Id,
			Username:    userResp.User.Username,
			DisplayName: userResp.User.Nickname,
			AvatarURL:   userResp.User.Avatar,
		}
	}

	// Fetch related videos (best-effort, non-fatal)
	relatedResp, _ := h.clients.Media.ListMedias(ctx, &mediav1.ListMediasRequest{
		PageSize:   5,
		CategoryId: &video.CategoryId,
	})
	if relatedResp != nil {
		for _, m := range relatedResp.Medias {
			if m.Id == req.ID {
				continue // exclude the current video
			}
			resp.RelatedVideos = append(resp.RelatedVideos, &MediaItem{
				ID:           m.Id,
				Title:        m.Title,
				Description:  m.Description,
				ThumbnailURL: m.Thumbnail,
				ViewCount:    m.ViewCount,
				AuthorID:     m.UserId,
			})
		}
	}

	return resp, nil
}
