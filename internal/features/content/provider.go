/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package content

import (
	"github.com/google/wire"
	"origadmin/application/origstudio/internal/features/content/biz"
	"origadmin/application/origstudio/internal/features/content/dal"
	"origadmin/application/origstudio/internal/features/content/service"
)

// ProviderSet is the wire provider set for the content feature module.
var ProviderSet = wire.NewSet(
	dal.ProviderSet,
	biz.ProviderSet,
	service.ProviderSet,
)
