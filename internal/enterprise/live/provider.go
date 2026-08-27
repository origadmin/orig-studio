package live

import (
	"github.com/google/wire"
	"origadmin/application/origstudio/internal/enterprise/live/biz"
	"origadmin/application/origstudio/internal/enterprise/live/dal"
	"origadmin/application/origstudio/internal/enterprise/live/dto"
	"origadmin/application/origstudio/internal/enterprise/live/service"
)

var ProviderSet = wire.NewSet(
	dto.ProviderSet,
	dal.ProviderSet,
	biz.ProviderSet,
	service.ProviderSet,
)