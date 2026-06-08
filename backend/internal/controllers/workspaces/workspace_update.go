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

	routeMap := make(map[string]bool)
	for _, r := range req.Simulation.Routes {
		if r.Id == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "route id cannot be empty",
			})
			return
		}
		routeMap[r.Id] = true
	}

	// Validate population groups reference valid routes
	for _, g := range req.Simulation.PopulationGroups {
		if g.RouteId == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "population group routeId cannot be empty",
			})
			return
		}

		if !routeMap[g.RouteId] {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid routeId in population group",
			})
			return
		}
	}

	areaMap := make(map[string]bool)
	for _, a := range req.Simulation.Areas {
		if a.Id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "area id cannot be empty"})
			return
		}
		if a.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "area name cannot be empty"})
			return
		}
		if a.Kind != "start" && a.Kind != "end" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid area kind"})
			return
		}
		areaMap[a.Id] = true
	}

	// Validate routes reference existing areas + have segments
	for _, r := range req.Simulation.Routes {
		if r.StartAreaId == "" || r.EndAreaId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "route must have startAreaId and endAreaId"})
			return
		}
		if !areaMap[r.StartAreaId] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid startAreaId in route"})
			return
		}
		if !areaMap[r.EndAreaId] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid endAreaId in route"})
			return
		}

		if len(r.Segments) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "route must have at least one segment"})
			return
		}

		for _, seg := range r.Segments {
			if seg.Type != "line" && seg.Type != "bezier" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid segment type"})
				return
			}
			if len(seg.Points) == 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "segment must have points"})
				return
			}
			// Optional: stricter point count
			if seg.Type == "line" && len(seg.Points) != 2 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "line segment must have exactly 2 points"})
				return
			}
			if seg.Type == "bezier" && len(seg.Points) != 4 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "bezier segment must have exactly 4 points"})
				return
			}
		}
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
