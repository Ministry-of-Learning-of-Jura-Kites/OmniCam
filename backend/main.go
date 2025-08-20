package main

import (
	config_env "omnicam.com/backend/config"
	api_routes "omnicam.com/backend/internal/routes"
	db_client "omnicam.com/backend/pkg/db"
	"omnicam.com/backend/pkg/logger"
)

func main() {

	logger := logger.InitLogger()
	defer logger.Sync()
	env := config_env.InitAppEnv(logger)
	client_db := db_client.InitDatabase(env)
	router := api_routes.InitRoutes(logger, env, client_db.Queries)
	router.Run()
}
