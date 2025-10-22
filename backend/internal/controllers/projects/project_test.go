package controller_projects_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	config_env "omnicam.com/backend/config"
	controller_projects "omnicam.com/backend/internal/controllers/projects"
	"omnicam.com/backend/internal/middleware"
	"omnicam.com/backend/internal/testutils"
	"omnicam.com/backend/internal/utils"
	db_client "omnicam.com/backend/pkg/db"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
)

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

func setupProjectTest(t *testing.T, logger *zap.Logger) *testContext {
	t.Helper()
	ctx := context.Background()

	env := config_env.InitAppEnv(logger)
	env.JWTSecret = "123"
	env.JWTExpireTime = 168 * time.Hour

	dbName, conn, err, cleanup := testutils.GetTestDb(ctx, env)
	require.NoError(t, err)

	db := &db_client.DB{
		Queries: db_sqlc_gen.New(conn),
		Pool:    conn,
	}

	// --- Seed user ---
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

	token, err := utils.GenerateJWT(user.FirstName, user.LastName, user.ID.String(), user.Username, env.JWTSecret, 168*time.Hour)
	require.NoError(t, err)

	// --- Router ---
	router := gin.Default()
	apiV1 := router.Group("/api/v1")
	protected := apiV1.Group("/")
	authMiddleware := middleware.AuthMiddleware{Env: env, Logger: logger}
	protected.Use(authMiddleware.CreateHandler())

	deleteProjectRoute := controller_projects.DeleteProjectRoute{
		Logger: logger,
		Env:    env,
		DB:     db,
	}
	deleteProjectRoute.InitDeleteProjectRoute(protected)

	getProjectRoute := controller_projects.GetProjectRoute{
		Logger: logger,
		Env:    env,
		DB:     db,
	}
	getProjectRoute.InitGetProjectRoute(protected)

	postProjectRoute := controller_projects.PostProjectRoute{
		Logger: logger,
		Env:    env,
		DB:     db,
	}
	postProjectRoute.InitCreateProjectRoute(protected)

	updateProjectRoute := controller_projects.PutProjectRoute{
		Logger: logger,
		Env:    env,
		DB:     db,
	}
	updateProjectRoute.InitUpdateProjectRoute(protected)

	t.Cleanup(func() { cleanup(t) })

	return &testContext{
		Ctx:    ctx,
		Env:    env,
		DB:     db,
		User:   user,
		Token:  token,
		Router: router,
		DBName: dbName,
	}
}

func TestProjectsCRUD(t *testing.T) {
	var testLogger = zaptest.NewLogger(t)

	tests := []struct {
		name string
		run  func(t *testing.T, tc *testContext)
	}{
		{
			name: "Create project successfully",
			run: func(t *testing.T, tc *testContext) {
				reqBody := `{"name":"Project Alpha","description":"Test project"}`
				req, _ := http.NewRequest("POST", "/api/v1/projects", strings.NewReader(reqBody))
				req.Header.Set("Content-Type", "application/json")
				req.AddCookie(&http.Cookie{Name: "auth_token", Value: tc.Token})

				w := httptest.NewRecorder()
				tc.Router.ServeHTTP(w, req)

				require.Equal(t, http.StatusOK, w.Code)

				var resp struct {
					Data struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"data"`
				}
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				require.Equal(t, "Project Alpha", resp.Data.Name)
			},
		},
		{
			name: "Get project by ID",
			run: func(t *testing.T, tc *testContext) {
				// Create project first
				projectID := uuid.New()
				_, err := tc.DB.Queries.CreateProject(tc.Ctx, db_sqlc_gen.CreateProjectParams{
					ID: projectID, Name: "Project Beta",
				})
				require.NoError(t, err)

				projectIdBase64, err := utils.UuidToBase64(projectID)
				require.NoError(t, err)

				req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/projects/%s", projectIdBase64), nil)
				req.AddCookie(&http.Cookie{Name: "auth_token", Value: tc.Token})

				w := httptest.NewRecorder()
				tc.Router.ServeHTTP(w, req)

				require.Equal(t, http.StatusOK, w.Code)
				require.Contains(t, w.Body.String(), "Project Beta")
			},
		},
		{
			name: "Update project successfully",
			run: func(t *testing.T, tc *testContext) {
				projectID := uuid.New()
				_, err := tc.DB.Queries.CreateProject(tc.Ctx, db_sqlc_gen.CreateProjectParams{
					ID: projectID, Name: "Project Gamma",
				})
				require.NoError(t, err)

				_, err = tc.DB.Queries.AddUserToProject(tc.Ctx, db_sqlc_gen.AddUserToProjectParams{
					UserID:    tc.User.ID,
					ProjectID: projectID,
					Role:      db_sqlc_gen.RoleOwner,
				})
				require.NoError(t, err)

				projectIdBase64, err := utils.UuidToBase64(projectID)
				require.NoError(t, err)

				updateBody := `{"name":"Project Gamma Updated"}`
				req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/v1/projects/%s", projectIdBase64), strings.NewReader(updateBody))
				req.Header.Set("Content-Type", "application/json")
				req.AddCookie(&http.Cookie{Name: "auth_token", Value: tc.Token})

				w := httptest.NewRecorder()
				tc.Router.ServeHTTP(w, req)

				require.Equal(t, http.StatusOK, w.Code)
				require.Contains(t, w.Body.String(), "Project Gamma Updated")
			},
		},
		{
			name: "Update project successfully",
			run: func(t *testing.T, tc *testContext) {
				projectID := uuid.New()
				_, err := tc.DB.Queries.CreateProject(tc.Ctx, db_sqlc_gen.CreateProjectParams{
					ID: projectID, Name: "Project Gamma",
				})
				require.NoError(t, err)

				_, err = tc.DB.Queries.AddUserToProject(tc.Ctx, db_sqlc_gen.AddUserToProjectParams{
					UserID:    tc.User.ID,
					ProjectID: projectID,
					Role:      db_sqlc_gen.RoleOwner,
				})
				require.NoError(t, err)

				projectIdBase64, err := utils.UuidToBase64(projectID)
				require.NoError(t, err)

				updateBody := `{"name":"Project Gamma Updated"}`
				req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/v1/projects/%s", projectIdBase64), strings.NewReader(updateBody))
				req.Header.Set("Content-Type", "application/json")
				req.AddCookie(&http.Cookie{Name: "auth_token", Value: tc.Token})

				w := httptest.NewRecorder()
				tc.Router.ServeHTTP(w, req)

				require.Equal(t, http.StatusOK, w.Code)
				require.Contains(t, w.Body.String(), "Project Gamma Updated")
			},
		},
		{
			name: "Delete project successfully",
			run: func(t *testing.T, tc *testContext) {
				projectID := uuid.New()
				_, err := tc.DB.Queries.CreateProject(tc.Ctx, db_sqlc_gen.CreateProjectParams{
					ID: projectID, Name: "Project Delta",
				})
				require.NoError(t, err)

				_, err = tc.DB.Queries.AddUserToProject(tc.Ctx, db_sqlc_gen.AddUserToProjectParams{
					UserID:    tc.User.ID,
					ProjectID: projectID,
					Role:      db_sqlc_gen.RoleOwner,
				})
				require.NoError(t, err)

				projectIdBase64, err := utils.UuidToBase64(projectID)
				require.NoError(t, err)

				req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/projects/%s", projectIdBase64), nil)
				req.AddCookie(&http.Cookie{Name: "auth_token", Value: tc.Token})

				w := httptest.NewRecorder()
				tc.Router.ServeHTTP(w, req)
				require.Equal(t, http.StatusOK, w.Code)

				// Verify deletion
				_, err = tc.DB.Queries.GetProjectById(tc.Ctx, projectID)
				require.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			logger := zaptest.NewLogger(t)
			tc := setupProjectTest(t, logger)
			tt.run(t, tc)

			t.Cleanup(func() {
				if !t.Failed() {
					testLogger.Info("Test passed", zap.String("test", tt.name))
				}
			})
		})
	}
}
