package controller_camera

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgtype"
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

type CameraRequest struct {
	Name    string        `json:"name" binding:"required"`
	AngleX  pgtype.Float8 `json:"angle_x"`
	AngleY  pgtype.Float8 `json:"angle_y"`
	AngleZ  pgtype.Float8 `json:"angle_z"`
	AngleW  pgtype.Float8 `json:"angle_w"`
	PosX    pgtype.Float8 `json:"pos_x"`
	PosY    pgtype.Float8 `json:"pos_y"`
	PosZ    pgtype.Float8 `json:"pos_z"`
	ModelID string        `json:"model_id" binding:"required"`
	UserID  *string       `json:"user_id"`
}

func (t *CameraAutosaveRoute) get(c *gin.Context) {
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
				// eventContent := event.GetDeleteId()
				conn.WriteMessage(websocket.TextMessage, []byte("ok"))
			case camera.CameraEventType_CAMERA_EVENT_TYPE_UPSERT:
				eventContent := event.GetUpsert()
				t.Logger.Info("testt", zap.Any("name", eventContent.Name))
				conn.WriteMessage(websocket.TextMessage, []byte("ok"))
			}
		}
	}
}

func (t *CameraAutosaveRoute) InitRoute(router gin.IRouter) gin.IRouter {
	router.GET("/projects/:projectId/models/:modelId/autosave", t.get)
	return router
}
