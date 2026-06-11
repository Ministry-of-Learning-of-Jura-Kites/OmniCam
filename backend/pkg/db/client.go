package db_client

import (
	"context"
	"time"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"

	// "go.uber.org/zap/internal/pool"
	config_env "omnicam.com/backend/config"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
)

type DB struct {
	Queries *db_sqlc_gen.Queries
	Pool    *pgxpool.Pool
}

func InitDatabase(env *config_env.AppEnv, logger *zap.Logger) *DB {
	config, err := pgxpool.ParseConfig(env.DatabaseUrl)
	if err != nil {
		logger.Fatal("Unable to parse DATABASE_URL", zap.Error(err))
	}

	config.ConnConfig.Tracer = otelpgx.NewTracer(
		otelpgx.WithTracerProvider(otel.GetTracerProvider()),
		otelpgx.WithTrimSQLInSpanName(),
		otelpgx.WithDisableQuerySpanNamePrefix(),
	)

	var pool *pgxpool.Pool
	maxRetries := 5

	for i := 0; i < maxRetries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		pool, err = pgxpool.NewWithConfig(ctx, config)
		if err == nil {
			err = pool.Ping(ctx)
		}
		cancel()

		if err == nil {
			logger.Info("Successfully connected to the database")

			query := db_sqlc_gen.New(pool)

			return &DB{
				Queries: query,
				Pool:    pool,
			}
		}

		backoff := time.Duration(i*i) * time.Second

		logger.Warn("DB connection failed",
			zap.Int("attempt", i+1),
			zap.Int("max_retries", maxRetries),
			zap.Duration("backoff", backoff),
			zap.Error(err),
		)

		time.Sleep(backoff)
	}

	logger.Fatal("Exiting: Could not connect to DB after maximum attempts",
		zap.Int("total_attempts", maxRetries),
	)

	return nil
}
