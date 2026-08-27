package service

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/origadmin/runtime/errors"
	"github.com/origadmin/runtime/log"
	"google.golang.org/grpc/metadata"

	media "origadmin/application/origstudio/api/gen/v1/media"
	"origadmin/application/origstudio/api/gen/v1/types"
	repotypes "origadmin/application/origstudio/internal/domain/types"
	"origadmin/application/origstudio/internal/features/media/biz"
	"origadmin/application/origstudio/internal/features/media/dto"
	"origadmin/application/origstudio/internal/infra/auth"
)

const spriteGenerateTimeout = 30 * time.Minute

type AdminMediaService struct {
	media.UnimplementedAdminMediaServiceServer
	uc  *biz.MediaUseCase
	jwt *auth.Manager
	log *log.Helper
}

func NewAdminMediaService(uc *biz.MediaUseCase, jwt *auth.Manager, logger log.Logger) *AdminMediaService {
	return &AdminMediaService{
		uc:  uc,
		jwt: jwt,
		log: log.NewHelper(log.With(logger, "module", "media.service.admin-media")),
	}
}

// currentUserID resolves the authenticated user id from the JWT carried in
// gRPC incoming metadata (authorization / grpcgateway-authorization headers).
func (s *AdminMediaService) currentUserID(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	for _, key := range []string{"authorization", "grpcgateway-authorization"} {
		for _, header := range md.Get(key) {
			token := strings.TrimPrefix(header, "Bearer ")
			if token == header {
				continue
			}
			claims, err := s.jwt.Parse(token)
			if err == nil && claims != nil {
				return claims.GetUserID()
			}
		}
	}
	return ""
}

func (s *AdminMediaService) ListAdminMedias(ctx context.Context, req *media.ListAdminMediasRequest) (*media.ListAdminMediasResponse, error) {
	page, pageSize := repotypes.NormalizePagination(int(req.Page), int(req.PageSize))

	opts := &dto.MediaQueryOption{
		QueryOption: repotypes.QueryOption{
			Page:     int32(page),
			PageSize: int32(pageSize),
			Keyword:  req.Keyword,
		},
		AdminMode: true,
	}
	if req.State != "" {
		opts.State = req.State
	}
	if req.ReviewStatus != "" {
		opts.ReviewStatus = &req.ReviewStatus
	}
	// Media type filtering: proto field "type" maps to HTTP query param "type".
	// Empty/"all" = all types. Do NOT default to "video" here: the admin media
	// page's default video-only view is controlled by the frontend explicitly
	// sending type=video. Defaulting server-side to "video" silently hid
	// pending_review items whose type is image/file from the review filter
	// (BUG-233: review_status=pending_review returned 0 while DB had 4).
	if req.Type != "" && req.Type != "all" {
		opts.MediaType = req.Type
	}

	items, total, err := s.uc.ListMedias(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &media.ListAdminMediasResponse{
		Total:    total,
		Items:    items,
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}

func (s *AdminMediaService) GetAdminMedia(ctx context.Context, req *media.GetAdminMediaRequest) (*media.GetAdminMediaResponse, error) {
	item, err := s.uc.GetMedia(ctx, req.Id)
	if err != nil {
		return nil, errors.NotFound("MEDIA_NOT_FOUND", "media not found")
	}
	return &media.GetAdminMediaResponse{Media: item}, nil
}

func (s *AdminMediaService) UpdateMediaState(ctx context.Context, req *media.UpdateMediaStateRequest) (*media.UpdateMediaStateResponse, error) {
	err := s.uc.UpdateMediaState(ctx, req.Id, req.State)
	if err != nil {
		return nil, err
	}
	item, err := s.uc.GetMedia(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &media.UpdateMediaStateResponse{Media: item}, nil
}

func (s *AdminMediaService) GetMediaAdminStats(ctx context.Context, req *media.GetMediaAdminStatsRequest) (*media.GetMediaAdminStatsResponse, error) {
	item, err := s.uc.GetMedia(ctx, req.Id)
	if err != nil {
		return nil, errors.NotFound("MEDIA_NOT_FOUND", "media not found")
	}
	return &media.GetMediaAdminStatsResponse{
		ViewCount:     item.ViewCount,
		LikeCount:     item.LikeCount,
		CommentCount:  item.CommentCount,
		FavoriteCount: item.FavoriteCount,
	}, nil
}

func (s *AdminMediaService) GetAdminMediaVariants(ctx context.Context, req *media.GetAdminMediaVariantsRequest) (*media.GetAdminMediaVariantsResponse, error) {
	summary, err := s.uc.GetMediaVariantsByUUID(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	result := make([]*types.MediaVariant, len(summary.Variants))
	for i, v := range summary.Variants {
		result[i] = &types.MediaVariant{
			Id:         v.TaskID,
			MediaId:    req.Id,
			ProfileId:  strconv.Itoa(v.ProfileID),
			Resolution: v.Resolution,
			Url:        v.OutputPath,
			Status:     string(v.Status),
		}
	}
	return &media.GetAdminMediaVariantsResponse{Variants: result}, nil
}

func (s *AdminMediaService) ListAdminMediaTasks(ctx context.Context, req *media.ListAdminMediaTasksRequest) (*media.ListAdminMediaTasksResponse, error) {
	tasks, err := s.uc.ListEncodingTasks(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	result := make([]*types.EncodingTask, len(tasks))
	for i, t := range tasks {
		result[i] = &types.EncodingTask{
			Id:           t.Id,
			MediaId:      t.MediaId,
			ProfileId:    int64(t.ProfileId),
			Status:       string(t.Status),
			OutputPath:   t.OutputPath,
			ErrorMessage: t.ErrorMessage,
		}
	}
	return &media.ListAdminMediaTasksResponse{Tasks: result}, nil
}

func (s *AdminMediaService) RetryAdminMediaTask(ctx context.Context, req *media.RetryAdminMediaTaskRequest) (*media.RetryAdminMediaTaskResponse, error) {
	task, err := s.uc.RetryTask(ctx, req.TaskId)
	if err != nil {
		return nil, err
	}
	return &media.RetryAdminMediaTaskResponse{
		Task: &types.EncodingTask{
			Id:           task.Id,
			MediaId:      task.MediaId,
			ProfileId:    int64(task.ProfileId),
			Status:       string(task.Status),
			OutputPath:   task.OutputPath,
			ErrorMessage: task.ErrorMessage,
		},
	}, nil
}

func (s *AdminMediaService) ReviewMedia(ctx context.Context, req *media.ReviewMediaRequest) (*media.ReviewMediaResponse, error) {
	// BUG-242: gRPC-gateway JSON unmarshalling for the {status, reason} payload
	// previously set req.Status to "" for the approve path (the proto3 zero
	// value), so the old "== approved || reviewed" check failed and the
	// media was incorrectly routed to the reject branch. Treat any non-"rejected"
	// value (including "") as approve; the explicit "rejected" still wins.
	approve := req.Status != "rejected"
	// BUG-233: reviewerID must come from the JWT claims. Passing "" previously
	// violated the content_media_review_logs.reviewer_id FK (REFERENCES users(id))
	// so the audit log insert was rejected while the state update succeeded.
	reviewerID := s.currentUserID(ctx)
	_, err := s.uc.ReviewMedia(ctx, req.Id, approve, req.Reason, reviewerID)
	if err != nil {
		return nil, err
	}
	return &media.ReviewMediaResponse{
		Review: &types.Review{
			Id:       req.Id,
			MediaId:  req.Id,
			Status:   req.Status,
			Reason:   req.Reason,
			Notes:    req.Notes,
		},
	}, nil
}

func (s *AdminMediaService) GetMediaReviewLogs(ctx context.Context, req *media.GetMediaReviewLogsRequest) (*media.GetMediaReviewLogsResponse, error) {
	logs, err := s.uc.ListReviewLogs(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	result := make([]*types.Review, len(logs))
	for i, l := range logs {
		result[i] = &types.Review{
			Id:         l.ID,
			MediaId:    l.MediaID,
			ReviewerId: l.ReviewerID,
			Status:     l.NewStatus,
			Reason:     l.Comment,
			Notes:      l.Action,
		}
	}
	return &media.GetMediaReviewLogsResponse{Logs: result}, nil
}

func (s *AdminMediaService) RegenerateSprite(ctx context.Context, req *media.RegenerateSpriteRequest) (*media.RegenerateSpriteResponse, error) {
	// Run sprite generation asynchronously to avoid HTTP timeout.
	// The HTTP server has a 10-second timeout, but 4K video sprite generation
	// can take several minutes. Use context.Background() so the ffmpeg
	// subprocess is not killed when the HTTP request context is cancelled.
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), spriteGenerateTimeout)
		defer cancel()
		if err := s.uc.RegenerateSprite(bgCtx, req.Id); err != nil {
			s.log.Warnf("async sprite regeneration failed for media %s: %v", req.Id, err)
		}
	}()
	return &media.RegenerateSpriteResponse{Success: true}, nil
}

func (s *AdminMediaService) RegenerateThumbnail(ctx context.Context, req *media.RegenerateThumbnailRequest) (*media.RegenerateThumbnailResponse, error) {
	// Run with a background-derived context so the (potentially long) FFmpeg
	// extraction is not killed when the HTTP request context is cancelled
	// (e.g. client disconnect or reverse-proxy timeout). This is the real
	// microservice execution site for thumbnail regeneration - the gin
	// SpriteHandler in the CE monolith does not run here.
	bgCtx, cancel := context.WithTimeout(context.Background(), spriteGenerateTimeout)
	defer cancel()
	if err := s.uc.RegenerateThumbnail(bgCtx, req.Id, req.ThumbnailTime); err != nil {
		return nil, err
	}
	return &media.RegenerateThumbnailResponse{Success: true}, nil
}

func (s *AdminMediaService) UpdateAdminMedia(ctx context.Context, req *media.UpdateAdminMediaRequest) (*media.UpdateAdminMediaResponse, error) {
	existing, err := s.uc.GetByID(ctx, req.Id)
	if err != nil {
		return nil, errors.NotFound("MEDIA_NOT_FOUND", "media not found")
	}

	if req.Title != "" {
		existing.Title = req.Title
	}
	if req.Description != "" {
		existing.Description = req.Description
	}
	if req.CategoryId != 0 {
		existing.CategoryId = req.CategoryId
	}
	if req.Tags != nil {
		existing.Tags = req.Tags
	}

	// Persist the remaining editable fields. Previously the handler only
	// copied title/description/category_id/tags, so state/privacy/featured/
	// comments/download/listable submitted by the admin edit page were
	// silently dropped on save. The DAL already writes these columns.
	if req.State != "" {
		existing.State = req.State
	}
	if req.Privacy != 0 {
		existing.Privacy = types.Privacy(req.Privacy)
	}
	existing.Featured = req.Featured
	existing.EnableComments = req.EnableComments
	existing.AllowDownload = req.AllowDownload
	existing.Listable = req.Listable

	updated, err := s.uc.UpdateMedia(ctx, existing)
	if err != nil {
		return nil, err
	}

	// BUG-105: channel assignment (admin may assign / move A->B / clear). Only
	// act when the value actually changed. '' clears the assignment via the
	// dedicated UpdateMediaChannel path (generic Update cannot clear).
	if req.ChannelId != existing.ChannelId {
		if req.ChannelId != "" {
			ownerID, err := s.uc.GetChannelOwnerID(ctx, req.ChannelId)
			if err != nil {
				s.log.Errorf("Failed to resolve channel owner: %v", err)
				return nil, errors.InternalServer("RESOLVE_CHANNEL_OWNER_FAILED", "Failed to resolve channel owner")
			}
			if ownerID == "" {
				return nil, errors.BadRequest("CHANNEL_NOT_FOUND", "target channel does not exist")
			}
		}
		if err := s.uc.UpdateMediaChannel(ctx, existing.Id, req.ChannelId); err != nil {
			s.log.Errorf("Failed to update media channel: %v", err)
			return nil, errors.InternalServer("UPDATE_MEDIA_CHANNEL_FAILED", "Failed to update media channel")
		}
		// Keep the returned resource in sync with the dedicated channel write.
		updated.ChannelId = req.ChannelId
	}

	return &media.UpdateAdminMediaResponse{Media: updated}, nil
}

func (s *AdminMediaService) DeleteAdminMedia(ctx context.Context, req *media.DeleteAdminMediaRequest) (*media.DeleteAdminMediaResponse, error) {
	if err := s.uc.DeleteMedia(ctx, req.Id); err != nil {
		return nil, err
	}
	return &media.DeleteAdminMediaResponse{}, nil
}

func (s *AdminMediaService) GetTranscodingEvents(req *media.GetTranscodingEventsRequest, stream media.AdminMediaService_GetTranscodingEventsServer) error {
	ctx := stream.Context()
	events, cleanup := s.uc.Subscribe(ctx, "")
	defer cleanup()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case ev, ok := <-events:
			if !ok || ev == nil {
				return nil
			}
			if err := stream.Send(&media.TranscodingEvent{
				MediaId: ev.MediaId,
				TaskId:  ev.Task.Id,
				Status:  string(ev.Task.Status),
			}); err != nil {
				return err
			}
		}
	}
}
