package audit

import (
	"github.com/google/wire"

	"origadmin/application/origstudio/internal/enterprise/audit/dal"
	"origadmin/application/origstudio/internal/enterprise/audit/dto"
	"origadmin/application/origstudio/internal/enterprise/audit/service"
)

var ProviderSet = wire.NewSet(
	dto.ProviderSet,
	dal.ProviderSet,
	service.ProviderSet,
)