package dal

import (
	"github.com/google/wire"
	authbiz "origadmin/application/origstudio/internal/features/auth/biz"
)

var ProviderSet = wire.NewSet(
	NewGroupRepo,
	NewMemberRepo,
	NewUserPermRepo,
	NewCheckerAdapter,
	wire.Bind(new(authbiz.PermissionChecker), new(*CheckerAdapter)),
)