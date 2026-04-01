/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package client provides gRPC client connections to downstream services.
package client

import (
	"github.com/google/wire"
	mediav1 "origadmin/application/origcms/api/v1/services/media"
	userv1 "origadmin/application/origcms/api/v1/services/user"
)

// Clients holds all downstream gRPC service clients.
type Clients struct {
	Media mediav1.MediaServiceClient
	User  userv1.UserServiceClient
}

// ProviderSet is client providers for the gateway.
var ProviderSet = wire.NewSet(NewClients)

// NewClients creates a new Clients instance.
// TODO(M2): inject real gRPC connections via service discovery (Consul).
func NewClients(media mediav1.MediaServiceClient, user userv1.UserServiceClient) *Clients {
	return &Clients{
		Media: media,
		User:  user,
	}
}
