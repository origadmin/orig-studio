package server

import (
	"errors"
	stdhttp "net/http"

	"github.com/google/wire"

	grpcv1 "github.com/origadmin/runtime/api/gen/go/config/transport/grpc/v1"
	httpv1 "github.com/origadmin/runtime/api/gen/go/config/transport/http/v1"
	transportv1 "github.com/origadmin/runtime/api/gen/go/config/transport/v1"
	"github.com/origadmin/runtime/log"
	"github.com/origadmin/runtime/service/transport"
	"github.com/origadmin/runtime/service/transport/grpc"
	runtimehttp "github.com/origadmin/runtime/service/transport/http"
	transhttp "github.com/go-kratos/kratos/v2/transport/http"

	contentv1 "origadmin/application/origstudio/api/gen/v1/content"
	mediav1 "origadmin/application/origstudio/api/gen/v1/media"
	"origadmin/application/origstudio/internal/features/content/service"
	kratosadapter "origadmin/application/origstudio/internal/pkg/http/kratos"
	"origadmin/application/origstudio/internal/pkg/http/validate"
	"origadmin/application/origstudio/internal/server"
)

// ServerProviderSet provides the EE content microservice transport assembly.
// This was migrated from internal/features/content/server/server.go to comply
// with B2 layering (enterprise → features, not features → enterprise).
var ServerProviderSet = wire.NewSet(NewServers)

func NewServers(
	cfg *transportv1.Servers,
	svc *service.ContentService,
	channelSvc *service.ChannelServiceServer,
	playlistSvc *service.PlaylistServiceServer,
	categorySvc *service.CategoryServiceServer,
	tagSvc *service.TagServiceServer,
	commentHandler *service.CommentHandler,
	commentModerationHandler *service.CommentModerationHandler,
	notificationHandler *service.NotificationHandler,
	interactionHandler *service.InteractionHandler,
	subtitleHandler *service.SubtitleHandler,
	enterpriseContent *EnterpriseContentServer,
) ([]transport.Server, error) {
	if cfg == nil {
		return nil, errors.New("servers config is nil")
	}

	var servers []transport.Server
	for _, serverCfg := range cfg.GetConfigs() {
		if serverCfg.GetName() != "content" && serverCfg.GetName() != "origcms.content" {
			continue
		}
		switch serverCfg.GetProtocol() {
		case "grpc":
			srv, err := NewGRPCServer(serverCfg.GetGrpc(), svc, channelSvc, playlistSvc, categorySvc, tagSvc)
			if err != nil {
				return nil, err
			}
			servers = append(servers, srv)
		case "http":
			srv, err := NewHTTPServer(serverCfg.GetHttp(), svc, channelSvc, playlistSvc, categorySvc, tagSvc, commentHandler, commentModerationHandler, notificationHandler, interactionHandler, subtitleHandler, enterpriseContent)
			if err != nil {
				return nil, err
			}
			servers = append(servers, srv)
		default:
			log.Warnf("protocol '%s' not supported by content service, skipping", serverCfg.GetProtocol())
		}
	}

	if len(servers) == 0 {
		return nil, errors.New("no servers named 'content' were created")
	}
	return servers, nil
}

func NewGRPCServer(cfg *grpcv1.Server, svc *service.ContentService, channelSvc *service.ChannelServiceServer, playlistSvc *service.PlaylistServiceServer, categorySvc *service.CategoryServiceServer, tagSvc *service.TagServiceServer) (*transport.GRPCServer, error) {
	if cfg == nil {
		return nil, errors.New("grpc config is nil")
	}

	opts := &grpc.ServerOptions{
		ServerOptions: validate.GRPCServerOptions(),
	}
	srv, err := grpc.NewServer(cfg, opts)
	if err != nil {
		return nil, err
	}

	contentv1.RegisterContentServiceServer(srv, svc)
	mediav1.RegisterChannelServiceServer(srv, channelSvc)
	mediav1.RegisterPlaylistServiceServer(srv, playlistSvc)
	mediav1.RegisterCategoryServiceServer(srv, categorySvc)
	mediav1.RegisterTagServiceServer(srv, tagSvc)
	return srv, nil
}

// NewHTTPServer creates the HTTP server for the content microservice.
// Security: gRPC-Gateway routes are protected via Kratos Filter bridge.
// Enterprise and CE handler routes are registered via the kratos adapter
// (no Gin dependency).
func NewHTTPServer(cfg *httpv1.Server, svc *service.ContentService, channelSvc *service.ChannelServiceServer, playlistSvc *service.PlaylistServiceServer, categorySvc *service.CategoryServiceServer, tagSvc *service.TagServiceServer, commentHandler *service.CommentHandler, commentModerationHandler *service.CommentModerationHandler, notificationHandler *service.NotificationHandler, interactionHandler *service.InteractionHandler, subtitleHandler *service.SubtitleHandler, enterpriseContent *EnterpriseContentServer) (*transport.HTTPServer, error) {
	if cfg == nil {
		return nil, errors.New("http config is nil")
	}

	// Build Kratos server options: global base middleware via Filter bridge.
	serverOpts := &runtimehttp.ServerOptions{
		ServerOptions: []transhttp.ServerOption{
			transhttp.Filter(
				kratosadapter.MiddlewareToFilter(server.RecoveryCtx()),
				kratosadapter.MiddlewareToFilter(server.RequestIDCtx()),
			),
		},
	}

	srv, err := runtimehttp.NewServer(cfg, serverOpts)
	if err != nil {
		return nil, err
	}

	srv.Handle("/healthz", stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		w.WriteHeader(stdhttp.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","service":"content"}`))
	}))

	// Register Enterprise routes via kratos adapter FIRST (framework-agnostic).
	// Comment Handler routes MUST be registered before gRPC-Gateway so that
	// POST /api/v1/comments with nested {comment:{content,media_id}} can hit
	// the correct HTTP handler (which supports UUID media_id strings), instead
	// of falling through to the gRPC-gateway CreateComment which binds to a
	// flat {text, media_id:int64} protobuf schema and rejects UUID content IDs.
	router := kratosadapter.NewRouterAdapter(srv, "/api/v1")
	enterpriseContent.RegisterRoutes(router)
	if commentHandler != nil {
		commentHandler.RegisterRoutes(router)
	}
	if commentModerationHandler != nil {
		// BUG-243: admin/comments — CommentModerationHandler was built in wire but
		// never injected/registered here → /admin/comments 404.
		commentModerationHandler.RegisterRoutes(router)
	}
	if notificationHandler != nil {
		notificationHandler.RegisterRoutes(router)
	}
	if interactionHandler != nil {
		interactionHandler.RegisterRoutes(router)
	}
	// BUG-186: subtitle endpoints (list/create/delete/languages) — owner/admin manage.
	if subtitleHandler != nil {
		subtitleHandler.RegisterRoutes(router)
	}

	// Register gRPC-Gateway HTTP servers AFTER the HTTP-native handlers so
	// duplicate routes don't shadow the UUID-aware comment endpoints.
	contentv1.RegisterContentServiceHTTPServer(srv, svc)
	mediav1.RegisterChannelServiceHTTPServer(srv, channelSvc)
	mediav1.RegisterPlaylistServiceHTTPServer(srv, playlistSvc)
	mediav1.RegisterCategoryServiceHTTPServer(srv, categorySvc)
	mediav1.RegisterTagServiceHTTPServer(srv, tagSvc)

	return srv, nil
}
