/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package main

import (
	"flag"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/transport"

	_ "github.com/origadmin/contrib/config/consul"
	_ "github.com/origadmin/contrib/registry/consul"
	"github.com/origadmin/runtime"
	runtimebootstrap "github.com/origadmin/runtime/engine/bootstrap"
	"github.com/origadmin/runtime/log"

	"origadmin/application/origcms/internal/conf"
	_ "origadmin/application/origcms/internal/data/entity/runtime"
	_ "origadmin/application/origcms/internal/helpers/providers"
	confhelper "origadmin/application/origcms/internal/helpers/conf"
)

var (
	// Name is the name of the compiled software.
	Name = "origcms.svc-user"
	// Version is the version of the compiled software.
	Version = "v1.0.0"
	// envName is the environment name.
	envName = ".user"

	// flagconf is the config flag.
	flagconf string
)

func init() {
	flag.StringVar(&flagconf, "conf", "", "config path, eg: -conf bootstrap.yaml")
}

func NewApp(app *runtime.App, servers []transport.Server, kratosOpts ...kratos.Option) *kratos.App {
	log.SetLogger(app.Logger())
	return app.NewApp(servers, kratosOpts...)
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

	// NewFromBootstrap handles config loading, logging, and container setup.
	rt := runtime.New(Name, Version)
	if err := rt.Load(confPath, runtimebootstrap.WithConfigTransformer(conf.Transformer)); err != nil {
		log.Fatalf("failed to create runtime: %v", err)
	}
	defer func() {
		_ = rt.Decoder().Close()
	}()
	rt.ShowAppInfo()

	// Get bootstrap config
	bootstrapConfig, ok := rt.Config().(*conf.Config)
	if !ok {
		log.Fatalf("failed to get bootstrap config")
	}

	// wireApp now takes the runtime instance and builds the kratos app.
	app, cleanupApp, err := wireApp(rt, bootstrapConfig)
	if err != nil {
		log.Fatalf("failed to wire app: %v", err)
	}
	defer cleanupApp()

	// Run the application
	if err := app.Run(); err != nil {
		log.Fatalf("app run failed: %v", err)
	}
}
