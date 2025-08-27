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

type PostModelRoutes struct {
	Logger *zap.Logger
	Env    *config_env.AppEnv
	DB     *db_sqlc_gen.Queries
}

type CreateModelRequest struct {
	// ProjectId   string `form:"project_id" binding:"required"`
	Name        string `form:"name" binding:"required"`
	Description string `form:"description"`
}

func (t *PostModelRoutes) post(c *gin.Context) {
	var req CreateModelRequest

	strId := c.Param("projectId")
	projectId, err := uuid.Parse(strId)
	if err != nil {
		t.Logger.Error("error while converting str id to uuid", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model ID"})
		return
	}

	err = c.ShouldBind(&req)
	if err != nil {
		t.Logger.Debug("error while validating form", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	uploadDir := filepath.Join(t.Env.ModelFilePath, projectId.String())
	err = os.MkdirAll(uploadDir, os.ModePerm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload model file"})
		return
	}

	filePath := filepath.Join(uploadDir, file.Filename)
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}

	t.DB.CreateModel(c, db_sqlc_gen.CreateModelParams{
		ProjectID:   projectId,
		Name:        req.Name,
		Description: req.Description,
		FilePath:    filePath,
	})

	c.JSON(http.StatusOK, gin.H{
		"project_id":  projectId,
		"name":        req.Name,
		"description": req.Description,
	})
}

func (t *PostModelRoutes) InitCreateModelRoute(router gin.IRouter) gin.IRouter {
	router.POST("/projects/:projectId/models", t.post)
	return router
}
