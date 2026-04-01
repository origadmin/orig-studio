/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package service is the service layer for the user service.
package service

import (
	"github.com/google/wire"
)

// ProviderSet is the wire provider set for the service layer.
var ProviderSet = wire.NewSet(
	NewUserService,
)
