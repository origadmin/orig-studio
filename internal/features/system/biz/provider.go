/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package biz

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewSettingUseCase,
	NewEmailUseCase,
	NewFeatureFlagUseCase,
)
