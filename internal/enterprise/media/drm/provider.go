package drm

import (
	"github.com/google/wire"
	"origadmin/application/origstudio/internal/enterprise/media/drm/biz"
	"origadmin/application/origstudio/internal/enterprise/media/drm/dal"
	"origadmin/application/origstudio/internal/enterprise/media/drm/dto"
	"origadmin/application/origstudio/internal/enterprise/media/drm/service"
)

var ProviderSet = wire.NewSet(
	dto.ProviderSet,
	dal.ProviderSet,
	biz.ProviderSet,
	service.ProviderSet,
)