package controller_project

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
)

type GetProjectRoute struct {
	Logger *zap.Logger
	Env    *config_env.AppEnv
	DB     *db_sqlc_gen.Queries
}

func (t *GetProjectRoute) getAll(c *gin.Context) {
	project, err := t.DB.GetAllProjects(c)

	if err != nil {
		t.Logger.Error("Error while getting all project", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": project})
}

func (t *GetProjectRoute) getById(c *gin.Context) {
	strId := c.Query("projectId")

	id, err := uuid.Parse(strId)
	if err != nil {
		t.Logger.Error("error while converting str id to uuid", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid projectId"})
		return
	}

	project, err := t.DB.GetProjectById(c, id)
	if err != nil {
		t.Logger.Error("project not found", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"message": "not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": project})
}

func (t *GetProjectRoute) InitGetProjectRoute(router *gin.Engine) *gin.Engine {
	router.GET("/project/:projectId", t.getById)
	router.GET("/project", t.getAll)
	return router
}
