/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package service

import (
	"context"
	"fmt"
	stdhttp "net/http"
	"time"

	"origadmin/application/origcms/api/gen/v1/media"
	"origadmin/application/origcms/api/gen/v1/types"
	"origadmin/application/origcms/internal/svc-media/biz"

	"github.com/origadmin/runtime/errors"
	"github.com/origadmin/runtime/log"
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

func (s *MediaService) ListMedias(ctx context.Context, req *media.ListMediasRequest) (*media.ListMediasResponse, error) {
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

func (s *MediaService) GetMedia(ctx context.Context, req *media.GetMediaRequest) (*media.GetMediaResponse, error) {
	item, err := s.uc.GetMedia(ctx, req.Id)
	if err != nil {
		return nil, errors.NotFound("MEDIA_NOT_FOUND", "Media not found")
	}
	return &media.GetMediaResponse{Media: item}, nil
}

func (s *MediaService) CreateMedia(ctx context.Context, req *media.CreateMediaRequest) (*media.CreateMediaResponse, error) {
	item, err := s.uc.CreateMedia(ctx, req.Media)
	if err != nil {
		return nil, err
	}
	return &media.CreateMediaResponse{Media: item}, nil
}

func (s *MediaService) UpdateMedia(ctx context.Context, req *media.UpdateMediaRequest) (*media.UpdateMediaResponse, error) {
	item, err := s.uc.UpdateMedia(ctx, req.Media)
	if err != nil {
		return nil, err
	}
	return &media.UpdateMediaResponse{Media: item}, nil
}

func (s *MediaService) DeleteMedia(ctx context.Context, req *media.DeleteMediaRequest) (*media.DeleteMediaResponse, error) {
	err := s.uc.DeleteMedia(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &media.DeleteMediaResponse{}, nil
}

func (s *MediaService) IncrementViewCount(ctx context.Context, req *media.IncrementViewCountRequest) (*media.IncrementViewCountResponse, error) {
	count, err := s.uc.IncrementViewCount(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &media.IncrementViewCountResponse{ViewCount: count}, nil
}

func (s *MediaService) ListEncodingTasks(ctx context.Context, req *media.ListEncodingTasksRequest) (*media.ListEncodingTasksResponse, error) {
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

func (s *MediaService) GetTranscodingStatus(ctx context.Context, req *media.GetTranscodingStatusRequest) (*media.GetTranscodingStatusResponse, error) {
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
func (s *MediaService) ListEncodeProfiles(ctx context.Context, req *media.ListEncodeProfilesRequest) (*media.ListEncodeProfilesResponse, error) {
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
func (s *MediaService) GetEncodeProfile(ctx context.Context, req *media.GetEncodeProfileRequest) (*media.GetEncodeProfileResponse, error) {
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
func (s *MediaService) CreateEncodeProfile(ctx context.Context, req *media.CreateEncodeProfileRequest) (*media.CreateEncodeProfileResponse, error) {
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
func (s *MediaService) UpdateEncodeProfile(ctx context.Context, req *media.UpdateEncodeProfileRequest) (*media.UpdateEncodeProfileResponse, error) {
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
func (s *MediaService) DeleteEncodeProfile(ctx context.Context, req *media.DeleteEncodeProfileRequest) (*media.DeleteEncodeProfileResponse, error) {
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
			fmt.Fprintf(w, "event: transcoding_progress\ndata: {\"media_id\": %d, \"task_id\": %d, \"status\": \"%s\", \"progress\": %d}\n\n",
				ev.MediaId, ev.Task.Id, ev.Task.Status, ev.Task.Progress)
			flusher.Flush()
		}
	}
}
