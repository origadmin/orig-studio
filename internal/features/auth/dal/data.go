package dal

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewData,
	NewPermissionGroupRepo,
	NewGroupMemberRepo,
	NewUserPermRepo,
)
