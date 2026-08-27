package portal

import (
	"github.com/google/wire"
	"origadmin/application/origstudio/internal/enterprise/content/portal/biz"
	"origadmin/application/origstudio/internal/enterprise/content/portal/dal"
	"origadmin/application/origstudio/internal/enterprise/content/portal/dto"
	"origadmin/application/origstudio/internal/enterprise/content/portal/service"
)

var ProviderSet = wire.NewSet(
	dto.ProviderSet,
	dal.ProviderSet,
	biz.ProviderSet,
	service.ProviderSet,
)