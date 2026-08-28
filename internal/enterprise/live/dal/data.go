package dal

import (
	"context"
	"fmt"

	"origadmin/application/origstudio/internal/conf"
	"origadmin/application/origstudio/internal/data/entity"
)

func NewEntClient(cfg *conf.Config) (*entity.Client, func(), error) {
	dbDialect, dbSource := cfg.GetDefaultDB()
	if dbDialect == "" {
		dbDialect = "sqlite3"
	}
	if dbSource == "" {
		dbSource = "data/live.db?_fk=1"
	}

	db, err := entity.Open(dbDialect, dbSource)
	if err != nil {
		return nil, nil, fmt.Errorf("NewEntClient: failed to open database: %w", err)
	}

	if err := db.Schema.Create(context.Background()); err != nil {
		return nil, nil, fmt.Errorf("failed creating schema resources: %w", err)
	}

	cleanup := func() {
		_ = db.Close()
	}
	return db, cleanup, nil
}
