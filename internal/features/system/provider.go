/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package system

import (
	"github.com/google/wire"
	"origadmin/application/origcms/internal/features/system/biz"
	"origadmin/application/origcms/internal/features/system/dal"
	"origadmin/application/origcms/internal/features/system/service"
)

// ProviderSet is the wire provider set for the system feature module.
var ProviderSet = wire.NewSet(
	dal.ProviderSet,
	biz.ProviderSet,
	service.ProviderSet,
)
