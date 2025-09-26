package users

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
)

type UserRoute struct {
	Logger *zap.Logger
	Env    *config_env.AppEnv
	DB     *db_sqlc_gen.Queries
}

func (t *UserRoute) GetAll(c *gin.Context) {
	users, err := t.DB.GetAllUser(c)
	if err != nil {
		t.Logger.Error("failed to fetch users", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch users"})
		return
	}

	// build safe response (no password exposure)
	resp := make([]gin.H, 0, len(users))
	for _, u := range users {
		resp = append(resp, gin.H{
			"id":         u.ID,
			"name":       u.Name,
			"email":      u.Email,
			"created_at": u.CreatedAt,
			"updated_at": u.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (t *UserRoute) InitUserRouter(router gin.IRouter) gin.IRouter {
	router.GET("/users", t.GetAll)
	return router
}
