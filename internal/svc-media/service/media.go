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
	"origadmin/application/origcms/internal/svc-media/biz"
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
	items, total, err := s.uc.ListMedias(ctx)
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
			Id:           int32(t.Id),
			MediaId:      t.MediaId,
			ProfileId:    int32(t.ProfileId),
			Status:       t.Status,
			Progress:     int32(t.Progress),
			OutputPath:   t.OutputPath,
			ErrorMessage: t.ErrorMessage,
		}
	}
	return &media.ListEncodingTasksResponse{Tasks: result}, nil
}

func (s *MediaService) GetTranscodingStatus(
	ctx context.Context,
	req *media.GetTranscodingStatusRequest,
) (*media.GetTranscodingStatusResponse, error) {
	status, err := s.uc.GetTranscodingStatus(ctx, nil)
	if err != nil {
		return nil, err
	}

	// Convert items
	items := make([]*media.TranscodingMediaItem, len(status.Items))
	for i, item := range status.Items {
		tasks := make([]*types.EncodingTask, len(item.Tasks))
		for j, t := range item.Tasks {
			tasks[j] = &types.EncodingTask{
				Id:           int32(t.Id),
				MediaId:      t.MediaId,
				ProfileId:    int32(t.ProfileId),
				Status:       t.Status,
				Progress:     int32(t.Progress),
				OutputPath:   t.OutputPath,
				ErrorMessage: t.ErrorMessage,
			}
		}

		items[i] = &media.TranscodingMediaItem{
			Media: item.Media,
			Tasks: tasks,
		}
	}

	return &media.GetTranscodingStatusResponse{
		ProcessingCount: int32(status.ProcessingCount),
		PendingCount:    int32(status.PendingCount),
		FailedCount:     int32(status.FailedCount),
		SuccessCount:    int32(status.SuccessCount),
		Items:           items,
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
			Id:          int32(p.Id),
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
	p, err := s.uc.GetEncodeProfile(ctx, int(req.Id))
	if err != nil {
		return nil, err
	}
	return &media.GetEncodeProfileResponse{
		Profile: &types.EncodeProfile{
			Id:          int32(p.Id),
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
			Id:          int32(p.Id),
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
	p, err := s.uc.UpdateEncodeProfile(ctx, &biz.EncodeProfile{
		Id:          int(req.Profile.Id),
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
			Id:          int32(p.Id),
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
	err := s.uc.DeleteEncodeProfile(ctx, int(req.Id))
	if err != nil {
		return nil, err
	}
	return &media.DeleteEncodeProfileResponse{}, nil
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
	events, cleanup := s.uc.Subscribe(ctx, mediaID)
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
				"event: transcoding_progress\ndata: {\"media_id\": %d, \"task_id\": %d, \"status\": \"%s\", \"progress\": %d}\n\n",
				ev.MediaId,
				ev.Task.Id,
				ev.Task.Status,
				ev.Task.Progress,
			)
			flusher.Flush()
		}
	}
}

// TranscodingStatusHTTPHandler handles GET /api/v1/media/transcoding/status with query parameters.
// This bypasses the gRPC gateway to properly pass status/page/page_size from query string.
func (s *MediaService) TranscodingStatusHTTPHandler(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	filter := &biz.TranscodingStatusFilter{
		Page:     1,
		PageSize: 20,
	}

	if q := r.URL.Query().Get("status"); q != "" {
		filter.Status = q
	} else {
		filter.Status = "active"
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

	status, err := s.uc.GetTranscodingStatus(r.Context(), filter)
	if err != nil {
		stdhttp.Error(w, err.Error(), stdhttp.StatusInternalServerError)
		return
	}

	// Convert to response
	items := make([]*media.TranscodingMediaItem, len(status.Items))
	for i, item := range status.Items {
		tasks := make([]*types.EncodingTask, len(item.Tasks))
		for j, t := range item.Tasks {
			tasks[j] = &types.EncodingTask{
				Id:           int32(t.Id),
				MediaId:      t.MediaId,
				ProfileId:    int32(t.ProfileId),
				Status:       t.Status,
				Progress:     int32(t.Progress),
				OutputPath:   t.OutputPath,
				ErrorMessage: t.ErrorMessage,
			}
		}
		items[i] = &media.TranscodingMediaItem{
			Media: item.Media,
			Tasks: tasks,
		}
	}

	resp := map[string]any{
		"processing_count": status.ProcessingCount,
		"pending_count":    status.PendingCount,
		"failed_count":     status.FailedCount,
		"success_count":    status.SuccessCount,
		"total_filtered":   status.TotalFiltered,
		"page":             status.Page,
		"page_size":        status.PageSize,
		"items":            items,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	_ = json.NewEncoder(w).Encode(resp)
}

// RetryTaskHTTPHandler handles POST /api/v1/media/retry with query parameter task_id.
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

	task, err := s.uc.RetryTask(r.Context(), taskID)
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
			"progress":      task.Progress,
			"error_message": task.ErrorMessage,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	_ = json.NewEncoder(w).Encode(resp)
}

// RetryAllFailedHTTPHandler handles POST /api/v1/media/retry-all-failed with query parameter media_id.
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

	var mediaID int64
	fmt.Sscanf(mediaIDStr, "%d", &mediaID)

	count, err := s.uc.RetryAllFailedTasks(r.Context(), mediaID)
	if err != nil {
		writeRetryError(w, err.Error(), 500)
		return
	}

	resp := map[string]any{
		"success":     true,
		"reset_count": count,
		"media_id":    mediaID,
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

// MediaVariantsHTTPHandler handles GET /api/v1/media/{id}/variants
// Returns aggregated transcoding status for a single media, including all variant details.
// This is the API that the "media management" page uses to display transcoding overview.
func (s *MediaService) MediaVariantsHTTPHandler(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	// Extract media ID from URL path: /api/v1/media/{id}/variants
	var mediaID int64
	_, err := fmt.Sscanf(r.URL.Path, "/api/v1/media/%d/variants", &mediaID)
	if err != nil || mediaID <= 0 {
		writeRetryError(w, "invalid media ID in path", 400)
		return
	}

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

	var mediaID *int64
	if m := r.URL.Query().Get("media_id"); m != "" {
		var id int64
		if _, err := fmt.Sscanf(m, "%d", &id); err == nil && id > 0 {
			mediaID = &id
		}
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
