/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package server

import (
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/google/wire"
	"github.com/origadmin/runtime"
	"origadmin/application/origcms/api/gen/v1/media"
	"origadmin/application/origcms/internal/svc-media/service"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(NewHTTPServer, NewGRPCServer)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(app *runtime.App, mediaSvc *service.MediaService) *http.Server {
	srv := app.NewHTTPServer()
	media.RegisterMediaServiceHTTPServer(srv, mediaSvc)
	return srv
}

// NewGRPCServer new a gRPC server.
func NewGRPCServer(app *runtime.App, mediaSvc *service.MediaService) *grpc.Server {
	srv := app.NewGRPCServer()
	media.RegisterMediaServiceServer(srv, mediaSvc)
	return srv
}
