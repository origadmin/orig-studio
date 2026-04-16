/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package service

import (
	"context"
	"encoding/json"
	"fmt"
	stdhttp "net/http"
	"strconv"
	"strings"
	"time"

	"github.com/origadmin/runtime/errors"
	"github.com/origadmin/runtime/log"
	"origadmin/application/origcms/api/gen/v1/media"
	"origadmin/application/origcms/api/gen/v1/types"
	"origadmin/application/origcms/internal/helpers/repo"
	"origadmin/application/origcms/internal/svc-media/biz"
	"origadmin/application/origcms/internal/svc-media/dto"
)

type MediaService struct {
	media.UnimplementedMediaServiceServer
	uc  *biz.MediaUseCase
	log *log.Helper
}

func NewMediaService(uc *biz.MediaUseCase, logger log.Logger) *MediaService {
	return &MediaService{
		uc:  uc,
		log: log.NewHelper(log.With(logger, "module", "media.service")),
	}
}

func (s *MediaService) ListMedias(
	ctx context.Context,
	req *media.ListMediasRequest,
) (*media.ListMediasResponse, error) {
	// Create query options from request
	opts := &dto.MediaQueryOption{
		QueryOption: repo.QueryOption{
			Page:     req.Page,
			PageSize: req.PageSize,
			Keyword:  req.Keyword,
		},
		OrderBy:    req.OrderBy,
		Descending: req.Descending,
	}
	
	// Set filters
	if req.Type != nil {
		opts.Type = req.Type
	}
	if req.Status != nil {
		opts.Status = req.Status
	}
	if req.UserId != nil {
		opts.UserID = req.UserId
	}
	if req.CategoryId != nil {
		opts.CategoryID = req.CategoryId
	}
	
	// For short videos, set MediaType filter
	if req.Type != nil && *req.Type == 4 { // Assuming 4 is the type code for short videos
		opts.MediaType = "short_video"
	}
	
	items, total, err := s.uc.ListMedias(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &media.ListMediasResponse{
		Medias:   items,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (s *MediaService) GetMedia(
	ctx context.Context,
	req *media.GetMediaRequest,
) (*media.GetMediaResponse, error) {
	item, err := s.uc.GetMedia(ctx, req.Id)
	if err != nil {
		return nil, errors.NotFound("MEDIA_NOT_FOUND", "Media not found")
	}
	return &media.GetMediaResponse{Media: item}, nil
}

func (s *MediaService) CreateMedia(
	ctx context.Context,
	req *media.CreateMediaRequest,
) (*media.CreateMediaResponse, error) {
	item, err := s.uc.CreateMedia(ctx, req.Media)
	if err != nil {
		return nil, err
	}
	return &media.CreateMediaResponse{Media: item}, nil
}

func (s *MediaService) UpdateMedia(
	ctx context.Context,
	req *media.UpdateMediaRequest,
) (*media.UpdateMediaResponse, error) {
	item, err := s.uc.UpdateMedia(ctx, req.Media)
	if err != nil {
		return nil, err
	}
	return &media.UpdateMediaResponse{Media: item}, nil
}

func (s *MediaService) DeleteMedia(
	ctx context.Context,
	req *media.DeleteMediaRequest,
) (*media.DeleteMediaResponse, error) {
	err := s.uc.DeleteMedia(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &media.DeleteMediaResponse{}, nil
}

func (s *MediaService) IncrementViewCount(
	ctx context.Context,
	req *media.IncrementViewCountRequest,
) (*media.IncrementViewCountResponse, error) {
	count, err := s.uc.IncrementViewCount(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &media.IncrementViewCountResponse{ViewCount: count}, nil
}

func (s *MediaService) ListEncodingTasks(
	ctx context.Context,
	req *media.ListEncodingTasksRequest,
) (*media.ListEncodingTasksResponse, error) {
	tasks, err := s.uc.ListEncodingTasks(ctx, req.MediaId)
	if err != nil {
		return nil, err
	}

	result := make([]*types.EncodingTask, len(tasks))
	for i, t := range tasks {
		result[i] = &types.EncodingTask{
			Id:           t.Id,
			MediaId:      t.MediaId,
			ProfileId:    strconv.Itoa(t.ProfileId),
			Status:       string(t.Status),
			OutputPath:   t.OutputPath,
			ErrorMessage: t.ErrorMessage,
		}
	}
	return &media.ListEncodingTasksResponse{Tasks: result}, nil
}

// GetTranscodingStatus returns the overall encoding status of the system.
func (s *MediaService) GetEncodingStatus(
	ctx context.Context,
	req *media.GetEncodingStatusRequest,
) (*media.GetEncodingStatusResponse, error) {
	status, err := s.uc.GetTranscodingStatus(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &media.GetEncodingStatusResponse{
		ProcessingCount: int32(status.ProcessingCount),
		PendingCount:    int32(status.PendingCount),
		FailedCount:     int32(status.FailedCount),
		SuccessCount:    int32(status.SuccessCount),
		TotalFiltered:   0,
		Page:            req.Page,
		PageSize:        req.PageSize,
		Items:           []*media.TranscodingMediaItem{},
	}, nil
}

// ListEncodeProfiles returns a list of encoding profiles.
func (s *MediaService) ListEncodeProfiles(
	ctx context.Context,
	req *media.ListEncodeProfilesRequest,
) (*media.ListEncodeProfilesResponse, error) {
	profiles, err := s.uc.ListEncodeProfiles(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*types.EncodeProfile, len(profiles))
	for i, p := range profiles {
		result[i] = &types.EncodeProfile{
			Id:          strconv.Itoa(p.Id),
			Name:        p.Name,
			Description: p.Description,
			Extension:   p.Extension,
			Resolution:  p.Resolution,
			VideoCodec:  p.VideoCodec,
			AudioCodec:  p.AudioCodec,
			IsActive:    p.IsActive,
		}
	}
	return &media.ListEncodeProfilesResponse{Profiles: result}, nil
}

// GetEncodeProfile returns an encoding profile by ID.
func (s *MediaService) GetEncodeProfile(
	ctx context.Context,
	req *media.GetEncodeProfileRequest,
) (*media.GetEncodeProfileResponse, error) {
	profileID, _ := strconv.Atoi(req.Id)
	p, err := s.uc.GetEncodeProfile(ctx, profileID)
	if err != nil {
		return nil, err
	}
	return &media.GetEncodeProfileResponse{
		Profile: &types.EncodeProfile{
			Id:          strconv.Itoa(p.Id),
			Name:        p.Name,
			Description: p.Description,
			Extension:   p.Extension,
			Resolution:  p.Resolution,
			VideoCodec:  p.VideoCodec,
			AudioCodec:  p.AudioCodec,
			IsActive:    p.IsActive,
		},
	}, nil
}

// CreateEncodeProfile creates a new encoding profile.
func (s *MediaService) CreateEncodeProfile(
	ctx context.Context,
	req *media.CreateEncodeProfileRequest,
) (*media.CreateEncodeProfileResponse, error) {
	p, err := s.uc.CreateEncodeProfile(ctx, &biz.EncodeProfile{
		Name:        req.Profile.Name,
		Description: req.Profile.Description,
		Extension:   req.Profile.Extension,
		Resolution:  req.Profile.Resolution,
		VideoCodec:  req.Profile.VideoCodec,
		AudioCodec:  req.Profile.AudioCodec,
		IsActive:    req.Profile.IsActive,
	})
	if err != nil {
		return nil, err
	}
	return &media.CreateEncodeProfileResponse{
		Profile: &types.EncodeProfile{
			Id:          strconv.Itoa(p.Id),
			Name:        p.Name,
			Description: p.Description,
			Extension:   p.Extension,
			Resolution:  p.Resolution,
			VideoCodec:  p.VideoCodec,
			AudioCodec:  p.AudioCodec,
			IsActive:    p.IsActive,
		},
	}, nil
}

// UpdateEncodeProfile updates an existing encoding profile.
func (s *MediaService) UpdateEncodeProfile(
	ctx context.Context,
	req *media.UpdateEncodeProfileRequest,
) (*media.UpdateEncodeProfileResponse, error) {
	profileID, _ := strconv.Atoi(req.Profile.Id)
	p, err := s.uc.UpdateEncodeProfile(ctx, &biz.EncodeProfile{
		Id:          profileID,
		Name:        req.Profile.Name,
		Description: req.Profile.Description,
		Extension:   req.Profile.Extension,
		Resolution:  req.Profile.Resolution,
		VideoCodec:  req.Profile.VideoCodec,
		AudioCodec:  req.Profile.AudioCodec,
		IsActive:    req.Profile.IsActive,
	})
	if err != nil {
		return nil, err
	}
	return &media.UpdateEncodeProfileResponse{
		Profile: &types.EncodeProfile{
			Id:          strconv.Itoa(p.Id),
			Name:        p.Name,
			Description: p.Description,
			Extension:   p.Extension,
			Resolution:  p.Resolution,
			VideoCodec:  p.VideoCodec,
			AudioCodec:  p.AudioCodec,
			IsActive:    p.IsActive,
		},
	}, nil
}

// DeleteEncodeProfile deletes an encoding profile.
func (s *MediaService) DeleteEncodeProfile(
	ctx context.Context,
	req *media.DeleteEncodeProfileRequest,
) (*media.DeleteEncodeProfileResponse, error) {
	profileID, _ := strconv.Atoi(req.Id)
	err := s.uc.DeleteEncodeProfile(ctx, profileID)
	if err != nil {
		return nil, err
	}
	return &media.DeleteEncodeProfileResponse{}, nil
}

func (s *MediaService) GetMediaVariants(
	ctx context.Context,
	req *media.GetMediaVariantsRequest,
) (*media.GetMediaVariantsResponse, error) {
	// 直接使用 req.Id 作为 mediaID，因为 ID 是 UUID 格式
	mediaIDStr := req.Id
	
	summary, err := s.uc.GetMediaVariantsByUUID(ctx, mediaIDStr)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, errors.NotFound("MEDIA_NOT_FOUND", "Media not found")
		}
		return nil, err
	}
	
	result := make([]*types.MediaVariant, len(summary.Variants))
	for i, v := range summary.Variants {
		result[i] = &types.MediaVariant{
			Id:         v.TaskID,
			MediaId:    mediaIDStr,
			ProfileId:  strconv.Itoa(v.ProfileID),
			Resolution: v.Resolution,
			Url:        v.OutputPath,
			Size:       0,
			Status:     string(v.Status),
		}
	}
	
	return &media.GetMediaVariantsResponse{Variants: result}, nil
}

// SSEHandler handles Server-Sent Events for transcoding progress.
func (s *MediaService) SSEHandler(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	mediaIdStr := r.URL.Query().Get("media_id")
	var mediaID int64
	fmt.Sscanf(mediaIdStr, "%d", &mediaID)

	flusher, ok := w.(stdhttp.Flusher)
	if !ok {
		stdhttp.Error(w, "Streaming unsupported!", stdhttp.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ctx := r.Context()
	mediaIDStr := strconv.FormatInt(mediaID, 10)
	events, cleanup := s.uc.Subscribe(ctx, mediaIDStr)
	defer cleanup()

	// Keep-alive ticker
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	s.log.Infof("SSE client connected for media %d", mediaID)

	for {
		select {
		case <-ctx.Done():
			s.log.Infof("SSE client disconnected for media %d", mediaID)
			return
		case <-ticker.C:
			fmt.Fprintf(w, "event: ping\ndata: %d\n\n", time.Now().Unix())
			flusher.Flush()
		case ev := <-events:
			if ev == nil {
				return
			}
			fmt.Fprintf(
				w,
				"event: transcoding_progress\ndata: {\"media_id\": %s, \"task_id\": %s, \"status\": \"%s\"}\n\n",
				ev.MediaId,
				ev.Task.Id,
				string(ev.Task.Status),
			)
			flusher.Flush()
		}
	}
}

// TranscodingStatusHTTPHandler handles GET /api/v1/medias/transcoding/status.
// Returns aggregated encoding status counts.
func (s *MediaService) TranscodingStatusHTTPHandler(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	status, err := s.uc.GetTranscodingStatus(r.Context(), nil)
	if err != nil {
		stdhttp.Error(w, err.Error(), stdhttp.StatusInternalServerError)
		return
	}

	resp := map[string]any{
		"processing_count": status.ProcessingCount,
		"pending_count":    status.PendingCount,
		"failed_count":     status.FailedCount,
		"success_count":    status.SuccessCount,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	_ = json.NewEncoder(w).Encode(resp)
}

// RetryTaskHTTPHandler handles POST /api/v1/medias/encoding/retry with query parameter task_id.
// Resets a single failed encoding task to "pending" for re-processing.
// Query params:
//   - task_id (required): the encoding task ID to retry
func (s *MediaService) RetryTaskHTTPHandler(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if r.Method != stdhttp.MethodPost {
		stdhttp.Error(w, "Method not allowed", stdhttp.StatusMethodNotAllowed)
		return
	}

	taskIDStr := r.URL.Query().Get("task_id")
	if taskIDStr == "" {
		writeRetryError(w, "task_id is required", 400)
		return
	}

	var taskID int
	if _, err := fmt.Sscanf(taskIDStr, "%d", &taskID); err != nil || taskID <= 0 {
		writeRetryError(w, "invalid task_id: must be a positive integer", 400)
		return
	}

	task, err := s.uc.RetryTask(r.Context(), taskIDStr)
	if err != nil {
		writeRetryError(w, err.Error(), 422) // Unprocessable Entity
		return
	}

	resp := map[string]any{
		"success": true,
		"task": map[string]any{
			"id":            task.Id,
			"media_id":      task.MediaId,
			"profile_id":    task.ProfileId,
			"status":        task.Status,
			"error_message": task.ErrorMessage,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	_ = json.NewEncoder(w).Encode(resp)
}

// RetryAllFailedHTTPHandler handles POST /api/v1/medias/encoding/retry-all-failed with query parameter media_id.
// Resets all failed tasks for a media back to "pending".
func (s *MediaService) RetryAllFailedHTTPHandler(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if r.Method != stdhttp.MethodPost {
		stdhttp.Error(w, "Method not allowed", stdhttp.StatusMethodNotAllowed)
		return
	}

	mediaIDStr := r.URL.Query().Get("media_id")
	if mediaIDStr == "" {
		writeRetryError(w, "media_id is required", 400)
		return
	}

	count, err := s.uc.RetryAllFailedTasks(r.Context(), mediaIDStr)
	if err != nil {
		writeRetryError(w, err.Error(), 500)
		return
	}

	resp := map[string]any{
		"success":     true,
		"reset_count": count,
		"media_id":    mediaIDStr,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	_ = json.NewEncoder(w).Encode(resp)
}

func writeRetryError(w stdhttp.ResponseWriter, message string, code int) {
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

// MediaVariantsHTTPHandler handles GET /api/v1/medias/{id}/variants
// Returns aggregated transcoding status for a single media, including all variant details.
// This is the API that the "media management" page uses to display transcoding overview.
func (s *MediaService) MediaVariantsHTTPHandler(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	// Extract media ID from URL path: /api/v1/medias/{id}/variants
	path := r.URL.Path
	// Find the positions of "/medias/" and "/variants"
	mediasIndex := strings.Index(path, "/medias/")
	variantsIndex := strings.Index(path, "/variants")
	if mediasIndex == -1 || variantsIndex == -1 || mediasIndex >= variantsIndex {
		writeRetryError(w, "invalid media ID in path", 400)
		return
	}
	// Extract the media ID
	mediaIDStr := path[mediasIndex+8 : variantsIndex]
	if mediaIDStr == "" {
		writeRetryError(w, "invalid media ID in path", 400)
		return
	}

	mediaID, _ := strconv.ParseInt(mediaIDStr, 10, 64)
	summary, err := s.uc.GetMediaVariants(r.Context(), mediaID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeRetryError(w, "media not found", 404)
			return
		}
		writeRetryError(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	_ = json.NewEncoder(w).Encode(summary)
}

// EncodingTasksHTTPHandler handles GET /api/v1/media/encoding/tasks.
// Returns a flat, paginated list of encoding tasks (one row per task).
// Query params:
//   - status: "active" | "processing" | "pending" | "partial" | "failed" | "success" | "all"
//   - page: page number (default 1)
//   - page_size: items per page (default 25, max 100)
//   - media_id: optional filter to a specific media
func (s *MediaService) EncodingTasksHTTPHandler(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	filter := &biz.TranscodingStatusFilter{
		Page:     1,
		PageSize: 25,
	}

	if q := r.URL.Query().Get("status"); q != "" {
		filter.Status = q
	} else {
		filter.Status = "all"
	}
	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			filter.Page = v
		}
	}
	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 100 {
			filter.PageSize = v
		}
	}
	if pr := r.URL.Query().Get("profile"); pr != "" {
		filter.ProfileFilter = pr
	}
	if ch := r.URL.Query().Get("chunk"); ch != "" {
		filter.ChunkFilter = ch
	}
	if se := r.URL.Query().Get("search"); se != "" {
		filter.SearchQuery = se
	}
	if os := r.URL.Query().Get("only_stats"); os == "true" {
		filter.OnlyStats = true
	}

	var mediaID *string
	if m := r.URL.Query().Get("media_id"); m != "" {
		mediaID = &m
	}

	result, err := s.uc.ListEncodingTasksFlat(r.Context(), filter, mediaID)
	if err != nil {
		stdhttp.Error(w, err.Error(), stdhttp.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	_ = json.NewEncoder(w).Encode(result)
}
