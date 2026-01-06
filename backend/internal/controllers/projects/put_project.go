package controller_projects

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"
	"omnicam.com/backend/internal/utils"
	db_client "omnicam.com/backend/pkg/db"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
)

type PutProjectRoute struct {
	Logger *zap.Logger
	Env    *config_env.AppEnv
	DB     *db_client.DB
}

type UpdateProjectRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

func (t *PutProjectRoute) put(c *gin.Context) {
	strProjectId := c.Param("projectId")
	projectId, err := utils.ParseUuidBase64(strProjectId)
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

	var req UpdateProjectRequest

	err = c.ShouldBindJSON(&req)
	if err != nil {
		t.Logger.Debug("error while validating body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	params := db_sqlc_gen.UpdateProjectParams{
		ID: projectId,
	}

	if req.Name != nil {
		params.Name = pgtype.Text{String: *req.Name, Valid: true}
	} else {
		params.Name = pgtype.Text{Valid: false}
	}

	if req.Description != nil {
		params.Description = pgtype.Text{String: *req.Description, Valid: true}
	} else {
		params.Description = pgtype.Text{Valid: false}
	}

	project, err := t.DB.Queries.UpdateProject(c, params)
	if err != nil {
		t.Logger.Error("error while updating project", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": Project{
		Id:             project.ID,
		Name:           project.Name,
		Description:    project.Description,
		ImagePath:      project.ImagePath,
		ImageExtension: project.ImageExtension,
		CreatedAt:      project.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:      project.UpdatedAt.Time.Format(time.RFC3339),
	}})
}

func (t *PutProjectRoute) InitUpdateProjectRoute(router gin.IRouter) gin.IRouter {
	router.PUT("/projects/:projectId", t.put)
	return router
}
