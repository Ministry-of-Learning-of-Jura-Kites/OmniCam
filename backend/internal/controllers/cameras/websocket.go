package controller_camera

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	config_env "omnicam.com/backend/config"
	"omnicam.com/backend/internal/utils"
	db_client "omnicam.com/backend/pkg/db"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
	messages_cameras "omnicam.com/backend/pkg/messages/cameras"
	"omnicam.com/backend/pkg/messages/protobufs"
	messsages_trapezoids "omnicam.com/backend/pkg/messages/trapezoids"
)

type UpdateEventRoute struct {
	Logger          *zap.Logger
	Env             *config_env.AppEnv
	DB              *db_client.DB
	RedisClient     *redis.Client
	Upgrader        websocket.Upgrader
	OptimizeRespMap *sync.Map
}

// Camera handlers
func (t *UpdateEventRoute) handleEventDelete(
	c *gin.Context, conn *websocket.Conn,
	modelId uuid.UUID, userId uuid.UUID, deleteId string,
) {
	_, err := uuid.Parse(deleteId)
	if err != nil {
		return
	}

	newVersion, err := t.DB.Queries.UpdateWorkspaceCams(c, db_sqlc_gen.UpdateWorkspaceCamsParams{
		Key:     []string{deleteId},
		UserID:  userId,
		ModelID: modelId,
	})
	if err != nil {
		t.Logger.Error("error while updating workspace", zap.Error(err))
		return
	}

	resp := &protobufs.AutosaveEventResponse{
		LastUpdatedVersion: newVersion,
	}
	sendAutosaveEventResponse(t.Logger, conn, resp)
}

func (t *UpdateEventRoute) handleEventUpsert(
	c *gin.Context, conn *websocket.Conn,
	modelId uuid.UUID, userId uuid.UUID, upsert *protobufs.Camera,
) {
	cam := messages_cameras.ProtoCamToCam(upsert)
	marshalled, err := json.Marshal(cam)
	if err != nil {
		t.Logger.Error("error while marshaling camera", zap.Error(err))
		return
	}
	newVersion, err := t.DB.Queries.UpdateWorkspaceCams(c, db_sqlc_gen.UpdateWorkspaceCamsParams{
		Key:     []string{upsert.Id},
		Value:   marshalled,
		UserID:  userId,
		ModelID: modelId,
	})
	if err != nil {
		t.Logger.Error("error while updating workspace cameras", zap.Error(err))
		return
	}

	resp := &protobufs.AutosaveEventResponse{
		LastUpdatedVersion: newVersion,
	}
	sendAutosaveEventResponse(t.Logger, conn, resp)
}

// Calibration handler
func (t *UpdateEventRoute) handleCalibration(
	c *gin.Context, conn *websocket.Conn,
	modelId uuid.UUID, userId uuid.UUID,
	event *protobufs.AutosaveEvent_Calibrate,
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
		t.Logger.Error("error updating calibration", zap.Error(err))
		return
	}

	*currentVersion = row.Version

	resp := &protobufs.AutosaveEventResponse{
		LastUpdatedVersion: row.Version,
	}
	sendAutosaveEventResponse(t.Logger, conn, resp)
}

func (t *UpdateEventRoute) handleFaceUpsert(
	c *gin.Context, conn *websocket.Conn,
	modelId uuid.UUID, userId uuid.UUID,
	event *protobufs.AutosaveEvent_FaceUpsert,
	inputVersion uint32,
	currentVersion *int32) {
	if inputVersion <= uint32(*currentVersion) {
		return // stale/duplicate
	}

	trapezoid := messsages_trapezoids.ProtoTrapezoidToTrapezoid(event.FaceUpsert.CoverageFace)
	marshalled, err := json.Marshal(trapezoid)
	if err != nil {
		t.Logger.Error("error while marshaling camera", zap.Error(err))
		return
	}

	newVersion, err := t.DB.Queries.UpdateWorkspaceTargetTrapezoids(c, db_sqlc_gen.UpdateWorkspaceTargetTrapezoidsParams{
		Key:     []string{event.FaceUpsert.CoverageFace.Id},
		Value:   marshalled,
		UserID:  userId,
		ModelID: modelId,
	})
	if err != nil {
		t.Logger.Error("error updating face", zap.Error(err))
		return
	}

	resp := &protobufs.AutosaveEventResponse{
		LastUpdatedVersion: newVersion,
	}
	sendAutosaveEventResponse(t.Logger, conn, resp)
}

func (t *UpdateEventRoute) handleFaceDelete(
	c *gin.Context, conn *websocket.Conn,
	modelId uuid.UUID, userId uuid.UUID,
	event *protobufs.AutosaveEvent_FaceDelete,
	inputVersion uint32,
	currentVersion *int32) {
	if inputVersion <= uint32(*currentVersion) {
		return // stale/duplicate
	}

	newVersion, err := t.DB.Queries.UpdateWorkspaceTargetTrapezoids(c, db_sqlc_gen.UpdateWorkspaceTargetTrapezoidsParams{
		Key:     []string{event.FaceDelete.Id},
		Value:   nil,
		UserID:  userId,
		ModelID: modelId,
	})
	if err != nil {
		t.Logger.Error("error updating face", zap.Error(err))
		return
	}

	resp := &protobufs.AutosaveEventResponse{
		LastUpdatedVersion: newVersion,
	}
	sendAutosaveEventResponse(t.Logger, conn, resp)
}

func sendAutosaveEventResponse(logger *zap.Logger, conn *websocket.Conn, resp *protobufs.AutosaveEventResponse) {
	bytes, err := proto.Marshal(&protobufs.WorkspaceEventResponse{
		Resp: &protobufs.WorkspaceEventResponse_Autosave{
			Autosave: resp,
		},
	})
	if err != nil {
		logger.Error("error marshalling response", zap.Error(err))
		return
	}
	conn.WriteMessage(websocket.BinaryMessage, bytes)
}

func (t *UpdateEventRoute) handleAutosaveEvent(
	c *gin.Context, conn *websocket.Conn,
	modelId uuid.UUID, userId uuid.UUID, currentVersion *int32, casted *protobufs.AutosaveEventRequest) {
	if casted.Version <= uint32(*currentVersion) {
		return // stale
	}
	for _, camEvent := range casted.Events {
		switch ce := camEvent.GetEvent().(type) {
		case *protobufs.AutosaveEvent_Delete:
			t.handleEventDelete(c, conn, modelId, userId, ce.Delete.Id)
		case *protobufs.AutosaveEvent_Upsert:
			t.handleEventUpsert(c, conn, modelId, userId, ce.Upsert.Camera)
		case *protobufs.AutosaveEvent_Calibrate:
			t.handleCalibration(c, conn, modelId, userId, ce, casted.Version, currentVersion)
		case *protobufs.AutosaveEvent_FaceDelete:
			t.handleFaceDelete(c, conn, modelId, userId, ce, casted.Version, currentVersion)
		case *protobufs.AutosaveEvent_FaceUpsert:
			t.handleFaceUpsert(c, conn, modelId, userId, ce, casted.Version, currentVersion)
		}
	}
}

func (t *UpdateEventRoute) sendOptimizationEventResp(conn *websocket.Conn, optiResp *protobufs.OptimizationEventResp) {
	// 3. Wrap it in the top-level WorkspaceEventResponse (the oneof)
	eventResp := &protobufs.WorkspaceEventResponse{
		Resp: &protobufs.WorkspaceEventResponse_Optimize{
			Optimize: optiResp,
		},
	}

	bytes, err := proto.Marshal(eventResp)
	if err != nil {
		t.Logger.Error("error marshalling response", zap.Error(err))
		return
	}
	conn.WriteMessage(websocket.BinaryMessage, bytes)
}

func (t *UpdateEventRoute) handleOptimizeEvent(projectId uuid.UUID, modelId uuid.UUID, conn *websocket.Conn, casted *protobufs.OptimizationEventReq) {
	ctx := context.Background()

	if len(casted.GetCoverageFace()) == 0 {
		t.Logger.Warn("optimization aborted: no coverage faces provided")
		return
	}
	faces := make([][][]float64, 0, len(casted.GetCoverageFace()))

	for _, face := range casted.GetCoverageFace() {
		var points [][]float64
		for _, p := range face.GetPoints() {
			// Each point is a slice of 3 floats to represent the Tuple
			points = append(points, []float64{p.GetX(), p.GetY(), p.GetZ()})
		}
		faces = append(faces, points)
	}

	camConfigs := make([]map[string]interface{}, 0, len(casted.GetCameraConfig()))
	for _, c := range casted.GetCameraConfig() {
		camConfigs = append(camConfigs, map[string]interface{}{
			"name":   c.GetName(),
			"vfov":   c.GetFov(),
			"pixels": []float64{c.GetWidthRes(), c.GetHeightRes()}, // Maps to Tuple[float, float]
			"amount": 1,                                            // 'amount' is required by your Pydantic model
		})
	}

	jobId := uuid.New().String()

	resChan := make(chan string, 1)
	t.OptimizeRespMap.Store(jobId, resChan)

	payload := map[string]interface{}{
		"faces":       faces,
		"cam_configs": camConfigs,
		"scale":       casted.Scale,
		"job_id":      jobId,
		"project_id":  projectId.String(),
		"model_id":    modelId.String(),
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		t.Logger.Error("failed to serialize optimize request",
			zap.Error(err),
			zap.Int("face_count", len(faces)),
		)
		return
	}

	err = t.RedisClient.XAdd(ctx, &redis.XAddArgs{
		Stream: t.Env.OptiReqTopic, // Use your actual env field name here
		Values: map[string]interface{}{
			"data": string(jsonData), // This matches what model_validate_json(data) expects
		},
	}).Err()

	if err != nil {
		t.Logger.Error("failed to publish to redis stream",
			zap.Error(err),
			zap.String("job_id", jobId),
		)
		return
	}

	go func() {
		defer t.OptimizeRespMap.Delete(jobId)
		select {
		case rawJSON := <-resChan:
			optiResp := &protobufs.OptimizationEventResp{}

			err := protojson.Unmarshal([]byte(rawJSON), optiResp)
			if err != nil {
				t.Logger.Error("failed to unmarshal proto-json", zap.Error(err))
				return
			}

			t.sendOptimizationEventResp(conn, optiResp)
		case <-time.After(10 * time.Minute):
			t.Logger.Warn("optimization timed out", zap.String("job_id", jobId))
		}
	}()
}

// Main WebSocket handler
func (t *UpdateEventRoute) get(c *gin.Context) {
	strProjectId := c.Param("projectId")
	projectId, err := utils.ParseUuidBase64(strProjectId)
	if err != nil {
		t.Logger.Error("error while converting str id to uuid", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model ID"})
		return
	}

	strModelId := c.Param("modelId")
	modelId, err := utils.ParseUuidBase64(strModelId)
	if err != nil {
		t.Logger.Error("error while converting str id to uuid", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model ID"})
		return
	}

	userId, err := utils.GetUuidFromCtx(c, "userId")
	if err != nil {
		t.Logger.Error("error while getting userId", zap.Error(err))
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
		initResp := &protobufs.AutosaveEventResponse{
			LastUpdatedVersion: currentVersion,
		}
		sendAutosaveEventResponse(t.Logger, conn, initResp)

		for {
			_, rawMsg, err := conn.ReadMessage()
			if err != nil {
				t.Logger.Error("error reading message", zap.Error(err))
				break
			}

			// Decode unified wrapper
			msg := &protobufs.WorkspaceEventRequest{}
			if err := proto.Unmarshal(rawMsg, msg); err != nil {
				t.Logger.Error("error unmarshalling event", zap.Error(err))
				continue
			}

			switch casted := msg.Event.(type) {
			case *protobufs.WorkspaceEventRequest_Autosave:
				t.handleAutosaveEvent(c, conn, modelId, userId, &currentVersion, casted.Autosave)
			case *protobufs.WorkspaceEventRequest_Optimize:
				t.handleOptimizeEvent(projectId, modelId, conn, casted.Optimize)
			}
		}
	}()
}

func (t *UpdateEventRoute) InitRoute(router gin.IRouter) gin.IRouter {
	router.GET("/projects/:projectId/models/:modelId/autosave", t.get)
	return router
}
