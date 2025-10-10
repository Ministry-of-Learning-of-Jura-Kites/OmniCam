package controller_test

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"

	// db_client "omnicam.com/backend/pkg/db"
	db_client "omnicam.com/backend/pkg/db"
)

type TestRoute struct {
	Logger *zap.Logger
	Env    *config_env.AppEnv
	DB     *db_client.DB
}

func (t *TestRoute) Get(c *gin.Context) {
	t.Logger.Info("testtest")
	c.String(200, "testnaja")
}
