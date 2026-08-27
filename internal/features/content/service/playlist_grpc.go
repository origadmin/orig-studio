/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Playlist module - gRPC implementation of PlaylistServiceServer
 *
 * PlaylistService (user-scoped playlists) is served by the content microservice,
 * exactly like ChannelService. The gateway bridges it from media -> content so
 * that /api/v1/playlists and /api/v1/me/playlists resolve here instead of
 * returning 501 (Not Implemented) from the media microservice.
 */

package service

import (
	"context"
	"errors"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	mediav1 "origadmin/application/origstudio/api/gen/v1/media"
	types "origadmin/application/origstudio/api/gen/v1/types"
	"origadmin/application/origstudio/internal/dal/entity"
	"origadmin/application/origstudio/internal/features/content/biz"
	"origadmin/application/origstudio/internal/infra/auth"
	systembiz "origadmin/application/origstudio/internal/features/system/biz"
)

// PlaylistServiceServer implements the gRPC PlaylistServiceServer interface.
type PlaylistServiceServer struct {
	mediav1.UnimplementedPlaylistServiceServer
	uc        *biz.PlaylistChannelUseCase
	jwt       *auth.Manager
	settingUC *systembiz.SettingUseCase
	log       *log.Helper
}

// NewPlaylistServiceServer creates a new PlaylistServiceServer.
func NewPlaylistServiceServer(uc *biz.PlaylistChannelUseCase, jwt *auth.Manager, settingUC *systembiz.SettingUseCase, logger log.Logger) *PlaylistServiceServer {
	return &PlaylistServiceServer{
		uc:        uc,
		jwt:       jwt,
		settingUC: settingUC,
		log:       log.NewHelper(log.With(logger, "module", "service/playlist_grpc")),
	}
}

// extractUserID extracts the user ID from the gRPC context.
// It first checks the context value, then falls back to JWT parsing from metadata.
func (s *PlaylistServiceServer) extractUserID(ctx context.Context) string {
	if id, ok := ctx.Value("user_id").(string); ok && id != "" {
		return id
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		authHeaders = md.Get("grpcgateway-authorization")
	}
	for _, header := range authHeaders {
		token := strings.TrimPrefix(header, "Bearer ")
		if token == header {
			continue
		}
		claims, err := s.jwt.Parse(token)
		if err != nil {
			continue
		}
		return claims.GetUserID()
	}
	return ""
}

// bizPlaylistToProto converts a biz.Playlist to a proto types.Playlist.
func bizPlaylistToProto(p *biz.Playlist) *types.Playlist {
	if p == nil {
		return nil
	}
	pp := &types.Playlist{
		Id:          p.ID,
		ShortToken:  p.ShortToken,
		Title:       p.Title,
		Description: p.Description,
		UserId:      p.UserID,
		MediaCount:  int64(len(p.MediaItems)),
	}
	if p.IsPublic {
		pp.Privacy = types.Privacy_PRIVACY_PUBLIC
	} else {
		pp.Privacy = types.Privacy_PRIVACY_PRIVATE
	}
	if !p.CreateTime.IsZero() {
		pp.CreateTime = timestamppb.New(p.CreateTime)
	}
	if !p.UpdateTime.IsZero() {
		pp.UpdateTime = timestamppb.New(p.UpdateTime)
	}
	return pp
}

// playlistMediaItemsToProto converts the playlist's resolved media details into
// proto Media items.
//
// BUG-128 root cause A: GetPlaylistResponse declares `repeated types.Media items`
// but the field was never populated, so the portal playlist detail page rendered
// an empty list even when media_count was non-zero. The repository already loads
// MediaDetails in GetByShortToken/Get, so the data was simply being discarded.
func playlistMediaItemsToProto(items []biz.PlaylistMediaItem) []*types.Media {
	if len(items) == 0 {
		return nil
	}
	out := make([]*types.Media, 0, len(items))
	for i := range items {
		it := items[i]
		m := &types.Media{
			Id:             it.ID,
			ShortToken:     it.ShortToken,
			Title:          it.Title,
			Thumbnail:      it.Thumbnail,
			Duration:       int32(it.Duration),
			Type:           it.Type,
			ViewCount:      it.ViewCount,
			EncodingStatus: it.EncodingStatus,
		}
		if !it.CreateTime.IsZero() {
			m.CreateTime = timestamppb.New(it.CreateTime)
		}
		out = append(out, m)
	}
	return out
}

// playlistMediaMutationError maps use-case errors from playlist media mutations
// onto proper gRPC status codes.
//
// BUG-128 root cause C: every caller used to wrap the raw error in
// codes.Internal, so an unknown media id came back as
// `500 {"message":"entity: constraint failed: pq: ... violates foreign key
// constraint \"content_playlist_media_content_media_media\""}` - both the wrong
// semantics and a leak of the database schema to the public API.
func playlistMediaMutationError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, biz.ErrMediaNotFound):
		return status.Error(codes.NotFound, "media not found")
	case entity.IsNotFound(err):
		return status.Error(codes.NotFound, "playlist not found")
	case strings.Contains(err.Error(), "permission denied"):
		return status.Error(codes.PermissionDenied, "permission denied")
	default:
		return status.Error(codes.Internal, "failed to update playlist media")
	}
}

// mediaOrdersFromIDs converts an ordered media id list into an order map
// (1-based ascending ordering) for the repository.
func mediaOrdersFromIDs(ids []string) map[string]int {
	orders := make(map[string]int, len(ids))
	for i, id := range ids {
		orders[id] = i + 1
	}
	return orders
}

// ========== General (playlist-scoped) endpoints ==========

// GetPlaylists lists all playlists (optionally filtered by user).
func (s *PlaylistServiceServer) GetPlaylists(ctx context.Context, req *mediav1.GetPlaylistsRequest) (*mediav1.GetPlaylistsResponse, error) {
	page, pageSize := normalizePagination(req.Page, req.PageSize)
	var (
		items []*biz.Playlist
		total int
		err   error
	)
	if req.UserId != "" {
		items, total, err = s.uc.ListUserPlaylists(ctx, req.UserId, page, pageSize)
	} else {
		items, total, err = s.uc.ListPlaylists(ctx, page, pageSize)
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	playlists := make([]*types.Playlist, 0, len(items))
	for _, p := range items {
		playlists = append(playlists, bizPlaylistToProto(p))
	}
	return &mediav1.GetPlaylistsResponse{
		Items:    playlists,
		Total:    int32(total),
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}

// GetPlaylist returns a single playlist by short_token.
func (s *PlaylistServiceServer) GetPlaylist(ctx context.Context, req *mediav1.GetPlaylistRequest) (*mediav1.GetPlaylistResponse, error) {
	if req.ShortToken == "" {
		return nil, status.Error(codes.InvalidArgument, "short_token is required")
	}
	p, err := s.uc.GetPlaylistByShortToken(ctx, req.ShortToken)
	if err != nil {
		return nil, status.Error(codes.NotFound, "playlist not found")
	}
	return &mediav1.GetPlaylistResponse{
		Playlist: bizPlaylistToProto(p),
		Items:    playlistMediaItemsToProto(p.MediaDetails),
	}, nil
}

// CreatePlaylist creates a playlist (admin/general scope).
func (s *PlaylistServiceServer) CreatePlaylist(ctx context.Context, req *mediav1.CreatePlaylistRequest) (*mediav1.CreatePlaylistResponse, error) {
	if req.Playlist == nil {
		return nil, status.Error(codes.InvalidArgument, "playlist is required")
	}
	userID := s.extractUserID(ctx)
	if userID == "" {
		userID = req.Playlist.UserId
	}
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	if req.Playlist.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}
	p, err := s.uc.CreatePlaylist(ctx, &biz.Playlist{
		Title:       req.Playlist.Title,
		Description: req.Playlist.Description,
		UserID:      userID,
		IsPublic:    req.Playlist.Privacy != types.Privacy_PRIVACY_PRIVATE,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &mediav1.CreatePlaylistResponse{Playlist: bizPlaylistToProto(p)}, nil
}

// UpdatePlaylist updates a playlist (admin/general scope).
func (s *PlaylistServiceServer) UpdatePlaylist(ctx context.Context, req *mediav1.UpdatePlaylistRequest) (*mediav1.UpdatePlaylistResponse, error) {
	if req.Playlist == nil || req.Playlist.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "playlist.id is required")
	}
	userID := s.extractUserID(ctx)
	p, err := s.uc.UpdatePlaylist(ctx, &biz.Playlist{
		ID:          req.Playlist.Id,
		Title:       req.Playlist.Title,
		Description: req.Playlist.Description,
	}, userID, false)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &mediav1.UpdatePlaylistResponse{Playlist: bizPlaylistToProto(p)}, nil
}

// DeletePlaylist deletes a playlist (admin/general scope).
func (s *PlaylistServiceServer) DeletePlaylist(ctx context.Context, req *mediav1.DeletePlaylistRequest) (*mediav1.DeletePlaylistResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	userID := s.extractUserID(ctx)
	if err := s.uc.DeletePlaylist(ctx, req.Id, userID, false); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &mediav1.DeletePlaylistResponse{Empty: &emptypb.Empty{}}, nil
}

// AddPlaylistMedia adds a media item to a playlist (admin/general scope).
func (s *PlaylistServiceServer) AddPlaylistMedia(ctx context.Context, req *mediav1.AddPlaylistMediaRequest) (*mediav1.AddPlaylistMediaResponse, error) {
	if req.PlaylistId == "" || req.MediaId == "" {
		return nil, status.Error(codes.InvalidArgument, "playlist_id and media_id are required")
	}
	userID := s.extractUserID(ctx)
	if err := s.uc.AddMediaToPlaylist(ctx, req.PlaylistId, req.MediaId, userID, false); err != nil {
		return nil, playlistMediaMutationError(err)
	}
	count := int64(0)
	if p, err := s.uc.GetPlaylist(ctx, req.PlaylistId); err == nil {
		count = int64(len(p.MediaItems))
	}
	return &mediav1.AddPlaylistMediaResponse{MediaCount: count}, nil
}

// ReorderPlaylistMedia reorders media items in a playlist (admin/general scope).
func (s *PlaylistServiceServer) ReorderPlaylistMedia(ctx context.Context, req *mediav1.ReorderPlaylistMediaRequest) (*mediav1.ReorderPlaylistMediaResponse, error) {
	if req.PlaylistId == "" {
		return nil, status.Error(codes.InvalidArgument, "playlist_id is required")
	}
	userID := s.extractUserID(ctx)
	if err := s.uc.ReorderMediaInPlaylist(ctx, req.PlaylistId, mediaOrdersFromIDs(req.MediaIds), userID, false); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &mediav1.ReorderPlaylistMediaResponse{Success: true}, nil
}

// RemovePlaylistMedia removes a media item from a playlist (admin/general scope).
func (s *PlaylistServiceServer) RemovePlaylistMedia(ctx context.Context, req *mediav1.RemovePlaylistMediaRequest) (*mediav1.RemovePlaylistMediaResponse, error) {
	if req.PlaylistId == "" || req.MediaId == "" {
		return nil, status.Error(codes.InvalidArgument, "playlist_id and media_id are required")
	}
	userID := s.extractUserID(ctx)
	if err := s.uc.RemoveMediaFromPlaylist(ctx, req.PlaylistId, req.MediaId, userID, false); err != nil {
		return nil, playlistMediaMutationError(err)
	}
	return &mediav1.RemovePlaylistMediaResponse{Success: true}, nil
}

// ========== My (user-scoped) endpoints ==========

// ListMyPlaylists returns the current user's playlists.
func (s *PlaylistServiceServer) ListMyPlaylists(ctx context.Context, req *mediav1.ListMyPlaylistsRequest) (*mediav1.ListMyPlaylistsResponse, error) {
	userID := s.extractUserID(ctx)
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	page, pageSize := normalizePagination(req.Page, req.PageSize)
	items, total, err := s.uc.ListUserPlaylists(ctx, userID, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	playlists := make([]*types.Playlist, 0, len(items))
	for _, p := range items {
		playlists = append(playlists, bizPlaylistToProto(p))
	}
	return &mediav1.ListMyPlaylistsResponse{
		Items:    playlists,
		Total:    int32(total),
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}

// CreateMyPlaylist creates a playlist for the current user.
// This is the endpoint the watch page calls (POST /api/v1/me/playlists).
func (s *PlaylistServiceServer) CreateMyPlaylist(ctx context.Context, req *mediav1.CreateMyPlaylistRequest) (*mediav1.CreateMyPlaylistResponse, error) {
	userID := s.extractUserID(ctx)
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	if req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}
	isPublic := true
	switch req.Privacy {
	case "PRIVATE", "private":
		isPublic = false
	case "":
		isPublic = true
	default:
		isPublic = req.Privacy != "PRIVATE"
	}
	description := req.Description
	if description == "" {
		description = " "
	}
	p, err := s.uc.CreatePlaylist(ctx, &biz.Playlist{
		Title:       req.Title,
		Description: description,
		UserID:      userID,
		IsPublic:    isPublic,
	})
	if err != nil {
		s.log.Errorf("CreateMyPlaylist failed: userID=%q title=%q err=%v", userID, req.Title, err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &mediav1.CreateMyPlaylistResponse{Playlist: bizPlaylistToProto(p)}, nil
}

// UpdateMyPlaylist updates the current user's playlist.
func (s *PlaylistServiceServer) UpdateMyPlaylist(ctx context.Context, req *mediav1.UpdateMyPlaylistRequest) (*mediav1.UpdateMyPlaylistResponse, error) {
	userID := s.extractUserID(ctx)
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	p, err := s.uc.UpdatePlaylist(ctx, &biz.Playlist{
		ID:          req.Id,
		Title:       req.Title,
		Description: req.Description,
	}, userID, false)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &mediav1.UpdateMyPlaylistResponse{Playlist: bizPlaylistToProto(p)}, nil
}

// DeleteMyPlaylist deletes the current user's playlist.
func (s *PlaylistServiceServer) DeleteMyPlaylist(ctx context.Context, req *mediav1.DeleteMyPlaylistRequest) (*mediav1.DeleteMyPlaylistResponse, error) {
	userID := s.extractUserID(ctx)
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if err := s.uc.DeletePlaylist(ctx, req.Id, userID, false); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &mediav1.DeleteMyPlaylistResponse{Empty: &emptypb.Empty{}}, nil
}

// AddMyPlaylistMedia adds a media item to the current user's playlist.
func (s *PlaylistServiceServer) AddMyPlaylistMedia(ctx context.Context, req *mediav1.AddMyPlaylistMediaRequest) (*mediav1.AddMyPlaylistMediaResponse, error) {
	userID := s.extractUserID(ctx)
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	if req.Id == "" || req.MediaId == "" {
		return nil, status.Error(codes.InvalidArgument, "id and media_id are required")
	}
	if err := s.uc.AddMediaToPlaylist(ctx, req.Id, req.MediaId, userID, false); err != nil {
		return nil, playlistMediaMutationError(err)
	}
	return &mediav1.AddMyPlaylistMediaResponse{Success: true}, nil
}

// RemoveMyPlaylistMedia removes a media item from the current user's playlist.
func (s *PlaylistServiceServer) RemoveMyPlaylistMedia(ctx context.Context, req *mediav1.RemoveMyPlaylistMediaRequest) (*mediav1.RemoveMyPlaylistMediaResponse, error) {
	userID := s.extractUserID(ctx)
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	if req.Id == "" || req.MediaId == "" {
		return nil, status.Error(codes.InvalidArgument, "id and media_id are required")
	}
	if err := s.uc.RemoveMediaFromPlaylist(ctx, req.Id, req.MediaId, userID, false); err != nil {
		return nil, playlistMediaMutationError(err)
	}
	return &mediav1.RemoveMyPlaylistMediaResponse{Success: true}, nil
}

// ReorderMyPlaylistMedia reorders media items in the current user's playlist.
func (s *PlaylistServiceServer) ReorderMyPlaylistMedia(ctx context.Context, req *mediav1.ReorderMyPlaylistMediaRequest) (*mediav1.ReorderMyPlaylistMediaResponse, error) {
	userID := s.extractUserID(ctx)
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if err := s.uc.ReorderMediaInPlaylist(ctx, req.Id, mediaOrdersFromIDs(req.MediaIds), userID, false); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &mediav1.ReorderMyPlaylistMediaResponse{Success: true}, nil
}
