package service

import (
	"context"

	"github.com/origadmin/runtime/log"

	media "origadmin/application/origstudio/api/gen/v1/media"
	"origadmin/application/origstudio/api/gen/v1/types"
)

// ========== AdminCommentService ==========

type AdminCommentService struct {
	media.UnimplementedAdminCommentServiceServer
	log *log.Helper
}

func NewAdminCommentService(logger log.Logger) *AdminCommentService {
	return &AdminCommentService{log: log.NewHelper(log.With(logger, "module", "media.service.admin-comment"))}
}

func (s *AdminCommentService) ListAdminComments(ctx context.Context, req *media.ListAdminCommentsRequest) (*media.ListAdminCommentsResponse, error) {
	return &media.ListAdminCommentsResponse{}, nil
}

func (s *AdminCommentService) GetAdminComment(ctx context.Context, req *media.GetAdminCommentRequest) (*media.GetAdminCommentResponse, error) {
	return &media.GetAdminCommentResponse{}, nil
}

func (s *AdminCommentService) ApproveComment(ctx context.Context, req *media.ApproveCommentRequest) (*media.ApproveCommentResponse, error) {
	return &media.ApproveCommentResponse{Success: true}, nil
}

func (s *AdminCommentService) RejectComment(ctx context.Context, req *media.RejectCommentRequest) (*media.RejectCommentResponse, error) {
	return &media.RejectCommentResponse{Success: true}, nil
}

func (s *AdminCommentService) BlockComment(ctx context.Context, req *media.BlockCommentRequest) (*media.BlockCommentResponse, error) {
	return &media.BlockCommentResponse{Success: true}, nil
}

func (s *AdminCommentService) UnblockComment(ctx context.Context, req *media.UnblockCommentRequest) (*media.UnblockCommentResponse, error) {
	return &media.UnblockCommentResponse{Success: true}, nil
}

func (s *AdminCommentService) DismissCommentReports(ctx context.Context, req *media.DismissCommentReportsRequest) (*media.DismissCommentReportsResponse, error) {
	return &media.DismissCommentReportsResponse{Success: true}, nil
}

func (s *AdminCommentService) GetCommentReports(ctx context.Context, req *media.GetCommentReportsRequest) (*media.GetCommentReportsResponse, error) {
	return &media.GetCommentReportsResponse{}, nil
}

func (s *AdminCommentService) GetCommentStats(ctx context.Context, req *media.GetCommentStatsRequest) (*media.GetCommentStatsResponse, error) {
	return &media.GetCommentStatsResponse{}, nil
}

func (s *AdminCommentService) BatchApproveComments(ctx context.Context, req *media.BatchApproveCommentsRequest) (*media.BatchApproveCommentsResponse, error) {
	return &media.BatchApproveCommentsResponse{}, nil
}

func (s *AdminCommentService) BatchRejectComments(ctx context.Context, req *media.BatchRejectCommentsRequest) (*media.BatchRejectCommentsResponse, error) {
	return &media.BatchRejectCommentsResponse{}, nil
}

// ========== AdminChannelService ==========

type AdminChannelService struct {
	media.UnimplementedAdminChannelServiceServer
	log *log.Helper
}

func NewAdminChannelService(logger log.Logger) *AdminChannelService {
	return &AdminChannelService{log: log.NewHelper(log.With(logger, "module", "media.service.admin-channel"))}
}

func (s *AdminChannelService) ListAdminChannels(ctx context.Context, req *media.ListAdminChannelsRequest) (*media.ListAdminChannelsResponse, error) {
	return &media.ListAdminChannelsResponse{}, nil
}

func (s *AdminChannelService) GetAdminChannel(ctx context.Context, req *media.GetAdminChannelRequest) (*media.GetAdminChannelResponse, error) {
	return &media.GetAdminChannelResponse{}, nil
}

func (s *AdminChannelService) CreateAdminChannel(ctx context.Context, req *media.CreateAdminChannelRequest) (*media.CreateAdminChannelResponse, error) {
	return &media.CreateAdminChannelResponse{}, nil
}

func (s *AdminChannelService) UpdateAdminChannel(ctx context.Context, req *media.UpdateAdminChannelRequest) (*media.UpdateAdminChannelResponse, error) {
	return &media.UpdateAdminChannelResponse{}, nil
}

func (s *AdminChannelService) DeleteAdminChannel(ctx context.Context, req *media.DeleteAdminChannelRequest) (*media.DeleteAdminChannelResponse, error) {
	return &media.DeleteAdminChannelResponse{}, nil
}

// ========== AdminPlaylistService ==========

type AdminPlaylistService struct {
	media.UnimplementedAdminPlaylistServiceServer
	log *log.Helper
}

func NewAdminPlaylistService(logger log.Logger) *AdminPlaylistService {
	return &AdminPlaylistService{log: log.NewHelper(log.With(logger, "module", "media.service.admin-playlist"))}
}

func (s *AdminPlaylistService) ListAdminPlaylists(ctx context.Context, req *media.ListAdminPlaylistsRequest) (*media.ListAdminPlaylistsResponse, error) {
	return &media.ListAdminPlaylistsResponse{}, nil
}

func (s *AdminPlaylistService) GetAdminPlaylist(ctx context.Context, req *media.GetAdminPlaylistRequest) (*media.GetAdminPlaylistResponse, error) {
	return &media.GetAdminPlaylistResponse{}, nil
}

func (s *AdminPlaylistService) CreateAdminPlaylist(ctx context.Context, req *media.CreateAdminPlaylistRequest) (*media.CreateAdminPlaylistResponse, error) {
	return &media.CreateAdminPlaylistResponse{}, nil
}

func (s *AdminPlaylistService) UpdateAdminPlaylist(ctx context.Context, req *media.UpdateAdminPlaylistRequest) (*media.UpdateAdminPlaylistResponse, error) {
	return &media.UpdateAdminPlaylistResponse{}, nil
}

func (s *AdminPlaylistService) DeleteAdminPlaylist(ctx context.Context, req *media.DeleteAdminPlaylistRequest) (*media.DeleteAdminPlaylistResponse, error) {
	return &media.DeleteAdminPlaylistResponse{}, nil
}

// ========== AdminUserService ==========

type AdminUserService struct {
	media.UnimplementedAdminUserServiceServer
	log *log.Helper
}

func NewAdminUserService(logger log.Logger) *AdminUserService {
	return &AdminUserService{log: log.NewHelper(log.With(logger, "module", "media.service.admin-user"))}
}

func (s *AdminUserService) ListAdminUsers(ctx context.Context, req *media.ListAdminUsersRequest) (*media.ListAdminUsersResponse, error) {
	return &media.ListAdminUsersResponse{}, nil
}

func (s *AdminUserService) GetAdminUser(ctx context.Context, req *media.GetAdminUserRequest) (*media.GetAdminUserResponse, error) {
	return &media.GetAdminUserResponse{}, nil
}

func (s *AdminUserService) CreateAdminUser(ctx context.Context, req *media.CreateAdminUserRequest) (*media.CreateAdminUserResponse, error) {
	return &media.CreateAdminUserResponse{}, nil
}

func (s *AdminUserService) UpdateAdminUser(ctx context.Context, req *media.UpdateAdminUserRequest) (*media.UpdateAdminUserResponse, error) {
	return &media.UpdateAdminUserResponse{}, nil
}

func (s *AdminUserService) DeleteAdminUser(ctx context.Context, req *media.DeleteAdminUserRequest) (*media.DeleteAdminUserResponse, error) {
	return &media.DeleteAdminUserResponse{}, nil
}

func (s *AdminUserService) UpdateAdminUserStatus(ctx context.Context, req *media.UpdateAdminUserStatusRequest) (*media.UpdateAdminUserStatusResponse, error) {
	return &media.UpdateAdminUserStatusResponse{Success: true}, nil
}

func (s *AdminUserService) UpdateAdminUserRole(ctx context.Context, req *media.UpdateAdminUserRoleRequest) (*media.UpdateAdminUserRoleResponse, error) {
	return &media.UpdateAdminUserRoleResponse{Success: true}, nil
}

func (s *AdminUserService) GetAdminUserPermissions(ctx context.Context, req *media.GetAdminUserPermissionsRequest) (*media.GetAdminUserPermissionsResponse, error) {
	return &media.GetAdminUserPermissionsResponse{}, nil
}

// ========== PermissionService ==========

type PermissionService struct {
	media.UnimplementedPermissionServiceServer
	log *log.Helper
}

func NewPermissionService(logger log.Logger) *PermissionService {
	return &PermissionService{log: log.NewHelper(log.With(logger, "module", "media.service.permission"))}
}

func (s *PermissionService) ListPermissions(ctx context.Context, req *media.ListPermissionsRequest) (*media.ListPermissionsResponse, error) {
	return &media.ListPermissionsResponse{
		Permissions: []string{},
	}, nil
}

func (s *PermissionService) ListPermissionGroups(ctx context.Context, req *media.ListPermissionGroupsRequest) (*media.ListPermissionGroupsResponse, error) {
	return &media.ListPermissionGroupsResponse{}, nil
}

func (s *PermissionService) GetPermissionGroup(ctx context.Context, req *media.GetPermissionGroupRequest) (*media.GetPermissionGroupResponse, error) {
	return &media.GetPermissionGroupResponse{}, nil
}

func (s *PermissionService) CreatePermissionGroup(ctx context.Context, req *media.CreatePermissionGroupRequest) (*media.CreatePermissionGroupResponse, error) {
	return &media.CreatePermissionGroupResponse{}, nil
}

func (s *PermissionService) UpdatePermissionGroup(ctx context.Context, req *media.UpdatePermissionGroupRequest) (*media.UpdatePermissionGroupResponse, error) {
	return &media.UpdatePermissionGroupResponse{}, nil
}

func (s *PermissionService) DeletePermissionGroup(ctx context.Context, req *media.DeletePermissionGroupRequest) (*media.DeletePermissionGroupResponse, error) {
	return &media.DeletePermissionGroupResponse{}, nil
}

func (s *PermissionService) TogglePermissionGroup(ctx context.Context, req *media.TogglePermissionGroupRequest) (*media.TogglePermissionGroupResponse, error) {
	return &media.TogglePermissionGroupResponse{Success: true}, nil
}

func (s *PermissionService) ListPermissionGroupMembers(ctx context.Context, req *media.ListPermissionGroupMembersRequest) (*media.ListPermissionGroupMembersResponse, error) {
	return &media.ListPermissionGroupMembersResponse{}, nil
}

func (s *PermissionService) AddPermissionGroupMember(ctx context.Context, req *media.AddPermissionGroupMemberRequest) (*media.AddPermissionGroupMemberResponse, error) {
	return &media.AddPermissionGroupMemberResponse{Success: true}, nil
}

func (s *PermissionService) RemovePermissionGroupMember(ctx context.Context, req *media.RemovePermissionGroupMemberRequest) (*media.RemovePermissionGroupMemberResponse, error) {
	return &media.RemovePermissionGroupMemberResponse{Success: true}, nil
}

// ========== PortalManagementService ==========

type PortalManagementService struct {
	media.UnimplementedPortalManagementServiceServer
	log *log.Helper
}

func NewPortalManagementService(logger log.Logger) *PortalManagementService {
	return &PortalManagementService{log: log.NewHelper(log.With(logger, "module", "media.service.portal-management"))}
}

func (s *PortalManagementService) GetPortalConfig(ctx context.Context, req *media.GetPortalConfigRequest) (*media.GetPortalConfigResponse, error) {
	return &media.GetPortalConfigResponse{
		Modules: &media.PortalModules{
			Articles: true,
			Videos:   true,
			Music:    false,
		},
		Layout: "video",
		Site: &media.PortalSite{
			SiteName:          "OrigStudio",
			SiteDescription:   "",
			AllowRegistration: true,
			AllowUpload:       true,
		},
	}, nil
}

func (s *PortalManagementService) GetCustomPage(ctx context.Context, req *media.GetCustomPageRequest) (*media.GetCustomPageResponse, error) {
	return &media.GetCustomPageResponse{}, nil
}

func (s *PortalManagementService) ListNavItems(ctx context.Context, req *media.ListNavItemsRequest) (*media.ListNavItemsResponse, error) {
	return &media.ListNavItemsResponse{}, nil
}

func (s *PortalManagementService) CreateNavItem(ctx context.Context, req *media.CreateNavItemRequest) (*media.CreateNavItemResponse, error) {
	return &media.CreateNavItemResponse{}, nil
}

func (s *PortalManagementService) UpdateNavItem(ctx context.Context, req *media.UpdateNavItemRequest) (*media.UpdateNavItemResponse, error) {
	return &media.UpdateNavItemResponse{}, nil
}

func (s *PortalManagementService) DeleteNavItem(ctx context.Context, req *media.DeleteNavItemRequest) (*media.DeleteNavItemResponse, error) {
	return &media.DeleteNavItemResponse{}, nil
}

func (s *PortalManagementService) ReorderNavItems(ctx context.Context, req *media.ReorderNavItemsRequest) (*media.ReorderNavItemsResponse, error) {
	return &media.ReorderNavItemsResponse{Success: true}, nil
}

func (s *PortalManagementService) ListBanners(ctx context.Context, req *media.ListBannersRequest) (*media.ListBannersResponse, error) {
	return &media.ListBannersResponse{}, nil
}

func (s *PortalManagementService) CreateBanner(ctx context.Context, req *media.CreateBannerRequest) (*media.CreateBannerResponse, error) {
	return &media.CreateBannerResponse{}, nil
}

func (s *PortalManagementService) UpdateBanner(ctx context.Context, req *media.UpdateBannerRequest) (*media.UpdateBannerResponse, error) {
	return &media.UpdateBannerResponse{}, nil
}

func (s *PortalManagementService) DeleteBanner(ctx context.Context, req *media.DeleteBannerRequest) (*media.DeleteBannerResponse, error) {
	return &media.DeleteBannerResponse{}, nil
}

func (s *PortalManagementService) ToggleBanner(ctx context.Context, req *media.ToggleBannerRequest) (*media.ToggleBannerResponse, error) {
	return &media.ToggleBannerResponse{Success: true}, nil
}

func (s *PortalManagementService) ListPages(ctx context.Context, req *media.ListPagesRequest) (*media.ListPagesResponse, error) {
	return &media.ListPagesResponse{}, nil
}

func (s *PortalManagementService) CreatePage(ctx context.Context, req *media.CreatePageRequest) (*media.CreatePageResponse, error) {
	return &media.CreatePageResponse{}, nil
}

func (s *PortalManagementService) UpdatePage(ctx context.Context, req *media.UpdatePageRequest) (*media.UpdatePageResponse, error) {
	return &media.UpdatePageResponse{}, nil
}

func (s *PortalManagementService) DeletePage(ctx context.Context, req *media.DeletePageRequest) (*media.DeletePageResponse, error) {
	return &media.DeletePageResponse{}, nil
}

// ========== ArticleService ==========

type ArticleService struct {
	media.UnimplementedArticleServiceServer
	log *log.Helper
}

func NewArticleService(logger log.Logger) *ArticleService {
	return &ArticleService{log: log.NewHelper(log.With(logger, "module", "media.service.article"))}
}

func (s *ArticleService) ListArticles(ctx context.Context, req *media.ListArticlesRequest) (*media.ListArticlesResponse, error) {
	return &media.ListArticlesResponse{}, nil
}

func (s *ArticleService) GetArticle(ctx context.Context, req *media.GetArticleRequest) (*media.GetArticleResponse, error) {
	return &media.GetArticleResponse{}, nil
}

func (s *ArticleService) GetFeaturedArticles(ctx context.Context, req *media.GetFeaturedArticlesRequest) (*media.GetFeaturedArticlesResponse, error) {
	return &media.GetFeaturedArticlesResponse{}, nil
}

func (s *ArticleService) GetLatestArticles(ctx context.Context, req *media.GetLatestArticlesRequest) (*media.GetLatestArticlesResponse, error) {
	return &media.GetLatestArticlesResponse{}, nil
}

func (s *ArticleService) GetMyArticles(ctx context.Context, req *media.GetMyArticlesRequest) (*media.GetMyArticlesResponse, error) {
	return &media.GetMyArticlesResponse{}, nil
}

func (s *ArticleService) CreateArticle(ctx context.Context, req *media.CreateArticleRequest) (*media.CreateArticleResponse, error) {
	return &media.CreateArticleResponse{}, nil
}

func (s *ArticleService) UpdateArticle(ctx context.Context, req *media.UpdateArticleRequest) (*media.UpdateArticleResponse, error) {
	return &media.UpdateArticleResponse{}, nil
}

func (s *ArticleService) DeleteArticle(ctx context.Context, req *media.DeleteArticleRequest) (*media.DeleteArticleResponse, error) {
	return &media.DeleteArticleResponse{}, nil
}

func (s *ArticleService) UpdateArticleState(ctx context.Context, req *media.UpdateArticleStateRequest) (*media.UpdateArticleStateResponse, error) {
	return &media.UpdateArticleStateResponse{}, nil
}

func (s *ArticleService) ListAdminArticles(ctx context.Context, req *media.ListAdminArticlesRequest) (*media.ListAdminArticlesResponse, error) {
	return &media.ListAdminArticlesResponse{}, nil
}

func (s *ArticleService) GetAdminArticle(ctx context.Context, req *media.GetAdminArticleRequest) (*media.GetAdminArticleResponse, error) {
	return &media.GetAdminArticleResponse{}, nil
}

func (s *ArticleService) CreateAdminArticle(ctx context.Context, req *media.CreateAdminArticleRequest) (*media.CreateAdminArticleResponse, error) {
	return &media.CreateAdminArticleResponse{}, nil
}

func (s *ArticleService) UpdateAdminArticle(ctx context.Context, req *media.UpdateAdminArticleRequest) (*media.UpdateAdminArticleResponse, error) {
	return &media.UpdateAdminArticleResponse{}, nil
}

func (s *ArticleService) DeleteAdminArticle(ctx context.Context, req *media.DeleteAdminArticleRequest) (*media.DeleteAdminArticleResponse, error) {
	return &media.DeleteAdminArticleResponse{}, nil
}

func (s *ArticleService) UpdateAdminArticleState(ctx context.Context, req *media.UpdateAdminArticleStateRequest) (*media.UpdateAdminArticleStateResponse, error) {
	return &media.UpdateAdminArticleStateResponse{}, nil
}

// ========== SystemConfigService ==========

type SystemConfigService struct {
	media.UnimplementedSystemConfigServiceServer
	log *log.Helper
}

func NewSystemConfigService(logger log.Logger) *SystemConfigService {
	return &SystemConfigService{log: log.NewHelper(log.With(logger, "module", "media.service.system-config"))}
}

func (s *SystemConfigService) GetSettingsByCategory(ctx context.Context, req *media.GetSettingsByCategoryRequest) (*media.GetSettingsByCategoryResponse, error) {
	return &media.GetSettingsByCategoryResponse{
		Category: req.Category,
		Settings: map[string]string{},
	}, nil
}

func (s *SystemConfigService) GetSettingByKey(ctx context.Context, req *media.GetSettingByKeyRequest) (*media.GetSettingByKeyResponse, error) {
	return &media.GetSettingByKeyResponse{
		Key: req.Key,
	}, nil
}

func (s *SystemConfigService) UpdateSettingByKey(ctx context.Context, req *media.UpdateSettingByKeyRequest) (*media.UpdateSettingByKeyResponse, error) {
	return &media.UpdateSettingByKeyResponse{
		Key:   req.Key,
		Value: req.Value,
	}, nil
}

func (s *SystemConfigService) DeleteSettingByKey(ctx context.Context, req *media.DeleteSettingByKeyRequest) (*media.DeleteSettingByKeyResponse, error) {
	return &media.DeleteSettingByKeyResponse{}, nil
}

func (s *SystemConfigService) GetSystemSetting(ctx context.Context, req *media.GetSystemSettingRequest) (*media.GetSystemSettingResponse, error) {
	return &media.GetSystemSettingResponse{
		Key: req.Key,
	}, nil
}

func (s *SystemConfigService) ResetSystemSetting(ctx context.Context, req *media.ResetSystemSettingRequest) (*media.ResetSystemSettingResponse, error) {
	return &media.ResetSystemSettingResponse{
		Key: req.Key,
	}, nil
}

func (s *SystemConfigService) GetEmailStatus(ctx context.Context, req *media.GetEmailStatusRequest) (*media.GetEmailStatusResponse, error) {
	return &media.GetEmailStatusResponse{
		Configured: false,
	}, nil
}

func (s *SystemConfigService) TestEmail(ctx context.Context, req *media.TestEmailRequest) (*media.TestEmailResponse, error) {
	return &media.TestEmailResponse{
		Success: false,
		Message: "email not configured",
	}, nil
}

func (s *SystemConfigService) GetChannelLimits(ctx context.Context, req *media.GetChannelLimitsRequest) (*media.GetChannelLimitsResponse, error) {
	// Return sensible defaults: unlimited channels (-1), 0 current, can create
	return &media.GetChannelLimitsResponse{
		Limits: &types.ChannelLimits{
			MaxChannels:  -1, // -1 = unlimited
			CurrentCount: 0,
			CanCreate:    true,
		},
	}, nil
}
