/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package admin

import (
	"github.com/google/wire"
	"origadmin/application/origstudio/internal/features/admin/biz"
	"origadmin/application/origstudio/internal/features/admin/dal"
	"origadmin/application/origstudio/internal/features/admin/service"
)

// ProviderSet is the wire provider set for the admin feature module.
var ProviderSet = wire.NewSet(
	dal.ProviderSet,
	biz.ProviderSet,
	service.ProviderSet,
)
