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

// ProfileHandler handles user public profile page aggregation.
// Migrated from: internal/svc-portal/biz/portal.go GetUserProfile
type ProfileHandler struct {
	clients *client.Clients
	log     *log.Helper
}

// NewProfileHandler creates a new ProfileHandler.
func NewProfileHandler(clients *client.Clients, logger log.Logger) *ProfileHandler {
	return &ProfileHandler{
		clients: clients,
		log:     log.NewHelper(log.With(logger, "module", "gateway.handler.profile")),
	}
}

// UserProfileRequest defines the request for a user public profile page.
type UserProfileRequest struct {
	ID int64 `json:"id"`
}

// UserProfileResponse aggregates user info + their latest published media.
type UserProfileResponse struct {
	User         *AuthorInfo  `json:"user"`
	LatestVideos []*MediaItem `json:"latest_videos"`
	VideoCount   int64        `json:"video_count"`
}

// GetUserProfile returns a user's public profile with their latest videos.
func (h *ProfileHandler) GetUserProfile(ctx context.Context, req *UserProfileRequest) (*UserProfileResponse, error) {
	userResp, err := h.clients.User.GetUser(ctx, &userv1.GetUserRequest{
		Id:          req.ID,
		WithProfile: true,
	})
	if err != nil {
		h.log.Errorf("failed to get user %d: %v", req.ID, err)
		return nil, err
	}

	u := userResp.User
	resp := &UserProfileResponse{
		User: &AuthorInfo{
			ID:          u.Id,
			Username:    u.Username,
			DisplayName: u.Nickname,
			AvatarURL:   u.Avatar,
		},
	}

	// Fetch user's latest published media (best-effort)
	mediaResp, _ := h.clients.Media.ListMedias(ctx, &mediav1.ListMediasRequest{
		UserId:     &req.ID,
		PageSize:   10,
		OrderBy:    "created_at",
		Descending: true,
	})
	if mediaResp != nil {
		resp.VideoCount = int64(mediaResp.Total)
		for _, m := range mediaResp.Medias {
			resp.LatestVideos = append(resp.LatestVideos, &MediaItem{
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
