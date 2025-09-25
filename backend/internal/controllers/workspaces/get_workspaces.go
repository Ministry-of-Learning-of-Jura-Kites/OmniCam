package controller_workspaces

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
	messages_cameras "omnicam.com/backend/pkg/messages/cameras"
)

type Workspace struct {
	Id          uuid.UUID                `json:"id"`
	ModelId     uuid.UUID                `json:"modelId"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Version     int                      `json:"version"`
	CreatedAt   string                   `json:"createdAt"`
	UpdatedAt   string                   `json:"updatedAt"`
	Cameras     messages_cameras.Cameras `json:"cameras"`
}

type GetWorkspaceRoute struct {
	Logger *zap.Logger
	Env    *config_env.AppEnv
	DB     *db_sqlc_gen.Queries
}

func (t *GetWorkspaceRoute) getWorkspaceMe(c *gin.Context) {
	strProjectId := c.Param("modelId")

	decodedBytes, err := base64.RawURLEncoding.DecodeString(strProjectId)
	if err != nil {
		t.Logger.Error("error decoding Base64", zap.Error(err))
		return
	}
	modelId, err := uuid.FromBytes(decodedBytes)
	if err != nil {
		t.Logger.Error("error while converting str id to uuid", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model ID"})
		return
	}

	includedFields := c.QueryArray("fields")

	data, err := t.DB.GetWorkspaceByID(c, db_sqlc_gen.GetWorkspaceByIDParams{
		Fields:  includedFields,
		UserID:  uuid.Nil,
		ModelID: modelId,
	})
	if err != nil {
		t.Logger.Error("model not found", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	version := 0
	if data.Version.Valid {
		version = int(data.Version.Int32)
	}

	var cameras messages_cameras.Cameras
	err = json.Unmarshal(data.Cameras, &cameras)
	if err != nil {
		t.Logger.Error("cameras jsonb are invalid", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": Workspace{
		Id:          modelId,
		ModelId:     modelId,
		Name:        data.Name,
		Description: data.Description,
		Version:     version,
		CreatedAt:   data.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:   data.UpdatedAt.Time.Format(time.RFC3339),
		Cameras:     cameras,
	}})
}

func (t *GetWorkspaceRoute) InitRoute(router gin.IRouter) gin.IRouter {
	router.GET("/projects/:projectId/models/:modelId/workspaces/me", t.getWorkspaceMe)
	return router
}
