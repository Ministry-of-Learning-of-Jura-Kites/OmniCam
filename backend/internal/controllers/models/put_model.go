package controller_model

import (
	"encoding/base64"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"
	db_client "omnicam.com/backend/pkg/db"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
)

type PutModelRoute struct {
	Logger *zap.Logger
	Env    *config_env.AppEnv
	DB     *db_client.DB
}

type UpdateModelRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

func (t *PutModelRoute) put(c *gin.Context) {
	strId := c.Param("modelId")

	decodedBytes, err := base64.RawURLEncoding.DecodeString(strId)
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

	var req UpdateModelRequest
	err = c.ShouldBindJSON(&req)
	if err != nil {
		t.Logger.Debug("error while validating body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	params := db_sqlc_gen.UpdateModelParams{
		ID: modelId,
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

	data, err := t.DB.Queries.UpdateModel(c, db_sqlc_gen.UpdateModelParams{
		ID:          modelId,
		Name:        params.Name,
		Description: params.Description,
	})
	if err != nil {
		t.Logger.Error("error while updating project", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": Model{
		Id:          modelId,
		ProjectId:   data.ProjectID,
		Name:        data.Name,
		Description: data.Description,
		ImagePath:   data.ImagePath,
		Version:     data.Version,
		CreatedAt:   data.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:   data.UpdatedAt.Time.Format(time.RFC3339),
	}})

}

func (t *PutModelRoute) InitUpdateModelRoute(router gin.IRouter) gin.IRouter {
	router.PUT("/projects/:projectId/models/:modelId", t.put)
	return router
}
