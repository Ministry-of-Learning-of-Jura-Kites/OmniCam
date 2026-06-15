package controller_workspaces

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
	"omnicam.com/backend/internal/utils"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
	"omnicam.com/backend/pkg/logger"
	messages_workspace "omnicam.com/backend/pkg/messages/workspace"
)

func (t *WorkspaceRoute) postWorkspaceMe(c *gin.Context) {
	strModelId := c.Param("modelId")
	modelId, err := utils.ParseUuidBase64(strModelId)
	if err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).
			Debug("error while converting str id to uuid", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model ID"})
		return
	}

	userId, err := utils.GetUuidFromCtx(c, "userId")
	if err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).
			Debug("error while getting userId from context", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	workspace, err := t.DB.Queries.CreateWorkspace(c.Request.Context(), db_sqlc_gen.CreateWorkspaceParams{
		UserID:  userId,
		ModelID: modelId,
	})

	if err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).
			Error("error while creating workspace", zap.Error(err))
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "workspace already exists"})
			return
		}
		var e *pgconn.PgError
		if errors.As(err, &e) && e.Code == pgerrcode.UniqueViolation {
			c.JSON(http.StatusBadRequest, gin.H{"error": "workspace already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	logger.WithTraceID(c.Request.Context(), t.Logger).Info("create workspace successfully",
		zap.String("modelId", strModelId),
		zap.String("userId", userId.String()),
	)
	c.JSON(http.StatusCreated, gin.H{
		"data": messages_workspace.WorkspaceRawCams{
			WorkspaceNoCams: messages_workspace.WorkspaceNoCams{
				ModelId:     workspace.ModelID,
				UserId:      workspace.UserID,
				Version:     workspace.Version,
				BaseVersion: workspace.BaseVersion,
				CreatedAt:   workspace.CreatedAt.Time.Format(time.RFC3339),
				UpdatedAt:   workspace.UpdatedAt.Time.Format(time.RFC3339),
				ScaleFactor: workspace.ScaleFactor,
				ModelHeight: workspace.ModelHeight,
			},
			Cameras:     json.RawMessage(workspace.Cameras),
			BaseCameras: json.RawMessage(workspace.BaseCameras),
		},
	})
}
