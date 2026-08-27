package server

import (
	"errors"
	stdhttp "net/http"
	"os"

	"github.com/google/wire"

	"github.com/origadmin/runtime/log"
	"github.com/origadmin/runtime/service/transport"
	"github.com/origadmin/runtime/service/transport/http"

	httpv1 "github.com/origadmin/runtime/api/gen/go/config/transport/http/v1"
	transportv1 "github.com/origadmin/runtime/api/gen/go/config/transport/v1"

	auditservice "origadmin/application/origstudio/internal/enterprise/audit/service"
	http2 "origadmin/application/origstudio/internal/pkg/http"
	std "origadmin/application/origstudio/internal/pkg/http/std"
)

var ProviderSet = wire.NewSet(NewServers, NewEnterpriseAuditServer)

type EnterpriseAuditServer struct {
	auditHandler *auditservice.Handler
}

func NewEnterpriseAuditServer(
	auditHandler *auditservice.Handler,
) *EnterpriseAuditServer {
	return &EnterpriseAuditServer{
		auditHandler: auditHandler,
	}
}

func (s *EnterpriseAuditServer) RegisterRoutes(r http2.Router) {
	s.auditHandler.RegisterRoutes(r)
}

func (s *EnterpriseAuditServer) HTTPHandler() stdhttp.Handler {
	router := std.NewRouter()
	apiV1 := router.Group("/api/v1")
	s.RegisterRoutes(apiV1)
	return router
}

func NewServers(cfg *transportv1.Servers, enterpriseAudit *EnterpriseAuditServer) ([]transport.Server, error) {
	if cfg == nil {
		return nil, errors.New("servers config is nil")
	}
	var servers []transport.Server
	for _, serverCfg := range cfg.GetConfigs() {
		if serverCfg.GetName() != "audit" && serverCfg.GetName() != "origcms.svc-audit" {
			continue
		}
		switch serverCfg.GetProtocol() {
		case "http":
			srv, err := NewHTTPServer(serverCfg.GetHttp(), enterpriseAudit)
			if err != nil {
				return nil, err
			}
			servers = append(servers, srv)
		default:
			log.Warnf("protocol '%s' not supported by audit service, skipping", serverCfg.GetProtocol())
		}
	}
	if len(servers) == 0 {
		// If no audit config found, create a default HTTP server on port from env
		port := os.Getenv("AUDIT_PORT")
		if port == "" {
			port = "8085"
		}
		srv, err := NewHTTPServer(&httpv1.Server{Addr: ":" + port}, enterpriseAudit)
		if err != nil {
			return nil, err
		}
		servers = append(servers, srv)
	}
	return servers, nil
}

func NewHTTPServer(cfg *httpv1.Server, enterpriseAudit *EnterpriseAuditServer) (*transport.HTTPServer, error) {
	if cfg == nil {
		return nil, errors.New("http config is nil")
	}
	opts := &http.ServerOptions{}
	srv, err := http.NewServer(cfg, opts)
	if err != nil {
		return nil, err
	}

	srv.Handle("/healthz", stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		w.WriteHeader(stdhttp.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","service":"audit"}`))
	}))

	srv.HandlePrefix("/api/v1/", enterpriseAudit.HTTPHandler())
	return srv, nil
}