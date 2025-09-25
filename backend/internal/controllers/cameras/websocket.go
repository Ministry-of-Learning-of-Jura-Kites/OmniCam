package controller_camera

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	config_env "omnicam.com/backend/config"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
	camera "omnicam.com/backend/pkg/messages/protobufs"
)

type CameraAutosaveRoute struct {
	Logger   *zap.Logger
	Env      *config_env.AppEnv
	DB       *db_sqlc_gen.Queries
	Upgrader websocket.Upgrader
}

type Cameras = map[string]CameraStruct

type CameraStruct struct {
	Name   string  `json:"name" binding:"required"`
	AngleX float64 `json:"angle_x"`
	AngleY float64 `json:"angle_y"`
	AngleZ float64 `json:"angle_z"`
	AngleW float64 `json:"angle_w"`
	PosX   float64 `json:"pos_x"`
	PosY   float64 `json:"pos_y"`
	PosZ   float64 `json:"pos_z"`
}

func (t *CameraAutosaveRoute) handleEventDelete(workspace db_sqlc_gen.GetModelWorkspaceCamsByIDRow, conn *websocket.Conn, modelId uuid.UUID, deleteId string) {
	// eventContent := event.GetDeleteId()
	var cameras Cameras
	err := json.Unmarshal(workspace.Cameras, &cameras)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("error"))
		return
	}
	delete(cameras, deleteId)
	marshalled, err := json.Marshal(cameras)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("error"))
		return
	}
	err = t.DB.UpdateWorkspaceCams(context.Background(), db_sqlc_gen.UpdateWorkspaceCamsParams{
		Cameras: marshalled,
		UserID:  uuid.Nil,
		ModelID: modelId,
	})
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("error"))
		return
	}
	conn.WriteMessage(websocket.TextMessage, []byte("ok"))
}

func (t *CameraAutosaveRoute) handleEventUpsert(workspace db_sqlc_gen.GetModelWorkspaceCamsByIDRow, conn *websocket.Conn, modelId uuid.UUID, upsert *camera.Camera) {
	// eventContent := event.GetDeleteId()
	var cameras Cameras
	err := json.Unmarshal(workspace.Cameras, &cameras)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("error"))
		return
	}
	cameras[upsert.Id] = CameraStruct{
		Name:   upsert.Name,
		AngleX: upsert.AngleX,
		AngleY: upsert.AngleY,
		AngleZ: upsert.AngleZ,
		AngleW: upsert.AngleW,
		PosX:   upsert.PosX,
		PosY:   upsert.PosY,
		PosZ:   upsert.PosZ,
	}
	marshalled, err := json.Marshal(cameras)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("error"))
		return
	}
	err = t.DB.UpdateWorkspaceCams(context.Background(), db_sqlc_gen.UpdateWorkspaceCamsParams{
		Cameras: marshalled,
		UserID:  uuid.Nil,
		ModelID: modelId,
	})
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("error"))
		return
	}
	conn.WriteMessage(websocket.TextMessage, []byte("ok"))
}

func (t *CameraAutosaveRoute) get(c *gin.Context) {
	strId := c.Param("modelId")
	modelId, err := uuid.Parse(strId)
	if err != nil {
		t.Logger.Error("error while converting str id to uuid", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model ID"})
		return
	}

	workspace, err := t.DB.GetModelWorkspaceCamsByID(c, db_sqlc_gen.GetModelWorkspaceCamsByIDParams{
		UserID:  uuid.Nil,
		ModelID: modelId,
	})
	if err != nil {
		t.Logger.Error("model not found", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	conn, err := t.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	for {
		// Read message from client
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
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
				t.handleEventDelete(workspace, conn, modelId, event.GetDeleteId())
			case camera.CameraEventType_CAMERA_EVENT_TYPE_UPSERT:
				t.handleEventUpsert(workspace, conn, modelId, event.GetUpsert())
			}
		}
	}
}

func (t *CameraAutosaveRoute) InitRoute(router gin.IRouter) gin.IRouter {
	router.GET("/projects/:projectId/models/:modelId/autosave", t.get)
	return router
}
