//go:build wireinject
// +build wireinject

package main

import (
	"database/sql"
	"github.com/ThreeDotsLabs/watermill/message"
	_ "github.com/lib/pq"

	"github.com/origadmin/runtime/log"
	_ "github.com/sqlite3ent/sqlite3"

	config "origadmin/application/origstudio/internal/conf"
	"origadmin/application/origstudio/internal/data/entity"
	adminservice "origadmin/application/origstudio/internal/features/admin/service"
	authservice "origadmin/application/origstudio/internal/features/auth/service"
	contentservice "origadmin/application/origstudio/internal/features/content/service"
	mediabiz "origadmin/application/origstudio/internal/features/media/biz"
	mediaservice "origadmin/application/origstudio/internal/features/media/service"
	systembiz "origadmin/application/origstudio/internal/features/system/biz"
	systemservice "origadmin/application/origstudio/internal/features/system/service"
	userservice "origadmin/application/origstudio/internal/features/user/service"
	infraauth "origadmin/application/origstudio/internal/infra/auth"
	"origadmin/application/origstudio/internal/infra/pubsub"
	"origadmin/application/origstudio/internal/server/middleware"
	"origadmin/application/origstudio/internal/enterprise"

	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	enterprise.CESet,
)

type AppDependencies struct {
	DB                       *entity.Client
	SQLDB                    *sql.DB
	PubSub                   *pubsub.PubSub
	Router                   *message.Router
	JWTManager               *infraauth.Manager
	StoragePaths             *config.StoragePaths
	AuthHandler              *authservice.AuthHandler
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
	PortalHandler           *contentservice.PortalHandler
	AdHandler               *contentservice.AdHandler
	AdminHandler             *adminservice.AdminHandler
	AdminTagHandler          *adminservice.AdminTagHandler
	StubHandler              *contentservice.StubHandler
	SpriteHandler            *mediaservice.SpriteHandler
	SubtitleHandler          *contentservice.SubtitleHandler
	SystemHandler            *systemservice.SystemHandler
	FeatureFlagHandler       *systemservice.FeatureFlagHandler
	RateLimiter              *middleware.RateLimiter
	UploadUC                 *mediabiz.UploadUseCase
	SettingUC                *systembiz.SettingUseCase
}

func (d *AppDependencies) Cleanup() {
	if d.DB != nil {
		d.DB.Close()
	}
}

func wireApp(cfg *config.Config, logger log.Logger) (*AppDependencies, func(), error) {
	wire.Build(
		wire.Struct(new(AppDependencies), "*"),
		ProviderSet,
	)
	return nil, nil, nil
}
