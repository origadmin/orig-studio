/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package user

import (
	"github.com/google/wire"
	"origadmin/application/origstudio/internal/features/user/biz"
	"origadmin/application/origstudio/internal/features/user/dal"
	"origadmin/application/origstudio/internal/features/user/service"
)

var ProviderSet = wire.NewSet(
	dal.ProviderSet,
	biz.ProviderSet,
	service.ProviderSet,
)
