/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package server provides gRPC and HTTP server setup for svc-user.
package server

import (
	"errors"
	stdhttp "net/http"

	"github.com/google/wire"

	"github.com/origadmin/runtime"
	grpcv1 "github.com/origadmin/runtime/api/gen/go/config/transport/grpc/v1"
	httpv1 "github.com/origadmin/runtime/api/gen/go/config/transport/http/v1"
	transportv1 "github.com/origadmin/runtime/api/gen/go/config/transport/v1"
	"github.com/origadmin/runtime/log"
	"github.com/origadmin/runtime/middleware"
	"github.com/origadmin/runtime/service/transport"
	"github.com/origadmin/runtime/service/transport/grpc"
	"github.com/origadmin/runtime/service/transport/http"

	userv1 "origadmin/application/origcms/api/v1/services/user"
	"origadmin/application/origcms/internal/svc-user/service"
)

// ProviderSet is the wire provider set for the server package.
var ProviderSet = wire.NewSet(NewServers)

// NewServers creates gRPC and HTTP servers for svc-user.
func NewServers(
	app *runtime.App,
	cfg *transportv1.Servers,
	svc *service.UserService,
) ([]transport.Server, error) {
	if cfg == nil {
		return nil, errors.New("servers config is nil")
	}

	var servers []transport.Server
	for _, serverCfg := range cfg.GetConfigs() {
		if serverCfg.GetName() != "user" && serverCfg.GetName() != "origcms.svc-user" {
			continue
		}
		switch serverCfg.GetProtocol() {
		case "grpc":
			srv, err := NewGRPCServer(app, serverCfg.GetGrpc(), svc)
			if err != nil {
				return nil, err
			}
			servers = append(servers, srv)
		case "http":
			srv, err := NewHTTPServer(app, serverCfg.GetHttp(), svc)
			if err != nil {
				return nil, err
			}
			servers = append(servers, srv)
		default:
			log.Warnf("protocol '%s' not supported by svc-user, skipping", serverCfg.GetProtocol())
		}
	}

	if len(servers) == 0 {
		return nil, errors.New("no servers named 'user' were created")
	}
	return servers, nil
}

// NewGRPCServer creates a new gRPC server for svc-user.
func NewGRPCServer(app *runtime.App, cfg *grpcv1.Server, svc *service.UserService) (*transport.GRPCServer, error) {
	if cfg == nil {
		return nil, errors.New("grpc config is nil")
	}

	h := app.Container().In(runtime.CategoryMiddleware,
		runtime.WithInScope(runtime.ServerScope),
	)
	mwMap, err := middleware.GetMiddlewares(app.Context(), h)
	if err != nil {
		return nil, err
	}

	opts := &grpc.ServerOptions{
		ServerMiddlewares: mwMap,
	}
	srv, err := grpc.NewServer(cfg, opts)
	if err != nil {
		return nil, err
	}

	userv1.RegisterUserServiceServer(srv, svc)
	return srv, nil
}

// NewHTTPServer creates a new HTTP server for svc-user.
// NOTE: grpc-gateway style registration requires *runtime.ServeMux, not *kratos/http.Server.
// For M1, we start the HTTP server and register a simple health endpoint only.
// Full HTTP routing will be added in M2 via kratos-compatible proto generation.
func NewHTTPServer(app *runtime.App, cfg *httpv1.Server, svc *service.UserService) (*transport.HTTPServer, error) {
	if cfg == nil {
		return nil, errors.New("http config is nil")
	}

	h := app.Container().In(runtime.CategoryMiddleware,
		runtime.WithInScope(runtime.ServerScope),
	)
	mwMap, err := middleware.GetMiddlewares(app.Context(), h)
	if err != nil {
		return nil, err
	}

	opts := &http.ServerOptions{
		ServerMiddlewares: mwMap,
	}
	srv, err := http.NewServer(cfg, opts)
	if err != nil {
		return nil, err
	}

	// Register a health check endpoint
	srv.Handle("/healthz", stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		w.WriteHeader(stdhttp.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","service":"svc-user"}`))
	}))

	_ = svc // HTTP route registration will be added in M2 when kratos-annotated proto is generated

	helper := log.NewHelper(app.Logger())
	helper.Infow(log.DefaultMessageKey, "HTTP server created", "addr", cfg.GetAddr())

	return srv, nil
}
