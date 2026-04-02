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
	"google.golang.org/protobuf/types/known/timestamppb"
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
	// Not ideal to use biz.MediaQueryOption here, but for simplicity:
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
	return &media.GetTranscodingStatusResponse{
		ProcessingCount: int32(status.ProcessingCount),
		PendingCount:    int32(status.PendingCount),
		FailedCount:     int32(status.FailedCount),
		SuccessCount:    int32(status.SuccessCount),
	}, nil
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
