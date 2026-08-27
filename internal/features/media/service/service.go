package service

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewMediaService,
	NewUploadService,
	NewUploadServiceV1,
	NewMediaHandler,
	NewUploadHandler,
	NewSearchHandler,
	NewExploreService,
	NewAdminMediaService,
	NewAdminCommentService,
	NewAdminTagService,
	NewAdminCategoryService,
	NewAdminChannelService,
	NewAdminPlaylistService,
	NewAdminUserService,
	NewPermissionService,
	NewPortalManagementService,
	NewArticleService,
	NewSystemConfigService,
	NewAdminService,
	NewStorageProxyService,
)

var MicroserviceProviderSet = wire.NewSet(
	NewMediaService,
	NewUploadService,
	NewUploadServiceV1,
	NewUploadHandler,
	NewMediaHandlerForMicroservice,
	NewExploreService,
	NewAdminMediaService,
	NewAdminCommentService,
	NewAdminTagService,
	NewAdminCategoryService,
	NewAdminChannelService,
	NewAdminPlaylistService,
	NewAdminUserService,
	NewPermissionService,
	NewPortalManagementService,
	NewArticleService,
	NewSystemConfigService,
	NewAdminService,
	NewStorageProxyService,
	NewSpriteHandler,
)
