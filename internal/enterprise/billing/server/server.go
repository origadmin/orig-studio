package server

import (
	"errors"
	stdhttp "net/http"

	paymentservice "origadmin/application/origstudio/internal/enterprise/billing/payment/service"
	http2 "origadmin/application/origstudio/internal/pkg/http"
	std "origadmin/application/origstudio/internal/pkg/http/std"

	httpv1 "github.com/origadmin/runtime/api/gen/go/config/transport/http/v1"
	transportv1 "github.com/origadmin/runtime/api/gen/go/config/transport/v1"
	"github.com/origadmin/runtime/log"
	"github.com/origadmin/runtime/service/transport"
	"github.com/origadmin/runtime/service/transport/http"
)

// EnterpriseBillingServer wraps the enterprise billing handler (payment)
// and provides an HTTP handler that registers its routes on a Gin engine.
type EnterpriseBillingServer struct {
	paymentHandler *paymentservice.Handler
}

func NewEnterpriseBillingServer(
	paymentHandler *paymentservice.Handler,
) *EnterpriseBillingServer {
	return &EnterpriseBillingServer{
		paymentHandler: paymentHandler,
	}
}

// RegisterRoutes registers all enterprise billing routes on the given router.
func (s *EnterpriseBillingServer) RegisterRoutes(r http2.Router) {
	s.paymentHandler.RegisterRoutes(r)
}

// HTTPHandler returns an http.Handler that serves all enterprise billing routes.
func (s *EnterpriseBillingServer) HTTPHandler() stdhttp.Handler {
	router := std.NewRouter()
	apiV1 := router.Group("/api/v1")
	s.RegisterRoutes(apiV1)
	return router
}

// NewServers creates Kratos transport servers for the billing service
// based on the provided servers configuration.
func NewServers(cfg *transportv1.Servers, enterpriseBilling *EnterpriseBillingServer) ([]transport.Server, error) {
	if cfg == nil {
		return nil, errors.New("servers config is nil")
	}
	var servers []transport.Server
	for _, serverCfg := range cfg.GetConfigs() {
		if serverCfg.GetName() != "billing" && serverCfg.GetName() != "origcms.svc-billing" {
			continue
		}
		switch serverCfg.GetProtocol() {
		case "http":
			srv, err := NewHTTPServer(serverCfg.GetHttp(), enterpriseBilling)
			if err != nil {
				return nil, err
			}
			servers = append(servers, srv)
		default:
			log.Warnf("protocol '%s' not supported by billing service, skipping", serverCfg.GetProtocol())
		}
	}
	if len(servers) == 0 {
		return nil, errors.New("no servers named 'billing' were created")
	}
	return servers, nil
}

// NewHTTPServer creates a Kratos HTTP transport server for the billing service.
func NewHTTPServer(cfg *httpv1.Server, enterpriseBilling *EnterpriseBillingServer) (*transport.HTTPServer, error) {
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
		_, _ = w.Write([]byte(`{"status":"ok","service":"billing"}`))
	}))

	srv.HandlePrefix("/api/v1/", enterpriseBilling.HTTPHandler())
	return srv, nil
}