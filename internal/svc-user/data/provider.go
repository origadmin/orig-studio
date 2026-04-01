/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package data provides the data access layer implementations.
package data

import (
	"fmt"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/google/wire"

	"github.com/origadmin/runtime"
	storageiface "github.com/origadmin/runtime/contracts/storage"
	"github.com/origadmin/runtime/helpers/comp"

	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/svc-user/dto"
)

// ProviderSet is the wire provider set for data layer.
var ProviderSet = wire.NewSet(
	NewEntClient,
	NewData,
	NewUserRepo,
	wire.Bind(new(dto.UserRepo), new(*userRepo)),
)

// NewEntClient creates a new *entity.Client from the runtime App container.
// It mirrors the pattern used in backend/data/provider.go (NewEnt + NewDatabase).
func NewEntClient(app *runtime.App) (*entity.Client, func(), error) {
	// Extract the raw infrastructure database from the runtime container.
	dbInst, err := comp.Get[storageiface.Database](app.Context(), app.Container().In(runtime.CategoryDatabase))
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
