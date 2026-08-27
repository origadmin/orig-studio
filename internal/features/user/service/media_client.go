/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package service

import (
	"context"
	"errors"
	"strings"

	"github.com/origadmin/runtime/log"
	runtimegrpc "github.com/origadmin/runtime/service/transport/grpc"
	transportv1 "github.com/origadmin/runtime/api/gen/go/config/transport/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	mediav1 "origadmin/application/origstudio/api/gen/v1/media"
	"origadmin/application/origstudio/internal/conf"
)

// NewMediaServiceClient builds a gRPC client to the media service so the user
// service can compute user-level video counts (GetUserStats/GetMyStats) without
// reaching into the media database directly. This keeps the profile header
// count on the dedicated user interface, decoupled from the content list.
func NewMediaServiceClient(c *conf.Config, logger log.Logger) (mediav1.MediaServiceClient, func(), error) {
	clientConfig := findClientConfig(c, "client.media")
	if clientConfig == nil {
		return nil, nil, errors.New("client config not found: client.media")
	}
	grpcConfig := clientConfig.GetGrpc()
	if grpcConfig == nil {
		return nil, nil, errors.New("grpc client config not found: client.media")
	}
	conn, err := runtimegrpc.NewClient(context.Background(), grpcConfig, &runtimegrpc.ClientOptions{
		DialOptions: []grpc.DialOption{
			grpc.WithUnaryInterceptor(mediaAuthForwardInterceptor),
		},
	})
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() { _ = conn.Close() }
	_ = logger
	return mediav1.NewMediaServiceClient(conn), cleanup, nil
}

func findClientConfig(c *conf.Config, clientName string) *transportv1.Client {
	for _, cli := range c.GetClients().GetConfigs() {
		if cli.Name == clientName {
			return cli
		}
	}
	return nil
}

// mediaAuthForwardInterceptor forwards the inbound Authorization metadata to the
// media service, mirroring the gateway's behaviour for service-to-service calls.
func mediaAuthForwardInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	var authValue string

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get("authorization"); len(vals) > 0 {
			authValue = vals[0]
		}
		if authValue == "" {
			if vals := md.Get("grpcgateway-authorization"); len(vals) > 0 {
				authValue = vals[0]
			}
		}
	}

	if authValue != "" {
		existing, _ := metadata.FromOutgoingContext(ctx)
		newMD := existing.Copy()
		newMD.Set("authorization", authValue)
		if strings.HasPrefix(authValue, "Bearer ") {
			newMD.Set("grpcgateway-authorization", authValue)
		}
		ctx = metadata.NewOutgoingContext(ctx, newMD)
	}

	return invoker(ctx, method, req, reply, cc, opts...)
}
