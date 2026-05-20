package dal

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origstudio/internal/dal/entity"
	"origadmin/application/origstudio/internal/dal/entity/portalbanner"
	"origadmin/application/origstudio/internal/dal/entity/portalcustompage"
	"origadmin/application/origstudio/internal/dal/entity/portalnavitem"
	"origadmin/application/origstudio/internal/features/content/biz"
	"origadmin/application/origstudio/internal/features/content/dto"
)

// ==================== Entity-to-DTO Conversion ====================

func entityToPortalNavItemDTO(e *entity.PortalNavItem) *dto.PortalNavItemDTO {
	if e == nil {
		return nil
	}
	return &dto.PortalNavItemDTO{
		ID:         e.ID,
		Type:       e.Type,
		Label:      e.Label,
		LabelI18n:  e.LabelI18n,
		URL:        e.URL,
		TargetType: e.TargetType,
		TargetID:   e.TargetID,
		Icon:       e.Icon,
		Color:      e.Color,
		Sequence:   e.Sequence,
		ParentID:   e.ParentID,
		IsVisible:  e.IsVisible,
		OpenNewTab: e.OpenNewTab,
		CSSClass:   e.CSSClass,
		// Legacy compatibility
		Target:    e.TargetType,
		SortOrder: e.Sequence,
		IsActive:  e.IsVisible,
	}
}

func entityToPortalBannerDTO(e *entity.PortalBanner) *dto.PortalBannerDTO {
	if e == nil {
		return nil
	}
	return &dto.PortalBannerDTO{
		ID:                e.ID,
		Title:             e.Title,
		TitleI18n:         e.TitleI18n,
		Subtitle:          e.Subtitle,
		SubtitleI18n:      e.SubtitleI18n,
		BadgeText:         e.BadgeText,
		ImageURL:          e.ImageURL,
		ImageMobileURL:    e.ImageMobileURL,
		BgColorStart:      e.BgColorStart,
		BgColorEnd:        e.BgColorEnd,
		BgOverlayOpacity:  e.BgOverlayOpacity,
		PrimaryBtnText:    e.PrimaryBtnText,
		PrimaryBtnURL:     e.PrimaryBtnURL,
		SecondaryBtnText:  e.SecondaryBtnText,
		SecondaryBtnURL:   e.SecondaryBtnURL,
		Sequence:          e.Sequence,
		IsActive:          e.IsActive,
		StartAt:           e.StartAt,
		EndAt:             e.EndAt,
		AutoSlideInterval: e.AutoSlideInterval,
		// Legacy compatibility
		LinkURL:  e.PrimaryBtnURL,
		SortOrder: e.Sequence,
		StartTime: e.StartAt,
		EndTime:   e.EndAt,
	}
}

func entityToPortalCustomPageDTO(e *entity.PortalCustomPage) *dto.PortalCustomPageDTO {
	if e == nil {
		return nil
	}
	return &dto.PortalCustomPageDTO{
		ID:             e.ID,
		Title:          e.Title,
		Slug:           e.Slug,
		Type:           e.Type,
		ContentFormat:  e.ContentFormat,
		Content:        e.Content,
		Layout:         e.Layout,
		IsPublished:    e.IsPublished,
		PublishedAt:    e.PublishedAt,
		SeoTitle:       e.SeoTitle,
		SeoDescription: e.SeoDescription,
		FeaturedImage:  e.FeaturedImage,
		ViewCount:      e.ViewCount,
	}
}

type portalRepo struct {
	data *Data
	log  *log.Helper
}

func NewPortalRepo(data *Data, logger log.Logger) biz.PortalRepo {
	return &portalRepo{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "portal.data")),
	}
}

// ==================== NavItem ====================

func (r *portalRepo) ListNavItems(ctx context.Context) ([]*dto.PortalNavItemDTO, error) {
	items, err := r.data.db.PortalNavItem.Query().
		Order(entity.Asc(portalnavitem.FieldSequence)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list nav items: %w", err)
	}
	result := make([]*dto.PortalNavItemDTO, len(items))
	for i, item := range items {
		result[i] = entityToPortalNavItemDTO(item)
	}
	return result, nil
}

func (r *portalRepo) CreateNavItem(ctx context.Context, item *dto.PortalNavItemDTO) (*dto.PortalNavItemDTO, error) {
	builder := r.data.db.PortalNavItem.Create().
		SetType(item.Type).
		SetLabel(item.Label).
		SetSequence(item.Sequence).
		SetIsVisible(item.IsVisible).
		SetOpenNewTab(item.OpenNewTab)

	if item.LabelI18n != nil {
		builder.SetLabelI18n(item.LabelI18n)
	}
	if item.URL != "" {
		builder.SetURL(item.URL)
	}
	if item.TargetType != "" {
		builder.SetTargetType(item.TargetType)
	}
	if item.TargetID != "" {
		builder.SetTargetID(item.TargetID)
	}
	if item.Icon != "" {
		builder.SetIcon(item.Icon)
	}
	if item.Color != "" {
		builder.SetColor(item.Color)
	}
	if item.ParentID != "" {
		builder.SetParentID(item.ParentID)
	}
	if item.CSSClass != "" {
		builder.SetCSSClass(item.CSSClass)
	}

	ent, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create nav item: %w", err)
	}
	return entityToPortalNavItemDTO(ent), nil
}

func (r *portalRepo) UpdateNavItem(ctx context.Context, item *dto.PortalNavItemDTO) (*dto.PortalNavItemDTO, error) {
	builder := r.data.db.PortalNavItem.UpdateOneID(item.ID).
		SetType(item.Type).
		SetLabel(item.Label).
		SetSequence(item.Sequence).
		SetIsVisible(item.IsVisible).
		SetOpenNewTab(item.OpenNewTab)

	if item.LabelI18n != nil {
		builder.SetLabelI18n(item.LabelI18n)
	}
	builder.SetURL(item.URL)
	builder.SetTargetType(item.TargetType)
	builder.SetTargetID(item.TargetID)
	builder.SetIcon(item.Icon)
	builder.SetColor(item.Color)
	builder.SetParentID(item.ParentID)
	builder.SetCSSClass(item.CSSClass)

	ent, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update nav item: %w", err)
	}
	return entityToPortalNavItemDTO(ent), nil
}

func (r *portalRepo) DeleteNavItem(ctx context.Context, id string) error {
	return r.data.db.PortalNavItem.DeleteOneID(id).Exec(ctx)
}

func (r *portalRepo) ReorderNavItems(ctx context.Context, ids []string) error {
	for i, id := range ids {
		err := r.data.db.PortalNavItem.UpdateOneID(id).
			SetSequence(i).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("reorder nav item %s: %w", id, err)
		}
	}
	return nil
}

func (r *portalRepo) GetNavItemByID(ctx context.Context, id string) (*dto.PortalNavItemDTO, error) {
	item, err := r.data.db.PortalNavItem.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get nav item: %w", err)
	}
	return entityToPortalNavItemDTO(item), nil
}

// ==================== Banner ====================

func (r *portalRepo) ListBanners(ctx context.Context) ([]*dto.PortalBannerDTO, error) {
	items, err := r.data.db.PortalBanner.Query().
		Order(entity.Asc(portalbanner.FieldSequence)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list banners: %w", err)
	}
	result := make([]*dto.PortalBannerDTO, len(items))
	for i, item := range items {
		result[i] = entityToPortalBannerDTO(item)
	}
	return result, nil
}

func (r *portalRepo) CreateBanner(ctx context.Context, b *dto.PortalBannerDTO) (*dto.PortalBannerDTO, error) {
	builder := r.data.db.PortalBanner.Create().
		SetTitle(b.Title).
		SetSequence(b.Sequence).
		SetIsActive(b.IsActive).
		SetAutoSlideInterval(b.AutoSlideInterval)

	if b.TitleI18n != nil {
		builder.SetTitleI18n(b.TitleI18n)
	}
	if b.Subtitle != "" {
		builder.SetSubtitle(b.Subtitle)
	}
	if b.SubtitleI18n != nil {
		builder.SetSubtitleI18n(b.SubtitleI18n)
	}
	if b.BadgeText != "" {
		builder.SetBadgeText(b.BadgeText)
	}
	if b.ImageURL != "" {
		builder.SetImageURL(b.ImageURL)
	}
	if b.ImageMobileURL != "" {
		builder.SetImageMobileURL(b.ImageMobileURL)
	}
	if b.BgColorStart != "" {
		builder.SetBgColorStart(b.BgColorStart)
	}
	if b.BgColorEnd != "" {
		builder.SetBgColorEnd(b.BgColorEnd)
	}
	if b.BgOverlayOpacity != 0 {
		builder.SetBgOverlayOpacity(b.BgOverlayOpacity)
	}
	if b.PrimaryBtnText != "" {
		builder.SetPrimaryBtnText(b.PrimaryBtnText)
	}
	if b.PrimaryBtnURL != "" {
		builder.SetPrimaryBtnURL(b.PrimaryBtnURL)
	}
	if b.SecondaryBtnText != "" {
		builder.SetSecondaryBtnText(b.SecondaryBtnText)
	}
	if b.SecondaryBtnURL != "" {
		builder.SetSecondaryBtnURL(b.SecondaryBtnURL)
	}
	if !b.StartAt.IsZero() {
		builder.SetStartAt(b.StartAt)
	}
	if !b.EndAt.IsZero() {
		builder.SetEndAt(b.EndAt)
	}

	ent, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create banner: %w", err)
	}
	return entityToPortalBannerDTO(ent), nil
}

func (r *portalRepo) UpdateBanner(ctx context.Context, b *dto.PortalBannerDTO) (*dto.PortalBannerDTO, error) {
	builder := r.data.db.PortalBanner.UpdateOneID(b.ID).
		SetTitle(b.Title).
		SetSequence(b.Sequence).
		SetIsActive(b.IsActive).
		SetAutoSlideInterval(b.AutoSlideInterval)

	if b.TitleI18n != nil {
		builder.SetTitleI18n(b.TitleI18n)
	}
	builder.SetSubtitle(b.Subtitle)
	if b.SubtitleI18n != nil {
		builder.SetSubtitleI18n(b.SubtitleI18n)
	}
	builder.SetBadgeText(b.BadgeText)
	builder.SetImageURL(b.ImageURL)
	builder.SetImageMobileURL(b.ImageMobileURL)
	builder.SetBgColorStart(b.BgColorStart)
	builder.SetBgColorEnd(b.BgColorEnd)
	builder.SetBgOverlayOpacity(b.BgOverlayOpacity)
	builder.SetPrimaryBtnText(b.PrimaryBtnText)
	builder.SetPrimaryBtnURL(b.PrimaryBtnURL)
	builder.SetSecondaryBtnText(b.SecondaryBtnText)
	builder.SetSecondaryBtnURL(b.SecondaryBtnURL)
	if !b.StartAt.IsZero() {
		builder.SetStartAt(b.StartAt)
	}
	if !b.EndAt.IsZero() {
		builder.SetEndAt(b.EndAt)
	}

	ent, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update banner: %w", err)
	}
	return entityToPortalBannerDTO(ent), nil
}

func (r *portalRepo) DeleteBanner(ctx context.Context, id string) error {
	return r.data.db.PortalBanner.DeleteOneID(id).Exec(ctx)
}

func (r *portalRepo) ToggleBanner(ctx context.Context, id string) (*dto.PortalBannerDTO, error) {
	ent, err := r.data.db.PortalBanner.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get banner for toggle: %w", err)
	}
	updated, err := r.data.db.PortalBanner.UpdateOneID(id).
		SetIsActive(!ent.IsActive).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("toggle banner: %w", err)
	}
	return entityToPortalBannerDTO(updated), nil
}

func (r *portalRepo) ListActiveBanners(ctx context.Context) ([]*dto.PortalBannerDTO, error) {
	now := time.Now()
	query := r.data.db.PortalBanner.Query().
		Where(portalbanner.IsActiveEQ(true)).
		Order(entity.Asc(portalbanner.FieldSequence))

	items, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list active banners: %w", err)
	}

	var filtered []*entity.PortalBanner
	for _, b := range items {
		if !b.StartAt.IsZero() && b.StartAt.After(now) {
			continue
		}
		if !b.EndAt.IsZero() && b.EndAt.Before(now) {
			continue
		}
		filtered = append(filtered, b)
	}

	result := make([]*dto.PortalBannerDTO, len(filtered))
	for i, b := range filtered {
		result[i] = entityToPortalBannerDTO(b)
	}
	return result, nil
}

// ==================== CustomPage ====================

func (r *portalRepo) ListCustomPages(ctx context.Context, page, pageSize int) ([]*dto.PortalCustomPageDTO, int, error) {
	query := r.data.db.PortalCustomPage.Query()
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count custom pages: %w", err)
	}

	items, err := query.
		Order(entity.Desc(portalcustompage.FieldID)).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list custom pages: %w", err)
	}
	result := make([]*dto.PortalCustomPageDTO, len(items))
	for i, item := range items {
		result[i] = entityToPortalCustomPageDTO(item)
	}
	return result, total, nil
}

func (r *portalRepo) CreateCustomPage(ctx context.Context, p *dto.PortalCustomPageDTO) (*dto.PortalCustomPageDTO, error) {
	builder := r.data.db.PortalCustomPage.Create().
		SetTitle(p.Title).
		SetSlug(p.Slug).
		SetIsPublished(p.IsPublished)

	if p.Type != "" {
		builder.SetType(p.Type)
	}
	if p.ContentFormat != "" {
		builder.SetContentFormat(p.ContentFormat)
	}
	if p.Content != "" {
		builder.SetContent(p.Content)
	}
	if p.Layout != "" {
		builder.SetLayout(p.Layout)
	}
	if !p.PublishedAt.IsZero() {
		builder.SetPublishedAt(p.PublishedAt)
	}
	if p.SeoTitle != "" {
		builder.SetSeoTitle(p.SeoTitle)
	}
	if p.SeoDescription != "" {
		builder.SetSeoDescription(p.SeoDescription)
	}
	if p.FeaturedImage != "" {
		builder.SetFeaturedImage(p.FeaturedImage)
	}

	ent, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create custom page: %w", err)
	}
	return entityToPortalCustomPageDTO(ent), nil
}

func (r *portalRepo) GetCustomPageByID(ctx context.Context, id string) (*dto.PortalCustomPageDTO, error) {
	ent, err := r.data.db.PortalCustomPage.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get custom page: %w", err)
	}
	return entityToPortalCustomPageDTO(ent), nil
}

func (r *portalRepo) GetCustomPageBySlug(ctx context.Context, slug string) (*dto.PortalCustomPageDTO, error) {
	ent, err := r.data.db.PortalCustomPage.Query().
		Where(portalcustompage.SlugEQ(slug)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get custom page by slug: %w", err)
	}
	return entityToPortalCustomPageDTO(ent), nil
}

func (r *portalRepo) UpdateCustomPage(ctx context.Context, p *dto.PortalCustomPageDTO) (*dto.PortalCustomPageDTO, error) {
	builder := r.data.db.PortalCustomPage.UpdateOneID(p.ID).
		SetTitle(p.Title).
		SetSlug(p.Slug).
		SetIsPublished(p.IsPublished)

	if p.Type != "" {
		builder.SetType(p.Type)
	}
	if p.ContentFormat != "" {
		builder.SetContentFormat(p.ContentFormat)
	}
	builder.SetContent(p.Content)
	if p.Layout != "" {
		builder.SetLayout(p.Layout)
	}
	if !p.PublishedAt.IsZero() {
		builder.SetPublishedAt(p.PublishedAt)
	}
	builder.SetSeoTitle(p.SeoTitle)
	builder.SetSeoDescription(p.SeoDescription)
	builder.SetFeaturedImage(p.FeaturedImage)

	ent, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update custom page: %w", err)
	}
	return entityToPortalCustomPageDTO(ent), nil
}

func (r *portalRepo) DeleteCustomPage(ctx context.Context, id string) error {
	return r.data.db.PortalCustomPage.DeleteOneID(id).Exec(ctx)
}

func (r *portalRepo) IncrementPageViewCount(ctx context.Context, id string) error {
	_, err := r.data.db.PortalCustomPage.UpdateOneID(id).
		AddViewCount(1).
		Save(ctx)
	return err
}

func (r *portalRepo) ListPublishedCustomPages(ctx context.Context) ([]*dto.PortalCustomPageDTO, error) {
	items, err := r.data.db.PortalCustomPage.Query().
		Where(portalcustompage.IsPublishedEQ(true)).
		Order(entity.Desc(portalcustompage.FieldPublishedAt)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list published custom pages: %w", err)
	}
	result := make([]*dto.PortalCustomPageDTO, len(items))
	for i, item := range items {
		result[i] = entityToPortalCustomPageDTO(item)
	}
	return result, nil
}
