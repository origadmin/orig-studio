package permission

import (
	"github.com/google/wire"
	"origadmin/application/origstudio/internal/enterprise/auth/permission/biz"
	"origadmin/application/origstudio/internal/enterprise/auth/permission/dal"
	"origadmin/application/origstudio/internal/enterprise/auth/permission/dto"
	"origadmin/application/origstudio/internal/enterprise/auth/permission/service"
)

var ProviderSet = wire.NewSet(
	dto.ProviderSet,
	dal.ProviderSet,
	biz.ProviderSet,
	service.ProviderSet,
)