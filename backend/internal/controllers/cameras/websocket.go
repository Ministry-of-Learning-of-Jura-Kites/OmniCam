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

	resp := &camera.AutosaveResponse{
		LastUpdatedVersion: newVersion,
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

	resp := &camera.AutosaveResponse{
		LastUpdatedVersion: newVersion,
	}
	sendResponse(t.Logger, conn, resp)
}

// Calibration handler
func (t *CameraAutosaveRoute) handleCalibration(
	c *gin.Context, conn *websocket.Conn,
	modelId uuid.UUID, userId uuid.UUID,
	event *camera.AutosaveEvent_Calibrate,
	inputVersion uint32,
	currentVersion *int32,
) {
	if inputVersion <= uint32(*currentVersion) {
		return // stale/duplicate
	}

	row, err := t.DB.Queries.UpdateWorkspaceCalibration(c, db_sqlc_gen.UpdateWorkspaceCalibrationParams{
		UserID:      userId,
		ModelID:     modelId,
		ScaleFactor: event.Calibrate.ScaleFactor,
		ModelHeight: event.Calibrate.ModelHeight,
	})
	if err != nil {
		t.Logger.Error("error while marshelling first response", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	*currentVersion = row.Version

	resp := &camera.AutosaveResponse{
		LastUpdatedVersion: row.Version,
	}
	sendResponse(t.Logger, conn, resp)
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
		defer conn.Close()

		currentVersion := workspace.Version

		// Send initial state on connect — both camera version + calibration values
		initResp := &camera.AutosaveResponse{
			LastUpdatedVersion: currentVersion,
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
			_, rawMsg, err := conn.ReadMessage()
			if err != nil {
				t.Logger.Error("error from reading message from client", zap.Error(err))
				break
			}

			// Decode unified wrapper
			msg := &camera.AutosaveMessage{}
			if err := proto.Unmarshal(rawMsg, msg); err != nil {
				t.Logger.Error("error unmarshalling event", zap.Error(err))
				continue
			}

			if msg.Version <= uint32(currentVersion) {
				continue // stale
			}
			for _, camEvent := range msg.Events {
				switch ce := camEvent.GetEvent().(type) {
				case *camera.AutosaveEvent_Delete:
					t.handleEventDelete(c, conn, modelId, userId, ce.Delete.Id)
				case *camera.AutosaveEvent_Upsert:
					t.handleEventUpsert(c, conn, modelId, userId, ce.Upsert.Camera)
				case *camera.AutosaveEvent_Calibrate:
					t.handleCalibration(c, conn, modelId, userId, ce, msg.Version, &currentVersion)
				}
			}
		}
	}()
}

func (t *CameraAutosaveRoute) InitRoute(router gin.IRouter) gin.IRouter {
	router.GET("/projects/:projectId/models/:modelId/autosave", t.get)
	return router
}
