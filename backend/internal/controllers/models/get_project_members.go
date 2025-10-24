package controller_model

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"
	"omnicam.com/backend/internal/utils"
	db_client "omnicam.com/backend/pkg/db"
)

type GetProjectMembersRoute struct {
	Logger *zap.Logger
	Env    *config_env.AppEnv
	DB     *db_client.DB
}

func (t *GetProjectMembersRoute) getProjectMembers(c *gin.Context) {
	strId := c.Param("projectId")
	projectID, err := utils.ParseUuidBase64(strId)
	if err != nil {
		t.Logger.Error("error decoding Base64", zap.Error(err))
		return
	}

	members, err := t.DB.Queries.GetProjectMembers(c, projectID)
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
