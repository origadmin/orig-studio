/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package data provides the data access layer implementations.
package dal

import (
	"fmt"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/google/wire"

	"github.com/origadmin/runtime"
	storageiface "github.com/origadmin/runtime/contracts/storage"
	"github.com/origadmin/runtime/helpers/comp"

	"origadmin/application/origstudio/internal/data/entity"
)

// ProviderSet is the wire provider set for data layer in monolith mode
// (without NewEntClient, which is provided by infra.NewDatabase instead).
// For microservice mode, use MicroserviceProviderSet.
var ProviderSet = wire.NewSet(
	NewData,
	NewUserRepo,
)

// MicroserviceProviderSet is the wire provider set for standalone microservice mode,
// which includes NewEntClient that uses runtime.App to obtain the database connection.
var MicroserviceProviderSet = wire.NewSet(
	NewEntClient,
	NewData,
	NewUserRepo,
)

// NewEntClient creates a new *entity.Client from the runtime App container.
// It mirrors the pattern used in backend/data/provider.go (NewEnt + NewDatabase).
func NewEntClient(app *runtime.App) (*entity.Client, func(), error) {
	// Extract the raw infrastructure database from the runtime container.
	dbInst, err := comp.Get[storageiface.Database](app.Context(), app.Container().In(runtime.CategoryDatabase), "")
	if err != nil {
		return nil, nil, fmt.Errorf("NewEntClient: failed to get database from container: %w", err)
	}

	// Open an ent SQL driver using the infrastructure DB.
	drv := entsql.OpenDB(dbInst.Dialect(), dbInst.DB())
	client := entity.NewClient(entity.Driver(drv))

	cleanup := func() {
		if err := client.Close(); err != nil {
			_ = err // best-effort close
		}
	}
	return client, cleanup, nil
}
