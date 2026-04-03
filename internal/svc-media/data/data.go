/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package data

import (
	"context" // Import context
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
	NewUploadRepo,
	NewEncodeProfileRepo,
	NewEncodingTaskRepo,
	NewLocalStorage,
)

// NewEntClient creates a new *entity.Client for svc-media.
func NewEntClient(app *runtime.App) (*entity.Client, func(), error) {
	dbInst, err := comp.Get[storageiface.Database](app.Context(), app.Container().In(runtime.CategoryDatabase))
	if err != nil {
		return nil, nil, fmt.Errorf("NewEntClient: failed to get database: %w", err)
	}

	drv := entsql.OpenDB(dbInst.Dialect(), dbInst.DB())
	client := entity.NewClient(entity.Driver(drv))

	// Ensure the database schema is created
	if err := client.Schema.Create(context.Background()); err != nil {
		return nil, nil, fmt.Errorf("failed creating schema resources: %w", err)
	}

	// Call SeedEncodeProfiles to initialize default data
	if err := SeedEncodeProfiles(context.Background(), client); err != nil {
		return nil, nil, fmt.Errorf("failed to seed encode profiles: %w", err)
	}

	cleanup := func() {
		_ = client.Close()
	}
	return client, cleanup, nil
}
