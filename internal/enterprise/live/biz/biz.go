package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origstudio/internal/enterprise/live/dto"
)

type Repo interface {
	ListLiveRooms(ctx context.Context, page, pageSize int) ([]*dto.LiveRoomDTO, int, error)
	GetLiveRoomByID(ctx context.Context, id string) (*dto.LiveRoomDTO, error)
	CreateLiveRoom(ctx context.Context, room *dto.LiveRoomDTO) (*dto.LiveRoomDTO, error)
	UpdateLiveRoom(ctx context.Context, room *dto.LiveRoomDTO) (*dto.LiveRoomDTO, error)
	DeleteLiveRoom(ctx context.Context, id string) error
	UpdateStatus(ctx context.Context, id string, status string) (*dto.LiveRoomDTO, error)
	ListPublicLiveRooms(ctx context.Context) ([]*dto.LiveRoomDTO, error)
}

type UseCase struct {
	repo Repo
	log  *log.Helper
}

func NewUseCase(repo Repo, logger log.Logger) *UseCase {
	return &UseCase{
		repo: repo,
		log:  log.NewHelper(log.With(logger, "module", "enterprise/live.biz")),
	}
}

func (uc *UseCase) ListLiveRooms(ctx context.Context, page, pageSize int) ([]*dto.LiveRoomDTO, int, error) {
	return uc.repo.ListLiveRooms(ctx, page, pageSize)
}

func (uc *UseCase) GetLiveRoomByID(ctx context.Context, id string) (*dto.LiveRoomDTO, error) {
	return uc.repo.GetLiveRoomByID(ctx, id)
}

func (uc *UseCase) CreateLiveRoom(ctx context.Context, room *dto.LiveRoomDTO) (*dto.LiveRoomDTO, error) {
	return uc.repo.CreateLiveRoom(ctx, room)
}

func (uc *UseCase) UpdateLiveRoom(ctx context.Context, room *dto.LiveRoomDTO) (*dto.LiveRoomDTO, error) {
	return uc.repo.UpdateLiveRoom(ctx, room)
}

func (uc *UseCase) DeleteLiveRoom(ctx context.Context, id string) error {
	return uc.repo.DeleteLiveRoom(ctx, id)
}

func (uc *UseCase) StartLiveRoom(ctx context.Context, id string) (*dto.LiveRoomDTO, error) {
	return uc.repo.UpdateStatus(ctx, id, "live")
}

func (uc *UseCase) EndLiveRoom(ctx context.Context, id string) (*dto.LiveRoomDTO, error) {
	return uc.repo.UpdateStatus(ctx, id, "ended")
}

func (uc *UseCase) ListPublicLiveRooms(ctx context.Context) ([]*dto.LiveRoomDTO, error) {
	return uc.repo.ListPublicLiveRooms(ctx)
}