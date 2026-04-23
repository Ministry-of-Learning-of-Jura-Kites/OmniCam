package controller_workspaces

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"
	db_client "omnicam.com/backend/pkg/db"
)

type WorkspaceRoute struct {
	Logger *zap.Logger
	Env    *config_env.AppEnv
	DB     *db_client.DB
}

type FieldConflict struct {
	Base      any `json:"base"`
	Main      any `json:"main"`
	Workspace any `json:"workspace"`
}

type CamProperty string

const MaxMergeDepth = 10

func (t *WorkspaceRoute) InitRoute(router gin.IRouter) gin.IRouter {
	router.GET("/projects/:projectId/models/:modelId/workspaces/:workspaceId", t.getWorkspace)
	router.POST("/projects/:projectId/models/:modelId/workspaces/me", t.postWorkspaceMe)
	router.DELETE("/projects/:projectId/models/:modelId/workspaces/me", t.deleteWorkspaceMe)

	router.POST("/projects/:projectId/models/:modelId/workspaces/me/resolve", t.postResolveWorkspaceMe)
	router.POST("/projects/:projectId/models/:modelId/workspaces/me/merge", t.postMergeWorkspace)
	return router
}
