package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origstudio/internal/features/content/dto"
)

type AdRepo interface {
	ListPlacements(ctx context.Context) ([]*dto.AdPlacementDTO, error)
	GetPlacementByID(ctx context.Context, id string) (*dto.AdPlacementDTO, error)
	GetPlacementBySlug(ctx context.Context, slug string) (*dto.AdPlacementDTO, error)
	CreatePlacement(ctx context.Context, p *dto.AdPlacementDTO) (*dto.AdPlacementDTO, error)
	UpdatePlacement(ctx context.Context, p *dto.AdPlacementDTO) (*dto.AdPlacementDTO, error)
	DeletePlacement(ctx context.Context, id string) error
	CountAdsByPlacement(ctx context.Context, placementID string) (int, error)
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

	RecordClickLog(ctx context.Context, log *dto.AdClickLogDTO) error
	ListClickLogs(ctx context.Context, adID string, page, pageSize int) ([]*dto.AdClickLogDTO, int, error)
}

type AdUseCase struct {
	repo AdRepo
	log  *log.Helper
}

func NewAdUseCase(repo AdRepo, logger log.Logger) *AdUseCase {
	return &AdUseCase{
		repo: repo,
		log:  log.NewHelper(log.With(logger, "module", "ad.biz")),
	}
}

func (uc *AdUseCase) ListPlacements(ctx context.Context) ([]*dto.AdPlacementDTO, error) {
	return uc.repo.ListPlacements(ctx)
}

func (uc *AdUseCase) GetPlacementByID(ctx context.Context, id string) (*dto.AdPlacementDTO, error) {
	return uc.repo.GetPlacementByID(ctx, id)
}

func (uc *AdUseCase) GetPlacementBySlug(ctx context.Context, slug string) (*dto.AdPlacementDTO, error) {
	return uc.repo.GetPlacementBySlug(ctx, slug)
}

func (uc *AdUseCase) CreatePlacement(ctx context.Context, p *dto.AdPlacementDTO) (*dto.AdPlacementDTO, error) {
	return uc.repo.CreatePlacement(ctx, p)
}

func (uc *AdUseCase) UpdatePlacement(ctx context.Context, p *dto.AdPlacementDTO) (*dto.AdPlacementDTO, error) {
	return uc.repo.UpdatePlacement(ctx, p)
}

func (uc *AdUseCase) DeletePlacement(ctx context.Context, id string) error {
	return uc.repo.DeletePlacement(ctx, id)
}

func (uc *AdUseCase) CountAdsByPlacement(ctx context.Context, placementID string) (int, error) {
	return uc.repo.CountAdsByPlacement(ctx, placementID)
}

func (uc *AdUseCase) TogglePlacement(ctx context.Context, id string) (*dto.AdPlacementDTO, error) {
	return uc.repo.TogglePlacement(ctx, id)
}

func (uc *AdUseCase) ListAds(ctx context.Context, placementID string) ([]*dto.AdDTO, int, error) {
	return uc.repo.ListAds(ctx, placementID)
}

func (uc *AdUseCase) GetAdByID(ctx context.Context, id string) (*dto.AdDTO, error) {
	return uc.repo.GetAdByID(ctx, id)
}

func (uc *AdUseCase) CreateAd(ctx context.Context, a *dto.AdDTO) (*dto.AdDTO, error) {
	return uc.repo.CreateAd(ctx, a)
}

func (uc *AdUseCase) UpdateAd(ctx context.Context, a *dto.AdDTO) (*dto.AdDTO, error) {
	return uc.repo.UpdateAd(ctx, a)
}

func (uc *AdUseCase) DeleteAd(ctx context.Context, id string) error {
	return uc.repo.DeleteAd(ctx, id)
}

func (uc *AdUseCase) ToggleAd(ctx context.Context, id string) (*dto.AdDTO, error) {
	return uc.repo.ToggleAd(ctx, id)
}

func (uc *AdUseCase) ListActiveAdsByPlacement(ctx context.Context, placementSlug string) ([]*dto.AdDTO, error) {
	return uc.repo.ListActiveAdsByPlacement(ctx, placementSlug)
}

func (uc *AdUseCase) ListActivePlacementsWithAds(ctx context.Context) ([]*dto.AdPlacementWithAdsDTO, error) {
	return uc.repo.ListActivePlacementsWithAds(ctx)
}

func (uc *AdUseCase) RecordImpression(ctx context.Context, id string) error {
	return uc.repo.IncrementImpressions(ctx, id)
}

func (uc *AdUseCase) RecordClick(ctx context.Context, adID string, clickLog *dto.AdClickLogDTO) error {
	if err := uc.repo.IncrementClicks(ctx, adID); err != nil {
		uc.log.Warnf("increment clicks failed: %v", err)
	}
	return uc.repo.RecordClickLog(ctx, clickLog)
}

func (uc *AdUseCase) ListClickLogs(ctx context.Context, adID string, page, pageSize int) ([]*dto.AdClickLogDTO, int, error) {
	return uc.repo.ListClickLogs(ctx, adID, page, pageSize)
}
