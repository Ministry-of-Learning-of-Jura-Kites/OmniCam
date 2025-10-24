package controller_users

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"
	db_client "omnicam.com/backend/pkg/db"
)

type GetMeRoute struct {
	Logger *zap.Logger
	Env    *config_env.AppEnv
	DB     *db_client.DB
}

func (t *GetMeRoute) GetMe(c *gin.Context) {
	username, exists := c.Get("username")
	if !exists {
		t.Logger.Error("username not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, err := t.DB.Queries.GetUserByUsername(c, username.(string))
	if err != nil {
		t.Logger.Error("failed to fetch user", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"id":         user.ID,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"username":   user.Username,
			"email":      user.Email,
		},
	})
}

func (t *GetMeRoute) InitGetMeRouter(router gin.IRouter) gin.IRouter {
	router.GET("/me", t.GetMe)
	return router
}
