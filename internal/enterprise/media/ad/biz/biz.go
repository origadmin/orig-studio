package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origstudio/internal/enterprise/media/ad/dto"
)

type Repo interface {
	ListPlacements(ctx context.Context) ([]*dto.AdPlacementDTO, error)
	GetPlacementByID(ctx context.Context, id string) (*dto.AdPlacementDTO, error)
	GetPlacementBySlug(ctx context.Context, slug string) (*dto.AdPlacementDTO, error)
	CreatePlacement(ctx context.Context, p *dto.AdPlacementDTO) (*dto.AdPlacementDTO, error)
	UpdatePlacement(ctx context.Context, p *dto.AdPlacementDTO) (*dto.AdPlacementDTO, error)
	DeletePlacement(ctx context.Context, id string) error
	TogglePlacement(ctx context.Context, id string) (*dto.AdPlacementDTO, error)

	ListAds(ctx context.Context, placementID string) ([]*dto.AdDTO, int, error)
	GetAdByID(ctx context.Context, id string) (*dto.AdDTO, error)
	CreateAd(ctx context.Context, a *dto.AdDTO) (*dto.AdDTO, error)
	UpdateAd(ctx context.Context, a *dto.AdDTO) (*dto.AdDTO, error)
	DeleteAd(ctx context.Context, id string) error
	ToggleAd(ctx context.Context, id string) (*dto.AdDTO, error)
	ListActiveAdsByPlacement(ctx context.Context, placementSlug string) ([]*dto.AdDTO, error)
	ListActivePlacementsWithAds(ctx context.Context) ([]*dto.AdPlacementWithAdsDTO, error)
	IncrementImpressions(ctx context.Context, id string) error
	IncrementClicks(ctx context.Context, id string) error
	IncrementCreativeImpressions(ctx context.Context, id string) error
	IncrementCreativeClicks(ctx context.Context, id string) error

	RecordClickLog(ctx context.Context, log *dto.AdClickLogDTO) error
	ListClickLogs(ctx context.Context, adID string, page, pageSize int) ([]*dto.AdClickLogDTO, int, error)

	// Creative Library (G6-3): 创意一次定义、可被多个广告位复用
	ListCreatives(ctx context.Context) ([]*dto.AdCreativeDTO, error)
	GetCreative(ctx context.Context, id string) (*dto.AdCreativeDTO, error)
	CreateCreative(ctx context.Context, c *dto.AdCreativeDTO) (*dto.AdCreativeDTO, error)
	UpdateCreative(ctx context.Context, c *dto.AdCreativeDTO) (*dto.AdCreativeDTO, error)
	DeleteCreative(ctx context.Context, id string) error
	ListCreativeIDsByPlacement(ctx context.Context, placementID string) ([]string, error)
	AssignCreative(ctx context.Context, placementID, creativeID string) error
	UnassignCreative(ctx context.Context, placementID, creativeID string) error
}

type UseCase struct {
	repo Repo
	log  *log.Helper
}

func NewUseCase(repo Repo, logger log.Logger) *UseCase {
	return &UseCase{
		repo: repo,
		log:  log.NewHelper(log.With(logger, "module", "enterprise/ad.biz")),
	}
}

func (uc *UseCase) ListPlacements(ctx context.Context) ([]*dto.AdPlacementDTO, error) {
	return uc.repo.ListPlacements(ctx)
}

func (uc *UseCase) GetPlacementByID(ctx context.Context, id string) (*dto.AdPlacementDTO, error) {
	return uc.repo.GetPlacementByID(ctx, id)
}

func (uc *UseCase) GetPlacementBySlug(ctx context.Context, slug string) (*dto.AdPlacementDTO, error) {
	return uc.repo.GetPlacementBySlug(ctx, slug)
}

func (uc *UseCase) CreatePlacement(ctx context.Context, p *dto.AdPlacementDTO) (*dto.AdPlacementDTO, error) {
	return uc.repo.CreatePlacement(ctx, p)
}

func (uc *UseCase) UpdatePlacement(ctx context.Context, p *dto.AdPlacementDTO) (*dto.AdPlacementDTO, error) {
	return uc.repo.UpdatePlacement(ctx, p)
}

func (uc *UseCase) DeletePlacement(ctx context.Context, id string) error {
	return uc.repo.DeletePlacement(ctx, id)
}

func (uc *UseCase) TogglePlacement(ctx context.Context, id string) (*dto.AdPlacementDTO, error) {
	return uc.repo.TogglePlacement(ctx, id)
}

func (uc *UseCase) ListAds(ctx context.Context, placementID string) ([]*dto.AdDTO, int, error) {
	return uc.repo.ListAds(ctx, placementID)
}

func (uc *UseCase) GetAdByID(ctx context.Context, id string) (*dto.AdDTO, error) {
	return uc.repo.GetAdByID(ctx, id)
}

func (uc *UseCase) CreateAd(ctx context.Context, a *dto.AdDTO) (*dto.AdDTO, error) {
	return uc.repo.CreateAd(ctx, a)
}

func (uc *UseCase) UpdateAd(ctx context.Context, a *dto.AdDTO) (*dto.AdDTO, error) {
	return uc.repo.UpdateAd(ctx, a)
}

func (uc *UseCase) DeleteAd(ctx context.Context, id string) error {
	return uc.repo.DeleteAd(ctx, id)
}

func (uc *UseCase) ToggleAd(ctx context.Context, id string) (*dto.AdDTO, error) {
	return uc.repo.ToggleAd(ctx, id)
}

func (uc *UseCase) ListActiveAdsByPlacement(ctx context.Context, placementSlug string) ([]*dto.AdDTO, error) {
	return uc.repo.ListActiveAdsByPlacement(ctx, placementSlug)
}

func (uc *UseCase) ListActivePlacementsWithAds(ctx context.Context) ([]*dto.AdPlacementWithAdsDTO, error) {
	return uc.repo.ListActivePlacementsWithAds(ctx)
}

func (uc *UseCase) RecordImpression(ctx context.Context, id string) error {
	return uc.repo.IncrementImpressions(ctx, id)
}

func (uc *UseCase) RecordClick(ctx context.Context, adID string, clickLog *dto.AdClickLogDTO) error {
	if err := uc.repo.IncrementClicks(ctx, adID); err != nil {
		uc.log.Warnf("increment clicks failed: %v", err)
	}
	return uc.repo.RecordClickLog(ctx, clickLog)
}

func (uc *UseCase) RecordCreativeImpression(ctx context.Context, id string) error {
	return uc.repo.IncrementCreativeImpressions(ctx, id)
}

func (uc *UseCase) RecordCreativeClick(ctx context.Context, id string) error {
	return uc.repo.IncrementCreativeClicks(ctx, id)
}

func (uc *UseCase) ListClickLogs(ctx context.Context, adID string, page, pageSize int) ([]*dto.AdClickLogDTO, int, error) {
	return uc.repo.ListClickLogs(ctx, adID, page, pageSize)
}

// ==================== Creative Library (G6-3) ====================

func (uc *UseCase) ListCreatives(ctx context.Context) ([]*dto.AdCreativeDTO, error) {
	return uc.repo.ListCreatives(ctx)
}

func (uc *UseCase) GetCreative(ctx context.Context, id string) (*dto.AdCreativeDTO, error) {
	return uc.repo.GetCreative(ctx, id)
}

func (uc *UseCase) CreateCreative(ctx context.Context, c *dto.AdCreativeDTO) (*dto.AdCreativeDTO, error) {
	return uc.repo.CreateCreative(ctx, c)
}

func (uc *UseCase) UpdateCreative(ctx context.Context, c *dto.AdCreativeDTO) (*dto.AdCreativeDTO, error) {
	return uc.repo.UpdateCreative(ctx, c)
}

func (uc *UseCase) DeleteCreative(ctx context.Context, id string) error {
	return uc.repo.DeleteCreative(ctx, id)
}

func (uc *UseCase) ListCreativeIDsByPlacement(ctx context.Context, placementID string) ([]string, error) {
	return uc.repo.ListCreativeIDsByPlacement(ctx, placementID)
}

func (uc *UseCase) AssignCreative(ctx context.Context, placementID, creativeID string) error {
	return uc.repo.AssignCreative(ctx, placementID, creativeID)
}

func (uc *UseCase) UnassignCreative(ctx context.Context, placementID, creativeID string) error {
	return uc.repo.UnassignCreative(ctx, placementID, creativeID)
}
