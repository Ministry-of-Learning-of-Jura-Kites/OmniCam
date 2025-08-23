package config

import (
	"os"
	"path"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"omnicam.com/backend/internal"
)

type AppEnv struct {
	Mode        string `env:"MODE"`
	DatabaseUrl string `env:"DATABASE_URL"`
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

	logger.Info("Loaded env successfully")

	return &cfg
}
