package controller_model

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"
	"omnicam.com/backend/internal/utils"
	db_client "omnicam.com/backend/pkg/db"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
)

type AddProjectMembersRequest struct {
	ProjectID uuid.UUID `json:"projectId"`
	UserID    uuid.UUID `json:"userId"`
	Role      string    `json:"role"`
}

type PostProjectMembersRoute struct {
	Logger *zap.Logger
	Env    *config_env.AppEnv
	DB     *db_client.DB
}

func (t *PostProjectMembersRoute) addProjectMembers(c *gin.Context) {
	strProjectId := c.Param("projectId")

	projectID, err := utils.ParseUuidBase64(strProjectId)
	if err != nil {
		t.Logger.Error("error decoding Base64 projectId", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid base64 projectId"})
		return
	}

	var req []AddProjectMembersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		t.Logger.Error("invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}
	var userIDs []uuid.UUID
	var roles []string

	for _, m := range req {
		userIDs = append(userIDs, m.UserID)
		roles = append(roles, m.Role)
	}

	err = t.DB.Queries.PostProjectMembers(c, db_sqlc_gen.PostProjectMembersParams{
		ProjectID: projectID,
		UserIds:   userIDs,
		Roles:     roles,
	})
	if err != nil {
		t.Logger.Error("add member to project fail", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// InitProjectMemberRouter registers project member routes
func (t *PostProjectMembersRoute) InitProjectMemberRouter(router gin.IRouter) gin.IRouter {
	router.POST("/projects/:projectId/members", t.addProjectMembers)
	return router
}
