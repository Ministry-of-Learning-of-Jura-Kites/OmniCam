//go:build unit_test
// +build unit_test

package testutils

import (
	"context"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"sync"

	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"

	tc "github.com/testcontainers/testcontainers-go"
	"omnicam.com/backend/internal"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
)

var TestDb *TestDbStruct

var DbSem = make(chan struct{}, 5)

var LockTemplateDb sync.Mutex

var initOnce sync.Once

type TestDbStruct struct {
	DSN      string
	Queries  db_sqlc_gen.Queries
	pool     *pgxpool.Pool
	Cleanup  func()
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

func startTestDB(ctx context.Context) (*TestDbStruct, error) {
	net, err := network.New(ctx)
	if err != nil {
		return nil, err
	}

	container, err := tc.Run(ctx, "postgres:17",
		tc.WithExposedPorts("5432/tcp"),
		tc.WithEnv(map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "password",
			"POSTGRES_DB":       "omnicam",
		}),
		network.WithNetwork([]string{"db"}, net),
		tc.WithWaitStrategy(wait.ForSQL("5432/tcp", "pgx", func(host string, port nat.Port) string {
			return fmt.Sprintf("postgres://postgres:password@%s:%s/omnicam?sslmode=disable", host, port.Port())
		})),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	mappedPort, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		return nil, err
	}

	// Build DSN
	inNetDsn := "postgres://postgres:password@db:5432/omnicam?sslmode=disable"

	dsn := fmt.Sprintf("postgres://postgres:password@%s:%s/omnicam?sslmode=disable", host, mappedPort.Port())

	// Wait for readiness (already ensured by wait.ForSQL)
	log.Printf("Postgres started\nDSN in network: %s\nDSN: %s", inNetDsn, dsn)

	// ---- Run migrations (using migrate CLI image) ----
	migrationsPath := filepath.Join(internal.Root, "db", "migrations")
	migrateCon, err := tc.Run(
		ctx,
		"migrate/migrate",
		network.WithNetwork([]string{"migrate"}, net),
		tc.WithFiles(tc.ContainerFile{
			HostFilePath:      migrationsPath,
			ContainerFilePath: "/migrations",
			FileMode:          0o644,
		}),
		tc.WithCmdArgs(
			"-path",
			"/migrations",
			"-database",
			inNetDsn,
			"up",
		),
		tc.WithWaitStrategy(wait.ForExit()),
	)
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to start migration container: %w", err)
	}

	state, err := migrateCon.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %v", err)
	}

	if state.ExitCode != 0 {
		logsReader, _ := migrateCon.Logs(ctx)
		defer logsReader.Close()

		rawLog, _ := io.ReadAll(logsReader)
		return nil, fmt.Errorf("error while running migration container\n Logs:%s", string(rawLog))
	}
	log.Printf("Finished migrating db")

	adminConn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		log.Fatalf("connect admin: %v", err)
	}
	defer adminConn.Close(ctx)

	_, err = adminConn.Exec(ctx, `ALTER DATABASE omnicam WITH IS_TEMPLATE TRUE;`)
	if err != nil {
		log.Fatalf("create database: %v", err)
	}

	return &TestDbStruct{
		DSN:  dsn,
		Host: host,
		Port: mappedPort.Port(),
		Cleanup: func() {
			if err := container.Terminate(ctx); err != nil {
				log.Printf("failed to terminate container: %v", err)
			}

			if err := net.Remove(ctx); err != nil {
				log.Printf("failed to remove network: %s", err)
			}
		},
	}, nil
}

func InitTestDbIfNotAlready() {
	initOnce.Do(func() {
		ctx := context.Background()

		testDb, err := startTestDB(ctx)
		if err != nil {
			log.Fatalln(err)
		}

		TestDb = testDb
	})
}

func GetTestDb(ctx context.Context) (*pgxpool.Pool, error, func()) {
	// Allocate 1 db
	DbSem <- struct{}{}

	id, err := uuid.NewUUID()
	if err != nil {
		return nil, err, nil
	}

	newDbName := fmt.Sprintf("omnicam_%s", id.String())

	semReleased := false
	defer func() {
		if !semReleased {
			<-DbSem
		}
	}()

	// Create database
	err = func() error {
		LockTemplateDb.Lock()
		defer func() {
			LockTemplateDb.Unlock()
		}()
		log.Printf("Creating DB %s: Acquired a lock", newDbName)

		adminConn, err := pgx.Connect(ctx, TestDb.DSN)
		if err != nil {
			return fmt.Errorf("connect admin: %w", err)
		}
		defer adminConn.Close(ctx)

		_, err = adminConn.Exec(ctx, fmt.Sprintf(`CREATE DATABASE "%s" WITH TEMPLATE omnicam OWNER postgres;`, newDbName))
		if err != nil {
			return fmt.Errorf("create database: %w", err)
		}

		return nil
	}()
	if err != nil {
		return nil, err, nil
	}

	log.Printf("Created DB %s: Release a lock", newDbName)

	// Connect to new database
	cfg, err := pgxpool.ParseConfig(TestDb.DSN)
	if err != nil {
		return nil, err, nil
	}

	cfg.ConnConfig.Database = newDbName

	conn, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err, nil
	}

	semReleased = true
	return conn, nil, func() {
		adminConn, err := pgx.Connect(ctx, TestDb.DSN)
		if err != nil {
			log.Fatalf("connect admin: %v", err)
		}
		defer adminConn.Close(ctx)

		_, err = adminConn.Exec(ctx, fmt.Sprintf(`DROP DATABASE "%s";`, newDbName))
		if err != nil {
			log.Fatalf("error while dropping database %s: %v", newDbName, err)
		}
		log.Printf("Destroyed DB %s: releasing semaphore...", newDbName)
		<-DbSem
	}
}
