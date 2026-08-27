// Deprecated: This is the old monolith provider for cmd/enterprise.
// All enterprise modules have been migrated to individual Kratos microservices.
// See the respective cmd/{service}/wire.go files for the new wire injection setup.
// EESet has been removed; only CESet remains for cmd/server.
package enterprise

import (
	"context"

	"github.com/google/wire"
	"github.com/origadmin/runtime/log"

	"origadmin/application/origstudio/internal/conf"
	featureauth "origadmin/application/origstudio/internal/features/auth"
	authbiz "origadmin/application/origstudio/internal/features/auth/biz"
	authservice "origadmin/application/origstudio/internal/features/auth/service"
	"origadmin/application/origstudio/internal/features/admin"
	"origadmin/application/origstudio/internal/features/content"
	contentbiz "origadmin/application/origstudio/internal/features/content/biz"
	contentservice "origadmin/application/origstudio/internal/features/content/service"
	"origadmin/application/origstudio/internal/features/media"
	mediabiz "origadmin/application/origstudio/internal/features/media/biz"
	mediaservice "origadmin/application/origstudio/internal/features/media/service"
	"origadmin/application/origstudio/internal/features/system"
	systembiz "origadmin/application/origstudio/internal/features/system/biz"
	"origadmin/application/origstudio/internal/features/user"
	userbiz "origadmin/application/origstudio/internal/features/user/biz"
	"origadmin/application/origstudio/internal/infra"
	infraauth "origadmin/application/origstudio/internal/infra/auth"
	"origadmin/application/origstudio/internal/server/middleware"
)

var CESet = wire.NewSet(
	infra.ProviderSet,
	media.ProviderSet,
	content.ProviderSet,
	user.ProviderSet,
	featureauth.ProviderSet,
	admin.ProviderSet,
	system.ProviderSet,

	conf.NewStorageConfigFromSettings,
	conf.NewUploadConfigFromDefaults,
	conf.NewStoragePathsFromConfig,
	conf.NewTranscodeConfigFromDefaults,

	NewWorker,
	NewRateLimiter,
	NewAuthHandler,
	NewMediaReportHandler,
	NewStubHandler,
	NewSpriteHandler,
	NewNoopPermissionChecker,

	wire.Bind(new(authbiz.PermissionChecker), new(*noopPermissionChecker)),
	wire.Bind(new(contentbiz.MediaUseCaseInterface), new(*mediabiz.MediaUseCase)),
	wire.Bind(new(contentbiz.SettingReader), new(*systembiz.SettingUseCase)),
	wire.Bind(new(systembiz.ConfigProvider), new(*systembiz.SettingUseCase)),
)

type noopPermissionChecker struct{}

func NewNoopPermissionChecker() *noopPermissionChecker {
	return &noopPermissionChecker{}
}

func (n *noopPermissionChecker) CheckPermission(_ context.Context, _ string, _ string, _ string) (bool, error) {
	return true, nil
}

func (n *noopPermissionChecker) InvalidateUserCache(_ context.Context, _ string) error {
	return nil
}

func (n *noopPermissionChecker) InvalidateGroupCache(_ context.Context, _ string) error {
	return nil
}

func NewWorker(cfg *conf.Config, encodingRepo mediabiz.EncodingTaskRepo, mediaUC *mediabiz.MediaUseCase, logger log.Logger) mediabiz.TranscodeWorker {
	if cfg.GetAsynq() != nil {
		a := cfg.GetAsynq()
		workerConfig := mediabiz.DefaultAsynqWorkerConfig()
		workerConfig.RedisAddr = a.GetRedisAddr()
		workerConfig.RedisPassword = a.GetRedisPassword()
		if a.GetConcurrency() > 0 {
			workerConfig.Concurrency = int(a.GetConcurrency())
		}
		return mediabiz.NewAsynqWorker(workerConfig, encodingRepo, mediaUC, log.NewHelper(log.With(logger, "module", "asynq.worker")))
	}
	maxWorkers := int32(infra.EnvInt("TRANSCODE_MAX_WORKERS", 3))
	return mediabiz.NewGoroutineWorker(maxWorkers, log.NewHelper(log.With(logger, "module", "transcode.worker")))
}

func NewRateLimiter(settingUC *systembiz.SettingUseCase) *middleware.RateLimiter {
	rpm, excludePrefixes := conf.NewRateLimiterConfig(settingUC)
	return middleware.NewRateLimiter(rpm, excludePrefixes...)
}

func NewAuthHandler(uc *userbiz.UserUseCase, jwt *infraauth.Manager, settingUC *systembiz.SettingUseCase) *authservice.AuthHandler {
	// cmd/server is the deprecated CE monolith. Audit logging (BUG-265) is
	// served by the standalone cmd/audit microservice; the monolith keeps
	// auth audit dormant (no SetAuditFunc) to match prior behavior.
	return authservice.NewAuthHandler(uc, jwt, settingUC)
}

func NewMediaReportHandler(
	mediaReportUC *contentbiz.MediaReportUseCase,
	jwt *infraauth.Manager,
) *contentservice.MediaReportHandler {
	return contentservice.NewMediaReportHandler(mediaReportUC, jwt)
}

func NewStubHandler(jwt *infraauth.Manager, mediaUC *mediabiz.MediaUseCase, sp *conf.StoragePaths, logger log.Logger) *contentservice.StubHandler {
	// cmd/server is the deprecated CE monolith (microservices are the runtime).
	// Subtitle (BUG-186) requires a content Data handle; pass nil here so the
	// monolith still compiles — the real content microservice wires it fully.
	return contentservice.NewStubHandler(jwt, mediaUC, sp, logger, nil)
}

func NewSpriteHandler(mediaUC *mediabiz.MediaUseCase, sp *conf.StoragePaths, jwt *infraauth.Manager, logger log.Logger) *mediaservice.SpriteHandler {
	return mediaservice.NewSpriteHandler(mediaUC, sp, jwt, logger)
}
