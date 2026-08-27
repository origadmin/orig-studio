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
	"github.com/origadmin/runtime/service/transport/http"

	livev1 "origadmin/application/origstudio/api/gen/v1/live"
	liveservice "origadmin/application/origstudio/internal/enterprise/live/service"
)

var ProviderSet = wire.NewSet(NewServers)

func NewServers(
	cfg *transportv1.Servers,
	liveSvc *liveservice.LiveService,
) ([]transport.Server, error) {
	if cfg == nil {
		return nil, errors.New("servers config is nil")
	}

	var servers []transport.Server
	for _, serverCfg := range cfg.GetConfigs() {
		if serverCfg.GetName() != "live" && serverCfg.GetName() != "origcms.svc-live" {
			continue
		}
		switch serverCfg.GetProtocol() {
		case "grpc":
			srv, err := NewGRPCServer(serverCfg.GetGrpc(), liveSvc)
			if err != nil {
				return nil, err
			}
			servers = append(servers, srv)
		case "http":
			srv, err := NewHTTPServer(serverCfg.GetHttp())
			if err != nil {
				return nil, err
			}
			servers = append(servers, srv)
		default:
			log.Warnf("protocol '%s' not supported by live service, skipping", serverCfg.GetProtocol())
		}
	}

	if len(servers) == 0 {
		return nil, errors.New("no servers named 'live' were created")
	}
	return servers, nil
}

func NewGRPCServer(
	cfg *grpcv1.Server,
	liveSvc *liveservice.LiveService,
) (*transport.GRPCServer, error) {
	if cfg == nil {
		return nil, errors.New("grpc config is nil")
	}

	opts := &grpc.ServerOptions{}
	srv, err := grpc.NewServer(cfg, opts)
	if err != nil {
		return nil, err
	}

	livev1.RegisterLiveServiceServer(srv, liveSvc)
	return srv, nil
}

func NewHTTPServer(cfg *httpv1.Server) (*transport.HTTPServer, error) {
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
		_, _ = w.Write([]byte(`{"status":"ok","service":"live"}`))
	}))

	return srv, nil
}
