/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package user

import (
	"github.com/google/wire"
	"origadmin/application/origcms/internal/features/user/biz"
	"origadmin/application/origcms/internal/features/user/dal"
	"origadmin/application/origcms/internal/features/user/service"
)

// ProviderSet is the wire provider set for the user feature module.
var ProviderSet = wire.NewSet(
	dal.ProviderSet,
	biz.ProviderSet,
	service.ProviderSet,
)
