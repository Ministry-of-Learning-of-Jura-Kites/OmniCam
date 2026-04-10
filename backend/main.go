package main

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"
	api_routes "omnicam.com/backend/internal/routes"
	"omnicam.com/backend/internal/utils"
	db_client "omnicam.com/backend/pkg/db"
	"omnicam.com/backend/pkg/logger"
)

func main() {
	utils.RegisterCustomValidations()

	logger := logger.InitLogger(false)
	defer logger.Sync()

	env := config_env.InitAppEnv(logger)

	nc, err := nats.Connect(env.NatsUrl)
	if err != nil {
		logger.Fatal("Error while connecting to nats", zap.Error(err))
	}

	client_db := db_client.InitDatabase(env, logger)

	router := gin.Default()

	var allowOrigins []string = []string{env.FrontendHost}

	if env.Mode == "DEV" {
		allowOrigins = append(allowOrigins, "http://localhost:8000")
		logger.Info("Enabled cors for swagger")
	}

	router.Use(cors.New(cors.Config{
		AllowOrigins:     allowOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	apiV1 := router.Group("/api/v1")
	api_routes.InitRoutes(api_routes.Dependencies{
		Logger: logger,
		Env:    env,
		Nc:     nc,
		DB:     client_db,
	}, apiV1)

	router.Run()
}
