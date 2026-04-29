/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package biz

import "github.com/google/wire"

// ProviderSet is the wire provider set for the admin biz layer.
var ProviderSet = wire.NewSet(
	NewTagUseCase,
)
