package controller_test

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"
)

type TestRoute struct {
	Logger *zap.Logger
	Env    *config_env.AppEnv
}

func (t *TestRoute) Get(c *gin.Context) {
	t.Logger.Info("testtest")
	c.String(200, "testnaja")
}
