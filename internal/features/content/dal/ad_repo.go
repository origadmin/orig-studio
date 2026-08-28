package dal

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origstudio/internal/data/entity"
	"origadmin/application/origstudio/internal/data/entity/ad"
	"origadmin/application/origstudio/internal/data/entity/adclicklog"
	"origadmin/application/origstudio/internal/data/entity/adplacement"
	"origadmin/application/origstudio/internal/domain/types"
	"origadmin/application/origstudio/internal/features/content/biz"
	"origadmin/application/origstudio/internal/features/content/dto"
)

type adRepo struct {
	data *Data
	log  *log.Helper
}

func NewAdRepo(data *Data, logger log.Logger) biz.AdRepo {
	return &adRepo{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "ad.data")),
	}
}

// ==================== Entity-to-DTO Conversion ====================

func entityToAdPlacementDTO(e *entity.AdPlacement) *dto.AdPlacementDTO {
	if e == nil {
		return nil
	}
	return &dto.AdPlacementDTO{
		ID:          e.ID,
		Name:        e.Name,
		Slug:        e.Slug,
		Type:        e.Type,
		Description: e.Description,
		Width:       e.Width,
		Height:      e.Height,
		MaxAds:      e.MaxAds,
		IsActive:    e.IsActive,
		Sequence:    e.Sequence,
	}
}

func entityToAdDTO(e *entity.Ad) *dto.AdDTO {
	if e == nil {
		return nil
	}
	return &dto.AdDTO{
		ID:             e.ID,
		PlacementID:    e.PlacementID,
		Title:          e.Title,
		TitleI18n:      e.TitleI18n,
		ImageURL:       e.ImageURL,
		ImageMobileURL: e.ImageMobileURL,
		LinkURL:        e.LinkURL,
		LinkTarget:     e.LinkTarget,
		BadgeText:      e.BadgeText,
		Priority:       e.Priority,
		IsActive:       e.IsActive,
		StartAt:        e.StartAt,
		EndAt:          e.EndAt,
		Impressions:    e.Impressions,
		Clicks:         e.Clicks,
		SortOrder:      e.Priority,
		StartTime:      e.StartAt,
		EndTime:        e.EndAt,
	}
}

func entityToAdCreativeDTO(e *entity.AdCreative) *dto.AdCreativeDTO {
	if e == nil {
		return nil
	}
	return &dto.AdCreativeDTO{
		ID:             e.ID,
		Title:          e.Title,
		TitleI18n:      e.TitleI18n,
		ImageURL:       e.ImageURL,
		ImageMobileURL: e.ImageMobileURL,
		LinkURL:        e.LinkURL,
		LinkTarget:     e.LinkTarget,
		BadgeText:      e.BadgeText,
		IsActive:       e.IsActive,
		Priority:       e.Priority,
		Impressions:    e.Impressions,
		Clicks:         e.Clicks,
	}
}

func entityToAdClickLogDTO(e *entity.AdClickLog) *dto.AdClickLogDTO {
	if e == nil {
		return nil
	}
	return &dto.AdClickLogDTO{
		ID:          e.ID,
		AdID:        e.AdID,
		PlacementID: e.PlacementID,
		IP:          e.IP,
		IPAddress:   e.IP,
		UserAgent:   e.UserAgent,
		UserID:      e.UserID,
		Referer:     e.Referer,
	}
}

// ==================== Placement ====================

func (r *adRepo) ListPlacements(ctx context.Context) ([]*dto.AdPlacementDTO, error) {
	items, err := r.data.db.AdPlacement.Query().
		Order(entity.Asc(adplacement.FieldSequence)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list ad placements: %w", err)
	}
	result := make([]*dto.AdPlacementDTO, len(items))
	for i, item := range items {
		result[i] = entityToAdPlacementDTO(item)
	}
	return result, nil
}

func (r *adRepo) GetPlacementByID(ctx context.Context, id string) (*dto.AdPlacementDTO, error) {
	item, err := r.data.db.AdPlacement.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get ad placement: %w", err)
	}
	return entityToAdPlacementDTO(item), nil
}

func (r *adRepo) GetPlacementBySlug(ctx context.Context, slug string) (*dto.AdPlacementDTO, error) {
	item, err := r.data.db.AdPlacement.Query().
		Where(adplacement.SlugEQ(slug)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get ad placement by slug: %w", err)
	}
	return entityToAdPlacementDTO(item), nil
}

func (r *adRepo) CreatePlacement(ctx context.Context, p *dto.AdPlacementDTO) (*dto.AdPlacementDTO, error) {
	builder := r.data.db.AdPlacement.Create().
		SetName(p.Name).
		SetSlug(p.Slug).
		SetType(p.Type).
		SetWidth(p.Width).
		SetHeight(p.Height).
		SetMaxAds(p.MaxAds).
		SetIsActive(p.IsActive).
		SetSequence(p.Sequence)

	if p.Description != "" {
		builder.SetDescription(p.Description)
	}

	ent, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create ad placement: %w", err)
	}
	return entityToAdPlacementDTO(ent), nil
}

func (r *adRepo) UpdatePlacement(ctx context.Context, p *dto.AdPlacementDTO) (*dto.AdPlacementDTO, error) {
	builder := r.data.db.AdPlacement.UpdateOneID(p.ID).
		SetName(p.Name).
		SetSlug(p.Slug).
		SetType(p.Type).
		SetWidth(p.Width).
		SetHeight(p.Height).
		SetMaxAds(p.MaxAds).
		SetIsActive(p.IsActive).
		SetSequence(p.Sequence)

	builder.SetDescription(p.Description)

	ent, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update ad placement: %w", err)
	}
	return entityToAdPlacementDTO(ent), nil
}

func (r *adRepo) DeletePlacement(ctx context.Context, id string) error {
	// 级联删除：先删除该广告位下的所有广告，再删除广告位
	_, err := r.data.db.Ad.Delete().
		Where(ad.PlacementIDEQ(id)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("cascade delete ads for placement %s: %w", id, err)
	}
	return r.data.db.AdPlacement.DeleteOneID(id).Exec(ctx)
}

// CountAdsByPlacement 统计广告位下的广告数量（用于删除前提示）
func (r *adRepo) CountAdsByPlacement(ctx context.Context, placementID string) (int, error) {
	return r.data.db.Ad.Query().
		Where(ad.PlacementIDEQ(placementID)).
		Count(ctx)
}

func (r *adRepo) TogglePlacement(ctx context.Context, id string) (*dto.AdPlacementDTO, error) {
	ent, err := r.data.db.AdPlacement.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get ad placement for toggle: %w", err)
	}
	updated, err := r.data.db.AdPlacement.UpdateOneID(id).
		SetIsActive(!ent.IsActive).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("toggle ad placement: %w", err)
	}
	return entityToAdPlacementDTO(updated), nil
}

// ==================== Ad ====================

func (r *adRepo) ListAds(ctx context.Context, placementID string) ([]*dto.AdDTO, int, error) {
	query := r.data.db.Ad.Query().
		Where(ad.PlacementIDEQ(placementID))

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count ads: %w", err)
	}

	items, err := query.
		Order(entity.Desc(ad.FieldPriority), entity.Asc(ad.FieldID)).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list ads: %w", err)
	}
	result := make([]*dto.AdDTO, len(items))
	for i, item := range items {
		result[i] = entityToAdDTO(item)
	}
	return result, total, nil
}

func (r *adRepo) GetAdByID(ctx context.Context, id string) (*dto.AdDTO, error) {
	item, err := r.data.db.Ad.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get ad: %w", err)
	}
	return entityToAdDTO(item), nil
}

func (r *adRepo) CreateAd(ctx context.Context, a *dto.AdDTO) (*dto.AdDTO, error) {
	builder := r.data.db.Ad.Create().
		SetPlacementID(a.PlacementID).
		SetTitle(a.Title).
		SetPriority(a.Priority).
		SetIsActive(a.IsActive).
		SetLinkTarget(a.LinkTarget)

	if a.TitleI18n != nil {
		builder.SetTitleI18n(a.TitleI18n)
	}
	if a.ImageURL != "" {
		builder.SetImageURL(a.ImageURL)
	}
	if a.ImageMobileURL != "" {
		builder.SetImageMobileURL(a.ImageMobileURL)
	}
	if a.LinkURL != "" {
		builder.SetLinkURL(a.LinkURL)
	}
	if a.BadgeText != "" {
		builder.SetBadgeText(a.BadgeText)
	}
	if !a.StartAt.IsZero() {
		builder.SetStartAt(a.StartAt)
	}
	if !a.EndAt.IsZero() {
		builder.SetEndAt(a.EndAt)
	}

	ent, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create ad: %w", err)
	}
	return entityToAdDTO(ent), nil
}

func (r *adRepo) UpdateAd(ctx context.Context, a *dto.AdDTO) (*dto.AdDTO, error) {
	builder := r.data.db.Ad.UpdateOneID(a.ID).
		SetPlacementID(a.PlacementID).
		SetTitle(a.Title).
		SetPriority(a.Priority).
		SetIsActive(a.IsActive).
		SetLinkTarget(a.LinkTarget)

	if a.TitleI18n != nil {
		builder.SetTitleI18n(a.TitleI18n)
	}
	builder.SetImageURL(a.ImageURL)
	builder.SetImageMobileURL(a.ImageMobileURL)
	builder.SetLinkURL(a.LinkURL)
	builder.SetBadgeText(a.BadgeText)
	if !a.StartAt.IsZero() {
		builder.SetStartAt(a.StartAt)
	}
	if !a.EndAt.IsZero() {
		builder.SetEndAt(a.EndAt)
	}

	ent, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update ad: %w", err)
	}
	return entityToAdDTO(ent), nil
}

func (r *adRepo) DeleteAd(ctx context.Context, id string) error {
	return r.data.db.Ad.DeleteOneID(id).Exec(ctx)
}

func (r *adRepo) ToggleAd(ctx context.Context, id string) (*dto.AdDTO, error) {
	ent, err := r.data.db.Ad.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get ad for toggle: %w", err)
	}
	updated, err := r.data.db.Ad.UpdateOneID(id).
		SetIsActive(!ent.IsActive).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("toggle ad: %w", err)
	}
	return entityToAdDTO(updated), nil
}

func (r *adRepo) ListActiveAdsByPlacement(ctx context.Context, placementSlug string) ([]*dto.AdDTO, error) {
	placement, err := r.data.db.AdPlacement.Query().
		Where(adplacement.SlugEQ(placementSlug)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get placement by slug: %w", err)
	}
	if !placement.IsActive {
		return []*dto.AdDTO{}, nil
	}

	now := time.Now()
	items, err := r.data.db.Ad.Query().
		Where(
			ad.PlacementIDEQ(placement.ID),
			ad.IsActiveEQ(true),
		).
		Order(entity.Desc(ad.FieldPriority)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list active ads: %w", err)
	}

	var filtered []*entity.Ad
	for _, a := range items {
		if !a.StartAt.IsZero() && a.StartAt.After(now) {
			continue
		}
		if !a.EndAt.IsZero() && a.EndAt.Before(now) {
			continue
		}
		filtered = append(filtered, a)
	}

	if placement.MaxAds > 0 && len(filtered) > placement.MaxAds {
		filtered = filtered[:placement.MaxAds]
	}

	result := make([]*dto.AdDTO, len(filtered))
	for i, a := range filtered {
		result[i] = entityToAdDTO(a)
	}
	return result, nil
}

func (r *adRepo) ListActivePlacementsWithAds(ctx context.Context) ([]*dto.AdPlacementWithAdsDTO, error) {
	now := time.Now()
	placements, err := r.data.db.AdPlacement.Query().
		Where(adplacement.IsActiveEQ(true)).
		Order(entity.Asc(adplacement.FieldSequence)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list active placements: %w", err)
	}

	result := make([]*dto.AdPlacementWithAdsDTO, 0, len(placements))
	for _, p := range placements {
		ads, err := r.data.db.Ad.Query().
			Where(
				ad.PlacementIDEQ(p.ID),
				ad.IsActiveEQ(true),
			).
			Order(entity.Desc(ad.FieldPriority)).
			All(ctx)
		if err != nil {
			continue
		}
		var filtered []*entity.Ad
		for _, a := range ads {
			if !a.StartAt.IsZero() && a.StartAt.After(now) {
				continue
			}
			if !a.EndAt.IsZero() && a.EndAt.Before(now) {
				continue
			}
			filtered = append(filtered, a)
		}
		if p.MaxAds > 0 && len(filtered) > p.MaxAds {
			filtered = filtered[:p.MaxAds]
		}
		if len(filtered) == 0 {
			continue
		}
		adDTOs := make([]*dto.AdDTO, len(filtered))
		for i, a := range filtered {
			adDTOs[i] = entityToAdDTO(a)
		}
		result = append(result, &dto.AdPlacementWithAdsDTO{
			ID:          p.ID,
			Name:        p.Name,
			Slug:        p.Slug,
			Type:        p.Type,
			Description: p.Description,
			Width:       p.Width,
			Height:      p.Height,
			Sequence:    p.Sequence,
			Ads:         adDTOs,
		})
	}
	return result, nil
}

func (r *adRepo) IncrementImpressions(ctx context.Context, id string) error {
	_, err := r.data.db.Ad.UpdateOneID(id).
		AddImpressions(1).
		Save(ctx)
	return err
}

func (r *adRepo) IncrementClicks(ctx context.Context, id string) error {
	_, err := r.data.db.Ad.UpdateOneID(id).
		AddClicks(1).
		Save(ctx)
	return err
}

// ==================== ClickLog ====================

func (r *adRepo) RecordClickLog(ctx context.Context, cl *dto.AdClickLogDTO) error {
	builder := r.data.db.AdClickLog.Create().
		SetAdID(cl.AdID).
		SetPlacementID(cl.PlacementID)

	if cl.IP != "" {
		builder.SetIP(cl.IP)
	}
	if cl.UserAgent != "" {
		builder.SetUserAgent(cl.UserAgent)
	}
	if cl.UserID != "" {
		builder.SetUserID(cl.UserID)
	}
	if cl.Referer != "" {
		builder.SetReferer(cl.Referer)
	}

	_, err := builder.Save(ctx)
	if err != nil {
		return fmt.Errorf("record ad click log: %w", err)
	}
	return nil
}

func (r *adRepo) ListClickLogs(ctx context.Context, adID string, page, pageSize int) ([]*dto.AdClickLogDTO, int, error) {
	query := r.data.db.AdClickLog.Query().
		Where(adclicklog.AdIDEQ(adID))

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count ad click logs: %w", err)
	}

	items, err := query.
		Order(entity.Desc(adclicklog.FieldID)).
		Offset(types.CalcOffset(page, pageSize)).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list ad click logs: %w", err)
	}
	result := make([]*dto.AdClickLogDTO, len(items))
	for i, item := range items {
		result[i] = entityToAdClickLogDTO(item)
	}
	return result, total, nil
}
