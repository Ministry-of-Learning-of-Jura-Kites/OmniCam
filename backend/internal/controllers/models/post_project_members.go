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

	decodedBytes, err := base64.RawURLEncoding.DecodeString(strProjectId)
	if err != nil {
		t.Logger.Error("error decoding Base64 projectId", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid base64 projectId"})
		return
	}

	projectID, err := uuid.FromBytes(decodedBytes)
	if err != nil {
		t.Logger.Error("error converting decoded projectId to UUID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid projectId"})
		return
	}

	var req []AddProjectMembersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		t.Logger.Error("invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}

	for _, m := range req {
		if m.ProjectID != uuid.Nil && m.ProjectID != projectID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "projectId mismatch"})
			return
		}
	}

	for _, m := range req {
		if m.UserID == uuid.Nil {
			continue
		}

		param := db_sqlc_gen.PostProjectMembersParams{
			ProjectID: projectID,
			UserID:    m.UserID,
			Role:      db_sqlc_gen.Role(m.Role),
		}

		if err := t.DB.Queries.PostProjectMembers(c, param); err != nil {
			t.Logger.Error("failed to upsert project member", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// InitProjectMemberRouter registers project member routes
func (t *PostProjectMembersRoute) InitProjectMemberRouter(router gin.IRouter) gin.IRouter {
	router.POST("/projects/:projectId/members", t.addProjectMembers)
	return router
}
