package controller_camera

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	config_env "omnicam.com/backend/config"
	"omnicam.com/backend/internal/utils"
	db_client "omnicam.com/backend/pkg/db"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
	messages_cameras "omnicam.com/backend/pkg/messages/cameras"
	camera "omnicam.com/backend/pkg/messages/protobufs"
)

type CameraAutosaveRoute struct {
	Logger   *zap.Logger
	Env      *config_env.AppEnv
	DB       *db_client.DB
	Upgrader websocket.Upgrader
}

func (t *CameraAutosaveRoute) handleEventDelete(c *gin.Context, conn *websocket.Conn, modelId uuid.UUID, deleteId string) {
	userId, err := utils.GetUuidFromCtx(c, "userId")
	if err != nil {
		t.Logger.Error("error while getting userId form", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	_, err = uuid.Parse(deleteId)
	if err != nil {
		// conn.WriteMessage(websocket.TextMessage, []byte("error"))
		return
	}

	newVersion, err := t.DB.Queries.UpdateWorkspaceCams(c, db_sqlc_gen.UpdateWorkspaceCamsParams{
		Key:     []string{deleteId},
		UserID:  userId,
		ModelID: modelId,
	})
	if err != nil {
		t.Logger.Error("Error while updating workspace", zap.Error(err))
		return
	}

	resp := &camera.CameraAutosaveResponse{
		Payload: &camera.CameraAutosaveResponse_Ack{
			Ack: &camera.CameraAutosaveResponseAck{
				LastUpdatedVersion: newVersion,
			},
		},
	}
	respMarshalled, err := proto.Marshal(resp)
	if err != nil {
		t.Logger.Error("error while marshelling first response", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	conn.WriteMessage(websocket.BinaryMessage, respMarshalled)
}

func (t *CameraAutosaveRoute) handleEventUpsert(c *gin.Context, userId uuid.UUID, conn *websocket.Conn, modelId uuid.UUID, upsert *camera.Camera) {
	cam := messages_cameras.ProtoCamToCam(upsert)

	marshalled, err := json.Marshal(cam)
	if err != nil {
		t.Logger.Error("Error while marshaling camera", zap.Error(err))
		// conn.WriteMessage(websocket.TextMessage, []byte("error"))
		return
	}
	newVersion, err := t.DB.Queries.UpdateWorkspaceCams(c, db_sqlc_gen.UpdateWorkspaceCamsParams{
		Key:     []string{upsert.Id},
		Value:   marshalled,
		UserID:  userId,
		ModelID: modelId,
	})
	if err != nil {
		t.Logger.Error("Error while updating workspace cameras", zap.Error(err))
		// conn.WriteMessage(websocket.TextMessage, []byte("error"))
		return
	}
	resp := &camera.CameraAutosaveResponse{
		Payload: &camera.CameraAutosaveResponse_Ack{
			Ack: &camera.CameraAutosaveResponseAck{
				LastUpdatedVersion: newVersion,
			},
		},
	}
	firstRespMarshalled, err := proto.Marshal(resp)
	if err != nil {
		t.Logger.Error("error while marshelling first response", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	conn.WriteMessage(websocket.BinaryMessage, firstRespMarshalled)
}

func (t *CameraAutosaveRoute) get(c *gin.Context) {
	strModelId := c.Param("modelId")
	modelId, err := utils.ParseUuidBase64(strModelId)
	if err != nil {
		t.Logger.Error("error while converting str id to uuid", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model ID"})
		return
	}

	userId, err := utils.GetUuidFromCtx(c, "userId")
	if err != nil {
		t.Logger.Error("error while getting userId form", zap.Error(err))
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
		firstResponse := &camera.CameraAutosaveResponse{
			Payload: &camera.CameraAutosaveResponse_Ack{
				Ack: &camera.CameraAutosaveResponseAck{
					LastUpdatedVersion: workspace.Version,
				},
			},
		}
		firstRespMarshalled, err := proto.Marshal(firstResponse)
		if err != nil {
			t.Logger.Error("error while marshelling first response", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}
		conn.WriteMessage(websocket.BinaryMessage, firstRespMarshalled)

		defer conn.Close()

		for {
			// Read message from client
			_, msg, err := conn.ReadMessage()
			if err != nil {
				t.Logger.Error("error from reading message from client", zap.Error(err))
				break
			}

			events := &camera.CameraSaveEventSeries{}
			err = proto.Unmarshal(msg, events)
			if err != nil {
				continue
			}

			if events.Version <= uint32(workspace.Version) {
				// Repeated event sent (From Websocket reconnect)
				continue
			}

			for _, event := range events.Events {
				switch v := event.GetEvent().(type) {
				case *camera.CameraSaveEvent_Delete:
					t.handleEventDelete(c, conn, modelId, v.Delete.Id)
				case *camera.CameraSaveEvent_Upsert:
					t.handleEventUpsert(c, userId, conn, modelId, v.Upsert.Camera)
				}
			}
		}
	}()
}

func (t *CameraAutosaveRoute) InitRoute(router gin.IRouter) gin.IRouter {
	router.GET("/projects/:projectId/models/:modelId/autosave", t.get)
	return router
}
