package controller_projects

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"
	"omnicam.com/backend/internal"
	"omnicam.com/backend/internal/utils"
	db_client "omnicam.com/backend/pkg/db"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
)

type PostProjectRoute struct {
	Logger *zap.Logger
	Env    *config_env.AppEnv
	DB     *db_client.DB
}

type CreateProjectRequest struct {
	Name        string `form:"name" binding:"required"`
	Description string `form:"description"`
}

func (t *PostProjectRoute) post(c *gin.Context) {
	username, exists := c.Get("username")
	if !exists {
		t.Logger.Error("username not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, err := t.DB.Queries.GetUserByUsername(c, username.(string))
	if err != nil {
		t.Logger.Error("failed to get user by username", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
		return
	}
	var req CreateProjectRequest

	_, err = utils.GetUuidFromCtx(c, "userId")
	if err != nil {
		t.Logger.Error("error while getting userId form", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	if err := c.ShouldBind(&req); err != nil {
		t.Logger.Debug("error while validating form", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid form data"})
		return
	}

	projectID := uuid.New()

	var imageWebPath string
	imageFile, err := c.FormFile("image")
	if err == nil {
		ext := filepath.Ext(imageFile.Filename)
		if ext != ".jpg" && ext != ".png" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "image must be .jpg or .png"})
			return
		}

		uploadDir := filepath.Join(internal.Root, "uploads", "project", projectID.String())
		if mkErr := os.MkdirAll(uploadDir, os.ModePerm); mkErr != nil {
			t.Logger.Error("failed to create upload dir", zap.Error(mkErr))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create upload dir"})
			return
		}

		fsImagePath := filepath.Join(uploadDir, "image"+ext) // local filesystem path
		if saveErr := c.SaveUploadedFile(imageFile, fsImagePath); saveErr != nil {
			t.Logger.Error("failed to save project image", zap.Error(saveErr))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save project image"})
			return
		}
		imageWebPath = "/uploads/project/" + projectID.String() + "/image" + ext
	}

	tx, err := t.DB.Pool.Begin(c)
	if err != nil {
		t.Logger.Error("error while creating transaction", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	queries := t.DB.Queries.WithTx(tx)

	project, err := queries.CreateProject(c, db_sqlc_gen.CreateProjectParams{
		ID:          projectID,
		Name:        req.Name,
		Description: req.Description,
		ImagePath:   imageWebPath,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "project with this name already exists",
			})
			return
		}
		t.Logger.Error("error while creating project", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	if _, err := queries.AddUserToProject(c, db_sqlc_gen.AddUserToProjectParams{
		ProjectID: project.ID,
		UserID:    user.ID,
		Role:      db_sqlc_gen.RoleOwner,
	}); err != nil {
		t.Logger.Error("failed to add user to project", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}
	tx.Commit(c)

	// --- Response ---
	c.JSON(http.StatusOK, gin.H{"data": Project{
		Id:          project.ID,
		Name:        project.Name,
		Description: project.Description,
		CreatedAt:   project.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:   project.UpdatedAt.Time.Format(time.RFC3339),
	}})
}

func (t *PostProjectRoute) InitCreateProjectRoute(router gin.IRouter) gin.IRouter {
	router.POST("/projects", t.post)
	return router
}
