/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package admin

import (
	"github.com/google/wire"
	"origadmin/application/origcms/internal/features/admin/biz"
	"origadmin/application/origcms/internal/features/admin/dal"
	"origadmin/application/origcms/internal/features/admin/service"
)

// ProviderSet is the wire provider set for the admin feature module.
var ProviderSet = wire.NewSet(
	dal.ProviderSet,
	biz.ProviderSet,
	service.ProviderSet,
)
