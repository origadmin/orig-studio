//go:build wireinject
// +build wireinject

/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	_ "github.com/lib/pq"           // PostgreSQL driver
	_ "github.com/sqlite3ent/sqlite3" // SQLite3 driver
	"github.com/origadmin/runtime/log"
	"github.com/origadmin/toolkits/crypto/hash"
	hashtypes "github.com/origadmin/toolkits/crypto/hash/types"

	"origadmin/application/origcms/internal/auth"
	config "origadmin/application/origcms/internal/conf"
	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/data/entity/migrate"
	"origadmin/application/origcms/internal/pubsub"
	"origadmin/application/origcms/internal/server"
	mediabiz "origadmin/application/origcms/internal/svc-media/biz"
	mediadata "origadmin/application/origcms/internal/svc-media/data"
	"origadmin/application/origcms/internal/svc-user/biz"
	"origadmin/application/origcms/internal/svc-user/data"
	dto "origadmin/application/origcms/internal/svc-user/dto"

	contentbiz "origadmin/application/origcms/internal/svc-content/biz"
	contentdata "origadmin/application/origcms/internal/svc-content/data"
	systembiz "origadmin/application/origcms/internal/svc-system/biz"
	systemdata "origadmin/application/origcms/internal/svc-system/data"

	adminbiz "origadmin/application/origcms/internal/svc-admin/biz"
	admindata "origadmin/application/origcms/internal/svc-admin/data"
	adminservice "origadmin/application/origcms/internal/svc-admin/service"

	authbiz "origadmin/application/origcms/internal/svc-auth/biz"
	authdata "origadmin/application/origcms/internal/svc-auth/data"

	atlasmigrate "ariga.io/atlas/sql/migrate"
	"entgo.io/ent/dialect"
	dbschema "entgo.io/ent/dialect/sql/schema"
	"github.com/google/wire"
)

// ProviderSet is the wire provider set for the application.
var ProviderSet = wire.NewSet(
	NewDatabase,
	NewHasher,
	NewJWTManager,
	NewPubSub,
	NewMediaRepo,
	NewEncodeProfileRepo,
	NewEncodingTaskRepo,
	NewReviewLogRepo,
	NewUploadRepo,
	NewStorage,
	NewMediaUseCase,
	NewUploadUseCase,
	NewWorker,
	NewSpriteUseCase,
	NewTranscodeHandler,
	NewRouter,
	NewContentDB,
	NewCategoryRepo,
	NewTagRepo,
	NewArticleRepo,
	NewArticleUseCase,
	NewCommentRepo,
	NewPlaylistRepo,
	NewChannelRepo,
	NewFeedRepo,
	NewLikeRepo,
	NewFavoriteRepo,
	NewNotificationRepo,
	NewCategoryTagUseCase,
	NewCommentUseCase,
	NewCommentModerationRepo,
	NewCommentReportRepo,
	NewCommentModerationUseCase,
	NewCommentModerationHandler,
	NewPlaylistChannelUseCase,
	NewFeedUseCase,
	NewLikeFavoriteUseCase,
	NewNotificationUseCase,
	NewUserRepo,
	NewUserUseCase,
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
	NewStatsRepo,
	NewSettingRepo,
	NewSettingUseCase,
	NewAdminTagRepo,
	NewAdminTagUseCase,
	NewAdminTagService,
	NewAdminHandler,
	NewAuthData,
	NewPermissionGroupRepo,
	NewGroupMemberRepo,
	NewUserPermRepo,
	NewPermissionUseCase,
	NewPermissionHandler,
)

// NewDatabase creates a new database client.
func NewDatabase(cfg *config.Config, logger log.Logger) (*entity.Client, error) {
	dbDialect, dbSource := cfg.GetDefaultDB()
	db, err := openDB(dbSource, dbDialect, logger)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	// Build migration options based on database dialect
	migrateOpts := []dbschema.MigrateOption{
		migrate.WithDropIndex(true),
		migrate.WithDropColumn(true),
	}

	if dbDialect == "postgres" {
		// PostgreSQL: disable foreign keys during migration to avoid ordering issues
		migrateOpts = append(migrateOpts, migrate.WithForeignKeys(false))
		// Fix: PostgreSQL auto-migration may fail with "relation already exists" when
		// indexes already exist in the database but Atlas's InspectSchema does not
		// detect them (e.g., due to custom table names via entsql.Table annotation).
		// Use WithApplyHook to inject IF NOT EXISTS into CREATE INDEX statements,
		// making the migration idempotent on PostgreSQL.
		migrateOpts = append(migrateOpts, dbschema.WithApplyHook(func(next dbschema.Applier) dbschema.Applier {
			return dbschema.ApplyFunc(func(ctx context.Context, conn dialect.ExecQuerier, plan *atlasmigrate.Plan) error {
				for i, c := range plan.Changes {
					if strings.HasPrefix(c.Cmd, "CREATE INDEX ") || strings.HasPrefix(c.Cmd, "CREATE UNIQUE INDEX ") {
						// Insert "IF NOT EXISTS" after "CREATE INDEX" or "CREATE UNIQUE INDEX"
						c.Cmd = strings.Replace(c.Cmd, "CREATE INDEX ", "CREATE INDEX IF NOT EXISTS ", 1)
						c.Cmd = strings.Replace(c.Cmd, "CREATE UNIQUE INDEX ", "CREATE UNIQUE INDEX IF NOT EXISTS ", 1)
						plan.Changes[i] = c
					}
				}
				return next.Apply(ctx, conn, plan)
			})
		}))
	}

	if err := db.Schema.Create(ctx, migrateOpts...); err != nil {
		return nil, err
	}

	if err := mediadata.SeedEncodeProfiles(ctx, db); err != nil {
		return nil, err
	}

	if err := seedSettings(ctx, db); err != nil {
		return nil, err
	}

	return db, nil
}

// NewHasher creates a new hasher.
func NewHasher() (hash.Crypto, error) {
	return hash.NewCrypto(hashtypes.BCRYPT)
}

// NewJWTManager creates a new JWT manager.
func NewJWTManager(cfg *config.Config) *auth.Manager {
	jwtSigningKey, _, jwtTTL, refreshTokenTTL := cfg.GetJWTConfig()
	jwtExpire := config.ParseDuration(jwtTTL, 3600*time.Second)
	refreshTokenExpire := config.ParseDuration(refreshTokenTTL, 720*time.Hour)
	return auth.NewManager(
		jwtSigningKey,
		jwtExpire,
		refreshTokenExpire,
	)
}

// NewPubSub creates a new pubsub instance.
func NewPubSub(logger log.Logger) *pubsub.PubSub {
	wmLogger := watermill.NewStdLogger(true, true)
	return pubsub.NewGoChannel(64, wmLogger)
}

// NewPublisher creates a new message publisher.
func NewPublisher(ps *pubsub.PubSub) message.Publisher {
	return ps.Pub
}

// NewMediaRepo creates a new media repo.
func NewMediaRepo(db *entity.Client) mediabiz.MediaRepo {
	return mediadata.NewMediaRepo(db)
}

// NewEncodeProfileRepo creates a new encode profile repo.
func NewEncodeProfileRepo(db *entity.Client) mediabiz.EncodeProfileRepo {
	return mediadata.NewEncodeProfileRepo(db)
}

// NewEncodingTaskRepo creates a new encoding task repo.
func NewEncodingTaskRepo(db *entity.Client) mediabiz.EncodingTaskRepo {
	return mediadata.NewEncodingTaskRepo(db)
}

// NewUploadRepo creates a new upload repo.
func NewUploadRepo(db *entity.Client, logger log.Logger) mediabiz.UploadRepo {
	return mediadata.NewUploadRepo(db, logger)
}

// NewStorage creates a new storage.
func NewStorage() mediabiz.Storage {
	return mediadata.NewLocalStorage("./data/uploads")
}

// NewReviewLogRepo creates a new review log repo.
func NewReviewLogRepo(db *entity.Client) mediabiz.ReviewLogRepo {
	return mediadata.NewReviewLogRepo(db)
}

// NewMediaUseCase creates a new media use case.
func NewMediaUseCase(
	mediaRepo mediabiz.MediaRepo,
	profileRepo mediabiz.EncodeProfileRepo,
	taskRepo mediabiz.EncodingTaskRepo,
	reviewLogRepo mediabiz.ReviewLogRepo,
	storage mediabiz.Storage,
	publisher message.Publisher,
	logger log.Logger,
	spriteUC *mediabiz.SpriteUseCase,
) *mediabiz.MediaUseCase {
	return mediabiz.NewMediaUseCase(mediaRepo, profileRepo, taskRepo, reviewLogRepo, storage, publisher, logger, spriteUC)
}

// NewUploadUseCase creates a new upload use case.
func NewUploadUseCase(
	uploadRepo mediabiz.UploadRepo,
	mediaRepo mediabiz.MediaRepo,
	profileRepo mediabiz.EncodeProfileRepo,
	taskRepo mediabiz.EncodingTaskRepo,
	mediaUC *mediabiz.MediaUseCase,
	storage mediabiz.Storage,
	logger log.Logger,
) *mediabiz.UploadUseCase {
	uploadUC := mediabiz.NewUploadUseCase(
		uploadRepo,
		mediaRepo,
		profileRepo,
		taskRepo,
		mediaUC,
		storage,
		5*1024*1024, // 5MB chunk size
		logger,
	)
	return uploadUC
}

// NewWorker creates a new worker.
func NewWorker(logger log.Logger) mediabiz.TranscodeWorker {
	maxWorkers := int32(envInt("TRANSCODE_MAX_WORKERS", 3))
	return mediabiz.NewGoroutineWorker(maxWorkers, log.NewHelper(log.With(logger, "module", "transcode.worker")))
}

func NewSpriteUseCase(
	mediaRepo mediabiz.MediaRepo,
	settingUC *systembiz.SettingUseCase,
	logger log.Logger,
) *mediabiz.SpriteUseCase {
	return mediabiz.NewSpriteUseCase(mediaRepo, settingUC, "./data/uploads", logger)
}

// NewTranscodeHandler creates a new transcode handler.
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

// NewRouter creates a new Watermill router.
func NewRouter(
	transcodeHandler *mediabiz.TranscodeHandler,
	ps *pubsub.PubSub,
	logger log.Logger,
) (*message.Router, error) {
	wmLogger := watermill.NewStdLogger(true, true)
	router, err := message.NewRouter(message.RouterConfig{}, wmLogger)
	if err != nil {
		return nil, err
	}
	router.AddHandler(
		"media_transcode",
		pubsub.MediaEncodeRequestTopic,
		ps.Sub,
		"",  // no output topic (handler publishes directly)
		nil, // no output publisher needed
		func(msg *message.Message) ([]*message.Message, error) {
			return nil, transcodeHandler.Handle(msg)
		},
	)
	return router, nil
}

// NewContentDB creates a new content DB.
func NewContentDB(db *entity.Client) *contentdata.Data {
	return contentdata.NewData(db)
}

// NewCategoryRepo creates a new category repo.
func NewCategoryRepo(contentDB *contentdata.Data, logger log.Logger) contentbiz.CategoryRepo {
	return contentdata.NewCategoryRepo(contentDB, logger)
}

// NewTagRepo creates a new tag repo.
func NewTagRepo(contentDB *contentdata.Data, logger log.Logger) contentbiz.TagRepo {
	return contentdata.NewTagRepo(contentDB, logger)
}

// NewArticleRepo creates a new article repo.
func NewArticleRepo(contentDB *contentdata.Data, logger log.Logger) contentbiz.ArticleRepo {
	return contentdata.NewArticleRepo(contentDB, logger)
}

// NewArticleUseCase creates a new article use case.
func NewArticleUseCase(articleRepo contentbiz.ArticleRepo, logger log.Logger) *contentbiz.ArticleUseCase {
	return contentbiz.NewArticleUseCase(articleRepo, logger)
}

// NewCommentRepo creates a new comment repo.
func NewCommentRepo(contentDB *contentdata.Data, logger log.Logger) contentbiz.CommentRepo {
	return contentdata.NewCommentRepo(contentDB, logger)
}

// NewPlaylistRepo creates a new playlist repo.
func NewPlaylistRepo(contentDB *contentdata.Data, logger log.Logger) contentbiz.PlaylistRepo {
	return contentdata.NewPlaylistRepo(contentDB, logger)
}

// NewChannelRepo creates a new channel repo.
func NewChannelRepo(contentDB *contentdata.Data, logger log.Logger) contentbiz.ChannelRepo {
	return contentdata.NewChannelRepo(contentDB, logger)
}

// NewFeedRepo creates a new feed repo.
func NewFeedRepo(contentDB *contentdata.Data, logger log.Logger) contentbiz.FeedRepo {
	return contentdata.NewFeedRepo(contentDB, logger)
}

// NewLikeRepo creates a new like repo.
func NewLikeRepo(contentDB *contentdata.Data, logger log.Logger) contentbiz.LikeRepo {
	return contentdata.NewLikeRepo(contentDB, logger)
}

// NewFavoriteRepo creates a new favorite repo.
func NewFavoriteRepo(contentDB *contentdata.Data, logger log.Logger) contentbiz.FavoriteRepo {
	return contentdata.NewFavoriteRepo(contentDB, logger)
}

// NewNotificationRepo creates a new notification repo.
func NewNotificationRepo(contentDB *contentdata.Data, logger log.Logger) contentbiz.NotificationRepo {
	return contentdata.NewNotificationRepo(contentDB, logger)
}

// NewCategoryTagUseCase creates a new category tag use case.
func NewCategoryTagUseCase(
	categoryRepo contentbiz.CategoryRepo,
	tagRepo contentbiz.TagRepo,
	logger log.Logger,
) *contentbiz.CategoryTagUseCase {
	return contentbiz.NewCategoryTagUseCase(categoryRepo, tagRepo, logger)
}

// NewCommentUseCase creates a new comment use case.
func NewCommentUseCase(
	commentRepo contentbiz.CommentRepo,
	mediaUC *mediabiz.MediaUseCase,
	logger log.Logger,
) *contentbiz.CommentUseCase {
	return contentbiz.NewCommentUseCase(commentRepo, mediaUC, logger)
}

// NewCommentModerationRepo creates a new comment moderation repo.
func NewCommentModerationRepo(contentDB *contentdata.Data, logger log.Logger) contentbiz.CommentModerationRepo {
	return contentdata.NewCommentModerationRepo(contentDB, logger)
}

// NewCommentReportRepo creates a new comment report repo.
func NewCommentReportRepo(contentDB *contentdata.Data, logger log.Logger) contentbiz.CommentReportRepo {
	return contentdata.NewCommentReportRepo(contentDB, logger)
}

// NewCommentModerationUseCase creates a new comment moderation use case.
func NewCommentModerationUseCase(
	commentModerationRepo contentbiz.CommentModerationRepo,
	commentReportRepo contentbiz.CommentReportRepo,
	settingUC *systembiz.SettingUseCase,
	logger log.Logger,
) *contentbiz.CommentModerationUseCase {
	return contentbiz.NewCommentModerationUseCase(commentModerationRepo, commentReportRepo, settingUC, logger)
}

// NewCommentModerationHandler creates a new comment moderation handler.
func NewCommentModerationHandler(
	moderationUC *contentbiz.CommentModerationUseCase,
	db *entity.Client,
	jwt *auth.Manager,
) *server.CommentModerationHandler {
	return server.NewCommentModerationHandler(moderationUC, db, jwt)
}

// NewPlaylistChannelUseCase creates a new playlist channel use case.
func NewPlaylistChannelUseCase(
	playlistRepo contentbiz.PlaylistRepo,
	channelRepo contentbiz.ChannelRepo,
	logger log.Logger,
) *contentbiz.PlaylistChannelUseCase {
	return contentbiz.NewPlaylistChannelUseCase(playlistRepo, channelRepo, logger)
}

// NewFeedUseCase creates a new feed use case.
func NewFeedUseCase(
	feedRepo contentbiz.FeedRepo,
	logger log.Logger,
) *contentbiz.FeedUseCase {
	return contentbiz.NewFeedUseCase(feedRepo, logger)
}

// NewLikeFavoriteUseCase creates a new like favorite use case.
func NewLikeFavoriteUseCase(
	likeRepo contentbiz.LikeRepo,
	favoriteRepo contentbiz.FavoriteRepo,
	mediaUC *mediabiz.MediaUseCase,
	logger log.Logger,
) *contentbiz.LikeFavoriteUseCase {
	return contentbiz.NewLikeFavoriteUseCase(likeRepo, favoriteRepo, mediaUC, logger)
}

// NewNotificationUseCase creates a new notification use case.
func NewNotificationUseCase(
	notificationRepo contentbiz.NotificationRepo,
	logger log.Logger,
) *contentbiz.NotificationUseCase {
	return contentbiz.NewNotificationUseCase(notificationRepo, logger)
}

// NewUserRepo creates a new user repo.
func NewUserRepo(db *entity.Client) dto.UserRepo {
	return data.NewUserRepo(db)
}

// NewUserUseCase creates a new user use case.
func NewUserUseCase(
	userRepo dto.UserRepo,
	hasher hash.Crypto,
	logger log.Logger,
) *biz.UserUseCase {
	return biz.NewUserUseCase(userRepo, hasher, logger)
}

// NewAuthHandler creates a new auth handler.
func NewAuthHandler(
	userUC *biz.UserUseCase,
	jwt *auth.Manager,
) *server.AuthHandler {
	return server.NewAuthHandler(userUC, jwt)
}

// NewUserHandler creates a new user handler.
func NewUserHandler(
	userUC *biz.UserUseCase,
	jwt *auth.Manager,
) *server.UserHandler {
	return server.NewUserHandler(userUC, jwt)
}

// NewMediaHandler creates a new media handler.
func NewMediaHandler(
	jwt *auth.Manager,
	mediaUC *mediabiz.MediaUseCase,
	uploadUC *mediabiz.UploadUseCase,
	likeFavoriteUC *contentbiz.LikeFavoriteUseCase,
	playlistChannelUC *contentbiz.PlaylistChannelUseCase,
	userUC *biz.UserUseCase,
	entityClient *entity.Client,
) *server.MediaHandler {
	return server.NewMediaHandler(jwt, mediaUC, uploadUC, likeFavoriteUC, playlistChannelUC, userUC, entityClient)
}

// NewUploadHandler creates a new upload handler.
func NewUploadHandler(
	uploadUC *mediabiz.UploadUseCase,
	jwt *auth.Manager,
	logger log.Logger,
) *server.UploadHandler {
	return server.NewUploadHandler(uploadUC, jwt, logger)
}

// NewCategoryHandler creates a new category handler.
func NewCategoryHandler(
	categoryTagUC *contentbiz.CategoryTagUseCase,
	jwt *auth.Manager,
) *server.CategoryHandler {
	return server.NewCategoryHandler(categoryTagUC, jwt)
}

// NewTagHandler creates a new tag handler.
func NewTagHandler(
	categoryTagUC *contentbiz.CategoryTagUseCase,
	jwt *auth.Manager,
) *server.TagHandler {
	return server.NewTagHandler(categoryTagUC, jwt)
}

// NewFeedHandler creates a new feed handler.
func NewFeedHandler(
	feedUC *contentbiz.FeedUseCase,
) *server.FeedHandler {
	return server.NewFeedHandler(feedUC)
}

// NewNotificationHandler creates a new notification handler.
func NewNotificationHandler(
	notificationUC *contentbiz.NotificationUseCase,
	jwt *auth.Manager,
) *server.NotificationHandler {
	return server.NewNotificationHandler(notificationUC, jwt)
}

// NewChannelHandler creates a new channel handler.
func NewChannelHandler(
	playlistChannelUC *contentbiz.PlaylistChannelUseCase,
	jwt *auth.Manager,
	entityClient *entity.Client,
) *server.ChannelHandler {
	return server.NewChannelHandler(playlistChannelUC, jwt, entityClient)
}

// NewShareHandler creates a new share handler.
func NewShareHandler(
	likeFavoriteUC *contentbiz.LikeFavoriteUseCase,
	jwt *auth.Manager,
) *server.ShareHandler {
	return server.NewShareHandler(likeFavoriteUC, jwt)
}

// NewSystemHandler creates a new system handler.
func NewSystemHandler(
	jwt *auth.Manager,
	statsRepo *systemdata.StatsRepo,
	settingUC *systembiz.SettingUseCase,
) *server.SystemHandler {
	return server.NewSystemHandler(jwt, statsRepo, settingUC)
}

// NewStatsHandler creates a new stats handler.
func NewStatsHandler(
	mediaUC *mediabiz.MediaUseCase,
	likeFavoriteUC *contentbiz.LikeFavoriteUseCase,
	statsRepo *systemdata.StatsRepo,
	jwt *auth.Manager,
) *server.StatsHandler {
	return server.NewStatsHandler(mediaUC, likeFavoriteUC, statsRepo, jwt)
}

// NewSearchHandler creates a new search handler.
func NewSearchHandler(
	mediaUC *mediabiz.MediaUseCase,
) *server.SearchHandler {
	return server.NewSearchHandler(mediaUC)
}

// NewMeHandler creates a new me handler.
func NewMeHandler(
	userUC *biz.UserUseCase,
	likeFavoriteUC *contentbiz.LikeFavoriteUseCase,
	playlistChannelUC *contentbiz.PlaylistChannelUseCase,
	jwt *auth.Manager,
) *server.MeHandler {
	return server.NewMeHandler(userUC, likeFavoriteUC, playlistChannelUC, jwt)
}

// NewStatsRepo creates a new stats repo.
func NewStatsRepo(db *entity.Client) *systemdata.StatsRepo {
	return systemdata.NewStatsRepo(db)
}

func NewSettingRepo(db *entity.Client) *systemdata.SettingRepo {
	return systemdata.NewSettingRepo(db)
}

func NewSettingUseCase(settingRepo *systemdata.SettingRepo) *systembiz.SettingUseCase {
	return systembiz.NewSettingUseCase(settingRepo)
}

func seedSettings(ctx context.Context, db *entity.Client) error {
	repo := systemdata.NewSettingRepo(db)
	uc := systembiz.NewSettingUseCase(repo)
	return uc.SeedDefaults(ctx)
}

// NewAdminTagRepo creates a new admin tag repo.
func NewAdminTagRepo(db *entity.Client) admindata.TagRepository {
	return admindata.NewTagRepository(db)
}

// NewAdminTagUseCase creates a new admin tag use case.
func NewAdminTagUseCase(tagRepo admindata.TagRepository) *adminbiz.TagUseCase {
	return adminbiz.NewTagUseCase(tagRepo)
}

// NewAdminTagService creates a new admin tag service.
func NewAdminTagService(tagUC *adminbiz.TagUseCase) *adminservice.TagService {
	return adminservice.NewTagService(tagUC)
}

// NewAdminHandler creates a new admin handler.
func NewAdminHandler(
	jwt *auth.Manager,
	mediaUC *mediabiz.MediaUseCase,
	channelUC *contentbiz.PlaylistChannelUseCase,
	tagService *adminservice.TagService,
	settingUC *systembiz.SettingUseCase,
	categoryUC *contentbiz.CategoryTagUseCase,
	articleUC *contentbiz.ArticleUseCase,
	userUC *biz.UserUseCase,
) *server.AdminHandler {
	return server.NewAdminHandler(jwt, mediaUC, channelUC, tagService, settingUC, categoryUC, articleUC, userUC)
}

// NewAuthData creates a new auth data layer.
func NewAuthData(db *entity.Client) *authdata.Data {
	return authdata.NewData(db)
}

// NewPermissionGroupRepo creates a new permission group repo.
func NewPermissionGroupRepo(data *authdata.Data, logger log.Logger) authbiz.PermissionGroupRepo {
	return authdata.NewPermissionGroupRepo(data, logger)
}

// NewGroupMemberRepo creates a new group member repo.
func NewGroupMemberRepo(data *authdata.Data, logger log.Logger) authbiz.GroupMemberRepo {
	return authdata.NewGroupMemberRepo(data, logger)
}

// NewUserPermRepo creates a new user perm repo.
func NewUserPermRepo(data *authdata.Data, logger log.Logger) authbiz.UserPermRepo {
	return authdata.NewUserPermRepo(data, logger)
}

// NewPermissionUseCase creates a new permission use case.
func NewPermissionUseCase(
	groupRepo authbiz.PermissionGroupRepo,
	memberRepo authbiz.GroupMemberRepo,
	userRepo authbiz.UserPermRepo,
	logger log.Logger,
) *authbiz.PermissionUseCase {
	return authbiz.NewPermissionUseCase(groupRepo, memberRepo, userRepo, logger)
}

// NewPermissionHandler creates a new permission handler.
func NewPermissionHandler(
	permUC *authbiz.PermissionUseCase,
	jwt *auth.Manager,
) *server.PermissionHandler {
	return server.NewPermissionHandler(permUC, jwt)
}

// AppDependencies holds all application dependencies.
type AppDependencies struct {
	DB                       *entity.Client
	PubSub                   *pubsub.PubSub
	Router                   *message.Router
	JWTManager               *auth.Manager
	AuthHandler              *server.AuthHandler
	UserHandler              *server.UserHandler
	MediaHandler             *server.MediaHandler
	UploadHandler            *server.UploadHandler
	CategoryHandler          *server.CategoryHandler
	TagHandler               *server.TagHandler
	FeedHandler              *server.FeedHandler
	NotificationHandler      *server.NotificationHandler
	ChannelHandler           *server.ChannelHandler
	ShareHandler             *server.ShareHandler
	SystemHandler            *server.SystemHandler
	StatsHandler             *server.StatsHandler
	SearchHandler            *server.SearchHandler
	MeHandler                *server.MeHandler
	AdminHandler             *server.AdminHandler
	UploadUC                 *mediabiz.UploadUseCase
	CommentLikeUC            *contentbiz.CommentLikeUseCase
	CommentModerationHandler *server.CommentModerationHandler
	PermissionHandler        *server.PermissionHandler
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
		NewDatabase,
		NewHasher,
		NewJWTManager,
		NewPubSub,
		NewPublisher,
		NewMediaRepo,
		NewEncodeProfileRepo,
		NewEncodingTaskRepo,
		NewReviewLogRepo,
		NewUploadRepo,
		NewStorage,
		NewMediaUseCase,
		NewUploadUseCase,
		NewWorker,
		NewTranscodeHandler,
		NewRouter,
		NewContentDB,
		NewCategoryRepo,
		NewTagRepo,
		NewPlaylistRepo,
		NewChannelRepo,
		NewFeedRepo,
		NewLikeRepo,
		NewFavoriteRepo,
		NewNotificationRepo,
		NewCategoryTagUseCase,
		NewCommentModerationRepo,
		NewCommentReportRepo,
		NewCommentModerationUseCase,
		NewCommentModerationHandler,
		NewPlaylistChannelUseCase,
		NewFeedUseCase,
		NewLikeFavoriteUseCase,
		NewNotificationUseCase,
		NewUserRepo,
		NewUserUseCase,
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
		NewStatsRepo,
		NewSettingRepo,
		NewSettingUseCase,
		NewAdminTagRepo,
		NewAdminTagUseCase,
		NewAdminTagService,
		NewAdminHandler,
		NewAuthData,
		NewPermissionGroupRepo,
		NewGroupMemberRepo,
		NewUserPermRepo,
		NewPermissionUseCase,
		NewPermissionHandler,
	)
	return nil, nil
}

// Helper functions

func openDB(dsn, dbType string, logger log.Logger) (*entity.Client, error) {
	driverName := "sqlite3"
	if dbType == "postgres" {
		driverName = "postgres"
		// Ensure database exists before connecting
		if err := ensurePostgresDB(dsn, logger); err != nil {
			return nil, err
		}
		// Add sslmode if not present
		if !strings.Contains(dsn, "sslmode") {
			if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
				// URI format: append as query param
				if strings.Contains(dsn, "?") {
					dsn = dsn + "&sslmode=disable"
				} else {
					dsn = dsn + "?sslmode=disable"
				}
			} else {
				// key=value format: append as param
				dsn = dsn + " sslmode=disable"
			}
		}
	} else {
		// SQLite3: ensure the parent directory for the database file exists
		if err := ensureSQLiteDir(dsn); err != nil {
			return nil, fmt.Errorf("failed to create sqlite data directory: %w", err)
		}
		// Enable foreign keys pragma if not already set
		if !strings.Contains(dsn, "_fk=") {
			if strings.Contains(dsn, "?") {
				dsn = dsn + "&_fk=1"
			} else {
				dsn = dsn + "?_fk=1"
			}
		}
	}
	return entity.Open(driverName, dsn)
}

// ensureSQLiteDir ensures the parent directory for the SQLite database file exists.
func ensureSQLiteDir(dsn string) error {
	// Extract file path from DSN (remove query parameters if present)
	dbPath := dsn
	if idx := strings.Index(dsn, "?"); idx >= 0 {
		dbPath = dsn[:idx]
	}
	dir := filepath.Dir(dbPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}

func ensurePostgresDB(dsn string, logger log.Logger) error {
	// Parse DSN to extract connection info
	_, dbName := parsePostgresDSN(dsn)
	if dbName == "" {
		return nil
	}

	// Build a DSN pointing to the default 'postgres' database
	var defaultDSN string
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		// URI format: replace the database name in the path
		defaultDSN = replaceDBNameInURI(dsn, "postgres")
	} else {
		// key=value format: append/override dbname
		defaultDSN = dsn + " dbname=postgres sslmode=disable"
	}

	db, err := sql.Open("postgres", defaultDSN)
	if err != nil {
		return err
	}
	defer db.Close()

	// Check if database exists
	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", dbName).
		Scan(&exists)
	if err != nil {
		return err
	}

	if !exists {
		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
		if err != nil {
			return fmt.Errorf("create database %s: %w", dbName, err)
		}
		log.Infof("Created database: %s", dbName)
	}
	return nil
}

func replaceDBNameInURI(dsn, newDBName string) string {
	scheme := "postgres://"
	if strings.HasPrefix(dsn, "postgresql://") {
		scheme = "postgresql://"
	}
	rest := dsn[len(scheme):]

	slashIdx := strings.Index(rest, "/")
	if slashIdx < 0 {
		// No path, append /newDBName
		return dsn + "/" + newDBName
	}

	authority := rest[:slashIdx]
	remainder := rest[slashIdx+1:]

	// Separate path from query
	qIdx := strings.Index(remainder, "?")
	var query string
	if qIdx >= 0 {
		query = "?" + remainder[qIdx+1:]
		remainder = remainder[:qIdx]
	}

	return scheme + authority + "/" + newDBName + query
}

func parsePostgresDSN(dsn string) (connStr, dbName string) {
	// URI format: postgres://user:pass@host/db?opts
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		return parsePostgresURIDSN(dsn)
	}
	// key=value format: host=x dbname=y
	return parsePostgresKVDSN(dsn)
}

func parsePostgresURIDSN(dsn string) (connStr, dbName string) {
	// Remove scheme
	rest := dsn
	if idx := strings.Index(rest, "://"); idx >= 0 {
		rest = rest[idx+3:]
	}

	// Split authority and path
	var _, pathPart, queryPart string
	if slashIdx := strings.Index(rest, "/"); slashIdx >= 0 {
		remainder := rest[slashIdx+1:]
		if qIdx := strings.Index(remainder, "?"); qIdx >= 0 {
			pathPart = remainder[:qIdx]
			queryPart = remainder[qIdx+1:]
		} else {
			pathPart = remainder
		}
	}

	dbName = pathPart

	// Rebuild connection string pointing to 'postgres' default database
	connStr = dsn
	// Replace dbname in URI
	if dbName != "" {
		if queryPart != "" {
			connStr = strings.Replace(connStr, "/"+dbName+"?", "/postgres?", 1)
		} else {
			connStr = strings.Replace(connStr, "/"+dbName, "/postgres", 1)
		}
	}
	return connStr, dbName
}

func parsePostgresKVDSN(dsn string) (connStr, dbName string) {
	// Find dbname
	if i := strings.Index(dsn, "dbname="); i >= 0 {
		start := i + 7
		end := strings.IndexAny(dsn[start:], " ")
		if end < 0 {
			dbName = dsn[start:]
		} else {
			dbName = dsn[start : start+end]
		}
	}

	// Extract connection params for default DB (remove dbname)
	connParts := []string{}
	for _, part := range strings.Split(dsn, " ") {
		if strings.HasPrefix(part, "dbname=") {
			continue
		}
		connParts = append(connParts, part)
	}
	connStr = strings.Join(connParts, " ")
	return connStr, dbName
}

func envInt(key string, defaultVal int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultVal
	}
	return n
}
