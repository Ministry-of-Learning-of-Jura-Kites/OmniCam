package controller_model

import (
	"encoding/json"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"
	"omnicam.com/backend/internal/utils"
	db_client "omnicam.com/backend/pkg/db"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
	messages_cameras "omnicam.com/backend/pkg/messages/cameras"
	messages_model_workspace "omnicam.com/backend/pkg/messages/model_workspace"
)

type Model struct {
	WorkspaceExists *bool `json:"workspaceExists,omitempty"`
	messages_model_workspace.ModelWorkspace
}

type GetModelRoute struct {
	Logger *zap.Logger
	Env    *config_env.AppEnv
	DB     *db_client.DB
}

func (t *GetModelRoute) getModelById(c *gin.Context) {
	strModelId := c.Param("modelId")
	modelId, err := utils.ParseUuidBase64(strModelId)
	if err != nil {
		t.Logger.Error("error while converting str id to uuid", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model ID"})
		return
	}

	strProjectId := c.Param("projectId")
	projectId, err := utils.ParseUuidBase64(strProjectId)
	if err != nil {
		t.Logger.Error("error while converting str id to uuid", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model ID"})
		return
	}

	includedFields := c.QueryArray("fields")

	username := c.GetString("username")

	userInfo, err := t.DB.Queries.GetUserOfProject(c, db_sqlc_gen.GetUserOfProjectParams{
		Username: pgtype.Text{
			String: username,
			Valid:  true,
		},
		Projectid: projectId,
	})
	if err != nil {
		t.Logger.Error("user of project not found", zap.String("projectId", strProjectId), zap.String("username", username), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	uuidBytes, err := userInfo.ID.MarshalBinary()
	if err != nil {
		t.Logger.Error("Failed to marshal uuid", zap.String("username", username), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	data, err := t.DB.Queries.GetModelByID(c, db_sqlc_gen.GetModelByIDParams{
		Fields: includedFields,
		ID:     modelId,
		UserID: pgtype.UUID{
			Bytes: [16]byte(uuidBytes),
			Valid: true,
		},
	})
	if err != nil {
		t.Logger.Error("model not found", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	cameras := messages_cameras.Cameras{}
	if data.Cameras != nil {
		err = json.Unmarshal(data.Cameras, &cameras)
		if err != nil {
			t.Logger.Error("cameras jsonb are invalid", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}
	}

	var workspaceExists *bool
	if slices.Contains(includedFields, "workspace_exists") {
		workspaceExists = &data.WorkspaceExists
	}

	c.JSON(http.StatusOK, gin.H{"data": Model{
		ModelWorkspace: messages_model_workspace.ModelWorkspace{
			ModelId:     modelId,
			ProjectId:   data.ProjectID,
			Name:        data.Name,
			Description: data.Description,
			FilePath:    data.FilePath,
			ImagePath:   data.ImagePath,
			Version:     data.Version,
			CreatedAt:   data.CreatedAt.Time.Format(time.RFC3339),
			UpdatedAt:   data.UpdatedAt.Time.Format(time.RFC3339),
			Cameras:     &cameras,
		},
		WorkspaceExists: workspaceExists,
	}})
}

func (t *GetModelRoute) getAllModel(c *gin.Context) {
	strProjectId := c.Param("projectId")

	projectId, err := utils.ParseUuidBase64(strProjectId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{})
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
	data, err := t.DB.Queries.GetAllfdfModels(c, db_sqlc_gen.GetAllfdfModelsParams{
		ProjectID:  projectId,
		PageSize:   int32(pageSize),
		PageOffset: int32(offset),
	})
	if err != nil {
		t.Logger.Error("models not found or database error", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	dataCount, err := t.DB.Queries.CountModels(c, projectId)
	if err != nil {
		t.Logger.Error("models not found or database error", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}
	var dataList []messages_model_workspace.ModelWorkspace

	for _, model := range data {
		dataList = append(dataList, messages_model_workspace.ModelWorkspace{
			ModelId:     model.ID,
			ProjectId:   model.ProjectID,
			Name:        model.Name,
			Description: model.Description,
			ImagePath:   model.ImagePath,
			Version:     model.Version,
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
