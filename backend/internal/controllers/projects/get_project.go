package controller_projects

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
)

type Project struct {
	Id          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   string    `json:"createdAt"`
	UpdatedAt   string    `json:"updatedAt"`
}

type GetProjectRoute struct {
	Logger *zap.Logger
	Env    *config_env.AppEnv
	DB     *db_sqlc_gen.Queries
}

func (t *GetProjectRoute) getAll(c *gin.Context) {
	project, err := t.DB.GetAllProjects(c)
	if err != nil {
		t.Logger.Error("Error while getting all project", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	if project == nil {
		project = []db_sqlc_gen.Project{}
	}

	c.JSON(http.StatusOK, gin.H{"data": project})
}

func (t *GetProjectRoute) getById(c *gin.Context) {
	strId := c.Param("projectId")

	id, err := uuid.Parse(strId)
	if err != nil {
		t.Logger.Error("error while converting str id to uuid", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid projectId"})
		return
	}

	project, err := t.DB.GetProjectById(c, id)
	if err != nil {
		t.Logger.Error("project not found", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": Project{
		Id:          id,
		Name:        project.Name,
		Description: project.Description,
		CreatedAt:   project.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:   project.UpdatedAt.Time.Format(time.RFC3339),
	}})
}

func (t *GetProjectRoute) InitGetProjectRoute(router gin.IRouter) gin.IRouter {
	router.GET("/projects/:projectId", t.getById)
	router.GET("/projects", t.getAll)
	return router
}
