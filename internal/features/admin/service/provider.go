/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package service

import "github.com/google/wire"

// ProviderSet is the wire provider set for the admin service layer.
var ProviderSet = wire.NewSet(
	NewTagService,
	NewAdminTagHandler,
)
