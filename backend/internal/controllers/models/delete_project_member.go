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

type DeleteProjectMemberRoute struct {
	Logger *zap.Logger
	Env    *config_env.AppEnv
	DB     *db_client.DB
}

func (t *DeleteProjectMemberRoute) deleteMember(c *gin.Context) {

	projectParam := c.Param("projectId")
	projectID, err := utils.ParseUuidBase64(projectParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	userParam := c.Param("userId")
	userID, err := utils.ParseUuidBase64(userParam)
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
