// Package server provides the EE media microservice transport assembly.
// This was migrated from internal/features/media/server/server.go to comply
// with B2 layering (enterprise → features, not features → enterprise).
// The implementation is Kratos-based (not Gin) and is moved here unchanged
// except for removing the self-referential enterprise import.
package server

import (
	"errors"
	stdhttp "net/http"

	grpcv1 "github.com/origadmin/runtime/api/gen/go/config/transport/grpc/v1"
	httpv1 "github.com/origadmin/runtime/api/gen/go/config/transport/http/v1"
	transportv1 "github.com/origadmin/runtime/api/gen/go/config/transport/v1"
	"github.com/origadmin/runtime/log"
	"github.com/origadmin/runtime/service/transport"
	"github.com/origadmin/runtime/service/transport/grpc"
	runtimehttp "github.com/origadmin/runtime/service/transport/http"

	media "origadmin/application/origstudio/api/gen/v1/media"
	uploadv1 "origadmin/application/origstudio/api/gen/v1/upload"
	"origadmin/application/origstudio/internal/features/media/service"
	"origadmin/application/origstudio/internal/infra/auth"
	"origadmin/application/origstudio/internal/server"
	kratosadapter "origadmin/application/origstudio/internal/pkg/http/kratos"
)

func NewServers(
	cfg *transportv1.Servers,
	jwtMgr *auth.Manager,
	svc *service.MediaService,
	uploadSvcV1 *service.UploadServiceV1,
	uploadHandler *service.UploadHandler,
	mediaHandler *service.MediaHandler,
	storageProxy *service.StorageProxyService,
	spriteHandler *service.SpriteHandler,
	exploreSvc *service.ExploreService,
	adminSvc *service.AdminService,
	adminMediaSvc *service.AdminMediaService,
	adminCommentSvc *service.AdminCommentService,
	adminTagSvc *service.AdminTagService,
	adminCategorySvc *service.AdminCategoryService,
	adminChannelSvc *service.AdminChannelService,
	adminPlaylistSvc *service.AdminPlaylistService,
	adminUserSvc *service.AdminUserService,
	permissionSvc *service.PermissionService,
	portalMgmtSvc *service.PortalManagementService,
	articleSvc *service.ArticleService,
	systemConfigSvc *service.SystemConfigService,
	enterpriseMedia *EnterpriseMediaServer,
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
			srv, err := NewGRPCServer(serverCfg.GetGrpc(), svc, uploadSvcV1, exploreSvc, adminSvc, adminMediaSvc, adminCommentSvc, adminTagSvc, adminCategorySvc, adminChannelSvc, adminPlaylistSvc, adminUserSvc, permissionSvc, portalMgmtSvc, articleSvc, systemConfigSvc)
			if err != nil {
				return nil, err
			}
			servers = append(servers, srv)
		case "http":
			srv, err := NewHTTPServer(serverCfg.GetHttp(), jwtMgr, svc, exploreSvc, adminSvc, adminMediaSvc, adminCommentSvc, adminTagSvc, adminCategorySvc, adminChannelSvc, adminPlaylistSvc, adminUserSvc, permissionSvc, portalMgmtSvc, articleSvc, systemConfigSvc, enterpriseMedia, uploadHandler, mediaHandler, storageProxy, spriteHandler)
			if err != nil {
				return nil, err
			}
			servers = append(servers, srv)
		default:
			log.Warnf("protocol '%s' not supported by media service, skipping", serverCfg.GetProtocol())
		}
	}

	if len(servers) == 0 {
		return nil, errors.New("no servers named 'media' were created")
	}
	return servers, nil
}

func NewGRPCServer(
	cfg *grpcv1.Server,
	mediaSvc *service.MediaService,
	uploadSvcV1 *service.UploadServiceV1,
	exploreSvc *service.ExploreService,
	adminSvc *service.AdminService,
	adminMediaSvc *service.AdminMediaService,
	adminCommentSvc *service.AdminCommentService,
	adminTagSvc *service.AdminTagService,
	adminCategorySvc *service.AdminCategoryService,
	adminChannelSvc *service.AdminChannelService,
	adminPlaylistSvc *service.AdminPlaylistService,
	adminUserSvc *service.AdminUserService,
	permissionSvc *service.PermissionService,
	portalMgmtSvc *service.PortalManagementService,
	articleSvc *service.ArticleService,
	systemConfigSvc *service.SystemConfigService,
) (*transport.GRPCServer, error) {
	if cfg == nil {
		return nil, errors.New("grpc config is nil")
	}

	opts := &grpc.ServerOptions{}
	srv, err := grpc.NewServer(cfg, opts)
	if err != nil {
		return nil, err
	}

	media.RegisterMediaServiceServer(srv, mediaSvc)
	uploadv1.RegisterUploadServiceServer(srv, uploadSvcV1)
	media.RegisterEncodingProfileServiceServer(srv, mediaSvc)
	media.RegisterExploreServiceServer(srv, exploreSvc)
	media.RegisterAdminServiceServer(srv, adminSvc)
	media.RegisterAdminMediaServiceServer(srv, adminMediaSvc)
	media.RegisterAdminCommentServiceServer(srv, adminCommentSvc)
	media.RegisterAdminTagServiceServer(srv, adminTagSvc)
	media.RegisterAdminCategoryServiceServer(srv, adminCategorySvc)
	media.RegisterAdminChannelServiceServer(srv, adminChannelSvc)
	media.RegisterAdminPlaylistServiceServer(srv, adminPlaylistSvc)
	media.RegisterAdminUserServiceServer(srv, adminUserSvc)
	media.RegisterPermissionServiceServer(srv, permissionSvc)
	media.RegisterPortalManagementServiceServer(srv, portalMgmtSvc)
	media.RegisterArticleServiceServer(srv, articleSvc)
	media.RegisterSystemConfigServiceServer(srv, systemConfigSvc)
	return srv, nil
}

func NewHTTPServer(
	cfg *httpv1.Server,
	jwtMgr *auth.Manager,
	mediaSvc *service.MediaService,
	exploreSvc *service.ExploreService,
	adminSvc *service.AdminService,
	adminMediaSvc *service.AdminMediaService,
	adminCommentSvc *service.AdminCommentService,
	adminTagSvc *service.AdminTagService,
	adminCategorySvc *service.AdminCategoryService,
	adminChannelSvc *service.AdminChannelService,
	adminPlaylistSvc *service.AdminPlaylistService,
	adminUserSvc *service.AdminUserService,
	permissionSvc *service.PermissionService,
	portalMgmtSvc *service.PortalManagementService,
	articleSvc *service.ArticleService,
	systemConfigSvc *service.SystemConfigService,
	enterpriseMedia *EnterpriseMediaServer,
	uploadHandler *service.UploadHandler,
	mediaHandler *service.MediaHandler,
	storageProxy *service.StorageProxyService,
	spriteHandler *service.SpriteHandler,
) (*transport.HTTPServer, error) {
	if cfg == nil {
		return nil, errors.New("http config is nil")
	}

	opts := &runtimehttp.ServerOptions{}
	srv, err := runtimehttp.NewServer(cfg, opts)
	if err != nil {
		return nil, err
	}

	srv.Handle("/healthz", stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		w.WriteHeader(stdhttp.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","service":"media"}`))
	}))

	// Admin-only SSE endpoint for transcoding progress (with JWT + admin role check).
	// MUST be registered BEFORE proto Register*HTTPServer calls to ensure exact-path
	// routes take priority over parameterised routes like /admin/medias/{id}/tasks.
	sseHandler := stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		mediaSvc.SSEHandler(w, r)
	})
	srv.Handle("/api/v1/admin/medias/transcoding/events", server.StdWithAdmin(jwtMgr, sseHandler))
	srv.Handle("/api/v1/medias/transcoding/events", stdhttp.HandlerFunc(mediaSvc.SSEHandler))

	// Admin encoding/transcoding endpoints (with JWT + admin role check).
	srv.Handle("/api/v1/admin/encoding/tasks", server.StdWithAdmin(jwtMgr, stdhttp.HandlerFunc(mediaSvc.EncodingTasksHTTPHandler)))
	srv.Handle("/api/v1/admin/medias/encoding/tasks", server.StdWithAdmin(jwtMgr, stdhttp.HandlerFunc(mediaSvc.EncodingTasksHTTPHandler)))
	srv.Handle("/api/v1/admin/encoding/retry", server.StdWithAdmin(jwtMgr, stdhttp.HandlerFunc(mediaSvc.RetryTaskHTTPHandler)))
	srv.Handle("/api/v1/admin/encoding/retry-all-failed", server.StdWithAdmin(jwtMgr, stdhttp.HandlerFunc(mediaSvc.RetryAllFailedHTTPHandler)))
	srv.Handle("/api/v1/admin/medias/encoding/retry", server.StdWithAdmin(jwtMgr, stdhttp.HandlerFunc(mediaSvc.RetryTaskHTTPHandler)))
	srv.Handle("/api/v1/admin/medias/encoding/retry-all-failed", server.StdWithAdmin(jwtMgr, stdhttp.HandlerFunc(mediaSvc.RetryAllFailedHTTPHandler)))
	srv.Handle("/api/v1/admin/medias/transcoding/status", server.StdWithAdmin(jwtMgr, stdhttp.HandlerFunc(mediaSvc.TranscodingStatusHTTPHandler)))
	srv.Handle("/api/v1/medias/transcoding/status", stdhttp.HandlerFunc(mediaSvc.TranscodingStatusHTTPHandler))

	// Thumbnail management endpoints (gin-free net/http). These are registered
	// BEFORE the proto routes so they take priority over any overlapping
	// parameterized proto route. Distinct path literals are used (regen-thumbnail
	// / set-thumbnail) because Kratos/httprouter panics on duplicate routes, so
	// the proto `regenerate-thumbnail` literal must not be re-registered here.
	srv.Handle("/api/v1/admin/medias/{id}/regen-thumbnail", server.StdWithAdmin(jwtMgr, stdhttp.HandlerFunc(adminMediaSvc.AdminRegenThumbnailHTTPHandler)))
	srv.Handle("/api/v1/admin/medias/{id}/set-thumbnail", server.StdWithAdmin(jwtMgr, stdhttp.HandlerFunc(adminMediaSvc.AdminUploadThumbnailHTTPHandler)))
	srv.Handle("/api/v1/admin/medias/{id}/thumbnail", server.StdWithAdmin(jwtMgr, stdhttp.HandlerFunc(adminMediaSvc.AdminGetThumbnailHTTPHandler)))
	srv.Handle("/api/v1/me/medias/{token}/regen-thumbnail", server.StdWithJWT(jwtMgr, stdhttp.HandlerFunc(adminMediaSvc.OwnerRegenThumbnailHTTPHandler)))
	srv.Handle("/api/v1/me/medias/{token}/set-thumbnail", server.StdWithJWT(jwtMgr, stdhttp.HandlerFunc(adminMediaSvc.OwnerUploadThumbnailHTTPHandler)))

	// BUG-138: admin review-log history endpoint. Registered BEFORE the proto
	// RegisterAdminMediaServiceHTTPServer call below so this plain net/http
	// route wins over the GetReviewHistory empty stub.
	// Returns ReviewItem-shaped JSON consumed by the admin ReviewFlow log page.
	srv.Handle("/api/v1/admin/medias/review/history", server.StdWithAdmin(jwtMgr, stdhttp.HandlerFunc(mediaHandler.ReviewHistoryList)))

	// BUG-022: register sprite (raw VTT/JPEG) read routes BEFORE the proto
	// routes so they shadow the (unimplemented) proto GetMediaSpriteVTT/JPG
	// handlers. The media microservice owns these assets (ffmpeg-generated,
	// S3-stored), consistent with media_service.proto and the gateway->bridge
	// auto-registration norm (this is the "must-be-HTTP manual part").
	if spriteHandler != nil {
		spriteRouter := kratosadapter.NewRouterAdapter(srv, "/api/v1")
		spriteHandler.RegisterSpriteRoutes(spriteRouter)
	}

	// Register proto-generated HTTP servers
	media.RegisterMediaServiceHTTPServer(srv, mediaSvc)
	media.RegisterExploreServiceHTTPServer(srv, exploreSvc)
	media.RegisterAdminServiceHTTPServer(srv, adminSvc)
	media.RegisterAdminMediaServiceHTTPServer(srv, adminMediaSvc)
	media.RegisterAdminCommentServiceHTTPServer(srv, adminCommentSvc)
	media.RegisterAdminTagServiceHTTPServer(srv, adminTagSvc)
	media.RegisterAdminCategoryServiceHTTPServer(srv, adminCategorySvc)
	media.RegisterAdminChannelServiceHTTPServer(srv, adminChannelSvc)
	media.RegisterAdminPlaylistServiceHTTPServer(srv, adminPlaylistSvc)
	media.RegisterAdminUserServiceHTTPServer(srv, adminUserSvc)
	media.RegisterPermissionServiceHTTPServer(srv, permissionSvc)
	media.RegisterPortalManagementServiceHTTPServer(srv, portalMgmtSvc)
	media.RegisterArticleServiceHTTPServer(srv, articleSvc)
	media.RegisterSystemConfigServiceHTTPServer(srv, systemConfigSvc)
	media.RegisterEncodingProfileServiceHTTPServer(srv, mediaSvc)

	// Raw-binary upload endpoints (multipart uploads) via Gin engine.
	srv.HandlePrefix("/api/v1/uploads/", uploadHandler.HTTPHandler())

	// Public media encoding/transcoding endpoints via Gin engine.
	// Note: Gin handler takes priority for /api/v1/medias/* routes that need custom handling
	srv.HandlePrefix("/api/v1/medias/", mediaHandler.HTTPHandler())

	// Enterprise media routes (ad, drm, live) via Gin engine.
	srv.HandlePrefix("/api/v1/", enterpriseMedia.HTTPHandler())

	// Static file serving via StorageProxy (video files, HLS streams, thumbnails, etc.)
	srv.HandlePrefix("/files/", storageProxy)

	return srv, nil
}
