package controller_model

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"
	"omnicam.com/backend/internal" // 👈 use Root
	"omnicam.com/backend/internal/utils"
	db_client "omnicam.com/backend/pkg/db"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
)

type PutImageModelRoute struct {
	Logger *zap.Logger
	Env    *config_env.AppEnv
	DB     *db_client.DB
}

func (t *PutImageModelRoute) updateImage(c *gin.Context) {
	strProjectId := c.Param("projectId")
	projectId, err := utils.ParseUuidBase64(strProjectId)
	if err != nil {
		t.Logger.Error("error while converting str id to uuid", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model ID"})
		return
	}

	strModelId := c.Param("modelId")
	modelId, err := utils.ParseUuidBase64(strModelId)
	if err != nil {
		t.Logger.Error("error while converting str id to uuid", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model ID"})
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
		Projectid: projectId,
	})
	if err != nil {
		t.Logger.Debug("user of project not found", zap.String("projectId", strProjectId), zap.String("userId", userId.String()), zap.Error(err))
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	imageFile, err := c.FormFile("image")
	if err != nil {
		t.Logger.Error("Image file is required", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "image file is required"})
		return
	}

	imageExt := filepath.Ext(imageFile.Filename)
	if imageExt != ".jpg" && imageExt != ".png" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image must be .jpg or .png"})
		return
	}

	t.Logger.Info("Received image file",
		zap.String("filename", imageFile.Filename),
		zap.Int64("size", imageFile.Size),
	)

	// Where to save on disk -> <project-root>/uploads/model/{projectId}/{modelId}
	imageDir := filepath.Join(internal.Root, "uploads", "images", projectId.String())
	if err := os.MkdirAll(imageDir, os.ModePerm); err != nil {
		t.Logger.Error("Failed to create image directory", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create folder for model image"})
		return
	}

	// Remove old images (if any)
	files, _ := os.ReadDir(imageDir)
	for _, f := range files {
		if !f.IsDir() {
			_ = os.Remove(filepath.Join(imageDir, f.Name()))
		}
	}

	// Save the new image
	fsImagePath := filepath.Join(imageDir, modelId.String()+imageExt)
	if err := c.SaveUploadedFile(imageFile, fsImagePath); err != nil {
		t.Logger.Error("Failed to save image file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save image"})
		return
	}

	// Web path for DB/frontend
	webImagePath := "/uploads/images/" + projectId.String() + modelId.String()
	_, err = t.DB.Queries.UpdateModelImage(c, db_sqlc_gen.UpdateModelImageParams{
		ID:             modelId,
		ImagePath:      webImagePath,
		ImageExtension: imageExt,
	})
	if err != nil {
		t.Logger.Error("Error while updating model image", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update model image in DB"})
		return
	}

	t.Logger.Info("Model image updated", zap.String("path", fsImagePath))

	c.JSON(http.StatusOK, gin.H{
		"message":        "image updated successfully",
		"imagePath":      webImagePath,
		"imageExtension": imageExt,
	})
}

func (t *PutImageModelRoute) InitUpdateImageRoute(router gin.IRouter) gin.IRouter {
	router.PUT("/projects/:projectId/models/:modelId/image", t.updateImage)
	return router
}
