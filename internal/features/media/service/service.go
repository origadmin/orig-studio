/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package service

import "github.com/google/wire"

// ProviderSet is service providers for monolith mode.
var ProviderSet = wire.NewSet(
	NewMediaService,
	NewUploadService,
	NewMediaHandler,
	NewUploadHandler,
	NewSearchHandler,
)

// MicroserviceProviderSet is service providers for standalone microservice mode.
// Excludes Gin-based handlers (MediaHandler, UploadHandler) that depend on
// cross-feature use cases only available in monolith mode.
var MicroserviceProviderSet = wire.NewSet(
	NewMediaService,
	NewUploadService,
)
