/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package svcmedia

import (
	"github.com/google/wire"
	"origadmin/application/origcms/internal/svc-media/biz"
	"origadmin/application/origcms/internal/svc-media/data"
	"origadmin/application/origcms/internal/svc-media/service"
)

// ProviderSet is svc-media providers.
var ProviderSet = wire.NewSet(
	data.ProviderSet,
	biz.ProviderSet,
	service.ProviderSet,
)
