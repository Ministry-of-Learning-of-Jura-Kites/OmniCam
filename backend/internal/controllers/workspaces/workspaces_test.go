//go:build unit_test
// +build unit_test

package controller_workspaces

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	config_env "omnicam.com/backend/config"
	"omnicam.com/backend/internal/middleware"
	"omnicam.com/backend/internal/testutils"
	"omnicam.com/backend/internal/utils"

	db_client "omnicam.com/backend/pkg/db"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
	"omnicam.com/backend/pkg/logger"
)

var testLogger = logger.InitLogger(true)

// --- Helper Struct ---
type testContext struct {
	Ctx      context.Context
	Env      *config_env.AppEnv
	DB       *db_client.DB
	User     db_sqlc_gen.CreateUserRow
	Project1 uuid.UUID
	Project2 uuid.UUID
	Model1   uuid.UUID
	Model2   uuid.UUID
	Token    string
	Router   *gin.Engine
	DBName   string
	Cleanup  func(*testing.T)
}

// --- Setup Helper ---
func setupTest(t *testing.T) *testContext {
	t.Helper()

	ctx := context.Background()
	env := config_env.InitAppEnv(testLogger)
	env.JWTSecret = "123"
	env.JWTExpireTime = 168 * time.Hour

	dbName, conn, err, cleanup := testutils.GetTestDb(ctx, env)
	require.NoError(t, err, "failed to get test DB")

	db := &db_client.DB{
		Queries: db_sqlc_gen.New(conn),
		Pool:    conn,
	}

	// --- Seed data ---
	hashedPassword, err := utils.HashPassword("test123")
	require.NoError(t, err)

	user, err := db.Queries.CreateUser(ctx, db_sqlc_gen.CreateUserParams{
		Email:     "test@example.com",
		FirstName: "test",
		LastName:  "naja",
		Username:  "test123_-.",
		Password:  []byte(hashedPassword),
	})
	require.NoError(t, err)

	project1 := uuid.New()
	_, err = db.Queries.CreateProject(ctx, db_sqlc_gen.CreateProjectParams{
		ID: project1, Name: "project 1",
	})
	require.NoError(t, err)

	project2 := uuid.New()
	_, err = db.Queries.CreateProject(ctx, db_sqlc_gen.CreateProjectParams{
		ID: project2, Name: "project 2",
	})
	require.NoError(t, err)

	model1 := uuid.New()
	_, err = db.Queries.CreateModel(ctx, db_sqlc_gen.CreateModelParams{
		ID: model1, ProjectID: project1, Name: "model 1",
	})
	require.NoError(t, err)

	model2 := uuid.New()
	_, err = db.Queries.CreateModel(ctx, db_sqlc_gen.CreateModelParams{
		ID: model2, ProjectID: project2, Name: "model 2",
	})
	require.NoError(t, err)

	token, err := utils.GenerateJWT(user.FirstName, user.LastName, user.ID.String(), user.Username, env.JWTSecret, 168*time.Hour)
	require.NoError(t, err)

	// --- Router setup ---
	router := gin.Default()
	apiV1 := router.Group("/api/v1")
	protected := apiV1.Group("/")
	authMiddleware := middleware.AuthMiddleware{Env: env, Logger: testLogger}
	protected.Use(authMiddleware.CreateHandler())

	route := WorkspaceRoute{
		Logger: testLogger,
		Env:    env,
		DB:     db,
	}
	route.InitRoute(protected)

	t.Cleanup(func() {
		cleanup(t) // drop DB
	})

	return &testContext{
		Ctx:      ctx,
		Env:      env,
		DB:       db,
		User:     user,
		Project1: project1,
		Project2: project2,
		Model1:   model1,
		Model2:   model2,
		Token:    token,
		Router:   router,
		DBName:   dbName,
		Cleanup:  cleanup,
	}
}

func TestWorkspacesEndpoints(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, tc *testContext)
	}{
		{
			name: "No workspace and no permission",
			run: func(t *testing.T, tc *testContext) {
				projectIdBase64, err := utils.UuidToBase64(tc.Project1)
				require.NoError(t, err)
				modelIdBase64, err := utils.UuidToBase64(tc.Model1)
				require.NoError(t, err)

				req, _ := http.NewRequest("GET",
					fmt.Sprintf("/api/v1/projects/%s/models/%s/workspaces/me", projectIdBase64, modelIdBase64),
					nil)
				req.AddCookie(&http.Cookie{Name: "auth_token", Value: tc.Token})

				w := httptest.NewRecorder()
				tc.Router.ServeHTTP(w, req)

				require.Equal(t, http.StatusNotFound, w.Code)
			},
		},
		{
			name: "Create and get workspace successfully",
			run: func(t *testing.T, tc *testContext) {
				projectIdBase64, _ := utils.UuidToBase64(tc.Project1)
				modelIdBase64, _ := utils.UuidToBase64(tc.Model1)

				_, err := tc.DB.Queries.AddUserToProject(tc.Ctx, db_sqlc_gen.AddUserToProjectParams{
					UserID: tc.User.ID, ProjectID: tc.Project1, Role: db_sqlc_gen.RoleCollaborator,
				})
				require.NoError(t, err)

				// Create workspace
				req, _ := http.NewRequest("POST",
					fmt.Sprintf("/api/v1/projects/%s/models/%s/workspaces/me", projectIdBase64, modelIdBase64),
					nil)
				req.AddCookie(&http.Cookie{Name: "auth_token", Value: tc.Token})

				w := httptest.NewRecorder()
				tc.Router.ServeHTTP(w, req)
				require.Equal(t, http.StatusCreated, w.Code)

				var response struct {
					Data struct {
						ModelId string `json:"modelId"`
						UserId  string `json:"userId"`
					} `json:"data"`
				}
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
				require.Equal(t, tc.Model1.String(), response.Data.ModelId)
				require.Equal(t, tc.User.ID.String(), response.Data.UserId)

				// Get workspace
				req, _ = http.NewRequest("GET",
					fmt.Sprintf("/api/v1/projects/%s/models/%s/workspaces/me", projectIdBase64, modelIdBase64),
					nil)
				req.AddCookie(&http.Cookie{Name: "auth_token", Value: tc.Token})

				w = httptest.NewRecorder()
				tc.Router.ServeHTTP(w, req)
				require.Equal(t, http.StatusOK, w.Code)
				require.Contains(t, w.Body.String(), tc.Model1.String())
				require.Contains(t, w.Body.String(), tc.Project1.String())
			},
		},
		{
			name: "Create workspace that already exists returns 400",
			run: func(t *testing.T, tc *testContext) {
				projectIdBase64, _ := utils.UuidToBase64(tc.Project1)
				modelIdBase64, _ := utils.UuidToBase64(tc.Model1)

				req, _ := http.NewRequest("POST",
					fmt.Sprintf("/api/v1/projects/%s/models/%s/workspaces/me", projectIdBase64, modelIdBase64),
					nil)
				req.AddCookie(&http.Cookie{Name: "auth_token", Value: tc.Token})
				w := httptest.NewRecorder()
				tc.Router.ServeHTTP(w, req)
				require.Equal(t, http.StatusCreated, w.Code)

				w = httptest.NewRecorder()
				tc.Router.ServeHTTP(w, req)
				require.Equal(t, http.StatusBadRequest, w.Code)
				require.Contains(t, w.Body.String(), "workspace already exists")
			},
		},
		{
			name: "Merge workspace no changes returns 404",
			run: func(t *testing.T, tc *testContext) {
				projectIdBase64, _ := utils.UuidToBase64(tc.Project1)
				modelIdBase64, _ := utils.UuidToBase64(tc.Model1)

				req, _ := http.NewRequest("POST",
					fmt.Sprintf("/api/v1/projects/%s/models/%s/workspaces/merge", projectIdBase64, modelIdBase64),
					nil)
				req.AddCookie(&http.Cookie{Name: "auth_token", Value: tc.Token})
				w := httptest.NewRecorder()
				tc.Router.ServeHTTP(w, req)
				require.Equal(t, http.StatusNotFound, w.Code)
			},
		},
		{
			name: "Merge workspace invalid model ID returns 400",
			run: func(t *testing.T, tc *testContext) {
				projectIdBase64, _ := utils.UuidToBase64(tc.Project1)
				invalidModelId := "!!!invalid"

				req, _ := http.NewRequest("POST",
					fmt.Sprintf("/api/v1/projects/%s/models/%s/workspaces/merge", projectIdBase64, invalidModelId),
					nil)
				req.AddCookie(&http.Cookie{Name: "auth_token", Value: tc.Token})

				w := httptest.NewRecorder()
				tc.Router.ServeHTTP(w, req)
				require.Equal(t, http.StatusBadRequest, w.Code)
				require.Contains(t, w.Body.String(), "invalid project ID")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tc := setupTest(t)
			tt.run(t, tc)
		})
	}
}
