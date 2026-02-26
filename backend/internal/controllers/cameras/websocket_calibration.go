package controller_camera

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	config_env "omnicam.com/backend/config"
	"omnicam.com/backend/internal/utils"
	db_client "omnicam.com/backend/pkg/db"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
	camera "omnicam.com/backend/pkg/messages/protobufs"
)

type CalibrationAutosaveRoute struct {
	Logger   *zap.Logger
	Env      *config_env.AppEnv
	DB       *db_client.DB
	Upgrader websocket.Upgrader
}

func (t *CalibrationAutosaveRoute) get(c *gin.Context) {
	strModelId := c.Param("modelId")
	modelId, err := utils.ParseUuidBase64(strModelId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model ID"})
		return
	}

	userId, err := utils.GetUuidFromCtx(c, "userId")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	workspace, err := t.DB.Queries.GetWorkspaceByID(c, db_sqlc_gen.GetWorkspaceByIDParams{
		UserID:  userId,
		ModelID: modelId,
	})
	if err != nil {
		t.Logger.Error("workspace not found", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	conn, err := t.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	go func() {
		defer conn.Close()

		// Send initial calibration values on connect
		firstResponse := &camera.CalibrationAutosaveResponse{
			Payload: &camera.CalibrationAutosaveResponse_Ack{
				Ack: &camera.CalibrationAutosaveResponseAck{
					LastUpdatedVersion: workspace.Version,
					ScaleFactor:        workspace.ScaleFactor,
					ModelHeight:        workspace.ModelHeight,
				},
			},
		}
		firstBytes, _ := proto.Marshal(firstResponse)
		conn.WriteMessage(websocket.BinaryMessage, firstBytes)

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				t.Logger.Error("read error", zap.Error(err))
				break
			}

			event := &camera.CalibrationSaveEvent{}
			if err := proto.Unmarshal(msg, event); err != nil {
				continue
			}

			// Reject stale/duplicate events
			if event.Version <= uint32(workspace.Version) {
				continue
			}

			row, err := t.DB.Queries.UpdateWorkspaceCalibration(c, db_sqlc_gen.UpdateWorkspaceCalibrationParams{
				UserID:      userId,
				ModelID:     modelId,
				ScaleFactor: event.Calibration.ScaleFactor,
				ModelHeight: event.Calibration.ModelHeight,
			})
			if err != nil {
				t.Logger.Error("error updating calibration", zap.Error(err))
				continue
			}

			// Update local version to prevent re-processing
			workspace.Version = row.Version

			resp := &camera.CalibrationAutosaveResponse{
				Payload: &camera.CalibrationAutosaveResponse_Ack{
					Ack: &camera.CalibrationAutosaveResponseAck{
						LastUpdatedVersion: row.Version,
						ScaleFactor:        row.ScaleFactor,
						ModelHeight:        row.ModelHeight,
					},
				},
			}
			respBytes, _ := proto.Marshal(resp)
			conn.WriteMessage(websocket.BinaryMessage, respBytes)
		}
	}()
}

func (t *CalibrationAutosaveRoute) InitRoute(router gin.IRouter) gin.IRouter {
	router.GET("/projects/:projectId/models/:modelId/autosave/calibration", t.get)
	return router
}
