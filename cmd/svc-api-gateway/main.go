/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/origadmin/runtime/log"
	"origadmin/application/origcms/internal/auth"
	"origadmin/application/origcms/internal/gateway/client"
	"origadmin/application/origcms/internal/gateway/router"
)

const (
	defaultAddr     = ":9080"
	shutdownTimeout = 10 * time.Second
)

func main() {
	logger := log.DefaultLogger

	log.NewHelper(logger).Info("svc-api-gateway starting...")

	// Initialize JWT manager for gateway middleware
	// TODO(M1): load secret from config.yaml
	jwtMgr := auth.NewManager(getEnv("JWT_SECRET", "change-me-in-production"), 24*time.Hour)

	// For now, gateway starts with nil clients (handlers will fail gracefully until M1 wires up downstreams)
	_ = client.ProviderSet

	// Build router
	// Real wiring of handlers (feed, detail, etc.) happens in M1/M2 via gRPC clients
	rt := router.NewRouter(nil, nil, nil, nil, jwtMgr, logger)
	h := rt.Build(logger)

	addr := defaultAddr
	if envAddr := os.Getenv("GATEWAY_ADDR"); envAddr != "" {
		addr = envAddr
	}

	srv := &http.Server{
		Addr:         addr,
		Handler:      h,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.NewHelper(logger).Infof("svc-api-gateway listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "server error: %v\n", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.NewHelper(logger).Info("svc-api-gateway shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "forced shutdown: %v\n", err)
	}
	log.NewHelper(logger).Info("svc-api-gateway stopped")
}

func getEnv(key, defaultVal string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return defaultVal
}
