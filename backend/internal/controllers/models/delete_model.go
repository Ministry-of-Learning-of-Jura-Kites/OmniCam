package controller_model

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings" // Import the strings package

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"
	"omnicam.com/backend/internal"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
)

type DeleteModelRoute struct {
	Logger *zap.Logger
	Env    *config_env.AppEnv
	DB     *db_sqlc_gen.Queries
}

func (t *DeleteModelRoute) delete(c *gin.Context) {
	strId := c.Param("modelId")
	modelId, err := uuid.Parse(strId)
	if err != nil {
		t.Logger.Error("error while converting str id to uuid", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model ID"})
		return
	}

	model, err := t.DB.GetModelByID(c, modelId)
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
		fmt.Println("Absolute delete path:", absPath)
		fmt.Println("Attempting to delete folder:", dirPath) // Debug log

		if err := os.RemoveAll(absPath); err != nil {
			t.Logger.Error("failed to remove folder", zap.String("path", absPath), zap.Error(err))
		}
	}

	deleteFile(model.FilePath)
	deleteFile(model.ImagePath)

	id, err := t.DB.DeleteModel(c, modelId)
	if err != nil {
		t.Logger.Error("something wrong with DB deletion", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "delete successfully", "data": id})
}

func (t *DeleteModelRoute) InitDeleteModelRoute(router gin.IRouter) gin.IRouter {
	router.DELETE("/projects/:projectId/models/:modelId", t.delete)
	return router
}
