package api_routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"

	// controller_test "omnicam.com/backend/internal/controllers"
	"omnicam.com/backend/internal/controllers/authentication"
	controller_files "omnicam.com/backend/internal/controllers/files"
	controller_users "omnicam.com/backend/internal/controllers/users"
	"omnicam.com/backend/internal/middleware"

	controller_camera "omnicam.com/backend/internal/controllers/cameras"
	controller_model "omnicam.com/backend/internal/controllers/models"
	controller_projects "omnicam.com/backend/internal/controllers/projects"
	controller_workspaces "omnicam.com/backend/internal/controllers/workspaces"
	db_client "omnicam.com/backend/pkg/db"
)

type Dependencies struct {
	Logger *zap.Logger
	Env    *config_env.AppEnv
	DB     *db_client.DB
}

func InitRoutes(deps Dependencies, router gin.IRouter) {
	publicRoute := router.Group("/")
	protectedRoute := router.Group("/")
	authMiddleware := middleware.AuthMiddleware{
		Env:    deps.Env,
		Logger: deps.Logger,
	}
	protectedRoute.Use(authMiddleware.CreateHandler())

	deleteProjectRoute := controller_projects.DeleteProjectRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	deleteProjectRoute.InitDeleteProjectRoute(protectedRoute)

	getProjectRoute := controller_projects.GetProjectRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	getProjectRoute.InitGetProjectRoute(protectedRoute)

	postProjectRoute := controller_projects.PostProjectRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	postProjectRoute.InitCreateProjectRoute(protectedRoute)

	updateProjectRoute := controller_projects.PutProjectRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	updateProjectRoute.InitUpdateProjectRoute(protectedRoute)

	postModelRoute := controller_model.PostModelRoutes{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	postModelRoute.InitCreateModelRoute(protectedRoute)

	getModelRoute := controller_model.GetModelRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	getModelRoute.InitGetModelRoute(protectedRoute)

	putModelRoute := controller_model.PutModelRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	putModelRoute.InitUpdateModelRoute(protectedRoute)

	deleteModelRoute := controller_model.DeleteModelRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	deleteModelRoute.InitDeleteModelRoute(protectedRoute)

	cameraAutosaveRoute := controller_camera.CameraAutosaveRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
		Upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return r.Header.Get("Origin") == deps.Env.FrontendHost
			},
		},
	}
	cameraAutosaveRoute.InitRoute(protectedRoute)

	workspaceRoute := controller_workspaces.WorkspaceRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	workspaceRoute.InitRoute(protectedRoute)

	putImageModelRoute := controller_model.PutImageModelRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	putImageModelRoute.InitUpdateImageRoute(protectedRoute)

	putImageProjectRoute := controller_projects.PutImageProjectRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	putImageProjectRoute.InitUpdateImageRoute(protectedRoute)

	registerRoute := authentication.AuthRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	registerRoute.InitRegisterRouter(publicRoute)

	loginRoute := authentication.AuthRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	loginRoute.InitLoginRouter(publicRoute)

	logoutRoute := authentication.AuthRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	logoutRoute.InitLogoutRouter(publicRoute)

	getUserRoute := controller_users.UserRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	getUserRoute.InitUserRouter(protectedRoute)

	GetProjectMembersRoute := controller_model.GetProjectMembersRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	GetProjectMembersRoute.InitProjectMemberRouter(protectedRoute)

	PostProjectMembersRoute := controller_model.PostProjectMembersRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	PostProjectMembersRoute.InitProjectMemberRouter(protectedRoute)

	UsersForAddMembersRoute := controller_model.UsersForAddMembersRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	UsersForAddMembersRoute.InitUserRouter(protectedRoute)

	DeleteProjectMemberRoute := controller_model.DeleteProjectMemberRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	DeleteProjectMemberRoute.InitDeleteProjectMemberRoute(protectedRoute)

	PutUserRoleRoute := controller_model.PutUserRoleRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	PutUserRoleRoute.InitPutUserRoleRoute(protectedRoute)

	meRoute := controller_users.GetMeRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	meRoute.InitGetMeRouter(protectedRoute)

	fileRoute := controller_files.FileRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	fileRoute.InitFileRouter(protectedRoute)
}
