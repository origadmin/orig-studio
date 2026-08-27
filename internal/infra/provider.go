/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package infra

import (
	"github.com/google/wire"

	"origadmin/application/origstudio/internal/infra/database"
)

// ProviderSet is the wire provider set for infrastructure components.
var ProviderSet = wire.NewSet(
	NewHasher,
	NewJWTManager,
	NewPubSub,
	NewPublisher,
	NewRouter,
	database.NewDatabaseBundle,
	database.NewEntityClient,
	database.NewSQLDB,
)
