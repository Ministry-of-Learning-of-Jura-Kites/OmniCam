package main

import (
	config_env "omnicam.com/backend/config"
	api_routes "omnicam.com/backend/internal/routes"
	"omnicam.com/backend/pkg/logger"
)

func main() {
	logger := logger.InitLogger()
	defer logger.Sync()
	env := config_env.InitAppEnv(logger)
	router := api_routes.InitRoutes(logger, env)
	router.Run()
}
