package controller_workspaces

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"omnicam.com/backend/internal/utils"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
	model_workspace "omnicam.com/backend/pkg/messages/model_workspace"
)

type UpdateSimulationRequest struct {
	Simulation model_workspace.Simulation `json:"simulation"`
}

func (t *WorkspaceRoute) putWorkspaceSimulation(c *gin.Context) {
	strModelId := c.Param("modelId")

	modelId, err := utils.ParseUuidBase64(strModelId)
	if err != nil {
		t.Logger.Error(
			"error while converting model id",
			zap.Error(err),
		)

		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid model ID",
		})

		return
	}

	userId, err := utils.GetUuidFromCtx(c, "userId")
	if err != nil {
		t.Logger.Error(
			"error while getting user id",
			zap.Error(err),
		)

		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	var req UpdateSimulationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		t.Logger.Error(
			"invalid request body",
			zap.Error(err),
		)

		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})

		return
	}

	simulationJson, err := json.Marshal(req.Simulation)
	if err != nil {
		t.Logger.Error(
			"failed to marshal simulation",
			zap.Error(err),
		)

		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	version, err := t.DB.Queries.UpdateWorkspaceSimulation(
		c,
		db_sqlc_gen.UpdateWorkspaceSimulationParams{
			UserID:     userId,
			ModelID:    modelId,
			Simulation: simulationJson,
		},
	)

	if err != nil {
		t.Logger.Error(
			"failed to update simulation",
			zap.Error(err),
		)

		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"version": version,
	})
}
