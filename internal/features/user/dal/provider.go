package dal

import (
	"fmt"

	"github.com/google/wire"

	"origadmin/application/origstudio/internal/conf"
	"origadmin/application/origstudio/internal/dal/entity"

	"github.com/origadmin/runtime/log"
)

var ProviderSet = wire.NewSet(
	NewData,
	NewUserRepo,
)

var MicroserviceProviderSet = wire.NewSet(
	NewEntClient,
	NewData,
	NewUserRepo,
)

func NewEntClient(cfg *conf.Config) (*entity.Client, func(), error) {
	dbDialect, dbSource := cfg.GetDefaultDB()
	if dbDialect == "" {
		dbDialect = "sqlite3"
	}
	if dbSource == "" {
		dbSource = "data/user.db?_fk=1"
	}

	db, err := entity.Open(dbDialect, dbSource)
	if err != nil {
		return nil, nil, fmt.Errorf("NewEntClient: failed to open database: %w", err)
	}

	cleanup := func() {
		if err := db.Close(); err != nil {
			log.Warnf("failed to close database: %v", err)
		}
	}
	return db, cleanup, nil
}
