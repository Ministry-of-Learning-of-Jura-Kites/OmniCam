package api_routes

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"
	controller_test "omnicam.com/backend/internal/controllers"
	controller_project "omnicam.com/backend/internal/controllers/project"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
)

func InitRoutes(logger *zap.Logger, env *config_env.AppEnv, db *db_sqlc_gen.Queries) *gin.Engine {
	// router := gin.New()

	// // log gin by zap
	// router.Use(ginzap.Ginzap(logger, time.RFC3339, false))

	// // log panics to zap
	// router.Use(ginzap.RecoveryWithZap(logger, true))

	router := gin.Default()

	testRoute := controller_test.TestRoute{
		Logger:  logger,
		Env:     env,
		Queries: db,
	}

	getUserRoute := controller_test.GetUserRoute{
		Logger: logger,
		Env:    env,
		DB:     db,
	}
	router.GET("/test", testRoute.Get)
	router.GET("/user", getUserRoute.Get)

	deleteProjectRoute := controller_project.DeleteProjectRoute{
		Logger: logger,
		Env:    env,
		DB:     db,
	}
	deleteProjectRoute.InitDeleteProjectRoute(router)

	getProjectRoute := controller_project.GetProjectRoute{
		Logger: logger,
		Env:    env,
		DB:     db,
	}
	getProjectRoute.InitGetProjectRoute(router)

	postProjectRoute := controller_project.PostProjectRoute{
		Logger: logger,
		Env:    env,
		DB:     db,
	}
	postProjectRoute.InitCreateProjectRoute(router)

	updateProjectRoute := controller_project.PutProjectRoute{
		Logger: logger,
		Env:    env,
		DB:     db,
	}
	updateProjectRoute.InitUpdateProjectRoute(router)

	return router
}
