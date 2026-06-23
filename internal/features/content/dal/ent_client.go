package dal

import (
	"context"
	"fmt"

	"origadmin/application/origstudio/internal/conf"
	"origadmin/application/origstudio/internal/data/entity"

	"github.com/origadmin/runtime/log"
)

func NewEntClient(cfg *conf.Config) (*entity.Client, func(), error) {
	dbDialect, dbSource := cfg.GetDefaultDB()
	if dbDialect == "" {
		dbDialect = "sqlite3"
	}
	if dbSource == "" {
		dbSource = "data/content.db?_fk=1"
	}

	db, err := entity.Open(dbDialect, dbSource)
	if err != nil {
		return nil, nil, fmt.Errorf("NewEntClient: failed to open database: %w", err)
	}

	if err := db.Schema.Create(context.Background()); err != nil {
		return nil, nil, fmt.Errorf("failed creating schema resources: %w", err)
	}

	if err := SeedCategories(context.Background(), db); err != nil {
		log.Warnf("failed to seed categories: %v", err)
	}

	if err := SeedTags(context.Background(), db); err != nil {
		log.Warnf("failed to seed tags: %v", err)
	}

	cleanup := func() {
		_ = db.Close()
	}
	return db, cleanup, nil
}
