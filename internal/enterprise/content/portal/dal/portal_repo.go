package dal

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origstudio/internal/dal/entity"
	"origadmin/application/origstudio/internal/dal/entity/media"
	"origadmin/application/origstudio/internal/dal/entity/portalbanner"
	"origadmin/application/origstudio/internal/dal/entity/portalcustompage"
	"origadmin/application/origstudio/internal/dal/entity/portalnavitem"
	"origadmin/application/origstudio/internal/enterprise/content/portal/biz"
	"origadmin/application/origstudio/internal/enterprise/content/portal/dto"
)

type repo struct {
	db  *entity.Client
	log *log.Helper
}

func NewRepo(db *entity.Client, logger log.Logger) biz.Repo {
	return &repo{
		db:  db,
		log: log.NewHelper(log.With(logger, "module", "enterprise/portal.repo")),
	}
}

func entityToNavItemDTO(e *entity.PortalNavItem) *dto.PortalNavItemDTO {
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
	}
}

func entityToBannerDTO(e *entity.PortalBanner) *dto.PortalBannerDTO {
	if e == nil {
		return nil
	}
	d := &dto.PortalBannerDTO{
		ID:                e.ID,
		Title:             e.Title,
		TitleI18n:         e.TitleI18n,
		Subtitle:          e.Subtitle,
		SubtitleI18n:      e.SubtitleI18n,
		BadgeText:         e.BadgeText,
		ImageURL:          e.ImageURL,
		ImageMobileURL:    e.ImageMobileURL,
		VideoURL:          e.VideoURL,
		BgColorStart:      e.BgColorStart,
		BgColorEnd:        e.BgColorEnd,
		BgOverlayOpacity:  e.BgOverlayOpacity,
		PrimaryBtnText:    e.PrimaryBtnText,
		PrimaryBtnURL:     e.PrimaryBtnURL,
		SecondaryBtnText:  e.SecondaryBtnText,
		SecondaryBtnURL:   e.SecondaryBtnURL,
		Sequence:          e.Sequence,
		IsActive:          e.IsActive,
		AutoSlideInterval: e.AutoSlideInterval,
		Type:              e.Type,
		Count:             e.Count,
		CategoryID:        e.CategoryID,
		DisplayMode:       e.DisplayMode,
	}
	if !e.StartAt.IsZero() {
		t := e.StartAt
		d.StartAt = &t
		d.StartTime = &t
	}
	if !e.EndAt.IsZero() {
		t := e.EndAt
		d.EndAt = &t
		d.EndTime = &t
	}
	return d
}

func entityToCustomPageDTO(e *entity.PortalCustomPage) *dto.PortalCustomPageDTO {
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

func (r *repo) ListNavItems(ctx context.Context) ([]*dto.PortalNavItemDTO, error) {
	items, err := r.db.PortalNavItem.Query().
		Order(entity.Asc(portalnavitem.FieldSequence)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list nav items: %w", err)
	}
	result := make([]*dto.PortalNavItemDTO, len(items))
	for i, item := range items {
		result[i] = entityToNavItemDTO(item)
	}
	return result, nil
}

func (r *repo) CreateNavItem(ctx context.Context, item *dto.PortalNavItemDTO) (*dto.PortalNavItemDTO, error) {
	builder := r.db.PortalNavItem.Create().
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
	return entityToNavItemDTO(ent), nil
}

func (r *repo) GetNavItemByID(ctx context.Context, id string) (*dto.PortalNavItemDTO, error) {
	item, err := r.db.PortalNavItem.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get nav item: %w", err)
	}
	return entityToNavItemDTO(item), nil
}

func (r *repo) UpdateNavItem(ctx context.Context, item *dto.PortalNavItemDTO) (*dto.PortalNavItemDTO, error) {
	builder := r.db.PortalNavItem.UpdateOneID(item.ID).
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
	return entityToNavItemDTO(ent), nil
}

func (r *repo) DeleteNavItem(ctx context.Context, id string) error {
	return r.db.PortalNavItem.DeleteOneID(id).Exec(ctx)
}

func (r *repo) ReorderNavItems(ctx context.Context, ids []string) error {
	for i, id := range ids {
		_, err := r.db.PortalNavItem.UpdateOneID(id).
			SetSequence(i).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("reorder nav item %s: %w", id, err)
		}
	}
	return nil
}

func (r *repo) ListBanners(ctx context.Context) ([]*dto.PortalBannerDTO, error) {
	items, err := r.db.PortalBanner.Query().
		Order(entity.Asc(portalbanner.FieldSequence)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list banners: %w", err)
	}
	result := make([]*dto.PortalBannerDTO, len(items))
	for i, item := range items {
		result[i] = entityToBannerDTO(item)
	}
	return result, nil
}

func (r *repo) GetBanner(ctx context.Context, id string) (*dto.PortalBannerDTO, error) {
	b, err := r.db.PortalBanner.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get banner: %w", err)
	}
	return entityToBannerDTO(b), nil
}

func (r *repo) CreateBanner(ctx context.Context, b *dto.PortalBannerDTO) (*dto.PortalBannerDTO, error) {
	builder := r.db.PortalBanner.Create().
		SetTitle(b.Title).
		SetSequence(b.Sequence).
		SetIsActive(b.IsActive).
		SetAutoSlideInterval(b.AutoSlideInterval).
		SetType(b.Type).
		SetCount(b.Count).
		SetDisplayMode(b.DisplayMode)
	if b.CategoryID != "" {
		builder.SetCategoryID(b.CategoryID)
	}
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
	if b.VideoURL != "" {
		builder.SetVideoURL(b.VideoURL)
	}
	if b.BgColorStart != "" {
		builder.SetBgColorStart(b.BgColorStart)
	}
	if b.BgColorEnd != "" {
		builder.SetBgColorEnd(b.BgColorEnd)
	}
	builder.SetBgOverlayOpacity(b.BgOverlayOpacity)
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
	if b.StartAt != nil && !b.StartAt.IsZero() {
		builder.SetStartAt(*b.StartAt)
	}
	if b.EndAt != nil && !b.EndAt.IsZero() {
		builder.SetEndAt(*b.EndAt)
	}
	ent, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create banner: %w", err)
	}
	return entityToBannerDTO(ent), nil
}

func (r *repo) UpdateBanner(ctx context.Context, b *dto.PortalBannerDTO) (*dto.PortalBannerDTO, error) {
	builder := r.db.PortalBanner.UpdateOneID(b.ID).
		SetTitle(b.Title).
		SetSequence(b.Sequence).
		SetIsActive(b.IsActive).
		SetAutoSlideInterval(b.AutoSlideInterval).
		SetType(b.Type).
		SetCount(b.Count).
		SetDisplayMode(b.DisplayMode)
	if b.CategoryID != "" {
		builder.SetCategoryID(b.CategoryID)
	}
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
	builder.SetVideoURL(b.VideoURL)
	builder.SetBgColorStart(b.BgColorStart)
	builder.SetBgColorEnd(b.BgColorEnd)
	builder.SetBgOverlayOpacity(b.BgOverlayOpacity)
	builder.SetPrimaryBtnText(b.PrimaryBtnText)
	builder.SetPrimaryBtnURL(b.PrimaryBtnURL)
	builder.SetSecondaryBtnText(b.SecondaryBtnText)
	builder.SetSecondaryBtnURL(b.SecondaryBtnURL)
	if b.ClearStartAt {
		builder.ClearStartAt()
	} else if b.StartAt != nil && !b.StartAt.IsZero() {
		builder.SetStartAt(*b.StartAt)
	}
	if b.ClearEndAt {
		builder.ClearEndAt()
	} else if b.EndAt != nil && !b.EndAt.IsZero() {
		builder.SetEndAt(*b.EndAt)
	}
	ent, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update banner: %w", err)
	}
	return entityToBannerDTO(ent), nil
}

func (r *repo) DeleteBanner(ctx context.Context, id string) error {
	return r.db.PortalBanner.DeleteOneID(id).Exec(ctx)
}

func (r *repo) ToggleBanner(ctx context.Context, id string) (*dto.PortalBannerDTO, error) {
	ent, err := r.db.PortalBanner.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get banner for toggle: %w", err)
	}
	updated, err := r.db.PortalBanner.UpdateOneID(id).
		SetIsActive(!ent.IsActive).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("toggle banner: %w", err)
	}
	return entityToBannerDTO(updated), nil
}

func (r *repo) ListActiveBanners(ctx context.Context) ([]*dto.PortalBannerDTO, error) {
	now := time.Now()
	items, err := r.db.PortalBanner.Query().
		Where(portalbanner.IsActiveEQ(true)).
		Order(entity.Asc(portalbanner.FieldSequence)).
		All(ctx)
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
		dtoItem := entityToBannerDTO(b)
		r.enrichDynamicBanner(ctx, dtoItem)
		result[i] = dtoItem
	}
	return result, nil
}

// enrichDynamicBanner：视频类 banner 在主页按「一条对应一条」展示。
// 方案 A（用户裁定）：banner 显式绑定 1 个视频 —— 绑定方式为 primary_btn_url=/watch?v=<short_token>
// （后台选视频时自动写入海报与跳转）；后端按 token 解析海报/标题（为空时填充），点击进该视频播放页。
// legacy hot_videos/new_videos（无 watch 链接）回退为规则取顶 1 条（最火/最新）。
func (r *repo) enrichDynamicBanner(ctx context.Context, b *dto.PortalBannerDTO) {
	// 方案 A：primary_btn_url 形如 /watch?v=<token> → 解析并填充该视频
	if token := watchTokenFromURL(b.PrimaryBtnURL); token != "" {
		top, err := r.db.Media.Query().Where(media.ShortTokenEQ(token)).Only(ctx)
		if err == nil && top != nil {
			r.applyVideoToBanner(b, top)
		}
		return
	}
	// legacy 兜底：hot_videos / new_videos 按规则取顶 1 条
	if b.Type != "hot_videos" && b.Type != "new_videos" {
		return
	}
	limit := b.Count
	if limit <= 0 {
		limit = 5
	}
	q := r.db.Media.Query().Where(media.StateEQ("active")).Limit(limit)
	if b.Type == "hot_videos" {
		q = q.Order(entity.Desc(media.FieldViewCount))
	} else {
		q = q.Order(entity.Desc(media.FieldCreateTime))
	}
	videos, err := q.All(ctx)
	if err != nil || len(videos) == 0 {
		return
	}
	r.applyVideoToBanner(b, videos[0])
}

// watchTokenFromURL：从 /watch?v=<token>[&...] 提取 short_token。
func watchTokenFromURL(u string) string {
	const prefix = "/watch?v="
	if !strings.HasPrefix(u, prefix) {
		return ""
	}
	tok := strings.TrimPrefix(u, prefix)
	if i := strings.IndexByte(tok, '&'); i >= 0 {
		tok = tok[:i]
	}
	return tok
}

// applyVideoToBanner：把单个视频的信息填进 banner（海报/标题/播放地址/跳转）。
// 海报/标题只在为空时填充（管理员自定义优先）；跳转始终指向该视频播放页（绑定视频权威）。
func (r *repo) applyVideoToBanner(b *dto.PortalBannerDTO, v *entity.Media) {
	if b.ImageURL == "" {
		if v.Poster != "" {
			b.ImageURL = v.Poster
		} else if v.Thumbnail != "" {
			b.ImageURL = v.Thumbnail
		}
	}
	if b.VideoURL == "" {
		if v.HlsFile != "" {
			b.VideoURL = v.HlsFile
		} else if v.URL != "" {
			b.VideoURL = v.URL
		}
	}
	if b.Title == "" && v.Title != "" {
		b.Title = v.Title
	}
	// BUG-228(v6)：不再用 /videos 死链；点击进绑定视频播放页。
	if v.ShortToken != "" {
		b.PrimaryBtnURL = "/watch?v=" + v.ShortToken
	} else {
		b.PrimaryBtnURL = "/latest"
	}
	if b.PrimaryBtnText == "" {
		b.PrimaryBtnText = "立即观看"
	}
}

func (r *repo) ListCustomPages(ctx context.Context, page, pageSize int) ([]*dto.PortalCustomPageDTO, int, error) {
	query := r.db.PortalCustomPage.Query()
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
		result[i] = entityToCustomPageDTO(item)
	}
	return result, total, nil
}

func (r *repo) CreateCustomPage(ctx context.Context, p *dto.PortalCustomPageDTO) (*dto.PortalCustomPageDTO, error) {
	builder := r.db.PortalCustomPage.Create().
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
	return entityToCustomPageDTO(ent), nil
}

func (r *repo) GetCustomPageByID(ctx context.Context, id string) (*dto.PortalCustomPageDTO, error) {
	item, err := r.db.PortalCustomPage.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get custom page: %w", err)
	}
	return entityToCustomPageDTO(item), nil
}

func (r *repo) GetCustomPageBySlug(ctx context.Context, slug string) (*dto.PortalCustomPageDTO, error) {
	item, err := r.db.PortalCustomPage.Query().
		Where(portalcustompage.SlugEQ(slug)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get custom page by slug: %w", err)
	}
	return entityToCustomPageDTO(item), nil
}

func (r *repo) UpdateCustomPage(ctx context.Context, p *dto.PortalCustomPageDTO) (*dto.PortalCustomPageDTO, error) {
	builder := r.db.PortalCustomPage.UpdateOneID(p.ID).
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
	return entityToCustomPageDTO(ent), nil
}

func (r *repo) DeleteCustomPage(ctx context.Context, id string) error {
	return r.db.PortalCustomPage.DeleteOneID(id).Exec(ctx)
}

func (r *repo) IncrementPageViewCount(ctx context.Context, id string) error {
	_, err := r.db.PortalCustomPage.UpdateOneID(id).
		AddViewCount(1).
		Save(ctx)
	return err
}

func (r *repo) ListPublishedCustomPages(ctx context.Context) ([]*dto.PortalCustomPageDTO, error) {
	items, err := r.db.PortalCustomPage.Query().
		Where(portalcustompage.IsPublishedEQ(true)).
		Order(entity.Desc(portalcustompage.FieldID)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list published custom pages: %w", err)
	}
	result := make([]*dto.PortalCustomPageDTO, len(items))
	for i, item := range items {
		result[i] = entityToCustomPageDTO(item)
	}
	return result, nil
}
