package controller_projects

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"
	"omnicam.com/backend/internal/utils"
	db_client "omnicam.com/backend/pkg/db"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
	"omnicam.com/backend/pkg/logger"
)

type Project struct {
	Id             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	ImagePath      string    `json:"imagePath"`
	ImageExtension string    `json:"imageExtension"`
	CreatedAt      string    `json:"createdAt"`
	UpdatedAt      string    `json:"updatedAt"`
}

type GetProjectRoute struct {
	Logger *zap.Logger
	Env    *config_env.AppEnv
	DB     *db_client.DB
}

func (t *GetProjectRoute) getAll(c *gin.Context) {
	username, exists := c.Get("username")
	if !exists {
		logger.WithTraceID(c.Request.Context(), t.Logger).Debug("username not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, err := t.DB.Queries.GetUserByUsername(c.Request.Context(), username.(string))
	if err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).Error("failed to get user by username", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
		return
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		logger.WithTraceID(c.Request.Context(), t.Logger).Error("invalid page number", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page number"})
		return
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if err != nil || pageSize < 1 {
		logger.WithTraceID(c.Request.Context(), t.Logger).Error("invalid page size", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page size"})
		return
	}

	pageOffset := (page - 1) * pageSize

	projects, err := t.DB.Queries.GetProjectsByUserId(c.Request.Context(), db_sqlc_gen.GetProjectsByUserIdParams{
		UserID:     user.ID,
		PageSize:   int32(pageSize),
		PageOffset: int32(pageOffset),
	})
	if err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).Error("failed to get projects by user id", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get projects"})
		return
	}

	var projectList []Project
	for _, data := range projects {
		projectList = append(projectList, Project{
			Id:             data.ID,
			Name:           data.Name,
			Description:    data.Description,
			ImagePath:      data.ImagePath,
			ImageExtension: data.ImageExtension,
			CreatedAt:      data.CreatedAt.Time.Format(time.RFC3339),
			UpdatedAt:      data.UpdatedAt.Time.Format(time.RFC3339),
		})
	}

	logger.WithTraceID(c.Request.Context(), t.Logger).Info("successfully get project")
	c.JSON(http.StatusOK, gin.H{
		"data":     projectList,
		"page":     page,
		"pageSize": pageSize,
	})
}

func (t *GetProjectRoute) getById(c *gin.Context) {
	strId := c.Param("projectId")
	id, err := utils.ParseUuidBase64(strId)
	if err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).Debug("error while convering str id to uuid", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	project, err := t.DB.Queries.GetProjectById(c.Request.Context(), id)
	if err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).Error("project not found", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	logger.WithTraceID(c.Request.Context(), t.Logger).Info("successfully get project by id",
		zap.String("projectId", strId),
	)
	c.JSON(http.StatusOK, gin.H{"data": Project{
		Id:          id,
		Name:        project.Name,
		Description: project.Description,
		ImagePath:   project.ImagePath,
		CreatedAt:   project.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:   project.UpdatedAt.Time.Format(time.RFC3339),
	}})
}

func (t *GetProjectRoute) InitGetProjectRoute(router gin.IRouter) gin.IRouter {
	router.GET("/projects/:projectId", t.getById)
	router.GET("/projects", t.getAll)
	return router
}
