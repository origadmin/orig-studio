package server

import (
	"net/http"

	"github.com/google/wire"

	permissionservice "origadmin/application/origstudio/internal/enterprise/auth/permission/service"
	http2 "origadmin/application/origstudio/internal/pkg/http"
	std "origadmin/application/origstudio/internal/pkg/http/std"
)

var ProviderSet = wire.NewSet(NewEnterpriseAuthServer)

// EnterpriseAuthServer wraps the enterprise auth handler (permission)
// and provides an HTTP handler that registers its routes on a Gin engine.
type EnterpriseAuthServer struct {
	permissionHandler *permissionservice.Handler
}

func NewEnterpriseAuthServer(
	permissionHandler *permissionservice.Handler,
) *EnterpriseAuthServer {
	return &EnterpriseAuthServer{
		permissionHandler: permissionHandler,
	}
}

// RegisterRoutes registers all enterprise auth routes on the given router.
func (s *EnterpriseAuthServer) RegisterRoutes(r http2.Router) {
	s.permissionHandler.RegisterRoutes(r)
}

// HTTPHandler returns an http.Handler that serves all enterprise auth routes.
func (s *EnterpriseAuthServer) HTTPHandler() http.Handler {
	router := std.NewRouter()
	apiV1 := router.Group("/api/v1")
	s.RegisterRoutes(apiV1)
	return router
}