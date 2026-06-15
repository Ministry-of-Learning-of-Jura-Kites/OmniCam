package controller_model

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings" // Import the strings package

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"
	"omnicam.com/backend/internal"
	"omnicam.com/backend/internal/utils"
	db_client "omnicam.com/backend/pkg/db"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
	"omnicam.com/backend/pkg/logger"
)

type DeleteModelRoute struct {
	Logger *zap.Logger
	Env    *config_env.AppEnv
	DB     *db_client.DB
}

func (t *DeleteModelRoute) delete(c *gin.Context) {
	strModelId := c.Param("modelId")
	modelId, err := utils.ParseUuidBase64(strModelId)
	if err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).Debug(
			"invalid model ID",
			zap.String("modelId", strModelId),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model ID"})
		return
	}

	strProjectId := c.Param("projectId")
	projectId, err := utils.ParseUuidBase64(strProjectId)
	if err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).Debug(
			"invalid project ID",
			zap.String("projectId", strProjectId),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	userId, err := utils.GetUuidFromCtx(c, "userId")
	if err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).Debug(
			"error while getting userId from context",
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	pgUserId, err := utils.UuidToPgUuid(userId)
	if err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).Debug(
			"failed to convert uuid to pg uuid",
			zap.String("userId", userId.String()),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	// Check if user is in project
	_, err = t.DB.Queries.GetUserOfProject(c.Request.Context(), db_sqlc_gen.GetUserOfProjectParams{
		UserID:    pgUserId,
		Projectid: projectId,
	})
	if err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).Warn(
			"access denied: user not in project",
			zap.String("projectId", projectId.String()),
			zap.String("userId", userId.String()),
			zap.Error(err),
		)
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	model, err := t.DB.Queries.GetModelByID(c.Request.Context(), db_sqlc_gen.GetModelByIDParams{
		ID: modelId,
	})
	if err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).Error(
			"failed to get model",
			zap.String("modelId", modelId.String()),
			zap.Error(err),
		)
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	deleteFile := func(dbPath string, ext string) {
		if dbPath == "" {
			return
		}

		relativePath := strings.TrimPrefix(dbPath, "/uploads/")
		// Append extension if provided
		if ext != "" {
			relativePath = relativePath + ext
		}

		fullPath := path.Join(internal.Root, "uploads", relativePath)
		absPath, _ := filepath.Abs(fullPath)

		logger.WithTraceID(c.Request.Context(), t.Logger).Info(
			"attempting to delete file",
			zap.String("path", absPath),
		)

		if err := os.Remove(absPath); err != nil {
			logger.WithTraceID(c.Request.Context(), t.Logger).Error(
				"failed to remove file",
				zap.String("path", absPath),
				zap.Error(err),
			)
		}
	}

	deleteFile(model.FilePath, model.ModelExtension)
	deleteFile(model.ImagePath, model.ImageExtension)
	_, err = t.DB.Queries.DeleteModel(c.Request.Context(), modelId)
	if err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).Error(
			"database deletion failed",
			zap.String("modelId", modelId.String()),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	logger.WithTraceID(c.Request.Context(), t.Logger).Info("successfully delete model", zap.String("modelId", strModelId))
	c.JSON(http.StatusOK, gin.H{"message": "delete successfully", "data": modelId})
}

func (t *DeleteModelRoute) InitDeleteModelRoute(router gin.IRouter) gin.IRouter {
	router.DELETE("/projects/:projectId/models/:modelId", t.delete)
	return router
}
