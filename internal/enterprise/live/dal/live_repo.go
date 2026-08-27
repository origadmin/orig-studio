package dal

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origstudio/internal/dal/entity"
	"origadmin/application/origstudio/internal/dal/entity/liveroom"
	"origadmin/application/origstudio/internal/enterprise/live/biz"
	"origadmin/application/origstudio/internal/enterprise/live/dto"
)

type repo struct {
	db  *entity.Client
	log *log.Helper
}

func NewRepo(db *entity.Client, logger log.Logger) biz.Repo {
	return &repo{
		db:  db,
		log: log.NewHelper(log.With(logger, "module", "enterprise/live.repo")),
	}
}

func entityToLiveRoomDTO(e *entity.LiveRoom) *dto.LiveRoomDTO {
	if e == nil {
		return nil
	}
	return &dto.LiveRoomDTO{
		ID:             e.ID,
		Title:          e.Title,
		Description:    e.Description,
		StreamKey:      e.StreamKey,
		RtmpURL:        e.RtmpURL,
		HlsURL:         e.HlsURL,
		Status:         string(e.Status),
		ScheduledAt:    e.ScheduledAt,
		StartedAt:      e.StartedAt,
		EndedAt:        e.EndedAt,
		MaxViewers:     e.MaxViewers,
		CurrentViewers: e.CurrentViewers,
		PeakViewers:    e.PeakViewers,
		Thumbnail:      e.Thumbnail,
		Category:       e.Category,
		Tags:           e.Tags,
		UserID:         e.UserID,
		CreateTime:     e.CreateTime,
		UpdateTime:     e.UpdateTime,
	}
}

func (r *repo) ListLiveRooms(ctx context.Context, page, pageSize int) ([]*dto.LiveRoomDTO, int, error) {
	query := r.db.LiveRoom.Query()
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count live rooms: %w", err)
	}
	items, err := query.
		Order(entity.Desc(liveroom.FieldCreateTime)).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list live rooms: %w", err)
	}
	result := make([]*dto.LiveRoomDTO, len(items))
	for i, item := range items {
		result[i] = entityToLiveRoomDTO(item)
	}
	return result, total, nil
}

func (r *repo) GetLiveRoomByID(ctx context.Context, id string) (*dto.LiveRoomDTO, error) {
	item, err := r.db.LiveRoom.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get live room: %w", err)
	}
	return entityToLiveRoomDTO(item), nil
}

func (r *repo) CreateLiveRoom(ctx context.Context, room *dto.LiveRoomDTO) (*dto.LiveRoomDTO, error) {
	builder := r.db.LiveRoom.Create().
		SetTitle(room.Title).
		SetStatus(liveroom.StatusIdle)
	if room.Description != "" {
		builder.SetDescription(room.Description)
	}
	if room.RtmpURL != "" {
		builder.SetRtmpURL(room.RtmpURL)
	}
	if room.HlsURL != "" {
		builder.SetHlsURL(room.HlsURL)
	}
	if room.MaxViewers > 0 {
		builder.SetMaxViewers(room.MaxViewers)
	}
	if room.Thumbnail != "" {
		builder.SetThumbnail(room.Thumbnail)
	}
	if room.Category != "" {
		builder.SetCategory(room.Category)
	}
	if len(room.Tags) > 0 {
		builder.SetTags(room.Tags)
	}
	if room.UserID != "" {
		builder.SetUserID(room.UserID)
	}
	if !room.ScheduledAt.IsZero() {
		builder.SetScheduledAt(room.ScheduledAt)
	}
	ent, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create live room: %w", err)
	}
	return entityToLiveRoomDTO(ent), nil
}

func (r *repo) UpdateLiveRoom(ctx context.Context, room *dto.LiveRoomDTO) (*dto.LiveRoomDTO, error) {
	builder := r.db.LiveRoom.UpdateOneID(room.ID).
		SetTitle(room.Title)
	if room.Description != "" {
		builder.SetDescription(room.Description)
	}
	builder.SetRtmpURL(room.RtmpURL)
	builder.SetHlsURL(room.HlsURL)
	if room.MaxViewers > 0 {
		builder.SetMaxViewers(room.MaxViewers)
	}
	builder.SetThumbnail(room.Thumbnail)
	builder.SetCategory(room.Category)
	if len(room.Tags) > 0 {
		builder.SetTags(room.Tags)
	}
	if !room.ScheduledAt.IsZero() {
		builder.SetScheduledAt(room.ScheduledAt)
	}
	ent, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update live room: %w", err)
	}
	return entityToLiveRoomDTO(ent), nil
}

func (r *repo) DeleteLiveRoom(ctx context.Context, id string) error {
	return r.db.LiveRoom.DeleteOneID(id).Exec(ctx)
}

func (r *repo) UpdateStatus(ctx context.Context, id string, status string) (*dto.LiveRoomDTO, error) {
	builder := r.db.LiveRoom.UpdateOneID(id).
		SetStatus(liveroom.Status(status))
	if status == "live" {
		now := time.Now()
		builder.SetStartedAt(now)
	}
	if status == "ended" {
		now := time.Now()
		builder.SetEndedAt(now)
	}
	ent, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update live room status: %w", err)
	}
	return entityToLiveRoomDTO(ent), nil
}

func (r *repo) ListPublicLiveRooms(ctx context.Context) ([]*dto.LiveRoomDTO, error) {
	items, err := r.db.LiveRoom.Query().
		Where(liveroom.StatusEQ(liveroom.StatusLive)).
		Order(entity.Desc(liveroom.FieldCurrentViewers)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list public live rooms: %w", err)
	}
	result := make([]*dto.LiveRoomDTO, len(items))
	for i, item := range items {
		room := entityToLiveRoomDTO(item)
		room.StreamKey = ""
		result[i] = room
	}
	return result, nil
}