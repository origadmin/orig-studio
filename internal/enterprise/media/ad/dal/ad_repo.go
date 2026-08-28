package dal

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origstudio/internal/data/entity"
	"origadmin/application/origstudio/internal/data/entity/ad"
	"origadmin/application/origstudio/internal/data/entity/adclicklog"
	"origadmin/application/origstudio/internal/data/entity/adcreative"
	"origadmin/application/origstudio/internal/data/entity/adplacement"
	"origadmin/application/origstudio/internal/enterprise/media/ad/biz"
	"origadmin/application/origstudio/internal/enterprise/media/ad/dto"
)

type repo struct {
	db  *entity.Client
	log *log.Helper
}

func NewRepo(db *entity.Client, logger log.Logger) biz.Repo {
	return &repo{
		db:  db,
		log: log.NewHelper(log.With(logger, "module", "enterprise/ad.repo")),
	}
}

func entityToAdPlacementDTO(e *entity.AdPlacement) *dto.AdPlacementDTO {
	if e == nil {
		return nil
	}
	return &dto.AdPlacementDTO{
		ID:           e.ID,
		Name:         e.Name,
		Slug:         e.Slug,
		Type:         e.Type,
		Description:  e.Description,
		Width:        e.Width,
		Height:       e.Height,
		MaxAds:       e.MaxAds,
		IsActive:     e.IsActive,
		Sequence:     e.Sequence,
		CreativeCount: len(e.Edges.Creatives),
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

func (r *repo) ListPlacements(ctx context.Context) ([]*dto.AdPlacementDTO, error) {
	items, err := r.db.AdPlacement.Query().
		Order(entity.Asc(adplacement.FieldSequence)).
		WithCreatives().
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

func (r *repo) GetPlacementByID(ctx context.Context, id string) (*dto.AdPlacementDTO, error) {
	item, err := r.db.AdPlacement.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get ad placement: %w", err)
	}
	return entityToAdPlacementDTO(item), nil
}

func (r *repo) GetPlacementBySlug(ctx context.Context, slug string) (*dto.AdPlacementDTO, error) {
	item, err := r.db.AdPlacement.Query().
		Where(adplacement.SlugEQ(slug)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get ad placement by slug: %w", err)
	}
	return entityToAdPlacementDTO(item), nil
}

func (r *repo) CreatePlacement(ctx context.Context, p *dto.AdPlacementDTO) (*dto.AdPlacementDTO, error) {
	builder := r.db.AdPlacement.Create().
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

func (r *repo) UpdatePlacement(ctx context.Context, p *dto.AdPlacementDTO) (*dto.AdPlacementDTO, error) {
	builder := r.db.AdPlacement.UpdateOneID(p.ID).
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

func (r *repo) DeletePlacement(ctx context.Context, id string) error {
	return r.db.AdPlacement.DeleteOneID(id).Exec(ctx)
}

func (r *repo) TogglePlacement(ctx context.Context, id string) (*dto.AdPlacementDTO, error) {
	ent, err := r.db.AdPlacement.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get ad placement for toggle: %w", err)
	}
	updated, err := r.db.AdPlacement.UpdateOneID(id).
		SetIsActive(!ent.IsActive).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("toggle ad placement: %w", err)
	}
	return entityToAdPlacementDTO(updated), nil
}

func (r *repo) ListAds(ctx context.Context, placementID string) ([]*dto.AdDTO, int, error) {
	query := r.db.Ad.Query().
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

func (r *repo) GetAdByID(ctx context.Context, id string) (*dto.AdDTO, error) {
	item, err := r.db.Ad.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get ad: %w", err)
	}
	return entityToAdDTO(item), nil
}

func (r *repo) CreateAd(ctx context.Context, a *dto.AdDTO) (*dto.AdDTO, error) {
	builder := r.db.Ad.Create().
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

func (r *repo) UpdateAd(ctx context.Context, a *dto.AdDTO) (*dto.AdDTO, error) {
	builder := r.db.Ad.UpdateOneID(a.ID).
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

func (r *repo) DeleteAd(ctx context.Context, id string) error {
	return r.db.Ad.DeleteOneID(id).Exec(ctx)
}

func (r *repo) ToggleAd(ctx context.Context, id string) (*dto.AdDTO, error) {
	ent, err := r.db.Ad.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get ad for toggle: %w", err)
	}
	updated, err := r.db.Ad.UpdateOneID(id).
		SetIsActive(!ent.IsActive).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("toggle ad: %w", err)
	}
	return entityToAdDTO(updated), nil
}

func (r *repo) ListActiveAdsByPlacement(ctx context.Context, placementSlug string) ([]*dto.AdDTO, error) {
	placement, err := r.db.AdPlacement.Query().
		Where(adplacement.SlugEQ(placementSlug)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get placement by slug: %w", err)
	}
	if !placement.IsActive {
		return []*dto.AdDTO{}, nil
	}
	now := time.Now()
	items, err := r.db.Ad.Query().
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

func (r *repo) ListActivePlacementsWithAds(ctx context.Context) ([]*dto.AdPlacementWithAdsDTO, error) {
	placements, err := r.db.AdPlacement.Query().
		Where(adplacement.IsActiveEQ(true)).
		Order(entity.Asc(adplacement.FieldSequence)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list active ad placements: %w", err)
	}
	now := time.Now()
	result := make([]*dto.AdPlacementWithAdsDTO, 0, len(placements))
	for _, p := range placements {
		ads := r.listActiveAdDTOs(ctx, p, now)
		creatives := r.listActiveCreativeDTOs(ctx, p.ID, p.MaxAds)
		if len(ads) == 0 && len(creatives) == 0 {
			continue
		}
		result = append(result, &dto.AdPlacementWithAdsDTO{
			AdPlacementDTO: *entityToAdPlacementDTO(p),
			Ads:            ads,
			Creatives:      creatives,
		})
	}
	return result, nil
}

// listActiveAdDTOs returns active, in-window ads for a placement, capped by MaxAds.
func (r *repo) listActiveAdDTOs(ctx context.Context, p *entity.AdPlacement, now time.Time) []*dto.AdDTO {
	adItems, err := r.db.Ad.Query().
		Where(
			ad.PlacementIDEQ(p.ID),
			ad.IsActiveEQ(true),
		).
		Order(entity.Desc(ad.FieldPriority)).
		All(ctx)
	if err != nil {
		return nil
	}
	var filtered []*entity.Ad
	for _, a := range adItems {
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
	ads := make([]*dto.AdDTO, len(filtered))
	for i, a := range filtered {
		ads[i] = entityToAdDTO(a)
	}
	return ads
}

// listActiveCreativeDTOs returns active creatives assigned to a placement, sorted by priority desc, capped by maxAds.
func (r *repo) listActiveCreativeDTOs(ctx context.Context, placementID string, maxAds int) []*dto.AdCreativeDTO {
	placement, err := r.db.AdPlacement.Query().
		Where(adplacement.IDEQ(placementID)).
		WithCreatives(func(q *entity.AdCreativeQuery) {
			q.Where(adcreative.IsActiveEQ(true)).
				Order(entity.Desc(adcreative.FieldPriority))
		}).
		Only(ctx)
	if err != nil {
		return nil
	}
	creatives := placement.Edges.Creatives
	if maxAds > 0 && len(creatives) > maxAds {
		creatives = creatives[:maxAds]
	}
	result := make([]*dto.AdCreativeDTO, len(creatives))
	for i, c := range creatives {
		result[i] = entityToAdCreativeDTO(c)
	}
	return result
}

func (r *repo) IncrementImpressions(ctx context.Context, id string) error {
	_, err := r.db.Ad.UpdateOneID(id).
		AddImpressions(1).
		Save(ctx)
	return err
}

func (r *repo) IncrementClicks(ctx context.Context, id string) error {
	_, err := r.db.Ad.UpdateOneID(id).
		AddClicks(1).
		Save(ctx)
	return err
}

func (r *repo) IncrementCreativeImpressions(ctx context.Context, id string) error {
	_, err := r.db.AdCreative.UpdateOneID(id).
		AddImpressions(1).
		Save(ctx)
	return err
}

func (r *repo) IncrementCreativeClicks(ctx context.Context, id string) error {
	_, err := r.db.AdCreative.UpdateOneID(id).
		AddClicks(1).
		Save(ctx)
	return err
}

func (r *repo) RecordClickLog(ctx context.Context, cl *dto.AdClickLogDTO) error {
	builder := r.db.AdClickLog.Create().
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

func (r *repo) ListClickLogs(ctx context.Context, adID string, page, pageSize int) ([]*dto.AdClickLogDTO, int, error) {
	query := r.db.AdClickLog.Query().
		Where(adclicklog.AdIDEQ(adID))
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count ad click logs: %w", err)
	}
	items, err := query.
		Order(entity.Desc(adclicklog.FieldID)).
		Offset((page - 1) * pageSize).
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

// ==================== Creative Library (G6-3) ====================

func (r *repo) ListCreatives(ctx context.Context) ([]*dto.AdCreativeDTO, error) {
	items, err := r.db.AdCreative.Query().
		Order(entity.Desc(adcreative.FieldPriority), entity.Desc(adcreative.FieldID)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list creatives: %w", err)
	}
	result := make([]*dto.AdCreativeDTO, len(items))
	for i, c := range items {
		result[i] = entityToAdCreativeDTO(c)
	}
	return result, nil
}

func (r *repo) GetCreative(ctx context.Context, id string) (*dto.AdCreativeDTO, error) {
	c, err := r.db.AdCreative.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get creative: %w", err)
	}
	return entityToAdCreativeDTO(c), nil
}

func (r *repo) CreateCreative(ctx context.Context, c *dto.AdCreativeDTO) (*dto.AdCreativeDTO, error) {
	linkTarget := c.LinkTarget
	if linkTarget == "" {
		linkTarget = "_blank"
	}
	builder := r.db.AdCreative.Create().
		SetTitle(c.Title).
		SetImageURL(c.ImageURL).
		SetImageMobileURL(c.ImageMobileURL).
		SetLinkURL(c.LinkURL).
		SetLinkTarget(linkTarget).
		SetBadgeText(c.BadgeText).
		SetIsActive(c.IsActive).
		SetPriority(c.Priority)
	if c.TitleI18n != nil {
		builder.SetTitleI18n(c.TitleI18n)
	}
	ent, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create creative: %w", err)
	}
	return entityToAdCreativeDTO(ent), nil
}

func (r *repo) UpdateCreative(ctx context.Context, c *dto.AdCreativeDTO) (*dto.AdCreativeDTO, error) {
	linkTarget := c.LinkTarget
	if linkTarget == "" {
		linkTarget = "_blank"
	}
	builder := r.db.AdCreative.UpdateOneID(c.ID).
		SetTitle(c.Title).
		SetImageURL(c.ImageURL).
		SetImageMobileURL(c.ImageMobileURL).
		SetLinkURL(c.LinkURL).
		SetLinkTarget(linkTarget).
		SetBadgeText(c.BadgeText).
		SetIsActive(c.IsActive).
		SetPriority(c.Priority)
	if c.TitleI18n != nil {
		builder.SetTitleI18n(c.TitleI18n)
	}
	ent, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update creative: %w", err)
	}
	return entityToAdCreativeDTO(ent), nil
}

func (r *repo) DeleteCreative(ctx context.Context, id string) error {
	return r.db.AdCreative.DeleteOneID(id).Exec(ctx)
}

func (r *repo) ListCreativeIDsByPlacement(ctx context.Context, placementID string) ([]string, error) {
	placement, err := r.db.AdPlacement.Query().
		Where(adplacement.IDEQ(placementID)).
		WithCreatives().
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get placement creatives: %w", err)
	}
	ids := make([]string, 0, len(placement.Edges.Creatives))
	for _, c := range placement.Edges.Creatives {
		ids = append(ids, c.ID)
	}
	return ids, nil
}

func (r *repo) AssignCreative(ctx context.Context, placementID, creativeID string) error {
	creative, err := r.db.AdCreative.Get(ctx, creativeID)
	if err != nil {
		return fmt.Errorf("get creative: %w", err)
	}
	if _, err := r.db.AdPlacement.UpdateOneID(placementID).
		AddCreatives(creative).
		Save(ctx); err != nil {
		return fmt.Errorf("assign creative: %w", err)
	}
	return nil
}

func (r *repo) UnassignCreative(ctx context.Context, placementID, creativeID string) error {
	creative, err := r.db.AdCreative.Get(ctx, creativeID)
	if err != nil {
		return fmt.Errorf("get creative: %w", err)
	}
	if _, err := r.db.AdPlacement.UpdateOneID(placementID).
		RemoveCreatives(creative).
		Save(ctx); err != nil {
		return fmt.Errorf("unassign creative: %w", err)
	}
	return nil
}
