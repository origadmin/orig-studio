//go:build wireinject
// +build wireinject

/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/google/wire"
	"github.com/origadmin/runtime"
	"origadmin/application/origcms/internal/conf"
	"origadmin/application/origcms/internal/helpers/providers"
	"origadmin/application/origcms/internal/svc-media/biz"
	"origadmin/application/origcms/internal/svc-media/data"
	"origadmin/application/origcms/internal/svc-media/server"
	"origadmin/application/origcms/internal/svc-media/service"
)

// wireApp init kratos application.
func wireApp(app *runtime.App, _ *conf.Config) (*kratos.App, func(), error) {
	panic(wire.Build(
		providers.CommonSet,
		data.ProviderSet,
		biz.ProviderSet,
		service.ProviderSet,
		server.ProviderSet,
		NewApp,
	))
}
