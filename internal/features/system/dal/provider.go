/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package dal

import "github.com/google/wire"

// ProviderSet is the wire provider set for the system data layer.
var ProviderSet = wire.NewSet(
	NewSettingRepo,
	NewStatsRepo,
)
