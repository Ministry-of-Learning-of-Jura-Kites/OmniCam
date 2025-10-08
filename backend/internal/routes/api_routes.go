package api_routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"

	// controller_test "omnicam.com/backend/internal/controllers"
	"omnicam.com/backend/internal/controllers/authentication"
	"omnicam.com/backend/internal/middleware"

	controller_camera "omnicam.com/backend/internal/controllers/cameras"
	controller_model "omnicam.com/backend/internal/controllers/models"
	controller_projects "omnicam.com/backend/internal/controllers/projects"
	"omnicam.com/backend/internal/controllers/users"
	controller_workspaces "omnicam.com/backend/internal/controllers/workspaces"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
)

type Dependencies struct {
	Logger *zap.Logger
	Env    *config_env.AppEnv
	DB     *db_sqlc_gen.Queries
}

func InitRoutes(deps Dependencies, router gin.IRouter) {
	publicRoute := router.Group("/")
	protectedRoute := router.Group("/")
	protectedRoute.Use(middleware.AuthMiddleware(deps.Env))

	deleteProjectRoute := controller_projects.DeleteProjectRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	deleteProjectRoute.InitDeleteProjectRoute(publicRoute)

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
	postProjectRoute.InitCreateProjectRoute(publicRoute)

	updateProjectRoute := controller_projects.PutProjectRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	updateProjectRoute.InitUpdateProjectRoute(publicRoute)

	postModelRoute := controller_model.PostModelRoutes{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	postModelRoute.InitCreateModelRoute(publicRoute)

	getModelRoute := controller_model.GetModelRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	getModelRoute.InitGetModelRoute(publicRoute)

	putModelRoute := controller_model.PutModelRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	putModelRoute.InitUpdateModelRoute(publicRoute)

	deleteModelRoute := controller_model.DeleteModelRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	deleteModelRoute.InitDeleteModelRoute(publicRoute)

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
	cameraAutosaveRoute.InitRoute(publicRoute)

	workspaceRoute := controller_workspaces.WorkspaceRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	workspaceRoute.InitRoute(publicRoute)

	putImageModelRoute := controller_model.PutImageModelRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	putImageModelRoute.InitUpdateImageRoute(publicRoute)

	putImageProjectRoute := controller_projects.PutImageProjectRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	putImageProjectRoute.InitUpdateImageRoute(publicRoute)

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

	getUserRoute := users.UserRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	getUserRoute.InitUserRouter(publicRoute)
}
