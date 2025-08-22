package controller_project

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
)

type DeleteProjectRoute struct {
	Logger *zap.Logger
	Env    *config_env.AppEnv
	DB     *db_sqlc_gen.Queries
}

func (t *DeleteProjectRoute) delete(c *gin.Context) {
	strId := c.Query("projectId")

	id, err := uuid.Parse(strId)
	if err != nil {
		t.Logger.Error("error while converting str id to uuid", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid projectId"})
		return
	}

	data, err := t.DB.DeleteProject(c, id)
	if err != nil {
		t.Logger.Error("something wrong", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete project"})
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "delete successfully", "data": data})

}

func (t *DeleteProjectRoute) InitDeleteProjectRoute(router gin.IRouter) gin.IRouter {
	router.DELETE("/project/:projectId", t.delete)
	return router
}
