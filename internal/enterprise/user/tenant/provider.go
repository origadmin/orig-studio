package tenant

import (
	"github.com/google/wire"
	"origadmin/application/origstudio/internal/enterprise/user/tenant/biz"
	"origadmin/application/origstudio/internal/enterprise/user/tenant/service"
	"origadmin/application/origstudio/internal/enterprise/user/tenant/dto"
)

var ProviderSet = wire.NewSet(
	dto.ProviderSet,
	biz.ProviderSet,
	service.ProviderSet,
)