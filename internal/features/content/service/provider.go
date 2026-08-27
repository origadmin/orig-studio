/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package service

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewContentService,
	NewCategoryHandler,
	NewTagHandler,
	NewArticleHandler,
	NewCommentHandler,
	NewCommentModerationHandler,
	NewFeedHandler,
	NewChannelHandler,
	NewPlaylistHandler,
	NewPlaylistServiceServer,
	NewInteractionHandler,
	NewNotificationHandler,
	NewShareHandler,
	NewExploreHandler,
	NewPortalHandler,
	NewAdHandler,
)

var MicroserviceProviderSet = wire.NewSet(
	NewContentService,
	NewChannelServiceServer,
	NewPlaylistServiceServer,
	NewCategoryServiceServer,
	NewTagServiceServer,
	NewCommentHandler,
	NewCommentModerationHandler,
	NewInteractionHandler,
	NewNotificationHandler,
)
