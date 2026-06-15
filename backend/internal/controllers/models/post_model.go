package controller_model

import (
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"
	"omnicam.com/backend/internal"
	"omnicam.com/backend/internal/utils"
	db_client "omnicam.com/backend/pkg/db"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
	"omnicam.com/backend/pkg/logger"
	messages_model_workspace "omnicam.com/backend/pkg/messages/model_workspace"
)

type PostModelRoutes struct {
	Logger *zap.Logger
	Env    *config_env.AppEnv
	DB     *db_client.DB
}

type CreateModelRequest struct {
	Name        string `form:"name" binding:"required"`
	Description string `form:"description"`
}

func (t *PostModelRoutes) post(c *gin.Context) {
	var req CreateModelRequest
	modelId := uuid.New()

	strProjectId := c.Param("projectId")
	projectId, err := utils.ParseUuidBase64(strProjectId)
	if err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).Debug("error while converting str id to uuid", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model ID"})
		return
	}

	if err := c.ShouldBind(&req); err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).Debug("error while validating form", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	// --- Save .glb ---
	file, err := c.FormFile("file")
	fileExt := filepath.Ext(file.Filename)
	if err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).Debug("file is required", zap.Error(err))
		return
	}
	if fileExt != ".glb" {
		logger.WithTraceID(c.Request.Context(), t.Logger).Debug("model file must have .glb extension", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "model file must have .glb extension"})
		return
	}

	// Local filesystem path
	uploadDir := filepath.Join(internal.Root, "uploads", "3d_models")
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).Error("failed to create upload dir", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create upload dir"})
		return
	}

	fsFilePath := filepath.Join(uploadDir, projectId.String(), modelId.String()+fileExt)
	if err := c.SaveUploadedFile(file, fsFilePath); err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).Error("failed to save model file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save model file"})
		return
	}

	// Web path for DB/frontend
	webFilePath := "/uploads/3d_models/" + projectId.String() + "/" + modelId.String()

	// --- Handle optional image ---
	var imageWebPath string
	var imageExt string
	imageFile, err := c.FormFile("image")
	if err == nil {
		imageExt = filepath.Ext(imageFile.Filename)
		if imageExt != ".jpg" && imageExt != ".png" {
			logger.WithTraceID(c.Request.Context(), t.Logger).Error("image must be .jpg or .png", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "image must be .jpg or .png"})
			return
		}

		imageDir := filepath.Join(internal.Root, "uploads", "images")
		if err := os.MkdirAll(imageDir, os.ModePerm); err != nil {
			logger.WithTraceID(c.Request.Context(), t.Logger).Error("failed to create folder for model image", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create folder for model image"})
			return
		}

		fsImagePath := filepath.Join(imageDir, projectId.String(), modelId.String()+imageExt) // local filesystem path
		if err := c.SaveUploadedFile(imageFile, fsImagePath); err != nil {
			logger.WithTraceID(c.Request.Context(), t.Logger).Error("failed to save model image", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save model image"})
			return
		}

		imageWebPath = "/uploads/images/" + projectId.String() + "/" + modelId.String()
		t.Logger.Info("model image uploaded", zap.String("path", fsImagePath))
	}

	// --- Insert into DB using web paths ---
	data, err := t.DB.Queries.CreateModel(c.Request.Context(), db_sqlc_gen.CreateModelParams{
		ID:             modelId,
		ProjectID:      projectId,
		Name:           req.Name,
		Description:    req.Description,
		FilePath:       webFilePath,  // web path
		ImagePath:      imageWebPath, // web path
		ModelExtension: fileExt,
		ImageExtension: imageExt,
	})
	if err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).Error("error while creating model", zap.Error(err))
		t.Logger.Error("error while creating model", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	logger.WithTraceID(c.Request.Context(), t.Logger).Info("successfully add model",
		zap.String("modelId", data.ID.String()),
		zap.String("projectId", data.ProjectID.String()),
	)
	c.JSON(http.StatusOK, gin.H{"data": messages_model_workspace.ModelWorkspace{
		ModelId:     data.ID,
		ProjectId:   data.ProjectID,
		Name:        data.Name,
		Description: data.Description,
		ImagePath:   data.ImagePath,
		Version:     data.Version,
		CreatedAt:   data.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:   data.UpdatedAt.Time.Format(time.RFC3339),
	}})
}

func (t *PostModelRoutes) InitCreateModelRoute(router gin.IRouter) gin.IRouter {
	router.POST("/projects/:projectId/models", t.post)
	return router
}
