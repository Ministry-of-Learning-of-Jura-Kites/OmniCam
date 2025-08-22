package api_routes

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"
	controller_test "omnicam.com/backend/internal/controllers"
	controller_project "omnicam.com/backend/internal/controllers/project"
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

	testRoute := controller_test.TestRoute{
		Logger:  deps.Logger,
		Env:     deps.Env,
		Queries: deps.DB,
	}

	getUserRoute := controller_test.GetUserRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	router.GET("/test", testRoute.Get)
	router.GET("/user", getUserRoute.Get)

	deleteProjectRoute := controller_project.DeleteProjectRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	deleteProjectRoute.InitDeleteProjectRoute(router)

	getProjectRoute := controller_project.GetProjectRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	getProjectRoute.InitGetProjectRoute(router)

	postProjectRoute := controller_project.PostProjectRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	postProjectRoute.InitCreateProjectRoute(router)

	updateProjectRoute := controller_project.PutProjectRoute{
		Logger: deps.Logger,
		Env:    deps.Env,
		DB:     deps.DB,
	}
	updateProjectRoute.InitUpdateProjectRoute(router)
}
