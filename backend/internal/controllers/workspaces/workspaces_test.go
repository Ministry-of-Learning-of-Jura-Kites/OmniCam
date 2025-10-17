//go:build unit_test
// +build unit_test

package controller_workspaces

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
	config_env "omnicam.com/backend/config"
	"omnicam.com/backend/internal/middleware"
	"omnicam.com/backend/internal/utils"

	tc "github.com/testcontainers/testcontainers-go"

	_ "github.com/jackc/pgx/v5/stdlib"
	"omnicam.com/backend/internal"
	db_client "omnicam.com/backend/pkg/db"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
	"omnicam.com/backend/pkg/logger"
)

var testDB *TestDB

var router *gin.Engine

type TestDB struct {
	DSN      string
	Queries  db_sqlc_gen.Queries
	Pool     *pgxpool.Pool
	Cleanup  func()
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

func StartTestDB(ctx context.Context) (*TestDB, error) {
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

	// ---- Create DB connection pool ----
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to test db: %w", err)
	}

	return &TestDB{
		DSN:  dsn,
		Pool: pool,
		Host: host,
		Port: mappedPort.Port(),
		Cleanup: func() {
			pool.Close()
			container.Terminate(ctx)
			if err := net.Remove(ctx); err != nil {
				log.Printf("failed to remove network: %s", err)
			}
		},
	}, nil
}

func TestMain(m *testing.M) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var err error
	testDB, err = StartTestDB(ctx)
	if err != nil {
		log.Fatalf("Failed to init DB containers %s", err)
	}

	env := config_env.AppEnv{
		JWTSecret:     "123",
		JWTExpireTime: 168 * time.Hour,
	}

	query := db_sqlc_gen.New(testDB.Pool)

	testDB.Queries = *query

	db := &db_client.DB{
		Queries: query,
		Pool:    testDB.Pool,
	}

	logger := logger.InitLogger()

	router = gin.Default()
	apiV1 := router.Group("/api/v1")
	protectedRoute := apiV1.Group("/")
	authMiddleware := middleware.AuthMiddleware{
		Env:    &env,
		Logger: logger,
	}
	protectedRoute.Use(authMiddleware.CreateHandler())

	route := WorkspaceRoute{
		Logger: logger,
		Env:    &env,
		DB:     db,
	}

	route.InitRoute(protectedRoute)

	exitCode := m.Run()

	defer func() {
		testDB.Cleanup()
		os.Exit(exitCode)
	}()
}

func TestHappy(t *testing.T) {
	hashedPassword, err := utils.HashPassword("test123")
	require.Nil(t, err)

	user, err := testDB.Queries.CreateUser(t.Context(), db_sqlc_gen.CreateUserParams{
		Email:     "test@example.com",
		FirstName: "test",
		LastName:  "naja",
		Username:  "test123_-.",
		Password:  []byte(hashedPassword),
	})
	require.Nil(t, err)

	projectId := uuid.New()
	_, err = testDB.Queries.CreateProject(t.Context(), db_sqlc_gen.CreateProjectParams{
		ID:          projectId,
		Name:        "project 1",
		Description: "",
		ImagePath:   "",
	})
	require.Nil(t, err)

	modelId := uuid.New()
	_, err = testDB.Queries.CreateModel(t.Context(), db_sqlc_gen.CreateModelParams{
		ID:          modelId,
		Name:        "project 1",
		Description: "",
		FilePath:    "",
		ImagePath:   "",
	})
	require.Nil(t, err)

	token, err := utils.GenerateJWT(user.FirstName, user.LastName, user.ID.String(), user.Username, "123", 168*time.Second)
	require.Nil(t, err)

	w := httptest.NewRecorder()
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("/api/v1/projects/%s/models/%s/workspaces/me", projectId.String(), modelId.String()),
		nil,
	)
	req.AddCookie(&http.Cookie{
		Name:  "auth_token",
		Value: token,
	})
	require.Nil(t, err)
	router.ServeHTTP(w, req)

	// req.Response.Body

	require.Equal(t, 403, w.Code)
}
