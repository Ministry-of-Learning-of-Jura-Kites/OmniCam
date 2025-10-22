package config

import (
	"os"
	"path"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"omnicam.com/backend/internal"
)

type AppEnv struct {
	Mode             string `env:"MODE"`
	DatabaseUrl      string `env:"DATABASE_URL"`
	ModelFilePath    string `env:"MODEL_FILE_PATH"`
	FrontendHost     string `env:"FRONTEND_HOST"`
	JWTSecret        string `env:"JWT_SECRET"`
	RawJWTExpireTime string `env:"JWT_EXPIRE_TIME"`
	JWTExpireTime    time.Duration
}

func transformAppEnv(logger *zap.Logger, cfg *AppEnv, isTest bool) {
	dur, err := time.ParseDuration(cfg.RawJWTExpireTime)
	if err != nil && !isTest {
		logger.Fatal("Invalid JWT_EXPIRE_TIME format", zap.Error(err))
	}
	cfg.JWTExpireTime = dur
}

func InitAppEnv(logger *zap.Logger) *AppEnv {
	onSuccess := func() {
		logger.Info("Loaded env successfully")
	}

	mode := os.Getenv("MODE")

	// For test, use env of process, and don't require all env
	if mode == "TEST" {
		var cfg AppEnv
		cfg, err := env.ParseAsWithOptions[AppEnv](env.Options{
			RequiredIfNoDef: false,
		})
		if err != nil {
			logger.Fatal("Error while reading env", zap.Error(err))
		}
		transformAppEnv(logger, &cfg, true)
		onSuccess()
		return &cfg
	}

	// For dev env (use getenv because this is before loading .env file)
	if mode == "" || mode == "DEV" {
		godotenv.Load(path.Join(internal.Root, "backend", ".env"))
	}

	var cfg AppEnv
	cfg, err := env.ParseAsWithOptions[AppEnv](env.Options{
		RequiredIfNoDef: true,
	})
	if err != nil {
		logger.Fatal("Error while reading env", zap.Error(err))
	}

	transformAppEnv(logger, &cfg, false)
	onSuccess()

	return &cfg
}
