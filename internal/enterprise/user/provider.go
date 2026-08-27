package user

import (
	"github.com/google/wire"

	enterpriseauth "origadmin/application/origstudio/internal/enterprise/auth/server"
	"origadmin/application/origstudio/internal/enterprise/auth/permission"
	"origadmin/application/origstudio/internal/enterprise/user/server"
	"origadmin/application/origstudio/internal/enterprise/user/tenant"
)

// ProviderSet is the top-level Wire provider set for the EE user module.
// It aggregates all sub-module providers so cmd/user/wire.go only needs to
// import enterprise/user (not features/user/server which violates B2 layering).
var ProviderSet = wire.NewSet(
	server.ServerProviderSet,  // NewServers (transport assembly, migrated from features/user/server)
	server.ProviderSet,        // NewEnterpriseUserServer (tenant + admin routes)
	tenant.ProviderSet,        // tenant biz/dal/dto/service
	permission.ProviderSet,    // permission subsystem
	enterpriseauth.ProviderSet, // enterprise auth server
)
