package controller_files

import (
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"
	"omnicam.com/backend/internal"
	"omnicam.com/backend/internal/utils"
	db_client "omnicam.com/backend/pkg/db"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
	"omnicam.com/backend/pkg/logger"
)

type FileRoute struct {
	Logger *zap.Logger
	Env    *config_env.AppEnv
	DB     *db_client.DB
}

func (t *FileRoute) userHasProjectAccess(c *gin.Context, userId uuid.UUID, projectId uuid.UUID) (bool, error) {
	projects, err := t.DB.Queries.GetProjectsByUserId(c.Request.Context(), db_sqlc_gen.GetProjectsByUserIdParams{
		UserID:     userId,
		PageSize:   1000, // or any large number
		PageOffset: 0,
	})

	if err != nil {
		return false, err
	}

	for _, p := range projects {

		if p.ID == projectId {
			return true, nil
		}
	}
	return false, nil
}

// Generic file serving
func (t *FileRoute) serveFile(c *gin.Context, pathSegments ...string) {
	filePath := filepath.Join(pathSegments...)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		logger.WithTraceID(c.Request.Context(), t.Logger).Error(
			"file not found",
			zap.String("filePath", filePath),
		)
		c.JSON(http.StatusNotFound, gin.H{"message": "file not found"})
		return
	}

	ext := filepath.Ext(filePath)
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	c.Header("Content-Type", mimeType)
	c.File(filePath)
}

func (t *FileRoute) getProjectFile(c *gin.Context) {
	projectIdStr := c.Param("projectId")
	fileExt := c.Param("fileExt")

	projectId, err := utils.ParseUuidBase64(projectIdStr)
	if err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).Error(
			"invalid projectId",
			zap.String("projectId", projectIdStr),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid projectId",
		})
		return
	}

	userId, err := utils.GetUuidFromCtx(c, "userId")
	if err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).Error(
			"failed to get userId from context",
			zap.Error(err),
		)
		c.JSON(http.StatusForbidden, gin.H{
			"message": "permission denied",
		})
		return
	}

	hasAccess, err := t.userHasProjectAccess(c, userId, projectId)
	if err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).Error(
			"failed to validate access",
			zap.String("projectId", projectId.String()),
			zap.String("userId", userId.String()),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to validate access",
		})
		return
	}

	if !hasAccess {
		logger.WithTraceID(c.Request.Context(), t.Logger).Warn(
			"access denied",
			zap.String("projectId", projectId.String()),
			zap.String("userId", userId.String()),
		)
		c.JSON(http.StatusForbidden, gin.H{
			"message": "permission denied",
		})
		return
	}

	fileType := "images"

	filePath := fmt.Sprintf(
		internal.Root+"/uploads/%s/%s.%s",
		fileType,
		projectId.String(),
		fileExt,
	)

	t.serveFile(c, filePath)
}

// Route: /:projectId/:modelId/:fileExt
func (t *FileRoute) getModelFile(c *gin.Context) {
	projectIdStr := c.Param("projectId")
	modelIdStr := c.Param("modelId")
	fileExt := c.Param("fileExt")

	projectId, err := utils.ParseUuidBase64(projectIdStr)
	if err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).Error(
			"invalid projectId",
			zap.String("projectId", projectIdStr),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid projectId"})
		return
	}

	modelId, err := utils.ParseUuidBase64(modelIdStr)
	if err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).Error(
			"invalid modelId",
			zap.String("modelId", modelIdStr),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid modelId"})
		return
	}

	userId, err := utils.GetUuidFromCtx(c, "userId")
	if err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).Error(
			"failed to get userId from context",
			zap.Error(err),
		)
		c.JSON(http.StatusForbidden, gin.H{"message": "permission denied"})
		return
	}

	hasAccess, err := t.userHasProjectAccess(c, userId, projectId)
	if err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).Error(
			"failed to validate access",
			zap.String("projectId", projectId.String()),
			zap.String("userId", userId.String()),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to validate access"})
		return
	}

	if !hasAccess {
		logger.WithTraceID(c.Request.Context(), t.Logger).Warn(
			"access denied",
			zap.String("projectId", projectId.String()),
			zap.String("userId", userId.String()),
		)
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	var fileType string
	if fileExt != "glb" {
		fileType = "images"
	} else {
		fileType = "3d_models"
	}

	filePath := fmt.Sprintf(internal.Root+"/uploads/%s/%s/%s.%s", fileType, projectId.String(), modelId.String(), fileExt)

	logger.WithTraceID(c.Request.Context(), t.Logger).Info("successfully serve file",
		zap.String("filePath", filePath),
	)
	t.serveFile(c, filePath)
}

// Initialize routes
func (t *FileRoute) InitFileRouter(router gin.IRouter) gin.IRouter {
	router.GET("/assets/projects/:projectId/models/:modelId/file/:fileExt", t.getModelFile)
	router.GET("/assets/projects/:projectId/file/:fileExt", t.getProjectFile)
	return router
}
