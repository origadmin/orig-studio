package content

import (
	"github.com/google/wire"

	"origadmin/application/origstudio/internal/enterprise/content/portal"
	"origadmin/application/origstudio/internal/enterprise/content/server"
)

// ProviderSet is the top-level Wire provider set for the EE content module.
// It aggregates all sub-module providers so cmd/content/wire.go only needs to
// import enterprise/content (not features/content/server which violates B2 layering).
var ProviderSet = wire.NewSet(
	server.ServerProviderSet,    // NewServers (transport assembly, migrated from features/content/server)
	server.ProviderSet,          // NewEnterpriseContentServer (portal routes)
	portal.ProviderSet,          // portal biz/dal/dto/service
)
