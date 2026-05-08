/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package biz

import "github.com/google/wire"

// ProviderSet is biz providers.
// Note: NewUploadUseCase, NewSpriteUseCase, and NewTranscodeHandler require
// *conf.StoragePaths and config values, so they are provided via bridge
// functions in wire.go.
var ProviderSet = wire.NewSet(
	NewMediaUseCase,
)
