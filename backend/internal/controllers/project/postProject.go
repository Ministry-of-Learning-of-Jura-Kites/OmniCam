package controller_project

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
)

type PostProjectRoute struct {
	Logger *zap.Logger
	Env    *config_env.AppEnv
	DB     *db_sqlc_gen.Queries
}

type CreateProjectRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

func (t *PostProjectRoute) post(c *gin.Context) {
	var req CreateProjectRequest

	err := c.ShouldBindJSON(&req)
	if err != nil {
		t.Logger.Debug("error while validating body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	data, err := t.DB.CreateProject(c, db_sqlc_gen.CreateProjectParams(req))
	if err != nil {
		t.Logger.Error("error while creating project", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": data})
}

func (t *PostProjectRoute) InitCreateProjectRoute(router gin.IRouter) gin.IRouter {
	router.POST("/project/:projectId", t.post)
	return router
}
