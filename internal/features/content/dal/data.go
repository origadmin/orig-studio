/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package dal

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewData,
	NewCommentRepo,
	NewNotificationRepo,
	NewLikeRepo,
	NewFavoriteRepo,
	NewCategoryRepo,
	NewTagRepo,
	NewPlaylistRepo,
	NewChannelRepo,
	NewSystemConfigRepo,
	NewChannelUserRepo,
	NewFeedRepo,
	NewArticleRepo,
	NewCommentModerationRepo,
	NewCommentReportRepo,
	NewCommentLikeRepo,
	NewHistoryRepo,
	NewMediaReportRepo,
	NewMediaReportModerationRepo,
	NewPortalRepo,
	NewAdRepo,
	NewSubtitleRepo,
	NewCommentQueryService,
	NewExploreQueryService,
)

var MicroserviceProviderSet = wire.NewSet(
	NewEntClient,
	NewData,
	NewCommentRepo,
	NewNotificationRepo,
	NewLikeRepo,
	NewFavoriteRepo,
	NewCategoryRepo,
	NewTagRepo,
	NewPlaylistRepo,
	NewChannelRepo,
	NewSystemConfigRepo,
	NewChannelUserRepo,
	NewFeedRepo,
	NewArticleRepo,
	NewCommentModerationRepo,
	NewCommentReportRepo,
	NewCommentLikeRepo,
	NewHistoryRepo,
	NewMediaReportRepo,
	NewMediaReportModerationRepo,
	NewPortalRepo,
	NewCommentQueryService,
)

var RepoProviderSet = wire.NewSet(
	NewData,
	NewCommentRepo,
	NewNotificationRepo,
	NewLikeRepo,
	NewFavoriteRepo,
	NewCategoryRepo,
	NewTagRepo,
	NewPlaylistRepo,
	NewChannelRepo,
	NewSystemConfigRepo,
	NewChannelUserRepo,
	NewFeedRepo,
	NewArticleRepo,
	NewCommentModerationRepo,
	NewCommentReportRepo,
	NewCommentLikeRepo,
	NewHistoryRepo,
	NewMediaReportRepo,
	NewMediaReportModerationRepo,
)
