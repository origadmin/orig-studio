//go:build wireinject
// +build wireinject

/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package main

import (
	"context"
	"fmt"

	"github.com/ThreeDotsLabs/watermill/message"
	_ "github.com/lib/pq" // PostgreSQL driver

	kratoslog "github.com/go-kratos/kratos/v2/log"
	"github.com/origadmin/runtime/log"
	_ "github.com/sqlite3ent/sqlite3" // SQLite3 driver

	config "origadmin/application/origstudio/internal/conf"
	"origadmin/application/origstudio/internal/dal/entity"
	dal4 "origadmin/application/origstudio/internal/features/content/dal"
	"origadmin/application/origstudio/internal/features/admin"
	adminservice "origadmin/application/origstudio/internal/features/admin/service"
	systemdal "origadmin/application/origstudio/internal/features/system/dal"
	featureauth "origadmin/application/origstudio/internal/features/auth"
	authbiz "origadmin/application/origstudio/internal/features/auth/biz"
	authservice "origadmin/application/origstudio/internal/features/auth/service"
	"origadmin/application/origstudio/internal/features/content"
	contentbiz "origadmin/application/origstudio/internal/features/content/biz"
	contentservice "origadmin/application/origstudio/internal/features/content/service"
	"origadmin/application/origstudio/internal/features/media"
	mediabiz "origadmin/application/origstudio/internal/features/media/biz"
	mediadal "origadmin/application/origstudio/internal/features/media/dal"
	mediaservice "origadmin/application/origstudio/internal/features/media/service"
	"origadmin/application/origstudio/internal/features/system"
	systembiz "origadmin/application/origstudio/internal/features/system/biz"
	systemservice "origadmin/application/origstudio/internal/features/system/service"
	"origadmin/application/origstudio/internal/features/user"
	userbiz "origadmin/application/origstudio/internal/features/user/biz"
	userservice "origadmin/application/origstudio/internal/features/user/service"
	"origadmin/application/origstudio/internal/infra"
	infraauth "origadmin/application/origstudio/internal/infra/auth"
	"origadmin/application/origstudio/internal/infra/pubsub"
	"origadmin/application/origstudio/internal/server/middleware"

	"github.com/google/wire"
)

// ProviderSet is the wire provider set for the application.
// It aggregates all module ProviderSets and retains bridge functions
// for constructors that require hardcoded configuration values or
// interface bindings that cannot be expressed in module ProviderSets.
var ProviderSet = wire.NewSet(
	infra.ProviderSet,
	media.ProviderSet,
	content.ProviderSet,
	user.ProviderSet,
	featureauth.ProviderSet,
	admin.ProviderSet,
	system.ProviderSet,

	// Config providers
	NewUploadConfig,
	NewTranscodeConfig,
	NewStoragePaths,
	NewStorageConfig,
	NewRateLimiter,

	// Bridge functions for media module
	NewStorage,
	NewStorageInterface,
	NewWorker,
	NewUploadUseCase,
	NewSpriteUseCase,
	NewTranscodeHandler,
	NewDatabaseBridge,
		// Bridge functions for handler constructors
	NewAuthHandler,
	NewMediaReportHandler,
	NewStubHandler,
	NewSpriteHandler,

	// Wire bindings
	wire.Bind(new(authbiz.PermissionChecker), new(*authbiz.PermissionUseCase)),
	wire.Bind(new(contentbiz.MediaUseCaseInterface), new(*mediabiz.MediaUseCase)),
	wire.Bind(new(systembiz.ConfigProvider), new(*systembiz.SettingUseCase)),
)

// NewDatabaseBridge wraps infra.NewDatabase to return a cleanup function
// instead of *sql.DB, which is required by wire's provider signature convention.
func NewDatabaseBridge(cfg *config.Config, logger log.Logger) (*entity.Client, error) {
	client, _, err := infra.NewDatabase(cfg, logger)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// NewAdminHandlerBridge creates a new admin handler with hardcoded string parameters
// that wire cannot inject automatically (multiple string params).
func NewAdminHandlerBridge(
	jwt *infraauth.Manager,
	mediaUC *mediabiz.MediaUseCase,
	mediaService *mediaservice.MediaService,
	channelUC *contentbiz.PlaylistChannelUseCase,
	tagService *adminservice.TagService,
	settingUC *systembiz.SettingUseCase,
	categoryUC *contentbiz.CategoryTagUseCase,
	articleUC *contentbiz.ArticleUseCase,
	userUC *userbiz.UserUseCase,
	permChecker authbiz.PermissionChecker,
	db *entity.Client,
	cfg *config.Config,
) *adminservice.AdminHandler {
	return adminservice.NewAdminHandler(jwt, mediaUC, mediaService, channelUC, tagService, settingUC, categoryUC, articleUC, userUC, permChecker, systemdal.NewStatsRepo(db), adminservice.NewAdminConfig(cfg))
}

// NewStorageConfig creates storage config from defaults.
func NewStorageConfig(settingUC *systembiz.SettingUseCase) *config.StorageConfig {
	cfg := config.DefaultStorageConfig()
	if basePath := settingUC.Get(context.Background(), "storage_base_path"); basePath != "" {
		cfg.BasePath = basePath
	}
	if storageType := settingUC.Get(context.Background(), "storage_type"); storageType != "" {
		cfg.Type = config.StorageType(storageType)
	}
	if endpoint := settingUC.Get(context.Background(), "s3_endpoint"); endpoint != "" {
		cfg.S3.Endpoint = endpoint
	}
	if region := settingUC.Get(context.Background(), "s3_region"); region != "" {
		cfg.S3.Region = region
	}
	if bucket := settingUC.Get(context.Background(), "s3_bucket"); bucket != "" {
		cfg.S3.Bucket = bucket
	}
	if accessKey := settingUC.Get(context.Background(), "s3_access_key"); accessKey != "" {
		cfg.S3.AccessKey = accessKey
	}
	if secretKey := settingUC.Get(context.Background(), "s3_secret_key"); secretKey != "" {
		cfg.S3.SecretKey = secretKey
	}
	if usePathStyle := settingUC.GetBool(context.Background(), "s3_use_path_style"); usePathStyle {
		cfg.S3.UsePathStyle = true
	}
	return cfg
}

// NewStorage creates a LocalStorage instance (used as the base for all storage types).
func NewStorage(sp *config.StoragePaths) *mediadal.LocalStorage {
	return mediadal.NewLocalStorage(sp)
}

func newContentData(client *entity.Client) *dal4.Data {
	return dal4.NewData(client)
}

// NewStorageInterface creates the appropriate Storage implementation based on
// StorageConfig.Type. It always creates LocalStorage; for "s3" type it also
// creates S3Storage; for "hybrid" type it creates HybridStorage with async sync.
func NewStorageInterface(
	local *mediadal.LocalStorage,
	sp *config.StoragePaths,
	cfg *config.StorageConfig,
	logger kratoslog.Logger,
) (mediabiz.Storage, func(), error) {
	switch cfg.Type {
	case config.StorageTypeS3:
		s3Storage, err := mediadal.NewS3Storage(&cfg.S3, logger)
		if err != nil {
			return nil, func() {}, fmt.Errorf("create S3 storage: %w", err)
		}
		return s3Storage, func() {}, nil

	case config.StorageTypeHybrid:
		s3Storage, err := mediadal.NewS3Storage(&cfg.S3, logger)
		if err != nil {
			return nil, func() {}, fmt.Errorf("create S3 storage for hybrid: %w", err)
		}
		hs := mediadal.NewHybridStorage(local, s3Storage, sp, cfg.Hybrid, logger)
		return hs, func() { hs.Close() }, nil

	case config.StorageTypeLocal:
		fallthrough
	default:
		return local, func() {}, nil
	}
}

// NewUploadConfig creates upload config from defaults.
func NewUploadConfig() *config.UploadConfig {
	return config.DefaultUploadConfig()
}

// NewStoragePaths creates StoragePaths from StorageConfig.
func NewStoragePaths(cfg *config.StorageConfig) *config.StoragePaths {
	return config.NewStoragePaths(cfg.BasePath)
}

// NewTranscodeConfig creates transcode config from defaults.
func NewTranscodeConfig() *config.TranscodeConfig {
	return config.DefaultTranscodeConfig()
}

// NewWorker creates a new transcode worker with config from environment.
func NewWorker(logger log.Logger) mediabiz.TranscodeWorker {
	maxWorkers := int32(infra.EnvInt("TRANSCODE_MAX_WORKERS", 3))
	return mediabiz.NewGoroutineWorker(maxWorkers, log.NewHelper(log.With(logger, "module", "transcode.worker")))
}

// NewRateLimiter creates a new rate limiter with rpm from settings.
// Upload endpoints are excluded from rate limiting to prevent upload failures.
func NewRateLimiter(settingUC *systembiz.SettingUseCase) *middleware.RateLimiter {
	defaultRPM := 60
	if settingUC != nil {
		if val := settingUC.Get(context.Background(), "api_rate_limit"); val != "" {
			var rpm int
			if _, err := fmt.Sscanf(val, "%d", &rpm); err == nil && rpm > 0 {
				defaultRPM = rpm
			}
		}
	}
	return middleware.NewRateLimiter(defaultRPM, "/api/v1/uploads")
}

// NewUploadUseCase creates a new upload use case with config from UploadConfig.
func NewUploadUseCase(
	uploadRepo mediabiz.UploadRepo,
	mediaRepo mediabiz.MediaRepo,
	profileRepo mediabiz.EncodeProfileRepo,
	taskRepo mediabiz.EncodingTaskRepo,
	mediaUC *mediabiz.MediaUseCase,
	storage mediabiz.Storage,
	sp *config.StoragePaths,
	cfg *config.UploadConfig,
	logger log.Logger,
	settingUC *systembiz.SettingUseCase,
) *mediabiz.UploadUseCase {
	return mediabiz.NewUploadUseCase(
		uploadRepo,
		mediaRepo,
		profileRepo,
		taskRepo,
		mediaUC,
		storage,
		sp,
		cfg.ChunkSize,
		logger,
		settingUC,
	)
}

// NewSpriteUseCase creates a new sprite use case with paths from StoragePaths.
func NewSpriteUseCase(
	mediaRepo mediabiz.MediaRepo,
	settingUC *systembiz.SettingUseCase,
	sp *config.StoragePaths,
	logger log.Logger,
) *mediabiz.SpriteUseCase {
	return mediabiz.NewSpriteUseCase(mediaRepo, settingUC, sp, logger)
}

// NewTranscodeHandler creates a new transcode handler with paths from StoragePaths.
func NewTranscodeHandler(
	mediaUC *mediabiz.MediaUseCase,
	profileRepo mediabiz.EncodeProfileRepo,
	taskRepo mediabiz.EncodingTaskRepo,
	mediaRepo mediabiz.MediaRepo,
	worker mediabiz.TranscodeWorker,
	publisher message.Publisher,
	logger log.Logger,
	sp *config.StoragePaths,
	cfg *config.TranscodeConfig,
	spriteUC *mediabiz.SpriteUseCase,
) *mediabiz.TranscodeHandler {
	return mediabiz.NewTranscodeHandler(
		mediaUC,
		profileRepo,
		taskRepo,
		mediaRepo,
		worker,
		publisher,
		logger,
		sp,
		cfg.TaskTimeout,
		spriteUC,
	)
}

// NewAuthHandler creates a new auth handler with config provider.
func NewAuthHandler(uc *userbiz.UserUseCase, jwt *infraauth.Manager, settingUC *systembiz.SettingUseCase) *authservice.AuthHandler {
	return authservice.NewAuthHandler(uc, jwt, settingUC)
}

// NewMediaReportHandler creates a new media report handler.
func NewMediaReportHandler(
	mediaReportUC *contentbiz.MediaReportUseCase,
	jwt *infraauth.Manager,
) *contentservice.MediaReportHandler {
	return contentservice.NewMediaReportHandler(mediaReportUC, jwt)
}

// NewStubHandler creates a new stub handler for missing routes.
func NewStubHandler(jwt *infraauth.Manager) *contentservice.StubHandler {
	return contentservice.NewStubHandler(jwt)
}

// NewSpriteHandler creates a new sprite handler for sprite sheet and VTT routes.
func NewSpriteHandler(mediaUC *mediabiz.MediaUseCase, sp *config.StoragePaths, jwt *infraauth.Manager, logger log.Logger) *contentservice.SpriteHandler {
	return contentservice.NewSpriteHandler(mediaUC, sp, jwt, logger)
}

// AppDependencies holds all application dependencies.
type AppDependencies struct {
	DB                       *entity.Client
	PubSub                   *pubsub.PubSub
	Router                   *message.Router
	JWTManager               *infraauth.Manager
	StoragePaths             *config.StoragePaths
	AuthHandler              *authservice.AuthHandler
	PermissionHandler        *authservice.PermissionHandler
	UserHandler              *userservice.UserHandler
	MeHandler                *userservice.MeHandler
	MediaHandler             *mediaservice.MediaHandler
	UploadHandler            *mediaservice.UploadHandler
	SearchHandler            *mediaservice.SearchHandler
	CategoryHandler          *contentservice.CategoryHandler
	TagHandler               *contentservice.TagHandler
	ArticleHandler           *contentservice.ArticleHandler
	CommentHandler           *contentservice.CommentHandler
	CommentModerationHandler *contentservice.CommentModerationHandler
	MediaReportHandler       *contentservice.MediaReportHandler
	FeedHandler              *contentservice.FeedHandler
	ChannelHandler           *contentservice.ChannelHandler
	PlaylistHandler          *contentservice.PlaylistHandler
	InteractionHandler       *contentservice.InteractionHandler
	NotificationHandler      *contentservice.NotificationHandler
	ShareHandler             *contentservice.ShareHandler
	ExploreHandler           *contentservice.ExploreHandler
	PortalHandler            *contentservice.PortalHandler
	AdminHandler            *adminservice.AdminHandler
	AdminTagHandler          *adminservice.AdminTagHandler
	StubHandler              *contentservice.StubHandler
	SpriteHandler            *contentservice.SpriteHandler
	SystemHandler            *systemservice.SystemHandler
	StatsHandler             *systemservice.StatsHandler
	RateLimiter              *middleware.RateLimiter
	UploadUC                 *mediabiz.UploadUseCase
	CommentLikeUC            *contentbiz.CommentLikeUseCase
	SettingUC                *systembiz.SettingUseCase
}

// Cleanup closes all resources.
func (d *AppDependencies) Cleanup() {
	if d.DB != nil {
		d.DB.Close()
	}
}

// wireApp initializes the application dependencies.
func wireApp(cfg *config.Config, logger log.Logger) (*AppDependencies, func(), error) {
	wire.Build(
		wire.Struct(new(AppDependencies), "*"),
		ProviderSet,
	)
	return nil, nil, nil
}
