package controller_files

import (
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
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
	Tracer trace.Tracer
}

func (t *FileRoute) userHasProjectAccess(c *gin.Context, userId uuid.UUID, projectId uuid.UUID) (bool, error) {
	projects, err := t.DB.Queries.GetProjectsByUserId(c, db_sqlc_gen.GetProjectsByUserIdParams{
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
	_, span := t.Tracer.Start(
		c.Request.Context(),
		"FileRoute.getProjectFile",
	)
	defer span.End()

	projectIdStr := c.Param("projectId")
	fileExt := c.Param("fileExt")

	span.SetAttributes(
		attribute.String("project.id_encoded", projectIdStr),
		attribute.String("file.extension", fileExt),
	)

	projectId, err := utils.ParseUuidBase64(projectIdStr)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid projectId")

		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid projectId",
		})
		return
	}

	userId, err := utils.GetUuidFromCtx(c, "userId")
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "permission denied")

		c.JSON(http.StatusForbidden, gin.H{
			"message": "permission denied",
		})
		return
	}

	hasAccess, err := t.userHasProjectAccess(c, userId, projectId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed access validation")

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to validate access",
		})
		return
	}

	if !hasAccess {
		span.SetAttributes(
			attribute.Bool("auth.allowed", false),
		)

		span.SetStatus(codes.Error, "permission denied")

		c.JSON(http.StatusForbidden, gin.H{
			"message": "permission denied",
		})
		return
	}

	span.SetAttributes(
		attribute.Bool("auth.allowed", true),
	)

	fileType := "images"

	filePath := fmt.Sprintf(
		internal.Root+"/uploads/%s/%s.%s",
		fileType,
		projectId.String(),
		fileExt,
	)

	span.SetAttributes(
		attribute.String("file.path", filePath),
		attribute.String("file.type", fileType),
	)

	t.serveFile(c, filePath)

	span.SetStatus(codes.Ok, "file served")
}

// Route: /:projectId/:modelId/:fileExt
func (t *FileRoute) getModelFile(c *gin.Context) {
	_, span := t.Tracer.Start(c.Request.Context(), "FileRoute.getModelFile")
	defer span.End()

	projectIdStr := c.Param("projectId")
	modelIdStr := c.Param("modelId")
	fileExt := c.Param("fileExt")

	span.SetAttributes(
		attribute.String("project.id_encoded", projectIdStr),
		attribute.String("model.id_encoded", modelIdStr),
		attribute.String("file.extension", fileExt),
	)

	projectId, err := utils.ParseUuidBase64(projectIdStr)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid projectId")
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid projectId"})
		return
	}

	modelId, err := utils.ParseUuidBase64(modelIdStr)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid modelId")
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid modelId"})
		return
	}

	userId, err := utils.GetUuidFromCtx(c, "userId")
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "permission denied")
		c.JSON(http.StatusForbidden, gin.H{"message": "permission denied"})
		return
	}

	hasAccess, err := t.userHasProjectAccess(c, userId, projectId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed access validation")
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to validate access"})
		return
	}

	if !hasAccess {
		span.SetAttributes(
			attribute.Bool("auth.allowed", false),
		)
		// Return not found for security
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

	span.SetAttributes(
		attribute.String("file.path", filePath),
		attribute.String("file.type", fileType),
	)

	t.serveFile(c, filePath)

	span.SetStatus(codes.Ok, "file served")
}

// Initialize routes
func (t *FileRoute) InitFileRouter(router gin.IRouter) gin.IRouter {
	router.GET("/assets/projects/:projectId/models/:modelId/file/:fileExt", t.getModelFile)
	router.GET("/assets/projects/:projectId/file/:fileExt", t.getProjectFile)
	return router
}
