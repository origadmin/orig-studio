package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origstudio/internal/enterprise/content/portal/dto"
	systembiz "origadmin/application/origstudio/internal/features/system/biz"
)

type Repo interface {
	ListNavItems(ctx context.Context) ([]*dto.PortalNavItemDTO, error)
	CreateNavItem(ctx context.Context, item *dto.PortalNavItemDTO) (*dto.PortalNavItemDTO, error)
	GetNavItemByID(ctx context.Context, id string) (*dto.PortalNavItemDTO, error)
	UpdateNavItem(ctx context.Context, item *dto.PortalNavItemDTO) (*dto.PortalNavItemDTO, error)
	DeleteNavItem(ctx context.Context, id string) error
	ReorderNavItems(ctx context.Context, ids []string) error

	ListBanners(ctx context.Context) ([]*dto.PortalBannerDTO, error)
	GetBanner(ctx context.Context, id string) (*dto.PortalBannerDTO, error)
	CreateBanner(ctx context.Context, b *dto.PortalBannerDTO) (*dto.PortalBannerDTO, error)
	UpdateBanner(ctx context.Context, b *dto.PortalBannerDTO) (*dto.PortalBannerDTO, error)
	DeleteBanner(ctx context.Context, id string) error
	ToggleBanner(ctx context.Context, id string) (*dto.PortalBannerDTO, error)
	ListActiveBanners(ctx context.Context) ([]*dto.PortalBannerDTO, error)

	ListCustomPages(ctx context.Context, page, pageSize int) ([]*dto.PortalCustomPageDTO, int, error)
	CreateCustomPage(ctx context.Context, p *dto.PortalCustomPageDTO) (*dto.PortalCustomPageDTO, error)
	GetCustomPageByID(ctx context.Context, id string) (*dto.PortalCustomPageDTO, error)
	GetCustomPageBySlug(ctx context.Context, slug string) (*dto.PortalCustomPageDTO, error)
	UpdateCustomPage(ctx context.Context, p *dto.PortalCustomPageDTO) (*dto.PortalCustomPageDTO, error)
	DeleteCustomPage(ctx context.Context, id string) error
	IncrementPageViewCount(ctx context.Context, id string) error
	ListPublishedCustomPages(ctx context.Context) ([]*dto.PortalCustomPageDTO, error)
}

type ConfigResponse struct {
	Modules    PortalModules               `json:"modules"`
	Layout     string                      `json:"layout"`
	Site       PortalSite                  `json:"site"`
	Navigation []*dto.PortalNavItemDTO     `json:"navigation"`
	Banners    []*dto.PortalBannerDTO      `json:"banners"`
	Pages      []*dto.PortalCustomPageDTO  `json:"pages"`
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
	SiteLogoURL       string   `json:"site_logo_url"`
}

type UseCase struct {
	repo       Repo
	settingUC  *systembiz.SettingUseCase
	log        *log.Helper
}

func NewUseCase(repo Repo, settingUC *systembiz.SettingUseCase, logger log.Logger) *UseCase {
	return &UseCase{
		repo:      repo,
		settingUC: settingUC,
		log:       log.NewHelper(log.With(logger, "module", "enterprise/portal.biz")),
	}
}

func (uc *UseCase) ListNavItems(ctx context.Context) ([]*dto.PortalNavItemDTO, error) {
	return uc.repo.ListNavItems(ctx)
}

func (uc *UseCase) CreateNavItem(ctx context.Context, item *dto.PortalNavItemDTO) (*dto.PortalNavItemDTO, error) {
	return uc.repo.CreateNavItem(ctx, item)
}

func (uc *UseCase) GetNavItemByID(ctx context.Context, id string) (*dto.PortalNavItemDTO, error) {
	return uc.repo.GetNavItemByID(ctx, id)
}

func (uc *UseCase) UpdateNavItem(ctx context.Context, item *dto.PortalNavItemDTO) (*dto.PortalNavItemDTO, error) {
	return uc.repo.UpdateNavItem(ctx, item)
}

func (uc *UseCase) DeleteNavItem(ctx context.Context, id string) error {
	return uc.repo.DeleteNavItem(ctx, id)
}

func (uc *UseCase) ReorderNavItems(ctx context.Context, ids []string) error {
	return uc.repo.ReorderNavItems(ctx, ids)
}

func (uc *UseCase) ListBanners(ctx context.Context) ([]*dto.PortalBannerDTO, error) {
	return uc.repo.ListBanners(ctx)
}

func (uc *UseCase) GetBanner(ctx context.Context, id string) (*dto.PortalBannerDTO, error) {
	return uc.repo.GetBanner(ctx, id)
}

func (uc *UseCase) CreateBanner(ctx context.Context, b *dto.PortalBannerDTO) (*dto.PortalBannerDTO, error) {
	return uc.repo.CreateBanner(ctx, b)
}

func (uc *UseCase) UpdateBanner(ctx context.Context, b *dto.PortalBannerDTO) (*dto.PortalBannerDTO, error) {
	return uc.repo.UpdateBanner(ctx, b)
}

func (uc *UseCase) DeleteBanner(ctx context.Context, id string) error {
	return uc.repo.DeleteBanner(ctx, id)
}

func (uc *UseCase) ToggleBanner(ctx context.Context, id string) (*dto.PortalBannerDTO, error) {
	return uc.repo.ToggleBanner(ctx, id)
}

func (uc *UseCase) ListActiveBanners(ctx context.Context) ([]*dto.PortalBannerDTO, error) {
	return uc.repo.ListActiveBanners(ctx)
}

func (uc *UseCase) ListCustomPages(ctx context.Context, page, pageSize int) ([]*dto.PortalCustomPageDTO, int, error) {
	return uc.repo.ListCustomPages(ctx, page, pageSize)
}

func (uc *UseCase) CreateCustomPage(ctx context.Context, p *dto.PortalCustomPageDTO) (*dto.PortalCustomPageDTO, error) {
	return uc.repo.CreateCustomPage(ctx, p)
}

func (uc *UseCase) GetCustomPageByID(ctx context.Context, id string) (*dto.PortalCustomPageDTO, error) {
	return uc.repo.GetCustomPageByID(ctx, id)
}

func (uc *UseCase) GetCustomPageBySlug(ctx context.Context, slug string) (*dto.PortalCustomPageDTO, error) {
	return uc.repo.GetCustomPageBySlug(ctx, slug)
}

func (uc *UseCase) UpdateCustomPage(ctx context.Context, p *dto.PortalCustomPageDTO) (*dto.PortalCustomPageDTO, error) {
	return uc.repo.UpdateCustomPage(ctx, p)
}

func (uc *UseCase) DeleteCustomPage(ctx context.Context, id string) error {
	return uc.repo.DeleteCustomPage(ctx, id)
}

func (uc *UseCase) IncrementPageViewCount(ctx context.Context, id string) error {
	return uc.repo.IncrementPageViewCount(ctx, id)
}

func (uc *UseCase) ListPublishedCustomPages(ctx context.Context) ([]*dto.PortalCustomPageDTO, error) {
	return uc.repo.ListPublishedCustomPages(ctx)
}
