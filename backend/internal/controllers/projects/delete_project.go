package controller_projects

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"
	"omnicam.com/backend/internal"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
)

type DeleteProjectRoute struct {
	Logger *zap.Logger
	Env    *config_env.AppEnv
	DB     *db_sqlc_gen.Queries
}

func (t *DeleteProjectRoute) delete(c *gin.Context) {
	strId := c.Param("projectId")

	decodedBytes, err := base64.StdEncoding.DecodeString(strId)
	if err != nil {
		t.Logger.Error("error decoding Base64", zap.Error(err))
		return
	}
	projectId, err := uuid.FromBytes(decodedBytes)
	if err != nil {
		t.Logger.Error("error while converting str id to uuid", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	// --- Get project ---
	_, err = t.DB.GetProjectById(c, projectId)
	if err != nil {
		t.Logger.Error("failed to get project", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	projectImageFolder := path.Join(internal.Root, "uploads", "project", projectId.String())
	t.deleteFolder(projectImageFolder)

	// --- Delete all models under this project ---
	models, err := t.DB.GetModelsByProjectID(c, projectId)
	if err != nil {
		t.Logger.Error("failed to get models for project", zap.Error(err))
	} else {
		for _, model := range models {
			if model.FilePath != "" {
				t.deleteFolder(path.Join(internal.Root, "uploads", "model", projectId.String()))
			}
			if model.ImagePath != "" {
				t.deleteFolder(path.Join(internal.Root, "uploads", "model", projectId.String()))
			}
			_, err := t.DB.DeleteModel(c, model.ID)
			if err != nil {
				t.Logger.Error("failed to delete model from DB", zap.Error(err))
			}
		}
	}

	// --- Delete project record from DB ---
	data, err := t.DB.DeleteProject(c, projectId)
	if err != nil {
		t.Logger.Error("failed to delete project", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete project"})
		return
	}

	// --- Delete project folder ---
	projectFolder := path.Join(internal.Root, "uploads", projectId.String())
	t.deleteFolder(projectFolder)

	c.JSON(http.StatusAccepted, gin.H{"data": data})
}

func (t *DeleteProjectRoute) deleteFolder(folderPath string) {
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		return
	}

	absPath, _ := filepath.Abs(folderPath)
	fmt.Println("Deleting folder:", absPath)

	if err := os.RemoveAll(absPath); err != nil {
		t.Logger.Error("failed to remove folder", zap.String("path", absPath), zap.Error(err))
	}
}

func (t *DeleteProjectRoute) InitDeleteProjectRoute(router gin.IRouter) gin.IRouter {
	router.DELETE("/projects/:projectId", t.delete)
	return router
}
