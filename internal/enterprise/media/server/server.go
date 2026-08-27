package server

import (
	"net/http"

	"github.com/google/wire"

	adservice "origadmin/application/origstudio/internal/enterprise/media/ad/service"
	drmservice "origadmin/application/origstudio/internal/enterprise/media/drm/service"
	http2 "origadmin/application/origstudio/internal/pkg/http"
	std "origadmin/application/origstudio/internal/pkg/http/std"
)

var ProviderSet = wire.NewSet(NewEnterpriseMediaServer, NewServers)

// EnterpriseMediaServer wraps the enterprise media handlers (ad, drm)
// and provides an HTTP handler that registers their routes on a Gin engine.
type EnterpriseMediaServer struct {
	adHandler  *adservice.Handler
	drmHandler *drmservice.Handler
}

func NewEnterpriseMediaServer(
	adHandler *adservice.Handler,
	drmHandler *drmservice.Handler,
) *EnterpriseMediaServer {
	return &EnterpriseMediaServer{
		adHandler:  adHandler,
		drmHandler: drmHandler,
	}
}

// RegisterRoutes registers all enterprise media routes on the given router.
func (s *EnterpriseMediaServer) RegisterRoutes(r http2.Router) {
	s.adHandler.RegisterRoutes(r)
	s.drmHandler.RegisterRoutes(r)
}

// HTTPHandler returns an http.Handler that serves all enterprise media routes.
// It creates a Gin engine, registers the routes under /api/v1, and returns
// the engine as a standard http.Handler.
func (s *EnterpriseMediaServer) HTTPHandler() http.Handler {
	router := std.NewRouter()
	apiV1 := router.Group("/api/v1")
	s.RegisterRoutes(apiV1)
	return router
}