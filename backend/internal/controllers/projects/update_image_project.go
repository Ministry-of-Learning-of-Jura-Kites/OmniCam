package controller_projects

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	config_env "omnicam.com/backend/config"
	"omnicam.com/backend/internal"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
)

type PutImageProjectRoute struct {
	Logger *zap.Logger
	Env    *config_env.AppEnv
	DB     *db_sqlc_gen.Queries
}

func (t *PutImageProjectRoute) updateImage(c *gin.Context) {
	projectIdStr := c.Param("projectId")
	projectId, err := uuid.Parse(projectIdStr)
	if err != nil {
		t.Logger.Error("Invalid projectId", zap.String("projectId", projectIdStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
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

	t.Logger.Info("Received project image",
		zap.String("filename", imageFile.Filename),
		zap.Int64("size", imageFile.Size),
	)

	imageDir := filepath.Join(internal.Root, "uploads", "project", projectId.String())
	if err := os.MkdirAll(imageDir, os.ModePerm); err != nil {
		t.Logger.Error("Failed to create image directory", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create folder for project image"})
		return
	}

	files, _ := os.ReadDir(imageDir)
	for _, f := range files {
		if !f.IsDir() {
			_ = os.Remove(filepath.Join(imageDir, f.Name()))
		}
	}

	fsImagePath := filepath.Join(imageDir, "image"+imageExt)
	if err := c.SaveUploadedFile(imageFile, fsImagePath); err != nil {
		t.Logger.Error("Failed to save image file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save image"})
		return
	}

	webImagePath := "/uploads/project/" + projectId.String() + "/image" + imageExt

	_, err = t.DB.UpdateProjectImage(c, db_sqlc_gen.UpdateProjectImageParams{
		ID:        projectId,
		ImagePath: webImagePath,
	})
	if err != nil {
		t.Logger.Error("Error while updating project image", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update project image in DB"})
		return
	}

	t.Logger.Info("Project image updated", zap.String("path", fsImagePath))
	c.JSON(http.StatusOK, gin.H{
		"message":   "project image updated successfully",
		"imagePath": webImagePath,
	})
}

func (t *PutImageProjectRoute) InitUpdateImageRoute(router gin.IRouter) gin.IRouter {
	router.PUT("/projects/:projectId/image", t.updateImage)
	return router
}
