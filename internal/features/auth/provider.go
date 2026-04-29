/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package auth

import (
	"github.com/google/wire"
	"origadmin/application/origcms/internal/features/auth/biz"
	"origadmin/application/origcms/internal/features/auth/dal"
	"origadmin/application/origcms/internal/features/auth/service"
)

// ProviderSet is the wire provider set for the auth feature module.
var ProviderSet = wire.NewSet(
	dal.ProviderSet,
	biz.ProviderSet,
	service.ProviderSet,
)
