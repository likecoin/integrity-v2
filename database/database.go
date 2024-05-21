package database

import (
	"context"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/starlinglab/integrity-v2/config"
	"github.com/starlinglab/integrity-v2/util"
)

var (
	pgPool *pgxpool.Pool
	pgOnce sync.Once
)

func GetDatabaseContext() context.Context {
	return context.Background()
}

func GetDatabaseConnectionPool() (*pgxpool.Pool, error) {
	pgOnce.Do(func() {
		connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
			config.GetConfig().Database.User,
			config.GetConfig().Database.Password,
			config.GetConfig().Database.Host,
			config.GetConfig().Database.Port,
			config.GetConfig().Database.Database,
		)
		db, err := pgxpool.New(GetDatabaseContext(), connString)
		if err != nil {
			util.Die("Failed to connect to the database: %v", err)
		}
		pgPool = db
	})
	return pgPool, nil
}

func CloseDatabaseConnectionPool() {
	if pgPool != nil {
		pgPool.Close()
	}
}
