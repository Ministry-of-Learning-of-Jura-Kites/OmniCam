package controller_project

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
)

type PutProjectRoute struct {
	Logger *zap.Logger
	Env    *config_env.AppEnv
	DB     *db_sqlc_gen.Queries
}

type UpdateProjectRequest struct {
	Id          string `json:"id" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

func (t *PutProjectRoute) put(c *gin.Context) {
	var req UpdateProjectRequest

	err := c.ShouldBindJSON(&req)
	if err != nil {
		t.Logger.Debug("error while validating body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := uuid.Parse(req.Id)
	if err != nil {
		t.Logger.Debug("error while validating body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	data, err := t.DB.UpdateProject(c, db_sqlc_gen.UpdateProjectParams{
		Name:        req.Name,
		Description: req.Description,
		ID:          id,
	})
	if err != nil {
		t.Logger.Error("error while creating project", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "update successful", "data": data})

}

func (t *PutProjectRoute) InitUpdateProjectRoute(router *gin.Engine) *gin.Engine {
	router.PUT("/project/:projectId", t.put)
	return router
}
