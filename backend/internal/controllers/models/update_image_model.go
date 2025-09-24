package controller_model

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
)

type PutImageModelRoute struct {
	Logger *zap.Logger
	Env    *config_env.AppEnv
	DB     *db_sqlc_gen.Queries
	// DB is not needed if we only update the image path on disk
}

func (t *PutImageModelRoute) updateImage(c *gin.Context) {
	projectIdStr := c.Param("projectId")
	modelIdStr := c.Param("modelId")

	projectId, err := uuid.Parse(projectIdStr)
	if err != nil {
		t.Logger.Error("Invalid projectId", zap.String("projectId", projectIdStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	modelId, err := uuid.Parse(modelIdStr)
	if err != nil {
		t.Logger.Error("Invalid modelId", zap.String("modelId", modelIdStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model ID"})
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

	// Where to save on disk
	imageDir := filepath.Join(t.Env.ModelFilePath, "model", projectId.String(), modelId.String())
	if err := os.MkdirAll(imageDir, os.ModePerm); err != nil {
		t.Logger.Error("Failed to create image directory", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create folder for model image"})
		return
	}

	// Remove old images
	files, _ := os.ReadDir(imageDir)
	for _, f := range files {
		if !f.IsDir() {
			_ = os.Remove(filepath.Join(imageDir, f.Name()))
		}
	}

	// Save the new one
	fsImagePath := filepath.Join(imageDir, "image"+imageExt) // filesystem
	if err := c.SaveUploadedFile(imageFile, fsImagePath); err != nil {
		t.Logger.Error("Failed to save image file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save image"})
		return
	}

	// Public web path for frontend
	webImagePath := "/uploads/model/" + projectId.String() + "/" + modelId.String() + "/image" + imageExt

	_, err = t.DB.UpdateModelImage(c, db_sqlc_gen.UpdateModelImageParams{
		ID:        modelId,
		ImagePath: webImagePath,
	})
	if err != nil {
		t.Logger.Error("error while updating project", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	t.Logger.Info("Model image updated", zap.String("path", fsImagePath))

	c.JSON(http.StatusOK, gin.H{
		"message":   "image updated successfully",
		"imagePath": webImagePath, // return web path instead of fs path
	})
}

// Register the route
func (t *PutImageModelRoute) InitUpdateImageRoute(router gin.IRouter) gin.IRouter {
	router.PUT("/projects/:projectId/models/:modelId/image", t.updateImage)
	return router
}
