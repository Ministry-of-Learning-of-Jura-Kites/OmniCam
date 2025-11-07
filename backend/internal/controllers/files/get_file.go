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
)

type FileRoute struct {
	Logger *zap.Logger
	Env    *config_env.AppEnv
	DB     *db_client.DB
}

func (t *FileRoute) userHasProjectAccess(c *gin.Context, userId uuid.UUID, projectId uuid.UUID) (bool, error) {
	projects, err := t.DB.Queries.GetProjectsByUserId(c, db_sqlc_gen.GetProjectsByUserIdParams{
		UserID:     userId,
		PageSize:   1000, // or any large number
		PageOffset: 0,
	})
	fmt.Println(len(projects))
	if err != nil {
		return false, err
	}

	for _, p := range projects {
		fmt.Println("project : ", p)
		if p.ID == projectId {
			return true, nil
		}
	}
	return false, nil
}

// Generic file serving
func (t *FileRoute) serveFile(c *gin.Context, pathSegments ...string) {
	filePath := filepath.Join(pathSegments...)
	fmt.Println("HELLLOOLOLOLs")
	fmt.Println(filePath)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"message": "file not found"})
		return
	}
	fmt.Println("exist")
	ext := filepath.Ext(filePath)
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	c.Header("Content-Type", mimeType)
	c.File(filePath)
	c.File(filePath)
}

func (t *FileRoute) getProjectFile(c *gin.Context) {
	projectIdStr := c.Param("projectId")
	fileExt := c.Param("fileExt")
	projectId, err := uuid.Parse(projectIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid projectId"})
		return
	}
	fmt.Println("2")
	userId, err := utils.GetUuidFromCtx(c, "userId")
	fmt.Println("user id : ", userId)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"message": "permission denied"})
		return
	}
	fmt.Println("3")
	hasAccess, err := t.userHasProjectAccess(c, userId, projectId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to validate access"})
		return
	}

	if !hasAccess {
		c.JSON(http.StatusForbidden, gin.H{"message": "permission denied"})
		return
	}
	fileType := "images"
	fmt.Println("4")
	// Construct file root/uplodas/images/projectId.ext
	filePath := fmt.Sprintf(internal.Root+"/uploads/%s/%s.%s", fileType, projectId.String(), fileExt)
	t.serveFile(c, filePath)
}

// Route: /:projectId/:modelId/:fileExt
func (t *FileRoute) getModelFile(c *gin.Context) {
	projectIdStr := c.Param("projectId")
	modelIdStr := c.Param("modelId")
	fileExt := c.Param("fileExt")
	fmt.Println("is in??")
	projectId, err := uuid.Parse(projectIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid projectId"})
		return
	}
	fmt.Println("1")
	modelId, err := uuid.Parse(modelIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid modelId"})
		return
	}
	fmt.Println("2")
	userId, err := utils.GetUuidFromCtx(c, "userId")
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"message": "permission denied"})
		return
	}

	hasAccess, err := t.userHasProjectAccess(c, userId, projectId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to validate access"})
		return
	}

	if !hasAccess {
		c.JSON(http.StatusForbidden, gin.H{"message": "permission denied"})
		return
	}
	fmt.Println("has access")
	var fileType string
	if fileExt != "glb" {
		fileType = "images"
	} else {
		fileType = "3d_models"
	}

	filePath := fmt.Sprintf(internal.Root+"/uploads/%s/%s/%s.%s", fileType, projectId.String(), modelId.String(), fileExt)
	fmt.Println("4")
	t.serveFile(c, filePath)
}

// Initialize routes
func (t *FileRoute) InitFileRouter(router gin.IRouter) gin.IRouter {
	router.GET("/assets/projects/:projectId/models/:modelId/file/:fileExt", t.getModelFile)
	router.GET("/assets/projects/:projectId/file/:fileExt", t.getProjectFile)
	return router
}
