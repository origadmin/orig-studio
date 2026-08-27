package ad

import (
	"github.com/google/wire"
	"origadmin/application/origstudio/internal/enterprise/media/ad/biz"
	"origadmin/application/origstudio/internal/enterprise/media/ad/dal"
	"origadmin/application/origstudio/internal/enterprise/media/ad/dto"
	"origadmin/application/origstudio/internal/enterprise/media/ad/service"
)

var ProviderSet = wire.NewSet(
	dto.ProviderSet,
	dal.ProviderSet,
	biz.ProviderSet,
	service.ProviderSet,
)