/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package gateway provides the API gateway internal components.
//
// Directory layout:
//
//	gateway/
//	├── client/      downstream gRPC service clients (svc-user, svc-media, svc-content)
//	├── handler/     aggregation handlers (feed / detail / search / profile)
//	│               migrated from the former svc-portal service
//	├── middleware/  HTTP middleware (logger / CORS / JWT auth)
//	└── router/     HTTP route registration
package gateway

import (
	"github.com/google/wire"
	"origadmin/application/origcms/internal/gateway/client"
	"origadmin/application/origcms/internal/gateway/handler"
	"origadmin/application/origcms/internal/gateway/router"
)

// ProviderSet aggregates all gateway providers for Wire injection.
var ProviderSet = wire.NewSet(
	client.ProviderSet,
	handler.ProviderSet,
	router.NewRouter,
)
