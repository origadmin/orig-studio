/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package dal

import "github.com/google/wire"

// ProviderSet is data providers.
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
)
