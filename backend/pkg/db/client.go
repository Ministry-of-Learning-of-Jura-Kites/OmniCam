package db_client

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	// "go.uber.org/zap/internal/pool"
	config_env "omnicam.com/backend/config"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
)

type DB struct {
	Queries *db_sqlc_gen.Queries
	Pool    *pgxpool.Pool
}

func InitDatabase(env *config_env.AppEnv) *DB {
	pool, err := pgxpool.New(context.Background(), env.DatabaseUrl)
	if err != nil {
		fmt.Println("Failed to create DB pool:", err)
		os.Exit(1)
	}

	err = pool.Ping(context.Background())

	if err != nil {
		fmt.Println("Failed to create DB pool:", err)
		os.Exit(1)
	}

	query := db_sqlc_gen.New(pool)

	return &DB{
		Queries: query,
		Pool:    pool,
	}
}
