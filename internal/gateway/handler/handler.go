/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package handler provides gateway aggregation handlers.
//
// These handlers replace the former svc-portal service and are responsible for
// aggregating data from multiple downstream services (svc-user, svc-media, svc-content).
//
// Migration note:
//   - GetHomeFeed  → FeedHandler.GetHomeFeed   (feed.go)
//   - GetVideoDetail → DetailHandler.GetVideoDetail (detail.go)
//   - Search       → SearchHandler.Search      (search.go)
//   - GetUserProfile → ProfileHandler.GetUserProfile (profile.go)
package handler

import "github.com/google/wire"

// ProviderSet is handler providers for the gateway.
var ProviderSet = wire.NewSet(
	NewFeedHandler,
	NewDetailHandler,
	NewSearchHandler,
	NewProfileHandler,
)
