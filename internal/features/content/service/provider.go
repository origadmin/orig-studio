/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package service

import "github.com/google/wire"

// ProviderSet is the wire provider set for the content service layer.
var ProviderSet = wire.NewSet(
	NewCategoryHandler,
	NewTagHandler,
	NewArticleHandler,
	NewCommentHandler,
	NewCommentModerationHandler,
	NewFeedHandler,
	NewChannelHandler,
	NewPlaylistHandler,
	NewInteractionHandler,
	NewNotificationHandler,
	NewShareHandler,
	NewExploreHandler,
)
