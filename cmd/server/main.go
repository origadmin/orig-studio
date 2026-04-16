/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// main is the M1 monolith entry point for origcms.
// Uses runtime for config loading and logger initialization.
// Wires up ent (SQLite/PostgreSQL), svc-user biz/data, and a Gin HTTP server.
// Run: go run ./cmd/server -conf configs/bootstrap.yaml
package main

import (
	"context"
	"flag"

	"github.com/origadmin/runtime"
	runtimebootstrap "github.com/origadmin/runtime/engine/bootstrap"
	"github.com/origadmin/runtime/log"
	"origadmin/application/origcms/internal/conf"
	confhelper "origadmin/application/origcms/internal/helpers/conf"
	"origadmin/application/origcms/internal/server"
)

var (
	// Name is the name of the compiled software.
	Name = "origcms.server"
	// Version is the version of the compiled software.
	Version = "v1.0.0"
	// envName is the environment name suffix for .env file
	envName = ".server"
	// flagconf is the config flag
	flagconf string
)

func init() {
	flag.StringVar(&flagconf, "conf", "", "config path, eg: -conf bootstrap.yaml")
}

func main() {
	flag.Parse()

	// Initialize environment variables and find configuration file
	confPath := confhelper.InitEnvAndConf(envName, flagconf)
	if confPath == "" {
		log.Fatalf(
			"Could not find configuration file. Searched -conf flag, executable path, and development path.",
		)
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

	// Debug: print actual database config
	dialect, source := cfg.GetDefaultDB()
	log.Infof("Database config: dialect=%s, source=%s", dialect, source)

	logger := rt.Logger()
	log.SetLogger(logger)

	// Initialize dependencies
	deps, err := wireApp(cfg, logger)
	if err != nil {
		log.Fatalf("failed to initialize dependencies: %v", err)
	}
	defer deps.Cleanup()

	// Start Watermill router
	go func() {
		if err := deps.Router.Run(context.Background()); err != nil {
			log.Fatalf("Watermill router error: %v", err)
		}
	}()

	// Inject publisher into UploadUseCase for async encoding requests
	deps.UploadUC.SetPublisher(deps.PubSub.Pub)

	// Create server
	srv := server.NewServer(
		deps.AuthHandler,
		deps.UserHandler,
		deps.MediaHandler,
		deps.UploadHandler,
		deps.CategoryHandler,
		deps.TagHandler,
		deps.FeedHandler,
		deps.NotificationHandler,
		deps.ChannelHandler,
		deps.ShareHandler,
		deps.SystemHandler,
		deps.StatsHandler,
		deps.SearchHandler,
		deps.MeHandler,
		deps.AdminHandler,
	)

	// Start server
	addr := cfg.Server.HTTP.Addr
	if addr == "" {
		addr = ":9090"
	}
	log.Infof("origcms server starting, addr: %s", addr)
	if err := srv.Start(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
