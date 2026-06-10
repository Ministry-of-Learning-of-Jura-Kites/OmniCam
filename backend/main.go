package main

import (
	"context"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"
	api_routes "omnicam.com/backend/internal/routes"
	"omnicam.com/backend/internal/utils"
	db_client "omnicam.com/backend/pkg/db"
	"omnicam.com/backend/pkg/logger"
	"omnicam.com/backend/pkg/telemetry"
)

func main() {
	ctx := context.Background()
	utils.RegisterCustomValidations()

	logger := logger.InitLogger(false)
	defer logger.Sync()

	env := config_env.InitAppEnv(logger)

	shutdown := telemetry.InitTelemetry(ctx, logger)
	defer shutdown()

	clientDB := db_client.InitDatabase(env, logger)

	router := gin.Default()
	router.Use(otelgin.Middleware(env.OtelServiceName))
	nc, err := nats.Connect(env.NatsUrl)
	if err != nil {
		logger.Fatal("Error while connecting to nats", zap.Error(err))
	}

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
		DB:     clientDB,
		Nc:     nc,
	}, apiV1)

	router.Run()
}
