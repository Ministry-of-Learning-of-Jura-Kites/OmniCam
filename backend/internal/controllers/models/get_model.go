package controller_model

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
)

type Model struct {
	Id          uuid.UUID `json:"id"`
	ProjectId   uuid.UUID `json:"projectId"`
	ImagePath   string    `json:"imagePath"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Version     int       `json:"version"`
	CreatedAt   string    `json:"createdAt"`
	UpdatedAt   string    `json:"updatedAt"`
}

type GetModelRoute struct {
	Logger *zap.Logger
	Env    *config_env.AppEnv
	DB     *db_sqlc_gen.Queries
}

func (t *GetModelRoute) getModelById(c *gin.Context) {
	strId := c.Param("modelId")
	id, err := uuid.Parse(strId)
	if err != nil {
		t.Logger.Error("error while converting str id to uuid", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model ID"})
		return
	}

	data, err := t.DB.GetModelByID(c, id)
	if err != nil {
		t.Logger.Error("model not found", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	version := 0
	if data.Version.Valid {
		version = int(data.Version.Int32)
	}

	c.JSON(http.StatusOK, gin.H{"data": Model{
		Id:          id,
		ProjectId:   data.ProjectID,
		Name:        data.Name,
		Description: data.Description,
		ImagePath:   data.ImagePath,
		Version:     version,
		CreatedAt:   data.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:   data.UpdatedAt.Time.Format(time.RFC3339),
	}})
}

func (t *GetModelRoute) getAllModel(c *gin.Context) {
	strProjectId := c.Param("projectId")
	projectId, err := uuid.Parse(strProjectId)
	if err != nil {
		t.Logger.Error("error while converting str id to uuid", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model ID"})
		return
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page number"})
		return
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if err != nil || pageSize < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page size"})
		return
	}

	offset := (page - 1) * pageSize
	// column1 -> projectId column2 -> page size column3 -> offset (the data from desc (createBy))
	data, err := t.DB.GetAllModels(c, db_sqlc_gen.GetAllModelsParams{
		Column1: projectId,
		Column2: int32(pageSize),
		Column3: int32(offset),
	})
	if err != nil {
		t.Logger.Error("models not found or database error", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	dataCount, err := t.DB.CountModels(c, projectId)
	if err != nil {
		t.Logger.Error("models not found or database error", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}
	var dataList []Model

	for _, model := range data {
		version := 0
		if model.Version.Valid {
			version = int(model.Version.Int32)
		}
		dataList = append(dataList, Model{
			Id:          model.ID,
			ProjectId:   model.ProjectID,
			Name:        model.Name,
			Description: model.Description,
			ImagePath:   model.ImagePath,
			Version:     version,
			CreatedAt:   model.CreatedAt.Time.Format(time.RFC3339),
			UpdatedAt:   model.UpdatedAt.Time.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": dataList, "count": dataCount})
}

func (t *GetModelRoute) InitGetModelRoute(router gin.IRouter) gin.IRouter {
	router.GET("/projects/:projectId/models/:modelId", t.getModelById)
	router.GET("/projects/:projectId/models", t.getAllModel)
	return router
}
