package controller_model

import (
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"
	db_client "omnicam.com/backend/pkg/db"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
)

type DeleteProjectMemberRoute struct {
	Logger *zap.Logger
	Env    *config_env.AppEnv
	DB     *db_client.DB
}

func (t *DeleteProjectMemberRoute) deleteMember(c *gin.Context) {

	projectParam := c.Param("projectId")
	decodedBytes, err := base64.RawURLEncoding.DecodeString(projectParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}
	projectID, err := uuid.FromBytes(decodedBytes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	userParam := c.Param("userId")
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

	err = t.DB.Queries.DeleteProjectMember(c, db_sqlc_gen.DeleteProjectMemberParams{
		UserID:    userID,
		ProjectID: projectID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete member"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "member removed successfully"})
}

func (t *DeleteProjectMemberRoute) InitDeleteProjectMemberRoute(router gin.IRouter) gin.IRouter {
	router.DELETE("/projects/:projectId/member/:userId", t.deleteMember)
	return router
}
