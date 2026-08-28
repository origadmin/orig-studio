package main

import (
	"context"
	"flag"

	"origadmin/application/origstudio/internal/conf"
	confhelper "origadmin/application/origstudio/internal/infra/bootstrap"
	"origadmin/application/origstudio/internal/server"

	"github.com/origadmin/runtime"
	runtimebootstrap "github.com/origadmin/runtime/engine/bootstrap"
	"github.com/origadmin/runtime/log"
)

var deps *AppDependencies

var (
	Name    = "origcms.server"
	Version = "v1.1.2"
	envName = ".server"
	flagconf string
)

func init() {
	flag.StringVar(&flagconf, "conf", "", "config path, eg: -conf bootstrap.yaml")
}

func main() {
	flag.Parse()

	confPath := confhelper.InitEnvAndConf(envName, flagconf)
	if confPath == "" {
		log.Fatalf("Could not find configuration file. Searched -conf flag, executable path, and development path.")
	}

	log.Infof("Loading configuration from: %s\n", confPath)

	rt := runtime.New(Name, Version)
	if err := rt.Load(confPath, runtimebootstrap.WithConfigTransformer(conf.Transformer), runtimebootstrap.WithDirectly(true)); err != nil {
		log.Fatalf("failed to create runtime: %v", err)
	}
	defer func() {
		_ = rt.Decoder().Close()
	}()
	rt.ShowAppInfo()

	cfg, ok := rt.Config().(*conf.Config)
	if !ok {
		log.Fatalf("failed to get config")
	}

	dialect, source := cfg.GetDefaultDB()
	log.Infof("Database config: dialect=%s, source=%s", dialect, source)

	logger := rt.Logger()
	log.SetLogger(logger)

	var err error
	deps, cleanup, err := wireApp(cfg, logger)
	if err != nil {
		log.Fatalf("failed to initialize dependencies: %v", err)
	}
	defer cleanup()
	defer deps.Cleanup()

	go func() {
		if err := deps.Router.Run(context.Background()); err != nil {
			log.Fatalf("Watermill router error: %v", err)
		}
	}()

	deps.UploadUC.SetPublisher(deps.PubSub.Pub)

	srv := server.NewServer(
		[]server.Module{
			deps.AuthHandler,
			deps.UserHandler,
			deps.MeHandler,
			deps.MediaHandler,
			deps.UploadHandler,
			deps.SearchHandler,
			deps.CategoryHandler,
			deps.TagHandler,
			deps.ArticleHandler,
			deps.CommentHandler,
			deps.CommentModerationHandler,
			deps.MediaReportHandler,
			deps.FeedHandler,
			deps.ChannelHandler,
			deps.PlaylistHandler,
			deps.InteractionHandler,
			deps.NotificationHandler,
			deps.ShareHandler,
			deps.ExploreHandler,
			deps.PortalHandler,
			deps.AdHandler,
			deps.AdminHandler,
			deps.AdminTagHandler,
			deps.StubHandler,
			deps.SpriteHandler,
			deps.SubtitleHandler,
			deps.SystemHandler,
			deps.FeatureFlagHandler,
		},
		deps.DB,
		deps.JWTManager,
		deps.StoragePaths,
	)

	srv.SetSettingProvider(deps.SettingUC)
	srv.SetSQLDB(deps.SQLDB)
	srv.SetRateLimiter(deps.RateLimiter)

	addr := cfg.Server.HTTP.Addr
	if addr == "" {
		addr = ":8080"
	}
	log.Infof("origcms server starting, addr: %s", addr)
	if err := srv.Start(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
