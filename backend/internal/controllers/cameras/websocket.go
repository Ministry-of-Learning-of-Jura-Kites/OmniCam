package controller_camera

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	config_env "omnicam.com/backend/config"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
	messages_cameras "omnicam.com/backend/pkg/messages/cameras"
	camera "omnicam.com/backend/pkg/messages/protobufs"
)

type CameraAutosaveRoute struct {
	Logger   *zap.Logger
	Env      *config_env.AppEnv
	DB       *db_sqlc_gen.Queries
	Upgrader websocket.Upgrader
}

func (t *CameraAutosaveRoute) handleEventDelete(c *gin.Context, conn *websocket.Conn, modelId uuid.UUID, deleteId string) {
	// eventContent := event.GetDeleteId()

	// var cameras Cameras
	// err := json.Unmarshal(workspace.Cameras, &cameras)
	// if err != nil {
	// 	conn.WriteMessage(websocket.TextMessage, []byte("error"))
	// 	return
	// }
	// delete(cameras, deleteId)
	// marshalled, err := json.Marshal(cameras)
	// if err != nil {
	// 	conn.WriteMessage(websocket.TextMessage, []byte("error"))
	// 	return
	// }

	_, err := uuid.Parse(deleteId)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("error"))
		return
	}

	err = t.DB.UpdateWorkspaceCams(c, db_sqlc_gen.UpdateWorkspaceCamsParams{
		Key:     []string{deleteId},
		UserID:  uuid.Nil,
		ModelID: modelId,
	})
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("error"))
		return
	}

	conn.WriteMessage(websocket.TextMessage, []byte("ok"))
}

func (t *CameraAutosaveRoute) handleEventUpsert(c *gin.Context, conn *websocket.Conn, modelId uuid.UUID, upsert *camera.Camera) {
	camera := messages_cameras.CameraStruct{
		Name:           upsert.Name,
		AngleX:         upsert.AngleX,
		AngleY:         upsert.AngleY,
		AngleZ:         upsert.AngleZ,
		AngleW:         upsert.AngleW,
		PosX:           upsert.PosX,
		PosY:           upsert.PosY,
		PosZ:           upsert.PosZ,
		Fov:            upsert.Fov,
		IsHidingArrows: upsert.IsHidingArrows,
		IsHidingWheels: upsert.IsHidingWheels,
	}

	marshalled, err := json.Marshal(camera)
	if err != nil {
		t.Logger.Error("Error while marshaling camera", zap.Error(err))
		conn.WriteMessage(websocket.TextMessage, []byte("error"))
		return
	}
	err = t.DB.UpdateWorkspaceCams(c, db_sqlc_gen.UpdateWorkspaceCamsParams{
		Key:     []string{upsert.Id},
		Value:   marshalled,
		UserID:  uuid.Nil,
		ModelID: modelId,
	})
	if err != nil {
		t.Logger.Error("Error while updating workspace cameras", zap.Error(err))
		conn.WriteMessage(websocket.TextMessage, []byte("error"))
		return
	}
	conn.WriteMessage(websocket.TextMessage, []byte("ok"))
}

func (t *CameraAutosaveRoute) get(c *gin.Context) {
	strId := c.Param("modelId")

	decodedBytes, err := base64.RawURLEncoding.DecodeString(strId)
	if err != nil {
		t.Logger.Error("error decoding Base64", zap.Error(err))
		return
	}
	modelId, err := uuid.FromBytes(decodedBytes)
	if err != nil {
		t.Logger.Error("error while converting str id to uuid", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model ID"})
		return
	}

	_, err = t.DB.GetWorkspaceByID(c, db_sqlc_gen.GetWorkspaceByIDParams{
		UserID:  uuid.Nil,
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

			for _, event := range events.Events {
				switch event.Type {
				case camera.CameraEventType_CAMERA_EVENT_TYPE_DELETE:
					t.handleEventDelete(c, conn, modelId, event.GetDeleteId())
				case camera.CameraEventType_CAMERA_EVENT_TYPE_UPSERT:
					t.handleEventUpsert(c, conn, modelId, event.GetUpsert())
				}
			}
		}
	}()
}

func (t *CameraAutosaveRoute) InitRoute(router gin.IRouter) gin.IRouter {
	router.GET("/projects/:projectId/models/:modelId/autosave", t.get)
	return router
}
