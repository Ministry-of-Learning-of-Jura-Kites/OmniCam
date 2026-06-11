package controller_workspaces

import (
	"encoding/json"
	"net/http"
	"slices"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
	"omnicam.com/backend/internal/utils"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
	messages_cameras "omnicam.com/backend/pkg/messages/cameras"
	messages_model_workspace "omnicam.com/backend/pkg/messages/model_workspace"
	messages_trapezoid "omnicam.com/backend/pkg/messages/trapezoids"
)

func (t *WorkspaceRoute) getWorkspace(c *gin.Context) {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	strWorkspaceId := c.Param("workspaceId")
	isWorkspaceMe := strWorkspaceId == "me"
	var workspaceId uuid.UUID
	var ownerID uuid.UUID
	if isWorkspaceMe {
		userID, err := utils.GetUuidFromCtx(c, "userId")
		if err != nil {
			t.Logger.Error("error while getting userId form", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}

		ownerID = userID
	} else {
		parsed, err := utils.ParseUuidBase64(strWorkspaceId)
		workspaceId = parsed
		if err != nil {
			t.Logger.Error("error while converting str id to uuid", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workspace ID"})
			return
		}
		ownerID = workspaceId
	}

	username := c.GetString("username")

	_, err = t.DB.Queries.GetUserOfProject(c.Request.Context(), db_sqlc_gen.GetUserOfProjectParams{
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

	includedFields := c.QueryArray("fields")

	data, err := t.DB.Queries.GetWorkspaceByID(c.Request.Context(), db_sqlc_gen.GetWorkspaceByIDParams{
		Fields:  includedFields,
		UserID:  ownerID,
		ModelID: modelId,
	})
	if err != nil {
		t.Logger.Error("model not found", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	var cameras messages_cameras.Cameras
	if slices.Contains(includedFields, "cameras") {
		cams, err := messages_cameras.UnmarshalCameras(data.Cameras)
		cameras = cams
		if err != nil {
			t.Logger.Error("cameras jsonb are invalid", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}
	}

	var targetTrapezoids messages_trapezoid.Trapezoids
	if slices.Contains(includedFields, "target_area_trapezoids") {
		err := json.Unmarshal(data.TargetAreaTrapezoids, &targetTrapezoids)
		if err != nil {
			t.Logger.Error("targetTrapezoids jsonb are invalid", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}
	}

	var simulation messages_model_workspace.Simulation

	if slices.Contains(includedFields, "simulation") {
		err := json.Unmarshal(data.Simulation, &simulation)
		if err != nil {
			t.Logger.Error(
				"simulation json invalid",
				zap.Error(err),
			)
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": messages_model_workspace.ModelWorkspace{
		ModelId:          modelId,
		Name:             data.Model.Name,
		Description:      data.Model.Description,
		ProjectId:        data.Model.ProjectID,
		FilePath:         data.Model.FilePath,
		ModelExtension:   data.Model.ModelExtension,
		ImagePath:        data.Model.ImagePath,
		ImageExtension:   data.Model.ImageExtension,
		Version:          data.Version,
		CreatedAt:        data.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:        data.UpdatedAt.Time.Format(time.RFC3339),
		Cameras:          &cameras,
		TargetTrapezoids: &targetTrapezoids,
		ScaleFactor:      data.ScaleFactor,
		ModelHeight:      data.ModelHeight,
		Simulation:       simulation,
	}})
}
