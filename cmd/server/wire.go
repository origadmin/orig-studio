//go:build wireinject
// +build wireinject

/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package main

import (
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	_ "github.com/lib/pq" // PostgreSQL driver

	"github.com/origadmin/runtime/log"
	_ "github.com/sqlite3ent/sqlite3" // SQLite3 driver

	config "origadmin/application/origcms/internal/conf"
	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/features/admin"
	adminservice "origadmin/application/origcms/internal/features/admin/service"
	featureauth "origadmin/application/origcms/internal/features/auth"
	authbiz "origadmin/application/origcms/internal/features/auth/biz"
	authservice "origadmin/application/origcms/internal/features/auth/service"
	"origadmin/application/origcms/internal/features/content"
	contentbiz "origadmin/application/origcms/internal/features/content/biz"
	contentservice "origadmin/application/origcms/internal/features/content/service"
	"origadmin/application/origcms/internal/features/media"
	mediabiz "origadmin/application/origcms/internal/features/media/biz"
	mediadal "origadmin/application/origcms/internal/features/media/dal"
	mediaservice "origadmin/application/origcms/internal/features/media/service"
	"origadmin/application/origcms/internal/features/system"
	systembiz "origadmin/application/origcms/internal/features/system/biz"
	systemservice "origadmin/application/origcms/internal/features/system/service"
	"origadmin/application/origcms/internal/features/user"
	userbiz "origadmin/application/origcms/internal/features/user/biz"
	userservice "origadmin/application/origcms/internal/features/user/service"
	"origadmin/application/origcms/internal/infra"
	infraauth "origadmin/application/origcms/internal/infra/auth"
	"origadmin/application/origcms/internal/infra/pubsub"

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

	// Bridge functions for media module (hardcoded config values)
	NewStorage,
	NewWorker,
	NewUploadUseCase,
	NewSpriteUseCase,
	NewTranscodeHandler,

	// Bridge functions for handler constructors
	NewAuthHandler,
	NewUserHandler,
	NewMediaHandler,
	NewUploadHandler,
	NewCategoryHandler,
	NewTagHandler,
	NewFeedHandler,
	NewNotificationHandler,
	NewChannelHandler,
	NewShareHandler,
	NewSystemHandler,
	NewStatsHandler,
	NewSearchHandler,
	NewMeHandler,
	NewAdminHandler,
	NewCommentModerationHandler,
	NewPermissionHandler,
	NewArticleHandler,
	NewCommentHandler,
	NewPlaylistHandler,
	NewInteractionHandler,
	NewAdminTagHandler,
	NewExploreHandler,
	NewStubHandler,
	NewSpriteHandler,

	// Wire bindings
	wire.Bind(new(authbiz.PermissionChecker), new(*authbiz.PermissionUseCase)),
	wire.Bind(new(mediabiz.Storage), new(*mediadal.LocalStorage)),
	wire.Bind(new(contentbiz.MediaUseCaseInterface), new(*mediabiz.MediaUseCase)),
	wire.Bind(new(systembiz.ConfigProvider), new(*systembiz.SettingUseCase)),
)

// NewStorage creates a new storage with hardcoded base path.
func NewStorage() *mediadal.LocalStorage {
	return mediadal.NewLocalStorage("./data/uploads")
}

// NewWorker creates a new transcode worker with config from environment.
func NewWorker(logger log.Logger) mediabiz.TranscodeWorker {
	maxWorkers := int32(infra.EnvInt("TRANSCODE_MAX_WORKERS", 3))
	return mediabiz.NewGoroutineWorker(maxWorkers, log.NewHelper(log.With(logger, "module", "transcode.worker")))
}

// NewUploadUseCase creates a new upload use case with hardcoded chunk size.
func NewUploadUseCase(
	uploadRepo mediabiz.UploadRepo,
	mediaRepo mediabiz.MediaRepo,
	profileRepo mediabiz.EncodeProfileRepo,
	taskRepo mediabiz.EncodingTaskRepo,
	mediaUC *mediabiz.MediaUseCase,
	storage mediabiz.Storage,
	logger log.Logger,
) *mediabiz.UploadUseCase {
	return mediabiz.NewUploadUseCase(
		uploadRepo,
		mediaRepo,
		profileRepo,
		taskRepo,
		mediaUC,
		storage,
		5*1024*1024, // 5MB chunk size
		logger,
	)
}

// NewSpriteUseCase creates a new sprite use case with hardcoded base directory.
func NewSpriteUseCase(
	mediaRepo mediabiz.MediaRepo,
	settingUC *systembiz.SettingUseCase,
	logger log.Logger,
) *mediabiz.SpriteUseCase {
	return mediabiz.NewSpriteUseCase(mediaRepo, settingUC, "./data/uploads", logger)
}

// NewTranscodeHandler creates a new transcode handler with hardcoded config values.
func NewTranscodeHandler(
	mediaUC *mediabiz.MediaUseCase,
	profileRepo mediabiz.EncodeProfileRepo,
	taskRepo mediabiz.EncodingTaskRepo,
	mediaRepo mediabiz.MediaRepo,
	worker mediabiz.TranscodeWorker,
	publisher message.Publisher,
	logger log.Logger,
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
		"./data/uploads",
		30*time.Minute,
		spriteUC,
	)
}

// NewArticleHandler creates a new article handler.
func NewArticleHandler(
	uc *contentbiz.ArticleUseCase,
	jwt *infraauth.Manager,
	settingUC *systembiz.SettingUseCase,
) *contentservice.ArticleHandler {
	return contentservice.NewArticleHandler(uc, jwt, settingUC)
}

// NewCommentHandler creates a new comment handler.
func NewCommentHandler(
	client *entity.Client,
	jwt *infraauth.Manager,
	commentLikeUC *contentbiz.CommentLikeUseCase,
	moderationUC *contentbiz.CommentModerationUseCase,
) *contentservice.CommentHandler {
	return contentservice.NewCommentHandler(client, jwt, commentLikeUC, moderationUC)
}

// NewPlaylistHandler creates a new playlist handler.
func NewPlaylistHandler(playlistUC *contentbiz.PlaylistChannelUseCase, settingUC *systembiz.SettingUseCase) *contentservice.PlaylistHandler {
	return contentservice.NewPlaylistHandler(playlistUC, settingUC)
}

// NewInteractionHandler creates a new interaction handler.
func NewInteractionHandler(
	jwt *infraauth.Manager,
	likeFavoriteUC *contentbiz.LikeFavoriteUseCase,
) *contentservice.InteractionHandler {
	return contentservice.NewInteractionHandler(jwt, likeFavoriteUC)
}

// NewAdminTagHandler creates a new admin tag handler.
func NewAdminTagHandler(service *adminservice.TagService) *adminservice.AdminTagHandler {
	return adminservice.NewAdminTagHandler(service)
}

// NewExploreHandler creates a new explore handler.
func NewExploreHandler(client *entity.Client) *contentservice.ExploreHandler {
	return contentservice.NewExploreHandler(client)
}

// NewStubHandler creates a new stub handler for missing routes.
func NewStubHandler(jwt *infraauth.Manager) *contentservice.StubHandler {
	return contentservice.NewStubHandler(jwt)
}

// NewSpriteHandler creates a new sprite handler for sprite sheet and VTT routes.
func NewSpriteHandler(mediaUC *mediabiz.MediaUseCase, jwt *infraauth.Manager, logger log.Logger) *contentservice.SpriteHandler {
	return contentservice.NewSpriteHandler(mediaUC, "./data/uploads", jwt, logger)
}

// NewMeHandler creates a new me handler.
func NewMeHandler(
	userUC *userbiz.UserUseCase,
	likeFavoriteUC *contentbiz.LikeFavoriteUseCase,
	playlistChannelUC *contentbiz.PlaylistChannelUseCase,
	historyUC *contentbiz.HistoryUseCase,
	jwt *infraauth.Manager,
) *userservice.MeHandler {
	return userservice.NewMeHandler(userUC, likeFavoriteUC, playlistChannelUC, historyUC, jwt)
}

// AppDependencies holds all application dependencies.
type AppDependencies struct {
	DB                       *entity.Client
	PubSub                   *pubsub.PubSub
	Router                   *message.Router
	JWTManager               *infraauth.Manager
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
	FeedHandler              *contentservice.FeedHandler
	ChannelHandler           *contentservice.ChannelHandler
	PlaylistHandler          *contentservice.PlaylistHandler
	InteractionHandler       *contentservice.InteractionHandler
	NotificationHandler      *contentservice.NotificationHandler
	ShareHandler             *contentservice.ShareHandler
	ExploreHandler           *contentservice.ExploreHandler
	AdminHandler             *adminservice.AdminHandler
	AdminTagHandler          *adminservice.AdminTagHandler
	StubHandler              *contentservice.StubHandler
	SpriteHandler            *contentservice.SpriteHandler
	SystemHandler            *systemservice.SystemHandler
	StatsHandler             *systemservice.StatsHandler
	UploadUC                 *mediabiz.UploadUseCase
	CommentLikeUC            *contentbiz.CommentLikeUseCase
}

// Cleanup closes all resources.
func (d *AppDependencies) Cleanup() {
	if d.DB != nil {
		d.DB.Close()
	}
}

// wireApp initializes the application dependencies.
func wireApp(cfg *config.Config, logger log.Logger) (*AppDependencies, error) {
	wire.Build(
		wire.Struct(new(AppDependencies), "*"),
		ProviderSet,
	)
	return nil, nil
}
