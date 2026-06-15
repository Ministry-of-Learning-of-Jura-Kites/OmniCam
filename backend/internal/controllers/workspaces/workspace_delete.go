package controller_workspaces

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"omnicam.com/backend/internal/utils"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
	"omnicam.com/backend/pkg/logger"
)

func (t *WorkspaceRoute) deleteWorkspaceMe(c *gin.Context) {
	strModelId := c.Param("modelId")
	modelId, err := utils.ParseUuidBase64(strModelId)
	if err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).Debug("error while converting str id to uuid", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model ID"})
		return
	}

	userId, err := utils.GetUuidFromCtx(c, "userId")
	if err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).Debug("error while getting userId form", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	err = t.DB.Queries.DeleteWorkspace(c.Request.Context(), db_sqlc_gen.DeleteWorkspaceParams{
		UserID:  userId,
		ModelID: modelId,
	})

	if err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).Error("error while deleting workspace", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	logger.WithTraceID(c.Request.Context(), t.Logger).Info("successfully delete workspace")
	c.Status(http.StatusNoContent)
}
