package controller_projects

import (
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"
	"omnicam.com/backend/internal"
	"omnicam.com/backend/internal/utils"
	db_client "omnicam.com/backend/pkg/db"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
	"omnicam.com/backend/pkg/logger"
)

type DeleteProjectRoute struct {
	Logger *zap.Logger
	Env    *config_env.AppEnv
	DB     *db_client.DB
}

func (t *DeleteProjectRoute) delete(c *gin.Context) {
	strProjectId := c.Param("projectId")
	projectId, err := utils.ParseUuidBase64(strProjectId)
	if err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).Error("error while converting str id to uuid", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model ID"})
		return
	}

	userId, err := utils.GetUuidFromCtx(c, "userId")
	if err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).Error("error while getting userId form", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	pgUserId, err := utils.UuidToPgUuid(userId)
	if err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).Error("error while convert uuid to pgtype", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	// Check if user is in project
	_, err = t.DB.Queries.GetUserOfProject(c.Request.Context(), db_sqlc_gen.GetUserOfProjectParams{
		UserID:    pgUserId,
		Projectid: projectId,
	})
	if err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).Error("user of project not found", zap.Error(err))
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	// --- Get project ---
	_, err = t.DB.Queries.GetProjectById(c.Request.Context(), projectId)
	if err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).Error("failed to get project", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	projectImageFolder := path.Join(internal.Root, "uploads", "project", projectId.String())
	t.deleteFolder(c, projectImageFolder)

	// --- Delete all models under this project ---
	models, err := t.DB.Queries.GetModelsByProjectID(c.Request.Context(), projectId)
	if err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).Error("failed to get models for project", zap.Error(err))
	} else {
		for _, model := range models {
			if model.FilePath != "" {
				t.deleteFolder(c, path.Join(internal.Root, "uploads", "model", projectId.String()))
			}
			if model.ImagePath != "" {
				t.deleteFolder(c, path.Join(internal.Root, "uploads", "model", projectId.String()))
			}
			_, err := t.DB.Queries.DeleteModel(c.Request.Context(), model.ID)
			if err != nil {
				logger.WithTraceID(c.Request.Context(), t.Logger).Error("failed to delete model from DB", zap.Error(err))
			}
		}
	}

	// --- Delete project record from DB ---
	data, err := t.DB.Queries.DeleteProject(c.Request.Context(), projectId)
	if err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).Error("failed to delete project", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete project"})
		return
	}

	// --- Delete project folder ---
	projectFolder := path.Join(internal.Root, "uploads", projectId.String())
	t.deleteFolder(c, projectFolder)

	logger.WithTraceID(c.Request.Context(), t.Logger).Info("successfully delete project",
		zap.String("projectId", projectId.String()),
	)
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func (t *DeleteProjectRoute) deleteFolder(c *gin.Context, folderPath string) {
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		return
	}

	absPath, _ := filepath.Abs(folderPath)
	logger.WithTraceID(c.Request.Context(), t.Logger).Info("Deleting folder",
		zap.String("absPath", absPath),
	)

	if err := os.RemoveAll(absPath); err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).Error("failed to remove folder", zap.Error(err))
	}
}

func (t *DeleteProjectRoute) InitDeleteProjectRoute(router gin.IRouter) gin.IRouter {
	router.DELETE("/projects/:projectId", t.delete)
	return router
}
