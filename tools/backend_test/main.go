package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"

	"github.com/docker/go-connections/nat"
	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
	"omnicam.com/backend/pkg/logger"
)

var Root string

func init() {
	_, filename, _, ok := runtime.Caller(0) // Get information about the current caller (this file)
	if !ok {
		fmt.Println("Unable to get the current filename.")
		return
	}
	Root = path.Dir(path.Dir(path.Dir(filename)))
}

func setTemplateDb(ctx context.Context, dsn string) error {
	adminConn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return fmt.Errorf("connect admin: %v", err)
	}
	defer adminConn.Close(ctx)

	_, err = adminConn.Exec(ctx, `ALTER DATABASE omnicam WITH IS_TEMPLATE TRUE;`)
	if err != nil {
		return fmt.Errorf("connect database: %v", err)
	}
	return nil
}

func main() {
	ctx := context.Background()
	logger := logger.InitLogger(true)
	defer logger.Sync()

	net, err := network.New(ctx)
	if err != nil {
		logger.Fatal("error while creating docker network", zap.Error(err))
	}
	defer func() {
		if err := net.Remove(ctx); err != nil {
			logger.Info("failed to remove network", zap.Error(err))
		}
	}()

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
		logger.Fatal("failed to start postgres container", zap.Error(err))
	}
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			logger.Info("failed to terminate container", zap.Error(err))
		}
	}()

	host, err := container.Host(ctx)
	if err != nil {
		logger.Fatal("failed to get host of postgres", zap.Error(err))
	}

	mappedPort, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		logger.Fatal("failed to get mapped port of postgres", zap.Error(err))
	}

	// Build DSN
	inNetDsn := "postgres://postgres:password@db:5432/omnicam?sslmode=disable"

	dsn := fmt.Sprintf("postgres://postgres:password@%s:%s/omnicam?sslmode=disable", host, mappedPort.Port())

	// Wait for readiness (already ensured by wait.ForSQL)
	logger.Info("Postgres started\nDSN in network: %s\nDSN: %s", zap.String("inNetDsn", inNetDsn), zap.String("dsn", dsn))

	// ---- Run migrations (using migrate CLI image) ----
	migrationsPath := filepath.Join(Root, "db", "migrations")
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
		logger.Fatal("failed to start migration container", zap.Error(err))
	}
	defer func() {
		if err := migrateCon.Terminate(ctx); err != nil {
			logger.Info("failed to terminate migration container", zap.Error(err))
		}
	}()

	state, err := migrateCon.State(ctx)
	if err != nil {
		logger.Fatal("failed to get container state", zap.Error(err))
	}

	if state.ExitCode != 0 {
		logsReader, _ := migrateCon.Logs(ctx)
		defer logsReader.Close()

		rawLog, _ := io.ReadAll(logsReader)
		logger.Fatal("error while running migration container", zap.String("logs", string(rawLog)))
	}
	logger.Info("Finished migrating db")

	setTemplateDb(ctx, dsn)

	// Target is after "--"
	testArgs := os.Args[2:]
	if len(testArgs) == 0 {
		testArgs = []string{"test", "./..."} // default
	} else if testArgs[0] != "test" {
		testArgs = append([]string{"test"}, testArgs...)
	}

	// ---- Run tests one by one ----
	env := append(os.Environ(), fmt.Sprintf("DATABASE_URL=%s", dsn))
	env = append(env, "MODE=TEST")

	cmd := exec.Command("go", testArgs...)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	logger.Info("Running go test", zap.Strings("args", testArgs))

	if err := cmd.Run(); err != nil {
		logger.Info("tests failed", zap.Error(err))
		os.Exit(1)
	}

	// // Run `go test` for each package with cloned DB
	// pkgs := getTestPackages(target, logger)
	// for _, pkg := range pkgs {
	// 	// Set environment variable for test to use
	// 	env := os.Environ()
	// 	env = append(env, fmt.Sprintf("DATABASE_URL=%s", dsn))
	// 	testArgs := append([]string{"test", pkg}, args...)
	// 	cmd := exec.Command("go", testArgs...)
	// 	cmd.Env = env
	// 	cmd.Stdout = os.Stdout
	// 	cmd.Stderr = os.Stderr

	// 	if err := cmd.Run(); err != nil {
	// 		logger.Info("tests failed for package", zap.String("pkg", pkg), zap.Error(err))
	// 	}
	// }
}
