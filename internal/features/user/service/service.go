/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package service is the service layer for the user service.
package service

import (
	"github.com/google/wire"
)

// ProviderSet is the wire provider set for the service layer (monolith mode).
var ProviderSet = wire.NewSet(
	NewUserService,
	NewMediaServiceClient,
	NewAuthHandler,
	NewUserHandler,
	NewMeHandler,
	NewAdminUserService,
	NewPermissionService,
	NewSystemConfigService,
)

// MicroserviceProviderSet is the wire provider set for standalone microservice mode.
var MicroserviceProviderSet = wire.NewSet(
	NewUserService,
	NewMediaServiceClient,
	NewAuthHandler,
	NewUserHandler,
	NewMeHandler,
	NewAdminUserService,
	NewPermissionService,
	NewSystemConfigService,
)
