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

func InitAppEnv(logger *zap.Logger) *AppEnv {
	// For dev env (use getenv because this is before loading .env file)
	if os.Getenv("MODE") == "" || os.Getenv("MODE") == "DEV" {
		godotenv.Load(path.Join(internal.Root, "backend", ".env"))
	}

	var cfg AppEnv
	// parse with generics
	cfg, err := env.ParseAsWithOptions[AppEnv](env.Options{
		RequiredIfNoDef: true,
	})

	if err != nil {
		logger.Fatal("Error while reading env", zap.Error(err))
	}

	dur, err := time.ParseDuration(cfg.RawJWTExpireTime)
	if err != nil {
		logger.Fatal("Invalid JWT_EXPIRE_TIME format", zap.Error(err))
	}
	cfg.JWTExpireTime = dur

	logger.Info("Loaded env successfully")

	return &cfg
}
