package api_routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"

	// controller_test "omnicam.com/backend/internal/controllers"
	"omnicam.com/backend/internal/controllers/authentication"

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
	// router := gin.New()

	// // log gin by zap
	// router.Use(ginzap.Ginzap(logger, time.RFC3339, false))

	// // log panics to zap
	// router.Use(ginzap.RecoveryWithZap(logger, true))

	// testRoute := controller_test.TestRoute{
	// 	Logger:  deps.Logger,
	// 	Env:     deps.Env,
	// 	Queries: deps.DB,
	// }

	// getUserRoute := controller_test.GetUserRoute{
	// 	Logger: deps.Logger,
	// 	Env:    deps.Env,
	// 	DB:     deps.DB,
	// }
	// router.GET("/test", testRoute.Get)
	// router.GET("/user", getUserRoute.Get)

	deleteProjectRoute := controller_projects.DeleteProjectRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	deleteProjectRoute.InitDeleteProjectRoute(router)

	getProjectRoute := controller_projects.GetProjectRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	getProjectRoute.InitGetProjectRoute(router)

	postProjectRoute := controller_projects.PostProjectRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	postProjectRoute.InitCreateProjectRoute(router)

	updateProjectRoute := controller_projects.PutProjectRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	updateProjectRoute.InitUpdateProjectRoute(router)

	postModelRoute := controller_model.PostModelRoutes{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	postModelRoute.InitCreateModelRoute(router)

	getModelRoute := controller_model.GetModelRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	getModelRoute.InitGetModelRoute(router)

	putModelRoute := controller_model.PutModelRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	putModelRoute.InitUpdateModelRoute(router)

	deleteModelRoute := controller_model.DeleteModelRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	deleteModelRoute.InitDeleteModelRoute(router)

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
	cameraAutosaveRoute.InitRoute(router)

	workspaceRoute := controller_workspaces.WorkspaceRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	workspaceRoute.InitRoute(router)

	putImageModelRoute := controller_model.PutImageModelRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	putImageModelRoute.InitUpdateImageRoute(router)

	putImageProjectRoute := controller_projects.PutImageProjectRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	putImageProjectRoute.InitUpdateImageRoute(router)

	registerRoute := authentication.AuthRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	registerRoute.InitRegisterRouter(router)

	loginRoute := authentication.AuthRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	loginRoute.InitLoginRouter(router)

	logoutRoute := authentication.AuthRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	logoutRoute.InitLogoutRouter(router)

	getUserRoute := users.UserRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	getUserRoute.InitUserRouter(router)
}
