package api_routes

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"
	controller_test "omnicam.com/backend/internal/controllers"
)

func InitRoutes(logger *zap.Logger, env *config_env.AppEnv) *gin.Engine {
	// router := gin.New()

	// // log gin by zap
	// router.Use(ginzap.Ginzap(logger, time.RFC3339, false))

	// // log panics to zap
	// router.Use(ginzap.RecoveryWithZap(logger, true))

	router := gin.Default()

	testRoute := controller_test.TestRoute{
		Logger: logger,
		Env:    env,
	}
	router.GET("/test", testRoute.Get)

	return router
}
