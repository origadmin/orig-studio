package dal

import (
	"context"
	"fmt"

	"github.com/google/wire"

	"origadmin/application/origstudio/internal/conf"
	"origadmin/application/origstudio/internal/dal/entity"

	"github.com/origadmin/runtime/log"
)

var ProviderSet = wire.NewSet(
	NewMediaRepo,
	NewUploadRepo,
	NewEncodeProfileRepo,
	NewEncodingTaskRepo,
	NewReviewLogRepo,
	)

var MicroserviceProviderSet = wire.NewSet(
	NewEntClient,
	NewMediaRepo,
	NewUploadRepo,
	NewEncodeProfileRepo,
	NewEncodingTaskRepo,
	NewReviewLogRepo,
	NewLocalStorage,
)

func NewEntClient(cfg *conf.Config) (*entity.Client, func(), error) {
	dbDialect, dbSource := cfg.GetDefaultDB()
	if dbDialect == "" {
		dbDialect = "sqlite3"
	}
	if dbSource == "" {
		dbSource = "data/media.db?_fk=1"
	}

	db, err := entity.Open(dbDialect, dbSource)
	if err != nil {
		return nil, nil, fmt.Errorf("NewEntClient: failed to open database: %w", err)
	}

	if err := db.Schema.Create(context.Background()); err != nil {
		return nil, nil, fmt.Errorf("failed creating schema resources: %w", err)
	}

	if err := SeedEncodeProfiles(context.Background(), db); err != nil {
		log.Warnf("failed to seed encode profiles: %v", err)
	}

	cleanup := func() {
		_ = db.Close()
	}
	return db, cleanup, nil
}
