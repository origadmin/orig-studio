/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package biz is the biz layer for the user service.
package biz

import (
	"github.com/google/wire"
)

// ProviderSet is the wire provider set for the biz layer.
var ProviderSet = wire.NewSet(
	NewUserUseCase,
)
