package main

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	config_env "omnicam.com/backend/config"
	api_routes "omnicam.com/backend/internal/routes"
	"omnicam.com/backend/internal/utils"
	db_client "omnicam.com/backend/pkg/db"
	"omnicam.com/backend/pkg/logger"
)

func StartResponseListener(redisClient *redis.Client, env *config_env.AppEnv, responseRegistry *sync.Map) {
	go func() {
		ctx := context.Background()
		for {
			// Read from the single response stream
			entries, err := redisClient.XRead(ctx, &redis.XReadArgs{
				Streams: []string{env.OptiResTopic, "$"},
				Block:   0, // Block indefinitely for new results
				Count:   10,
			}).Result()

			if err != nil {
				continue
			}

			for _, stream := range entries {
				for _, msg := range stream.Messages {
					rawJSON, ok := msg.Values["data"].(string)

					if !ok {
						// TODO: Error handling
						continue
					}

					// Peek at the Job ID in the JSON
					var temp struct {
						JobID string `json:"jobId"`
					}
					json.Unmarshal([]byte(rawJSON), &temp)

					// Find the waiting handler
					if ch, ok := responseRegistry.Load(temp.JobID); ok {
						// Send the raw data to that specific handler's channel
						ch.(chan string) <- rawJSON
						// Cleanup the registry
						responseRegistry.Delete(temp.JobID)
					}
				}
			}
		}
	}()
}

func main() {
	utils.RegisterCustomValidations()

	logger := logger.InitLogger(false)
	defer logger.Sync()

	env := config_env.InitAppEnv(logger)

	client_db := db_client.InitDatabase(env)

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	var optimizeRespMap sync.Map

	StartResponseListener(redisClient, env, &optimizeRespMap)

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
		Logger:          logger,
		Env:             env,
		DB:              client_db,
		RedisClient:     redisClient,
		OptimizeRespMap: &optimizeRespMap,
	}, apiV1)

	router.Run()
}
