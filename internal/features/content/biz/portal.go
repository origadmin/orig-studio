package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origstudio/internal/data/entity"
	systembiz "origadmin/application/origstudio/internal/features/system/biz"
)

type PortalRepo interface {
	ListNavItems(ctx context.Context) ([]*entity.PortalNavItem, error)
	CreateNavItem(ctx context.Context, item *entity.PortalNavItem) (*entity.PortalNavItem, error)
	GetNavItemByID(ctx context.Context, id string) (*entity.PortalNavItem, error)
	UpdateNavItem(ctx context.Context, item *entity.PortalNavItem) (*entity.PortalNavItem, error)
	DeleteNavItem(ctx context.Context, id string) error
	ReorderNavItems(ctx context.Context, ids []string) error

	ListBanners(ctx context.Context) ([]*entity.PortalBanner, error)
	CreateBanner(ctx context.Context, b *entity.PortalBanner) (*entity.PortalBanner, error)
	UpdateBanner(ctx context.Context, b *entity.PortalBanner) (*entity.PortalBanner, error)
	DeleteBanner(ctx context.Context, id string) error
	ToggleBanner(ctx context.Context, id string) (*entity.PortalBanner, error)
	ListActiveBanners(ctx context.Context) ([]*entity.PortalBanner, error)

	ListCustomPages(ctx context.Context, page, pageSize int) ([]*entity.PortalCustomPage, int, error)
	CreateCustomPage(ctx context.Context, p *entity.PortalCustomPage) (*entity.PortalCustomPage, error)
	GetCustomPageByID(ctx context.Context, id string) (*entity.PortalCustomPage, error)
	GetCustomPageBySlug(ctx context.Context, slug string) (*entity.PortalCustomPage, error)
	UpdateCustomPage(ctx context.Context, p *entity.PortalCustomPage) (*entity.PortalCustomPage, error)
	DeleteCustomPage(ctx context.Context, id string) error
	IncrementPageViewCount(ctx context.Context, id string) error
	ListPublishedCustomPages(ctx context.Context) ([]*entity.PortalCustomPage, error)
}

type PortalConfigResponse struct {
	Modules    PortalModules       `json:"modules"`
	Layout     string              `json:"layout"`
	Site       PortalSite          `json:"site"`
	Navigation []*entity.PortalNavItem `json:"navigation"`
	Banners    []*entity.PortalBanner  `json:"banners"`
	Pages      []*entity.PortalCustomPage `json:"pages"`
}

type PortalModules struct {
	Articles bool `json:"articles"`
	Videos   bool `json:"videos"`
	Music    bool `json:"music"`
}

type PortalSite struct {
	SiteName          string   `json:"site_name"`
	SiteDescription   string   `json:"site_description"`
	AllowRegistration bool     `json:"allow_registration"`
	AllowUpload       bool     `json:"allow_upload"`
	PrimaryURL        string   `json:"primary_url"`
	AllowedURLs       []string `json:"allowed_urls"`
}

type PortalUseCase struct {
	repo        PortalRepo
	settingUC   *systembiz.SettingUseCase
	log         *log.Helper
}

func NewPortalUseCase(repo PortalRepo, settingUC *systembiz.SettingUseCase, logger log.Logger) *PortalUseCase {
	return &PortalUseCase{
		repo:      repo,
		settingUC: settingUC,
		log:       log.NewHelper(log.With(logger, "module", "portal.biz")),
	}
}

func (uc *PortalUseCase) ListNavItems(ctx context.Context) ([]*entity.PortalNavItem, error) {
	return uc.repo.ListNavItems(ctx)
}

func (uc *PortalUseCase) CreateNavItem(ctx context.Context, item *entity.PortalNavItem) (*entity.PortalNavItem, error) {
	return uc.repo.CreateNavItem(ctx, item)
}

func (uc *PortalUseCase) GetNavItemByID(ctx context.Context, id string) (*entity.PortalNavItem, error) {
	return uc.repo.GetNavItemByID(ctx, id)
}

func (uc *PortalUseCase) UpdateNavItem(ctx context.Context, item *entity.PortalNavItem) (*entity.PortalNavItem, error) {
	return uc.repo.UpdateNavItem(ctx, item)
}

func (uc *PortalUseCase) DeleteNavItem(ctx context.Context, id string) error {
	return uc.repo.DeleteNavItem(ctx, id)
}

func (uc *PortalUseCase) ReorderNavItems(ctx context.Context, ids []string) error {
	return uc.repo.ReorderNavItems(ctx, ids)
}

func (uc *PortalUseCase) ListBanners(ctx context.Context) ([]*entity.PortalBanner, error) {
	return uc.repo.ListBanners(ctx)
}

func (uc *PortalUseCase) CreateBanner(ctx context.Context, b *entity.PortalBanner) (*entity.PortalBanner, error) {
	return uc.repo.CreateBanner(ctx, b)
}

func (uc *PortalUseCase) UpdateBanner(ctx context.Context, b *entity.PortalBanner) (*entity.PortalBanner, error) {
	return uc.repo.UpdateBanner(ctx, b)
}

func (uc *PortalUseCase) DeleteBanner(ctx context.Context, id string) error {
	return uc.repo.DeleteBanner(ctx, id)
}

func (uc *PortalUseCase) ToggleBanner(ctx context.Context, id string) (*entity.PortalBanner, error) {
	return uc.repo.ToggleBanner(ctx, id)
}

func (uc *PortalUseCase) ListActiveBanners(ctx context.Context) ([]*entity.PortalBanner, error) {
	return uc.repo.ListActiveBanners(ctx)
}

func (uc *PortalUseCase) ListCustomPages(ctx context.Context, page, pageSize int) ([]*entity.PortalCustomPage, int, error) {
	return uc.repo.ListCustomPages(ctx, page, pageSize)
}

func (uc *PortalUseCase) CreateCustomPage(ctx context.Context, p *entity.PortalCustomPage) (*entity.PortalCustomPage, error) {
	return uc.repo.CreateCustomPage(ctx, p)
}

func (uc *PortalUseCase) GetCustomPageByID(ctx context.Context, id string) (*entity.PortalCustomPage, error) {
	return uc.repo.GetCustomPageByID(ctx, id)
}

func (uc *PortalUseCase) GetCustomPageBySlug(ctx context.Context, slug string) (*entity.PortalCustomPage, error) {
	return uc.repo.GetCustomPageBySlug(ctx, slug)
}

func (uc *PortalUseCase) UpdateCustomPage(ctx context.Context, p *entity.PortalCustomPage) (*entity.PortalCustomPage, error) {
	return uc.repo.UpdateCustomPage(ctx, p)
}

func (uc *PortalUseCase) DeleteCustomPage(ctx context.Context, id string) error {
	return uc.repo.DeleteCustomPage(ctx, id)
}

func (uc *PortalUseCase) IncrementPageViewCount(ctx context.Context, id string) error {
	return uc.repo.IncrementPageViewCount(ctx, id)
}

func (uc *PortalUseCase) ListPublishedCustomPages(ctx context.Context) ([]*entity.PortalCustomPage, error) {
	return uc.repo.ListPublishedCustomPages(ctx)
}
