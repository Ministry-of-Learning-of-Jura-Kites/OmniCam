package api_routes

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"
	controller_test "omnicam.com/backend/internal/controllers"
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

	return router
}
