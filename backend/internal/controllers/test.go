package controller_test

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"

	// db_client "omnicam.com/backend/pkg/db"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
)

type TestRoute struct {
	Logger  *zap.Logger
	Env     *config_env.AppEnv
	Queries *db_sqlc_gen.Queries
}

func (t *TestRoute) Get(c *gin.Context) {
	t.Logger.Info("testtest")
	c.String(200, "testnaja")
}
