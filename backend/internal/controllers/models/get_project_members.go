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

type GetProjectMembersRoute struct {
	Logger *zap.Logger
	Env    *config_env.AppEnv
	DB     *db_sqlc_gen.Queries
}

func (t *GetProjectMembersRoute) getProjectMembers(c *gin.Context) {
	strId := c.Param("projectId")
	decodedBytes, err := base64.RawURLEncoding.DecodeString(strId)
	if err != nil {
		t.Logger.Error("error decoding Base64", zap.Error(err))
		return
	}
	projectId, err := uuid.FromBytes(decodedBytes)
	if err != nil {
		t.Logger.Error("error while converting str id to uuid", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	members, err := t.DB.GetProjectMembers(c, projectId)
	if err != nil {
		t.Logger.Error("failed to get project members", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve project members"})
		return
	}

	var memberList []gin.H
	for _, m := range members {
		memberList = append(memberList, gin.H{
			"userId":    m.UserID,
			"username":  m.Username,
			"email":     m.Email,
			"firstName": m.FirstName,
			"lastName":  m.LastName,
			"role":      m.Role,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  memberList,
		"count": len(memberList),
	})
}

func (t *GetProjectMembersRoute) InitProjectMemberRouter(router gin.IRouter) gin.IRouter {
	router.GET("/projects/:projectId/members", t.getProjectMembers)
	return router
}
