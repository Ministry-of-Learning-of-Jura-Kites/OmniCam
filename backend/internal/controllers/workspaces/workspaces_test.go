//go:build unit_test
// +build unit_test

package controller_workspaces

import (
	"context"
	"encoding/json"
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
var project1Id uuid.UUID
var project2Id uuid.UUID
var model1Id uuid.UUID
var model2Id uuid.UUID
var token string
var dbName string

func seedData(ctx context.Context) {
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

	project1Id = uuid.New()
	_, err = testutils.TestDb.Queries.CreateProject(ctx, db_sqlc_gen.CreateProjectParams{
		ID:          project1Id,
		Name:        "project 1",
		Description: "",
		ImagePath:   "",
	})
	if err != nil {
		log.Fatalln(err)
	}

	project2Id = uuid.New()
	_, err = testutils.TestDb.Queries.CreateProject(ctx, db_sqlc_gen.CreateProjectParams{
		ID:          project2Id,
		Name:        "project 2",
		Description: "",
		ImagePath:   "",
	})
	if err != nil {
		log.Fatalln(err)
	}

	model1Id = uuid.New()
	_, err = testutils.TestDb.Queries.CreateModel(ctx, db_sqlc_gen.CreateModelParams{
		ID:          model1Id,
		ProjectID:   project1Id,
		Name:        "model 1",
		Description: "",
		FilePath:    "",
		ImagePath:   "",
	})
	if err != nil {
		log.Fatalln(err)
	}

	model2Id = uuid.New()
	_, err = testutils.TestDb.Queries.CreateModel(ctx, db_sqlc_gen.CreateModelParams{
		ID:          model2Id,
		ProjectID:   project2Id,
		Name:        "model 2",
		Description: "",
		FilePath:    "",
		ImagePath:   "",
	})
	if err != nil {
		log.Fatalln(err)
	}

	token, err = utils.GenerateJWT(user.FirstName, user.LastName, user.ID.String(), user.Username, "123", 168*time.Second)
	if err != nil {
		log.Fatalln(err)
	}
}

func TestMain(m *testing.M) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	env := config_env.AppEnv{
		JWTSecret:     "123",
		JWTExpireTime: 168 * time.Hour,
	}

	foundDbName, conn, err, cleanup := testutils.GetTestDb(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	defer cleanup()

	dbName = foundDbName

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

	exitCode := m.Run()

	defer func() {
		testutils.TestDb.Cleanup()
		os.Exit(exitCode)
	}()
}

func TestNoWorkspaceNoPermission(t *testing.T) {
	require.NoError(t, testutils.Truncate(t.Context(), dbName))
	seedData(t.Context())

	projectIdBase64, err := utils.UuidToBase64(project1Id)
	require.Nil(t, err)
	modelIdBase64, err := utils.UuidToBase64(model1Id)
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

func TestCreateAndGetWorkspace(t *testing.T) {
	require.NoError(t, testutils.Truncate(t.Context(), dbName))
	seedData(t.Context())

	// --- Arrange ---
	projectIdBase64, err := utils.UuidToBase64(project1Id)
	require.NoError(t, err)

	modelIdBase64, err := utils.UuidToBase64(model1Id)
	require.NoError(t, err)

	// Add the user to the project, since getWorkspaceMe requires permission
	_, err = testutils.TestDb.Queries.AddUserToProject(t.Context(), db_sqlc_gen.AddUserToProjectParams{
		UserID:    user.ID,
		ProjectID: project1Id,
		Role:      db_sqlc_gen.RoleCollaborator,
	})
	require.NoError(t, err)

	// --- Act: create workspace ---
	w := httptest.NewRecorder()
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("/api/v1/projects/%s/models/%s/workspaces/me", projectIdBase64, modelIdBase64),
		nil,
	)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: "auth_token", Value: token})

	router.ServeHTTP(w, req)

	// --- Assert: workspace created ---
	require.Equal(t, http.StatusCreated, w.Code, "expected 201 Created")

	var response struct {
		Data struct {
			ModelId     string `json:"modelId"`
			UserId      string `json:"userId"`
			Version     int64  `json:"version"`
			BaseVersion int64  `json:"baseVersion"`
		} `json:"data"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	require.Equal(t, model1Id.String(), response.Data.ModelId)
	require.Equal(t, user.ID.String(), response.Data.UserId)

	// --- Act: get workspace ---
	w = httptest.NewRecorder()
	req, err = http.NewRequest(
		"GET",
		fmt.Sprintf("/api/v1/projects/%s/models/%s/workspaces/me?fields=cameras", projectIdBase64, modelIdBase64),
		nil,
	)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: "auth_token", Value: token})

	router.ServeHTTP(w, req)

	// --- Assert: workspace retrieved ---
	require.Equal(t, http.StatusOK, w.Code, "expected 200 OK")
	require.Contains(t, w.Body.String(), model1Id.String())
	require.Contains(t, w.Body.String(), project1Id.String())
}

func TestCreateWorkspaceAlreadyExists(t *testing.T) {
	require.NoError(t, testutils.Truncate(t.Context(), dbName))
	seedData(t.Context())

	projectIdBase64, _ := utils.UuidToBase64(project1Id)
	modelIdBase64, _ := utils.UuidToBase64(model1Id)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(
		"POST",
		fmt.Sprintf("/api/v1/projects/%s/models/%s/workspaces/me", projectIdBase64, modelIdBase64),
		nil,
	)
	req.AddCookie(&http.Cookie{Name: "auth_token", Value: token})
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(
		"POST",
		fmt.Sprintf("/api/v1/projects/%s/models/%s/workspaces/me", projectIdBase64, modelIdBase64),
		nil,
	)
	req.AddCookie(&http.Cookie{Name: "auth_token", Value: token})
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code, "expected 400 Bad Request for duplicate workspace")
	require.Contains(t, w.Body.String(), "workspace already exists")
}

func TestMergeWorkspaceNoChanges(t *testing.T) {
	require.NoError(t, testutils.Truncate(t.Context(), dbName))
	seedData(t.Context())

	projectIdBase64, _ := utils.UuidToBase64(project1Id)
	modelIdBase64, _ := utils.UuidToBase64(model1Id)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(
		"POST",
		fmt.Sprintf("/api/v1/projects/%s/models/%s/workspaces/merge", projectIdBase64, modelIdBase64),
		nil,
	)
	req.AddCookie(&http.Cookie{Name: "auth_token", Value: token})
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestMergeWorkspaceInvalidModelId(t *testing.T) {
	require.NoError(t, testutils.Truncate(t.Context(), dbName))
	seedData(t.Context())

	projectIdBase64, _ := utils.UuidToBase64(project1Id)
	invalidModelId := "!!!invalid"

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(
		"POST",
		fmt.Sprintf("/api/v1/projects/%s/models/%s/workspaces/merge", projectIdBase64, invalidModelId),
		nil,
	)
	req.AddCookie(&http.Cookie{Name: "auth_token", Value: token})
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Contains(t, w.Body.String(), "invalid project ID")
}
