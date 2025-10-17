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
		t.Logger.Error("error while converting str id to uuid", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model ID"})
		return
	}

	strProjectId := c.Param("projectId")
	projectId, err := utils.ParseUuidBase64(strProjectId)
	if err != nil {
		t.Logger.Error("error while converting str id to uuid", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	userId, err := utils.GetUuidFromCtx(c, "userId")
	if err != nil {
		t.Logger.Error("error while getting userId form", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	pgUserId, err := utils.UuidToPgUuid(userId)
	if err != nil {
		t.Logger.Error("Error while convert uuid to pgtype", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	// Check if user is in project
	_, err = t.DB.Queries.GetUserOfProject(c, db_sqlc_gen.GetUserOfProjectParams{
		UserID:    pgUserId,
		Projectid: *projectId,
	})
	if err != nil {
		t.Logger.Debug("user of project not found", zap.String("projectId", strProjectId), zap.String("userId", userId.String()), zap.Error(err))
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	model, err := t.DB.Queries.GetModelByID(c, db_sqlc_gen.GetModelByIDParams{
		ID: *modelId,
	})
	if err != nil {
		t.Logger.Error("failed to get model", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	deleteFile := func(dbPath string) {
		if dbPath == "" {
			return
		}

		relativePath := strings.TrimPrefix(dbPath, "/uploads/")

		// Use internal.Root from root.go
		fullPath := path.Join(internal.Root, "uploads", relativePath)
		dirPath := path.Dir(fullPath)

		absPath, _ := filepath.Abs(dirPath)
		t.Logger.Info("Absolute delete path:", zap.String("absPath", absPath))
		t.Logger.Info("Attempting to delete folder:", zap.String("dirPath", dirPath)) // Debug log

		if err := os.RemoveAll(absPath); err != nil {
			t.Logger.Error("failed to remove folder", zap.String("path", absPath), zap.Error(err))
		}
	}

	deleteFile(model.FilePath)
	deleteFile(model.ImagePath)

	_, err = t.DB.Queries.DeleteModel(c, *modelId)
	if err != nil {
		t.Logger.Error("something wrong with DB deletion", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "delete successfully", "data": modelId})
}

func (t *DeleteModelRoute) InitDeleteModelRoute(router gin.IRouter) gin.IRouter {
	router.DELETE("/projects/:projectId/models/:modelId", t.delete)
	return router
}
