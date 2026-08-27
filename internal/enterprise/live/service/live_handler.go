package service

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	livev1 "origadmin/application/origstudio/api/gen/v1/live"
	"origadmin/application/origstudio/internal/enterprise/live/biz"
	"origadmin/application/origstudio/internal/enterprise/live/dto"
)

type LiveService struct {
	livev1.UnimplementedLiveServiceServer
	uc  *biz.UseCase
	log *log.Helper
}

func NewLiveService(uc *biz.UseCase, logger log.Logger) *LiveService {
	return &LiveService{
		uc:  uc,
		log: log.NewHelper(log.With(logger, "module", "enterprise.live.service")),
	}
}

func dtoToLiveRoom(d *dto.LiveRoomDTO) *livev1.LiveRoom {
	if d == nil {
		return nil
	}
	status := livev1.LiveRoomStatus_LIVE_ROOM_STATUS_UNSPECIFIED
	switch d.Status {
	case "scheduled", "idle":
		status = livev1.LiveRoomStatus_LIVE_ROOM_STATUS_SCHEDULED
	case "live":
		status = livev1.LiveRoomStatus_LIVE_ROOM_STATUS_LIVE
	case "ended":
		status = livev1.LiveRoomStatus_LIVE_ROOM_STATUS_ENDED
	}
	return &livev1.LiveRoom{
		Id:             d.ID,
		Title:          d.Title,
		Description:    d.Description,
		StreamKey:      d.StreamKey,
		RtmpUrl:        d.RtmpURL,
		HlsUrl:         d.HlsURL,
		Status:         status,
		ScheduledAt:    safeTimestamp(d.ScheduledAt),
		StartedAt:      safeTimestamp(d.StartedAt),
		EndedAt:        safeTimestamp(d.EndedAt),
		MaxViewers:     int32(d.MaxViewers),
		CurrentViewers: int32(d.CurrentViewers),
		PeakViewers:    int32(d.PeakViewers),
		Thumbnail:      d.Thumbnail,
		Category:       d.Category,
		Tags:           d.Tags,
		UserId:         d.UserID,
		CreateTime:     safeTimestamp(d.CreateTime),
		UpdateTime:     safeTimestamp(d.UpdateTime),
	}
}

func dtoToPublicLiveRoom(d *dto.LiveRoomDTO) *livev1.PublicLiveRoom {
	if d == nil {
		return nil
	}
	status := livev1.LiveRoomStatus_LIVE_ROOM_STATUS_UNSPECIFIED
	switch d.Status {
	case "scheduled", "idle":
		status = livev1.LiveRoomStatus_LIVE_ROOM_STATUS_SCHEDULED
	case "live":
		status = livev1.LiveRoomStatus_LIVE_ROOM_STATUS_LIVE
	case "ended":
		status = livev1.LiveRoomStatus_LIVE_ROOM_STATUS_ENDED
	}
	return &livev1.PublicLiveRoom{
		Id:             d.ID,
		Title:          d.Title,
		Description:    d.Description,
		HlsUrl:         d.HlsURL,
		Status:         status,
		ScheduledAt:    safeTimestamp(d.ScheduledAt),
		StartedAt:      safeTimestamp(d.StartedAt),
		CurrentViewers: int32(d.CurrentViewers),
		PeakViewers:    int32(d.PeakViewers),
		Thumbnail:      d.Thumbnail,
		Category:       d.Category,
		Tags:           d.Tags,
		UserId:         d.UserID,
		CreateTime:     safeTimestamp(d.CreateTime),
	}
}

func (s *LiveService) ListLiveRooms(ctx context.Context, req *livev1.ListLiveRoomsRequest) (*livev1.ListLiveRoomsResponse, error) {
	page := int(req.Page)
	pageSize := int(req.PageSize)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	items, total, err := s.uc.ListLiveRooms(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}
	result := make([]*livev1.LiveRoom, len(items))
	for i, item := range items {
		result[i] = dtoToLiveRoom(item)
	}
	return &livev1.ListLiveRoomsResponse{
		Items: result,
		Total: int32(total),
	}, nil
}

func (s *LiveService) GetLiveRoom(ctx context.Context, req *livev1.GetLiveRoomRequest) (*livev1.GetLiveRoomResponse, error) {
	room, err := s.uc.GetLiveRoomByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &livev1.GetLiveRoomResponse{
		Room: dtoToLiveRoom(room),
	}, nil
}

func (s *LiveService) CreateLiveRoom(ctx context.Context, req *livev1.CreateLiveRoomRequest) (*livev1.CreateLiveRoomResponse, error) {
	room := &dto.LiveRoomDTO{
		Title:       req.Title,
		Description: req.Description,
		RtmpURL:     req.RtmpUrl,
		HlsURL:      req.HlsUrl,
		MaxViewers:  int(req.MaxViewers),
		Thumbnail:   req.Thumbnail,
		Category:    req.Category,
		Tags:        req.Tags,
		UserID:      req.UserId,
	}
	if req.ScheduledAt != nil {
		room.ScheduledAt = req.ScheduledAt.AsTime()
	}
	created, err := s.uc.CreateLiveRoom(ctx, room)
	if err != nil {
		return nil, err
	}
	return &livev1.CreateLiveRoomResponse{
		Room: dtoToLiveRoom(created),
	}, nil
}

func (s *LiveService) UpdateLiveRoom(ctx context.Context, req *livev1.UpdateLiveRoomRequest) (*livev1.UpdateLiveRoomResponse, error) {
	existing, err := s.uc.GetLiveRoomByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	room := &dto.LiveRoomDTO{ID: req.Id}
	if req.Title != "" {
		room.Title = req.Title
	} else {
		room.Title = existing.Title
	}
	room.Description = req.Description
	room.RtmpURL = req.RtmpUrl
	room.HlsURL = req.HlsUrl
	if req.MaxViewers > 0 {
		room.MaxViewers = int(req.MaxViewers)
	} else {
		room.MaxViewers = existing.MaxViewers
	}
	room.Thumbnail = req.Thumbnail
	room.Category = req.Category
	if req.Tags != nil {
		room.Tags = req.Tags
	} else {
		room.Tags = existing.Tags
	}
	if req.ScheduledAt != nil {
		room.ScheduledAt = req.ScheduledAt.AsTime()
	} else {
		room.ScheduledAt = existing.ScheduledAt
	}
	updated, err := s.uc.UpdateLiveRoom(ctx, room)
	if err != nil {
		return nil, err
	}
	return &livev1.UpdateLiveRoomResponse{
		Room: dtoToLiveRoom(updated),
	}, nil
}

func (s *LiveService) DeleteLiveRoom(ctx context.Context, req *livev1.DeleteLiveRoomRequest) (*emptypb.Empty, error) {
	if err := s.uc.DeleteLiveRoom(ctx, req.Id); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *LiveService) StartLiveRoom(ctx context.Context, req *livev1.StartLiveRoomRequest) (*livev1.StartLiveRoomResponse, error) {
	room, err := s.uc.StartLiveRoom(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &livev1.StartLiveRoomResponse{
		Room: dtoToLiveRoom(room),
	}, nil
}

func (s *LiveService) EndLiveRoom(ctx context.Context, req *livev1.EndLiveRoomRequest) (*livev1.EndLiveRoomResponse, error) {
	room, err := s.uc.EndLiveRoom(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &livev1.EndLiveRoomResponse{
		Room: dtoToLiveRoom(room),
	}, nil
}

func (s *LiveService) ListPublicLiveRooms(ctx context.Context, req *livev1.ListPublicLiveRoomsRequest) (*livev1.ListPublicLiveRoomsResponse, error) {
	items, err := s.uc.ListPublicLiveRooms(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*livev1.PublicLiveRoom, len(items))
	for i, item := range items {
		result[i] = dtoToPublicLiveRoom(item)
	}
	return &livev1.ListPublicLiveRoomsResponse{
		Items: result,
		Total: int32(len(items)),
	}, nil
}

func (s *LiveService) GetPublicLiveRoom(ctx context.Context, req *livev1.GetPublicLiveRoomRequest) (*livev1.GetPublicLiveRoomResponse, error) {
	room, err := s.uc.GetLiveRoomByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &livev1.GetPublicLiveRoomResponse{
		Room: dtoToPublicLiveRoom(room),
	}, nil
}

func safeTimestamp(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}
