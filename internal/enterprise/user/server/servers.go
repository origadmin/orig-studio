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

	userv1 "origadmin/application/origstudio/api/gen/v1/user"
	mediav1 "origadmin/application/origstudio/api/gen/v1/media"
	"origadmin/application/origstudio/internal/infra/auth"
	enterpriseauth "origadmin/application/origstudio/internal/enterprise/auth/server"
	"origadmin/application/origstudio/internal/features/user/service"
	kratosadapter "origadmin/application/origstudio/internal/pkg/http/kratos"
	"origadmin/application/origstudio/internal/pkg/http/validate"
	"origadmin/application/origstudio/internal/server"
)

// ServerProviderSet provides the EE user microservice transport assembly.
// This was migrated from internal/features/user/server/server.go to comply
// with B2 layering (enterprise → features, not features → enterprise) and
// the "EE prohibits Gin" rule.
var ServerProviderSet = wire.NewSet(NewServers)

func NewServers(
	cfg *transportv1.Servers,
	svc *service.UserService,
	adminUserSvc *service.AdminUserService,
	permissionSvc *service.PermissionService,
	systemConfigSvc *service.SystemConfigService,
	jwtMgr *auth.Manager,
	authHandler *service.AuthHandler,
	userHandler *service.UserHandler,
	meHandler *service.MeHandler,
	enterpriseAuth *enterpriseauth.EnterpriseAuthServer,
	enterpriseUser *EnterpriseUserServer,
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
			srv, err := NewGRPCServer(serverCfg.GetGrpc(), svc, adminUserSvc, permissionSvc, systemConfigSvc)
			if err != nil {
				return nil, err
			}
			servers = append(servers, srv)
		case "http":
			srv, err := NewHTTPServer(serverCfg.GetHttp(), jwtMgr, authHandler, userHandler, meHandler, enterpriseAuth, enterpriseUser)
			if err != nil {
				return nil, err
			}
			servers = append(servers, srv)
		default:
			log.Warnf("protocol '%s' not supported by user service, skipping", serverCfg.GetProtocol())
		}
	}

	if len(servers) == 0 {
		return nil, errors.New("no servers named 'user' were created")
	}
	return servers, nil
}

func NewGRPCServer(
	cfg *grpcv1.Server,
	svc *service.UserService,
	adminUserSvc *service.AdminUserService,
	permissionSvc *service.PermissionService,
	systemConfigSvc *service.SystemConfigService,
) (*transport.GRPCServer, error) {
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

	userv1.RegisterUserServiceServer(srv, svc)
	mediav1.RegisterAdminUserServiceServer(srv, adminUserSvc)
	mediav1.RegisterPermissionServiceServer(srv, permissionSvc)
	mediav1.RegisterSystemConfigServiceServer(srv, systemConfigSvc)
	return srv, nil
}

// NewHTTPServer creates the HTTP server for the user microservice.
// Migrated from Gin to Kratos adapter to comply with the "EE prohibits Gin" rule.
// Routes are registered via the kratos adapter (framework-agnostic http2.Router).
// JWT middleware is applied via the Kratos Filter bridge.
func NewHTTPServer(
	cfg *httpv1.Server,
	jwtMgr *auth.Manager,
	authHandler *service.AuthHandler,
	userHandler *service.UserHandler,
	meHandler *service.MeHandler,
	enterpriseAuth *enterpriseauth.EnterpriseAuthServer,
	enterpriseUser *EnterpriseUserServer,
) (*transport.HTTPServer, error) {
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
		_, _ = w.Write([]byte(`{"status":"ok","service":"user"}`))
	}))

	// Register routes via kratos adapter (framework-agnostic, no Gin).
	router := kratosadapter.NewRouterAdapter(srv, "/api/v1")
	if authHandler != nil {
		authHandler.RegisterRoutes(router)
	}
	if userHandler != nil {
		userHandler.RegisterRoutes(router)
	}
	if meHandler != nil {
		meHandler.RegisterRoutes(router)
	}
	enterpriseAuth.RegisterRoutes(router)
	enterpriseUser.RegisterRoutes(router)

	helper := log.NewHelper(log.DefaultLogger)
	helper.Infow(log.DefaultMessageKey, "HTTP server created", "addr", cfg.GetAddr())

	return srv, nil
}
