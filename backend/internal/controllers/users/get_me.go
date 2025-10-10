package controller_users

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
)

type GetMeRoute struct {
	Logger *zap.Logger
	Env    *config_env.AppEnv
	DB     *db_sqlc_gen.Queries
}

func (t *GetMeRoute) GetMe(c *gin.Context) {
	username, exists := c.Get("username")
	if !exists {
		t.Logger.Error("username not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": username})
}

func (t *GetMeRoute) InitGetMeRouter(router gin.IRouter) gin.IRouter {
	router.GET("/me", t.GetMe)
	return router
}
