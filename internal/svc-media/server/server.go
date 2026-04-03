/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package server

import (
	"errors"
	"strings"
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

	media "origadmin/application/origcms/api/gen/v1/media"
	"origadmin/application/origcms/internal/svc-media/service"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(NewServers)

// NewServers creates gRPC and HTTP servers for svc-media.
func NewServers(
	app *runtime.App,
	cfg *transportv1.Servers,
	svc *service.MediaService,
) ([]transport.Server, error) {
	if cfg == nil {
		return nil, errors.New("servers config is nil")
	}

	var servers []transport.Server
	for _, serverCfg := range cfg.GetConfigs() {
		if serverCfg.GetName() != "media" && serverCfg.GetName() != "origcms.svc-media" {
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
			log.Warnf("protocol '%s' not supported by svc-media, skipping", serverCfg.GetProtocol())
		}
	}

	if len(servers) == 0 {
		return nil, errors.New("no servers named 'media' were created")
	}
	return servers, nil
}

// NewGRPCServer new a gRPC server.
func NewGRPCServer(app *runtime.App, cfg *grpcv1.Server, mediaSvc *service.MediaService) (*transport.GRPCServer, error) {
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

	media.RegisterMediaServiceServer(srv, mediaSvc)
	return srv, nil
}

// NewHTTPServer new an HTTP server.
func NewHTTPServer(app *runtime.App, cfg *httpv1.Server, mediaSvc *service.MediaService) (*transport.HTTPServer, error) {
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
		_, _ = w.Write([]byte(`{"status":"ok","service":"svc-media"}`))
	}))

	// Register transcoding status HTTP handler (bypasses gRPC gateway for proper query param support)
	srv.Handle("/api/v1/media/transcoding/status", stdhttp.HandlerFunc(mediaSvc.TranscodingStatusHTTPHandler))

	// Register retry handlers
	srv.Handle("/api/v1/media/retry", stdhttp.HandlerFunc(mediaSvc.RetryTaskHTTPHandler))
	srv.Handle("/api/v1/media/retry-all-failed", stdhttp.HandlerFunc(mediaSvc.RetryAllFailedHTTPHandler))

	// Register media variant summary endpoint (for media management page)
	srv.Handle("/api/v1/media/", stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		pathSuffix := r.URL.Path[len("/api/v1/media/"):]
		if strings.HasPrefix(pathSuffix, "variants") {
			mediaSvc.MediaVariantsHTTPHandler(w, r)
			return
		}
		if strings.HasPrefix(pathSuffix, "encoding/tasks") {
			mediaSvc.EncodingTasksHTTPHandler(w, r)
			return
		}
		stdhttp.NotFound(w, r)
	}))

	// Register SSE endpoint
	srv.Handle("/api/v1/medias/transcoding/events", stdhttp.HandlerFunc(mediaSvc.SSEHandler))

	// Static file serving for media assets
	uploadsDir := "./data/uploads"
	fs := stdhttp.FileServer(stdhttp.Dir(uploadsDir))

	corsStatic := func(h stdhttp.Handler) stdhttp.Handler {
		return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			h.ServeHTTP(w, r)
		})
	}

	srv.Handle("/thumbnails/", corsStatic(stdhttp.StripPrefix("/thumbnails/", fs)))
	srv.Handle("/previews/", corsStatic(stdhttp.StripPrefix("/previews/", fs)))
	srv.Handle("/uploads/", corsStatic(stdhttp.StripPrefix("/uploads/", fs)))
	srv.Handle("/hls/", corsStatic(stdhttp.StripPrefix("/hls/", fs)))

	// media.RegisterMediaServiceHTTPServer(srv, mediaSvc) // Not available without http annotations in proto
	_ = mediaSvc

	helper := log.NewHelper(app.Logger())
	helper.Infow(log.DefaultMessageKey, "HTTP server created", "addr", cfg.GetAddr())

	return srv, nil
}
