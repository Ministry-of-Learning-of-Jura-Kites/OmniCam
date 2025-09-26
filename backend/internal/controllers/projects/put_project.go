package controller_projects

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
)

type PutProjectRoute struct {
	Logger *zap.Logger
	Env    *config_env.AppEnv
	DB     *db_sqlc_gen.Queries
}

type UpdateProjectRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

func (t *PutProjectRoute) put(c *gin.Context) {
	rawId := c.Param("projectId")

	var req UpdateProjectRequest

	err := c.ShouldBindJSON(&req)
	if err != nil {
		t.Logger.Debug("error while validating body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	id, err := uuid.Parse(rawId)
	if err != nil {
		t.Logger.Debug("error while validating body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid projectId"})
		return
	}
	params := db_sqlc_gen.UpdateProjectParams{
		ID: id,
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

	project, err := t.DB.UpdateProject(c, params)
	if err != nil {
		t.Logger.Error("error while updating project", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": Project{
		Id:          project.ID,
		Name:        project.Name,
		Description: project.Description,
		ImagePath:   project.ImagePath,
		CreatedAt:   project.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:   project.UpdatedAt.Time.Format(time.RFC3339),
	}})
}

func (t *PutProjectRoute) InitUpdateProjectRoute(router gin.IRouter) gin.IRouter {
	router.PUT("/projects/:projectId", t.put)
	return router
}
