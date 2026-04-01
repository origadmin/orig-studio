//go:build wireinject
// +build wireinject

/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// The build tag makes sure the stub is not built in the final build.
package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/google/wire"

	"github.com/origadmin/runtime"

	"origadmin/application/origcms/internal/conf"
	"origadmin/application/origcms/internal/helpers/providers"
	"origadmin/application/origcms/internal/svc-user/biz"
	"origadmin/application/origcms/internal/svc-user/data"
	"origadmin/application/origcms/internal/svc-user/server"
	"origadmin/application/origcms/internal/svc-user/service"
)

// wireApp init kratos application.
func wireApp(app *runtime.App, _ *conf.Config) (*kratos.App, func(), error) {
	panic(wire.Build(
		// Common providers (logger, hasher, publisher, servers config)
		providers.CommonSet,

		// Data layer (entity.Client + repo)
		data.ProviderSet,

		// Biz layer
		biz.ProviderSet,

		// Service layer
		service.ProviderSet,

		// Server layer (gRPC + HTTP)
		server.ProviderSet,

		// App constructor
		NewApp,
	))
}
