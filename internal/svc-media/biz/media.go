/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"origadmin/application/origcms/api/gen/v1/types" // Import the generated Media type
	"origadmin/application/origcms/internal/svc-media/dto"
	"sync"
)

// EncodingEvent represents a status change event for an encoding task.
type EncodingEvent struct {
	MediaId int64
	Task    *EncodingTask
}

// Media is a wrapper for types.Media for biz layer.
type Media = types.Media

// MediaRepo defines the storage operations for media.
type MediaRepo interface {
	Create(ctx context.Context, media *Media) (*Media, error)
	Get(ctx context.Context, id int64) (*Media, error)
	List(ctx context.Context, opts ...*dto.MediaQueryOption) ([]*Media, int32, error)
	Update(ctx context.Context, media *Media) (*Media, error)
	Delete(ctx context.Context, id int64) error
	IncrementViewCount(ctx context.Context, id int64) (int64, error)
}

// EncodeProfile represents an encoding preset.
type EncodeProfile struct {
	Id           int    `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Extension    string `json:"extension"`
	Resolution   string `json:"resolution"`
	VideoCodec   string `json:"video_codec"`
	VideoBitrate string `json:"video_bitrate"`
	AudioCodec   string `json:"audio_codec"`
	AudioBitrate string `json:"audio_bitrate"`
	IsActive     bool   `json:"is_active"`
}

// EncodeProfileRepo defines the storage operations for encode profiles.
type EncodeProfileRepo interface {
	ListActive(ctx context.Context) ([]*EncodeProfile, error)
	Get(ctx context.Context, id int) (*EncodeProfile, error)
}

// EncodingTask represents a transcoding sub-task for a specific media and profile.
type EncodingTask struct {
	Id           int    `json:"id"`
	MediaId      int64  `json:"media_id"`
	ProfileId    int    `json:"profile_id"`
	Status       string `json:"status"` // pending, processing, success, failed
	Progress     int    `json:"progress"`
	OutputPath   string `json:"output_path"`
	ErrorMessage string `json:"error_message"`
}

// EncodingTaskRepo defines the storage operations for encoding tasks.
type EncodingTaskRepo interface {
	Create(ctx context.Context, task *EncodingTask) (*EncodingTask, error)
	Update(ctx context.Context, task *EncodingTask) (*EncodingTask, error)
	Get(ctx context.Context, id int) (*EncodingTask, error)
	ListByMedia(ctx context.Context, mediaId int64) ([]*EncodingTask, error)
}

type MediaUseCase struct {
	repo         MediaRepo
	profileRepo  EncodeProfileRepo
	encodingRepo EncodingTaskRepo
	log          *log.Helper

	mu   sync.RWMutex
	subs map[int64][]chan *EncodingEvent
}

func NewMediaUseCase(repo MediaRepo, profileRepo EncodeProfileRepo, encodingRepo EncodingTaskRepo, logger log.Logger) *MediaUseCase {
	return &MediaUseCase{
		repo:         repo,
		profileRepo:  profileRepo,
		encodingRepo: encodingRepo,
		log:          log.NewHelper(log.With(logger, "module", "media.biz")),
		subs:         make(map[int64][]chan *EncodingEvent),
	}
}

func (uc *MediaUseCase) GetMedia(ctx context.Context, id int64) (*Media, error) {
	return uc.repo.Get(ctx, id)
}

func (uc *MediaUseCase) ListMedias(ctx context.Context, opts ...*dto.MediaQueryOption) ([]*Media, int32, error) {
	return uc.repo.List(ctx, opts...)
}

func (uc *MediaUseCase) CreateMedia(ctx context.Context, media *Media) (*Media, error) {
	return uc.repo.Create(ctx, media)
}

func (uc *MediaUseCase) UpdateMedia(ctx context.Context, media *Media) (*Media, error) {
	return uc.repo.Update(ctx, media)
}

func (uc *MediaUseCase) DeleteMedia(ctx context.Context, id int64) error {
	return uc.repo.Delete(ctx, id)
}

func (uc *MediaUseCase) IncrementViewCount(ctx context.Context, id int64) (int64, error) {
	return uc.repo.IncrementViewCount(ctx, id)
}

func (uc *MediaUseCase) ListEncodingTasks(ctx context.Context, mediaId int64) ([]*EncodingTask, error) {
	return uc.encodingRepo.ListByMedia(ctx, mediaId)
}

type TranscodingStatus struct {
	ProcessingCount int
	PendingCount    int
	FailedCount     int
	SuccessCount    int
}

func (uc *MediaUseCase) GetTranscodingStatus(ctx context.Context, userID *int64) (*TranscodingStatus, error) {
	// For now, list all tasks to aggregate. In production, this should be a optimized query.
	tasks, err := uc.encodingRepo.ListByMedia(ctx, 0) // 0 implies all if repo supports it, otherwise refine repo
	if err != nil {
		return nil, err
	}

	status := &TranscodingStatus{}
	for _, t := range tasks {
		switch t.Status {
		case "processing":
			status.ProcessingCount++
		case "pending":
			status.PendingCount++
		case "failed":
			status.FailedCount++
		case "success":
			status.SuccessCount++
		}
	}
	return status, nil
}

// Subscribe returns a channel that receives encoding events for a specific media.
func (uc *MediaUseCase) Subscribe(ctx context.Context, mediaID int64) (<-chan *EncodingEvent, func()) {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	ch := make(chan *EncodingEvent, 10)
	uc.subs[mediaID] = append(uc.subs[mediaID], ch)

	cleanup := func() {
		uc.mu.Lock()
		defer uc.mu.Unlock()
		subs := uc.subs[mediaID]
		for i, s := range subs {
			if s == ch {
				uc.subs[mediaID] = append(subs[:i], subs[i+1:]...)
				close(ch)
				break
			}
		}
	}

	return ch, cleanup
}

// Publish sends an encoding event to all subscribers of a media.
func (uc *MediaUseCase) Publish(mediaID int64, event *EncodingEvent) {
	uc.mu.RLock()
	defer uc.mu.RUnlock()

	for _, ch := range uc.subs[mediaID] {
		select {
		case ch <- event:
		default:
			// Buffer full, skip
		}
	}
}
