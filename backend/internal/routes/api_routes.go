package api_routes

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"
	controller_test "omnicam.com/backend/internal/controllers"
	controller_model "omnicam.com/backend/internal/controllers/models"
	controller_projects "omnicam.com/backend/internal/controllers/projects"
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
}
