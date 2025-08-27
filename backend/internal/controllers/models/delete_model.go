package controller_model

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
)

type DeleteModelRoute struct {
	Logger *zap.Logger
	Env    *config_env.AppEnv
	DB     *db_sqlc_gen.Queries
}

func (t *DeleteModelRoute) delete(c *gin.Context) {
	strId := c.Param("modelId")
	modelId, err := uuid.Parse(strId)
	if err != nil {
		t.Logger.Error("error while converting str id to uuid", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model ID"})
		return
	}

	model, err := t.DB.GetModelByID(c, modelId)
	if err != nil {
		t.Logger.Error("failed to get model", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get model"})
		return
	}

	id, err := t.DB.DeleteModel(c, modelId)
	if err != nil {
		t.Logger.Error("something wrong", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{})
	}

	if model.FilePath != "" {
		if err := os.Remove(model.FilePath); err != nil {
			t.Logger.Error("failed to remove model file", zap.Error(err))
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "delete successfully", "data": id})
}

func (t *DeleteModelRoute) InitDeleteModelRoute(router gin.IRouter) gin.IRouter {
	router.DELETE("/projects/:projectId/models/:modelId", t.delete)
	return router
}
