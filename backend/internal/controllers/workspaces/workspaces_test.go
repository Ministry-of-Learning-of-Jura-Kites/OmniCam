//go:build unit_test
// +build unit_test

package controller_workspaces_test

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
	config_env "omnicam.com/backend/config"
	controller_workspaces "omnicam.com/backend/internal/controllers/workspaces"
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
func setupTest(t *testing.T, source string) *testContext {
	t.Helper()

	testcaseLogger := testLogger.With(zap.String("testcase", source))

	ctx := context.Background()
	env := config_env.InitAppEnv(testcaseLogger)
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
	authMiddleware := middleware.AuthMiddleware{Env: env, Logger: testcaseLogger}
	protected.Use(authMiddleware.CreateHandler())

	route := controller_workspaces.WorkspaceRoute{
		Logger: testcaseLogger,
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

func TestWorkspacesMe(t *testing.T) {
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
					fmt.Sprintf("/api/v1/projects/%s/models/%s/workspaces/me/merge", projectIdBase64, modelIdBase64),
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
					fmt.Sprintf("/api/v1/projects/%s/models/%s/workspaces/me/merge", projectIdBase64, invalidModelId),
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
			tc := setupTest(t, tt.name)
			tt.run(t, tc)
		})
	}
}

func TestPostMergeWorkspace(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, tc *testContext)
	}{
		{
			name: "Invalid model ID returns 400",
			run: func(t *testing.T, tc *testContext) {
				projectIdBase64, _ := utils.UuidToBase64(tc.Project1)
				invalidModelId := "invalid-id"

				req, _ := http.NewRequest("POST",
					fmt.Sprintf("/api/v1/projects/%s/models/%s/workspaces/me/merge", projectIdBase64, invalidModelId),
					nil)
				req.AddCookie(&http.Cookie{Name: "auth_token", Value: tc.Token})

				w := httptest.NewRecorder()
				tc.Router.ServeHTTP(w, req)

				require.Equal(t, http.StatusBadRequest, w.Code)
				require.Contains(t, w.Body.String(), "invalid project ID")
			},
		},
		{
			name: "Workspace not found returns 404",
			run: func(t *testing.T, tc *testContext) {
				projectIdBase64, _ := utils.UuidToBase64(tc.Project1)
				modelIdBase64, _ := utils.UuidToBase64(tc.Model1)

				req, _ := http.NewRequest("POST",
					fmt.Sprintf("/api/v1/projects/%s/models/%s/workspaces/me/merge", projectIdBase64, modelIdBase64),
					nil)
				req.AddCookie(&http.Cookie{Name: "auth_token", Value: tc.Token})

				w := httptest.NewRecorder()
				tc.Router.ServeHTTP(w, req)

				require.Equal(t, http.StatusNotFound, w.Code)
			},
		},
		{
			name: "Workspace with no changes returns noChanges=true",
			run: func(t *testing.T, tc *testContext) {
				// Setup a workspace where version == baseVersion
				projectIdBase64, _ := utils.UuidToBase64(tc.Project1)
				modelIdBase64, _ := utils.UuidToBase64(tc.Model1)

				_, err := tc.DB.Queries.AddUserToProject(tc.Ctx, db_sqlc_gen.AddUserToProjectParams{
					UserID: tc.User.ID, ProjectID: tc.Project1, Role: db_sqlc_gen.RoleCollaborator,
				})
				require.NoError(t, err)

				// Create workspace manually in DB with equal versions
				_, err = tc.DB.Queries.CreateWorkspace(tc.Ctx, db_sqlc_gen.CreateWorkspaceParams{
					UserID:  tc.User.ID,
					ModelID: tc.Model1,
				})

				req, _ := http.NewRequest("POST",
					fmt.Sprintf("/api/v1/projects/%s/models/%s/workspaces/me/merge", projectIdBase64, modelIdBase64),
					nil)
				req.AddCookie(&http.Cookie{Name: "auth_token", Value: tc.Token})
				w := httptest.NewRecorder()
				tc.Router.ServeHTTP(w, req)

				require.Equal(t, http.StatusOK, w.Code)
				require.Contains(t, w.Body.String(), `"noChanges":true`)
			},
		},
		{
			name: "Model version equal to workspace baseVersion merges successfully",
			run: func(t *testing.T, tc *testContext) {
				projectIdBase64, _ := utils.UuidToBase64(tc.Project1)
				modelIdBase64, _ := utils.UuidToBase64(tc.Model1)

				_, err := tc.DB.Queries.AddUserToProject(tc.Ctx, db_sqlc_gen.AddUserToProjectParams{
					UserID: tc.User.ID, ProjectID: tc.Project1, Role: db_sqlc_gen.RoleCollaborator,
				})
				require.NoError(t, err)

				// Create model and workspace with equal baseVersion

				_, err = tc.DB.Queries.CreateWorkspace(tc.Ctx, db_sqlc_gen.CreateWorkspaceParams{
					UserID:  tc.User.ID,
					ModelID: tc.Model1,
				})
				require.NoError(t, err)

				_, err = tc.DB.Queries.UpdateWorkspaceCams(tc.Ctx, db_sqlc_gen.UpdateWorkspaceCamsParams{
					Key:     []string{"123"},
					Value:   []byte("{}"),
					UserID:  tc.User.ID,
					ModelID: tc.Model1,
				})
				require.NoError(t, err)

				req, _ := http.NewRequest("POST",
					fmt.Sprintf("/api/v1/projects/%s/models/%s/workspaces/me/merge", projectIdBase64, modelIdBase64),
					nil)
				req.AddCookie(&http.Cookie{Name: "auth_token", Value: tc.Token})
				w := httptest.NewRecorder()
				tc.Router.ServeHTTP(w, req)

				require.Equal(t, http.StatusOK, w.Code)
				require.Contains(t, w.Body.String(), `"noChanges":false`)
			},
		},
		{
			name: "Model version is higher than workspace baseVersion merges successfully",
			run: func(t *testing.T, tc *testContext) {
				projectIdBase64, _ := utils.UuidToBase64(tc.Project1)
				modelIdBase64, _ := utils.UuidToBase64(tc.Model1)

				_, err := tc.DB.Queries.AddUserToProject(tc.Ctx, db_sqlc_gen.AddUserToProjectParams{
					UserID: tc.User.ID, ProjectID: tc.Project1, Role: db_sqlc_gen.RoleCollaborator,
				})
				require.NoError(t, err)

				_, err = tc.DB.Queries.UpdateModelCams(tc.Ctx, db_sqlc_gen.UpdateModelCamsParams{
					Value:   []byte(`{"123":{"posX":1}}`),
					ModelID: tc.Model1,
				})
				require.NoError(t, err)

				// Create model and workspace with equal baseVersion
				_, err = tc.DB.Queries.CreateWorkspace(tc.Ctx, db_sqlc_gen.CreateWorkspaceParams{
					UserID:  tc.User.ID,
					ModelID: tc.Model1,
				})
				require.NoError(t, err)

				_, err = tc.DB.Queries.UpdateWorkspaceCams(tc.Ctx, db_sqlc_gen.UpdateWorkspaceCamsParams{
					Key:     []string{"123"},
					Value:   []byte(`{"posX":10}`),
					UserID:  tc.User.ID,
					ModelID: tc.Model1,
				})
				require.NoError(t, err)

				_, err = tc.DB.Queries.UpdateWorkspaceCams(tc.Ctx, db_sqlc_gen.UpdateWorkspaceCamsParams{
					Key:     []string{"456"},
					Value:   []byte(`{"posX":10}`),
					UserID:  tc.User.ID,
					ModelID: tc.Model1,
				})
				require.NoError(t, err)

				_, err = tc.DB.Queries.UpdateModelCams(tc.Ctx, db_sqlc_gen.UpdateModelCamsParams{
					Value:   []byte(`{"123":{"posX":12}}`),
					ModelID: tc.Model1,
				})
				require.NoError(t, err)

				req, _ := http.NewRequest("POST",
					fmt.Sprintf("/api/v1/projects/%s/models/%s/workspaces/me/merge", projectIdBase64, modelIdBase64),
					nil)
				req.AddCookie(&http.Cookie{Name: "auth_token", Value: tc.Token})
				w := httptest.NewRecorder()
				tc.Router.ServeHTTP(w, req)

				require.Equal(t, http.StatusOK, w.Code)
				require.Contains(t, w.Body.String(), `"conflicts":{"123":{"posX":{"base":1,"main":12,"workspace":10}}}`)
			},
		},
		// {
		// 	name: "Workspace version ahead of model returns 500",
		// 	run: func(t *testing.T, tc *testContext) {
		// 		projectIdBase64, _ := utils.UuidToBase64(tc.Project1)
		// 		modelIdBase64, _ := utils.UuidToBase64(tc.Model1)

		// 		_, err := tc.DB.Queries.CreateModel(tc.Ctx, db_sqlc_gen.CreateModelParams{
		// 			ID: tc.Model1, ProjectID: tc.Project1,
		// 		})
		// 		require.NoError(t, err)

		// 		_, err = tc.DB.Queries.CreateWorkspace(tc.Ctx, db_sqlc_gen.CreateWorkspaceParams{
		// 			UserID:  tc.User.ID,
		// 			ModelID: tc.Model1,
		// 		})
		// 		require.NoError(t, err)

		// 		req, _ := http.NewRequest("POST",
		// 			fmt.Sprintf("/api/v1/projects/%s/models/%s/workspaces/me/merge", projectIdBase64, modelIdBase64),
		// 			nil)
		// 		req.AddCookie(&http.Cookie{Name: "auth_token", Value: tc.Token})
		// 		w := httptest.NewRecorder()
		// 		tc.Router.ServeHTTP(w, req)

		// 		require.Equal(t, http.StatusInternalServerError, w.Code)
		// 	},
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tc := setupTest(t, tt.name)
			tt.run(t, tc)
		})
	}
}
func TestPostResolveWorkspaceMe(t *testing.T) {
	tests := []struct {
		name string
		body string
		run  func(t *testing.T, tc *testContext)
	}{
		{
			name: "Invalid model ID returns 400",
			run: func(t *testing.T, tc *testContext) {
				projectIdBase64, _ := utils.UuidToBase64(tc.Project1)
				invalidModelId := "!!!bad"
				// userIdBase64, _ := utils.UuidToBase64(tc.User.ID)

				req, _ := http.NewRequest("POST",
					fmt.Sprintf("/api/v1/projects/%s/models/%s/workspaces/me/resolve", projectIdBase64, invalidModelId),
					nil)
				req.AddCookie(&http.Cookie{Name: "auth_token", Value: tc.Token})

				w := httptest.NewRecorder()
				tc.Router.ServeHTTP(w, req)

				require.Equal(t, http.StatusBadRequest, w.Code)
				require.Contains(t, w.Body.String(), "invalid model ID")
			},
		},
		{
			name: "Workspace not found returns 404",
			run: func(t *testing.T, tc *testContext) {
				projectIdBase64, err := utils.UuidToBase64(tc.Project1)
				require.Nil(t, err)
				modelIdBase64, err := utils.UuidToBase64(tc.Model1)
				require.Nil(t, err)
				// userIdBase64, _ := utils.UuidToBase64(tc.User.ID)

				req, _ := http.NewRequest("POST",
					fmt.Sprintf("/api/v1/projects/%s/models/%s/workspaces/me/resolve", projectIdBase64, modelIdBase64),
					nil)
				req.AddCookie(&http.Cookie{Name: "auth_token", Value: tc.Token})

				w := httptest.NewRecorder()
				tc.Router.ServeHTTP(w, req)

				require.Equal(t, http.StatusNotFound, w.Code)
			},
		},
		{
			name: "Invalid JSON body returns 400",
			body: `invalid-json`,
			run: func(t *testing.T, tc *testContext) {
				projectIdBase64, _ := utils.UuidToBase64(tc.Project1)
				modelIdBase64, _ := utils.UuidToBase64(tc.Model1)
				// userIdBase64, _ := utils.UuidToBase64(tc.User.ID)

				// Create workspace and model
				_, err := tc.DB.Queries.CreateWorkspace(tc.Ctx, db_sqlc_gen.CreateWorkspaceParams{
					UserID:  tc.User.ID,
					ModelID: tc.Model1,
				})
				require.NoError(t, err)

				_, err = tc.DB.Queries.UpdateWorkspaceCams(tc.Ctx, db_sqlc_gen.UpdateWorkspaceCamsParams{
					Key:     []string{"123"},
					Value:   []byte("{}"),
					UserID:  tc.User.ID,
					ModelID: tc.Model1,
				})
				require.NoError(t, err)

				req, _ := http.NewRequest("POST",
					fmt.Sprintf("/api/v1/projects/%s/models/%s/workspaces/me/resolve",
						projectIdBase64, modelIdBase64),
					strings.NewReader(`invalid-json`))
				req.Header.Set("Content-Type", "application/json")
				req.AddCookie(&http.Cookie{Name: "auth_token", Value: tc.Token})

				w := httptest.NewRecorder()
				tc.Router.ServeHTTP(w, req)

				require.Equal(t, http.StatusBadRequest, w.Code)
				require.Contains(t, w.Body.String(), "invalid character")
			},
		},
		{
			name: "Successfully resolved all conflicts returns 200",
			body: `{"merged":{"CameraA":{"Field1":"sameValue"}}}`,
			run: func(t *testing.T, tc *testContext) {
				projectIdBase64, err := utils.UuidToBase64(tc.Project1)
				require.Nil(t, err)
				modelIdBase64, err := utils.UuidToBase64(tc.Model1)
				require.Nil(t, err)
				// userIdBase64, _ := utils.UuidToBase64(tc.User.ID)

				modelCams := []byte(`{"123":{"angleX":1}}`)
				workspaceCams := []byte(`{"angleX":2}`)

				// Create model and workspace with same data
				_, err = tc.DB.Queries.UpdateModelCams(tc.Ctx, db_sqlc_gen.UpdateModelCamsParams{
					ModelID: tc.Model1,
					Value:   modelCams,
				})
				require.NoError(t, err)

				_, err = tc.DB.Queries.CreateWorkspace(tc.Ctx, db_sqlc_gen.CreateWorkspaceParams{
					UserID:  tc.User.ID,
					ModelID: tc.Model1,
				})
				require.NoError(t, err)

				_, err = tc.DB.Queries.UpdateWorkspaceCams(tc.Ctx, db_sqlc_gen.UpdateWorkspaceCamsParams{
					Key:     []string{"123"},
					Value:   workspaceCams,
					UserID:  tc.User.ID,
					ModelID: tc.Model1,
				})
				require.NoError(t, err)

				req, _ := http.NewRequest("POST",
					fmt.Sprintf("/api/v1/projects/%s/models/%s/workspaces/me/resolve",
						projectIdBase64, modelIdBase64),
					strings.NewReader(`{"merged":{"123":{"angleX":1}}}`))
				req.Header.Set("Content-Type", "application/json")
				req.AddCookie(&http.Cookie{Name: "auth_token", Value: tc.Token})

				w := httptest.NewRecorder()
				tc.Router.ServeHTTP(w, req)

				testLogger.Info(w.Body.String())
				require.Equal(t, http.StatusOK, w.Code)
			},
		},
		{
			name: "Unresolved conflicts returns 400",
			body: `{"merged":{"CameraA":{"Field1":"differentValue"}}}`,
			run: func(t *testing.T, tc *testContext) {
				projectIdBase64, _ := utils.UuidToBase64(tc.Project1)
				modelIdBase64, _ := utils.UuidToBase64(tc.Model1)
				// userIdBase64, _ := utils.UuidToBase64(tc.User.ID)

				modelCams := []byte(`{"CameraA":{"angleX":1}}`)
				workspaceCams := []byte(`{"angleX":2}`)

				_, err := tc.DB.Queries.UpdateModelCams(tc.Ctx, db_sqlc_gen.UpdateModelCamsParams{
					ModelID: tc.Model1,
					Value:   modelCams,
				})
				require.NoError(t, err)

				_, err = tc.DB.Queries.CreateWorkspace(tc.Ctx, db_sqlc_gen.CreateWorkspaceParams{
					UserID:  tc.User.ID,
					ModelID: tc.Model1,
				})
				require.NoError(t, err)

				_, err = tc.DB.Queries.UpdateWorkspaceCams(tc.Ctx, db_sqlc_gen.UpdateWorkspaceCamsParams{
					Key:     []string{"CameraA"},
					Value:   workspaceCams,
					UserID:  tc.User.ID,
					ModelID: tc.Model1,
				})
				require.NoError(t, err)

				req, _ := http.NewRequest("POST",
					fmt.Sprintf("/api/v1/projects/%s/models/%s/workspaces/me/resolve",
						projectIdBase64, modelIdBase64),
					strings.NewReader(`{"merged":{}}`))
				req.Header.Set("Content-Type", "application/json")
				req.AddCookie(&http.Cookie{Name: "auth_token", Value: tc.Token})

				w := httptest.NewRecorder()
				tc.Router.ServeHTTP(w, req)

				require.Equal(t, http.StatusBadRequest, w.Code)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tc := setupTest(t, tt.name)
			tt.run(t, tc)
		})
	}
}
