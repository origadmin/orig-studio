/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package infra

import "github.com/google/wire"

// ProviderSet is the wire provider set for infrastructure components.
var ProviderSet = wire.NewSet(
	NewHasher,
	NewJWTManager,
	NewPubSub,
	NewPublisher,
	NewRouter,
)
