//go:build unit_test
// +build unit_test

package testutils

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"omnicam.com/backend/config"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
	"omnicam.com/backend/pkg/logger"
)

// TODO: Change docker container logic to scripts wrapper of go test instead to gain more control

var (
	// TestDb         *TestDbStruct
	DbNames        = make(chan string, maxPoolSize)
	DbPools        = map[string]*pgxpool.Pool{}
	LockTemplateDb sync.Mutex
	currentCount   = 0
	maxPoolSize    = 5
	initOnce       sync.Once
)

type TestDbStruct struct {
	Queries *db_sqlc_gen.Queries
	pool    *pgxpool.Pool
	Cleanup func()
}

func changeDatabase(dbURL string, newDB string) (string, error) {
	// Parse the URL
	u, err := url.Parse(dbURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	// The database name is typically the "path" component (e.g. /mydb)
	path := strings.TrimPrefix(u.Path, "/")
	if path == "" {
		return "", fmt.Errorf("no database name found in URL")
	}

	// Replace it with the new database
	u.Path = "/" + newDB

	return u.String(), nil
}

func GetTestDb(ctx context.Context, env *config.AppEnv) (string, *pgxpool.Pool, error, func()) {
	logger := logger.InitLogger(true)
	defer logger.Sync()

	// Allocate 1 db
	var dbName string
	select {
	case dbName = <-DbNames:
	default:
		LockTemplateDb.Lock()
		if currentCount < maxPoolSize {
			id, _ := uuid.NewUUID()
			dbName := fmt.Sprintf("omnicam_%s", id.String())

			log.Printf("Creating new DB %s", dbName)

			adminConn, err := pgx.Connect(ctx, env.DatabaseUrl)
			if err != nil {
				LockTemplateDb.Unlock()
				return "", nil, fmt.Errorf("connect admin with url %s failed: %w", env.DatabaseUrl, err), nil
			}
			_, err = adminConn.Exec(ctx, fmt.Sprintf(`CREATE DATABASE "%s" WITH TEMPLATE omnicam OWNER postgres;`, dbName))
			adminConn.Close(ctx)
			if err != nil {
				LockTemplateDb.Unlock()
				return "", nil, fmt.Errorf("create database %s failed: %w", dbName, err), nil
			}

			currentCount++
			LockTemplateDb.Unlock()

			dbUrl, err := changeDatabase(env.DatabaseUrl, dbName)
			if err != nil {
				return "", nil, err, nil
			}
			cfg, err := pgxpool.ParseConfig(dbUrl)
			if err != nil {
				return "", nil, err, nil
			}
			cfg.ConnConfig.Database = dbName

			conn, err := pgxpool.NewWithConfig(ctx, cfg)
			if err != nil {
				return "", nil, err, nil
			}
			DbPools[dbName] = conn
			return dbName, conn, nil, func() {
				DbNames <- dbName
			}
		} else {
			// Wait for an available DB from the pool
			LockTemplateDb.Unlock()
			dbName = <-DbNames
		}
	}

	return dbName, DbPools[dbName], nil, func() {
		DbNames <- dbName
	}

	// DbSem <- struct{}{}

	// id, err := uuid.NewUUID()
	// if err != nil {
	// 	return nil, err, nil
	// }

	// newDbName := fmt.Sprintf("omnicam_%s", id.String())

	// semReleased := false
	// defer func() {
	// 	if !semReleased {
	// 		<-DbSem
	// 	}
	// }()

	// // Create database
	// err = func() error {
	// 	LockTemplateDb.Lock()
	// 	defer func() {
	// 		LockTemplateDb.Unlock()
	// 	}()
	// 	log.Printf("Creating DB %s: Acquired a lock", newDbName)

	// 	adminConn, err := pgx.Connect(ctx, TestDb.DSN)
	// 	if err != nil {
	// 		return fmt.Errorf("connect admin: %w", err)
	// 	}
	// 	defer adminConn.Close(ctx)

	// 	_, err = adminConn.Exec(ctx, fmt.Sprintf(`CREATE DATABASE "%s" WITH TEMPLATE omnicam OWNER postgres;`, newDbName))
	// 	if err != nil {
	// 		return fmt.Errorf("create database: %w", err)
	// 	}

	// 	return nil
	// }()
	// if err != nil {
	// 	return nil, err, nil
	// }

	// log.Printf("Created DB %s: Release a lock", newDbName)

	// // Connect to new database
	// cfg, err := pgxpool.ParseConfig(TestDb.DSN)
	// if err != nil {
	// 	return nil, err, nil
	// }

	// cfg.ConnConfig.Database = newDbName

	// conn, err := pgxpool.NewWithConfig(ctx, cfg)
	// if err != nil {
	// 	return nil, err, nil
	// }

	// semReleased = true
	// return conn, nil, func() {
	// 	adminConn, err := pgx.Connect(ctx, TestDb.DSN)
	// 	if err != nil {
	// 		log.Fatalf("connect admin: %v", err)
	// 	}
	// 	defer adminConn.Close(ctx)

	// 	_, err = adminConn.Exec(ctx, fmt.Sprintf(`DROP DATABASE "%s";`, newDbName))
	// 	if err != nil {
	// 		log.Fatalf("error while dropping database %s: %v", newDbName, err)
	// 	}
	// 	log.Printf("Destroyed DB %s: releasing semaphore...", newDbName)
	// 	<-DbSem
	// }
}

func Truncate(ctx context.Context, dbName string) error {
	log.Printf("Truncating %s", dbName)

	conn := DbPools[dbName]
	rows, err := conn.Query(ctx, "SELECT tablename FROM pg_tables WHERE schemaname = 'public'")
	if err != nil {
		return err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return err
		}
		tables = append(tables, tableName)
	}

	// Truncate each table
	for _, table := range tables {
		_, err := conn.Exec(ctx, fmt.Sprintf(`TRUNCATE TABLE "%s" RESTART IDENTITY CASCADE`, table)) // RESTART IDENTITY for auto-increment reset
		if err != nil {
			log.Printf("Error truncating table %s: %v", table, err)
		} else {
			log.Printf("Table %s truncated.\n", table)
		}
	}
	return nil
}
