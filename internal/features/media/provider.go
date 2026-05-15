/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package media

import (
	"github.com/google/wire"
	"origadmin/application/origstudio/internal/features/media/biz"
	"origadmin/application/origstudio/internal/features/media/dal"
	"origadmin/application/origstudio/internal/features/media/service"
)

// ProviderSet is media providers.
var ProviderSet = wire.NewSet(
	dal.ProviderSet,
	biz.ProviderSet,
	service.ProviderSet,
)
