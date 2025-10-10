package controller_model

import (
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
)

type PutUserRoleRoute struct {
	Logger *zap.Logger
	Env    *config_env.AppEnv
	DB     *db_sqlc_gen.Queries
}

type UpdateMemberRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

func (t *PutUserRoleRoute) updateMemberRole(c *gin.Context) {
	projectParam := c.Param("projectId")
	userParam := c.Param("userId")

	projectBytes, err := base64.RawURLEncoding.DecodeString(projectParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}
	projectID, err := uuid.FromBytes(projectBytes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	userBytes, err := base64.RawURLEncoding.DecodeString(userParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}
	userID, err := uuid.FromBytes(userBytes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var req UpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	err = t.DB.PutUserRole(c, db_sqlc_gen.PutUserRoleParams{
		Role:      db_sqlc_gen.Role(req.Role),
		ProjectID: projectID,
		UserID:    userID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "role updated successfully"})
}

func (t *PutUserRoleRoute) InitPutUserRoleRoute(router gin.IRouter) gin.IRouter {
	router.PUT("/projects/:projectId/user/:userId/role", t.updateMemberRole)
	return router
}
