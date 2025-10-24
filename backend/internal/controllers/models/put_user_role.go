package controller_model

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"
	"omnicam.com/backend/internal/utils"
	db_client "omnicam.com/backend/pkg/db"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
)

type PutUserRoleRoute struct {
	Logger *zap.Logger
	Env    *config_env.AppEnv
	DB     *db_client.DB
}

type UpdateMemberRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

func (t *PutUserRoleRoute) updateMemberRole(c *gin.Context) {
	projectParam := c.Param("projectId")
	userParam := c.Param("userId")

	projectID, err := utils.ParseUuidBase64(projectParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	userID, err := utils.ParseUuidBase64(userParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var req UpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	err = t.DB.Queries.PutUserRole(c, db_sqlc_gen.PutUserRoleParams{
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
