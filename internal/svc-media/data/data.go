/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package data

import (
	"fmt"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/google/wire"

	"github.com/origadmin/runtime"
	storageiface "github.com/origadmin/runtime/contracts/storage"
	"github.com/origadmin/runtime/helpers/comp"

	"origadmin/application/origcms/internal/data/entity"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewEntClient,
	NewMediaRepo,
)

// NewEntClient creates a new *entity.Client for svc-media.
func NewEntClient(app *runtime.App) (*entity.Client, func(), error) {
	dbInst, err := comp.Get[storageiface.Database](app.Context(), app.Container().In(runtime.CategoryDatabase))
	if err != nil {
		return nil, nil, fmt.Errorf("NewEntClient: failed to get database: %w", err)
	}

	drv := entsql.OpenDB(dbInst.Dialect(), dbInst.DB())
	client := entity.NewClient(entity.Driver(drv))

	cleanup := func() {
		_ = client.Close()
	}
	return client, cleanup, nil
}
