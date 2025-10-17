//go:build unit_test
// +build unit_test

package controller_workspaces

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	config_env "omnicam.com/backend/config"
	"omnicam.com/backend/internal/middleware"
	"omnicam.com/backend/internal/testutils"
	"omnicam.com/backend/internal/utils"

	_ "github.com/jackc/pgx/v5/stdlib"
	db_client "omnicam.com/backend/pkg/db"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
	"omnicam.com/backend/pkg/logger"
)

var router *gin.Engine

var user db_sqlc_gen.CreateUserRow
var projectId uuid.UUID
var modelId uuid.UUID

func TestMain(m *testing.M) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	testutils.InitTestDbIfNotAlready()

	env := config_env.AppEnv{
		JWTSecret:     "123",
		JWTExpireTime: 168 * time.Hour,
	}

	conn, err, cleanup := testutils.GetTestDb(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	defer cleanup()

	query := db_sqlc_gen.New(conn)

	testutils.TestDb.Queries = *query

	db := &db_client.DB{
		Queries: query,
		Pool:    conn,
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

	hashedPassword, err := utils.HashPassword("test123")
	if err != nil {
		log.Fatalln(err)
	}

	user, err = testutils.TestDb.Queries.CreateUser(ctx, db_sqlc_gen.CreateUserParams{
		Email:     "test@example.com",
		FirstName: "test",
		LastName:  "naja",
		Username:  "test123_-.",
		Password:  []byte(hashedPassword),
	})
	if err != nil {
		log.Fatalln(err)
	}

	projectId = uuid.New()
	_, err = testutils.TestDb.Queries.CreateProject(ctx, db_sqlc_gen.CreateProjectParams{
		ID:          projectId,
		Name:        "project 1",
		Description: "",
		ImagePath:   "",
	})
	if err != nil {
		log.Fatalln(err)
	}

	modelId = uuid.New()
	_, err = testutils.TestDb.Queries.CreateModel(ctx, db_sqlc_gen.CreateModelParams{
		ID:          modelId,
		ProjectID:   projectId,
		Name:        "project 1",
		Description: "",
		FilePath:    "",
		ImagePath:   "",
	})
	if err != nil {
		log.Fatalln(err)
	}

	exitCode := m.Run()

	defer func() {
		testutils.TestDb.Cleanup()
		os.Exit(exitCode)
	}()
}

func TestNoWorkspaceNoPermission(t *testing.T) {
	token, err := utils.GenerateJWT(user.FirstName, user.LastName, user.ID.String(), user.Username, "123", 168*time.Second)
	require.Nil(t, err)

	projectIdBase64, err := utils.UuidToBase64(projectId)
	require.Nil(t, err)
	modelIdBase64, err := utils.UuidToBase64(modelId)
	require.Nil(t, err)

	w := httptest.NewRecorder()
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("/api/v1/projects/%s/models/%s/workspaces/me", projectIdBase64, modelIdBase64),
		nil,
	)
	req.AddCookie(&http.Cookie{
		Name:  "auth_token",
		Value: token,
	})
	require.Nil(t, err)
	router.ServeHTTP(w, req)

	require.Equal(t, 404, w.Code)
}
