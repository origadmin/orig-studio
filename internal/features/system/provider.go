/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package system

import (
	"github.com/google/wire"
	"origadmin/application/origstudio/internal/features/system/biz"
	"origadmin/application/origstudio/internal/features/system/dal"
	"origadmin/application/origstudio/internal/features/system/service"
)

// ProviderSet is the wire provider set for the system feature module.
var ProviderSet = wire.NewSet(
	dal.ProviderSet,
	biz.ProviderSet,
	service.ProviderSet,
)
