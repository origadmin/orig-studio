package database

import (
	"database/sql"

	"github.com/origadmin/runtime/log"

	config "origadmin/application/origstudio/internal/conf"
	"origadmin/application/origstudio/internal/dal/entity"
)

// DatabaseBundle bundles the ent client and the raw *sql.DB handle so that
// wire can inject both dependencies from a single provider (wire only allows
// one non-error return value plus an optional func() cleanup).
type DatabaseBundle struct {
	Client *entity.Client
	SQLDB  *sql.DB
}

func NewDatabaseBundle(cfg *config.Config, logger log.Logger) (*DatabaseBundle, func(), error) {
	client, sqlDB, err := NewDatabase(cfg, logger)
	if err != nil {
		return nil, nil, err
	}
	return &DatabaseBundle{Client: client, SQLDB: sqlDB}, func() { _ = sqlDB.Close() }, nil
}

func NewEntityClient(b *DatabaseBundle) *entity.Client {
	return b.Client
}

func NewSQLDB(b *DatabaseBundle) *sql.DB {
	return b.SQLDB
}
