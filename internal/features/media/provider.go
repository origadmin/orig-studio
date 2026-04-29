/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package media

import (
	"github.com/google/wire"
	"origadmin/application/origcms/internal/features/media/biz"
	"origadmin/application/origcms/internal/features/media/dal"
	"origadmin/application/origcms/internal/features/media/service"
)

// ProviderSet is media providers.
var ProviderSet = wire.NewSet(
	dal.ProviderSet,
	biz.ProviderSet,
	service.ProviderSet,
)
