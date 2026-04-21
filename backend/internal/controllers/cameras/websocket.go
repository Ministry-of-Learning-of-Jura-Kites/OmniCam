package controller_camera

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
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
	Logger   *zap.Logger
	Env      *config_env.AppEnv
	DB       *db_client.DB
	Nc       *nats.Conn
	Upgrader websocket.Upgrader
}

type EventContext struct {
	Gin     *gin.Context
	Conn    *websocket.Conn
	ModelID uuid.UUID
	UserID  uuid.UUID
	Subject string
}

// Camera handlers
func (t *UpdateEventRoute) handleEventDelete(
	e *EventContext, deleteId string,
) {
	_, err := uuid.Parse(deleteId)
	if err != nil {
		return
	}

	newVersion, err := t.DB.Queries.UpdateWorkspaceCams(e.Gin, db_sqlc_gen.UpdateWorkspaceCamsParams{
		Key:     []string{deleteId},
		UserID:  e.UserID,
		ModelID: e.ModelID,
	})
	if err != nil {
		t.Logger.Error("error while updating workspace", zap.Error(err))
		return
	}

	resp := &protobufs.AutosaveEventResponse{
		LastUpdatedVersion: newVersion,
	}
	t.sendAutosaveEventResponse(e, resp)
}

func (t *UpdateEventRoute) handleEventUpsert(
	e *EventContext, upsert *protobufs.Camera,
) {
	cam := messages_cameras.ProtoCamToCam(upsert)
	marshalled, err := json.Marshal(cam)
	if err != nil {
		t.Logger.Error("error while marshaling camera", zap.Error(err))
		return
	}
	newVersion, err := t.DB.Queries.UpdateWorkspaceCams(e.Gin, db_sqlc_gen.UpdateWorkspaceCamsParams{
		Key:     []string{upsert.Id},
		Value:   marshalled,
		UserID:  e.UserID,
		ModelID: e.ModelID,
	})
	if err != nil {
		t.Logger.Error("error while updating workspace cameras", zap.Error(err))
		return
	}

	resp := &protobufs.AutosaveEventResponse{
		LastUpdatedVersion: newVersion,
	}
	t.sendAutosaveEventResponse(e, resp)
}

// Calibration handler
func (t *UpdateEventRoute) handleCalibration(
	e *EventContext,
	event *protobufs.AutosaveEvent_Calibrate,
	inputVersion uint32,
	currentVersion *int32,
) {
	if inputVersion <= uint32(*currentVersion) {
		return // stale/duplicate
	}

	row, err := t.DB.Queries.UpdateWorkspaceCalibration(e.Gin, db_sqlc_gen.UpdateWorkspaceCalibrationParams{
		UserID:      e.UserID,
		ModelID:     e.ModelID,
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
	t.sendAutosaveEventResponse(e, resp)
}

func (t *UpdateEventRoute) handleFaceUpsert(
	e *EventContext,
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

	newVersion, err := t.DB.Queries.UpdateWorkspaceTargetTrapezoids(e.Gin, db_sqlc_gen.UpdateWorkspaceTargetTrapezoidsParams{
		Key:     []string{event.FaceUpsert.CoverageFace.Id},
		Value:   marshalled,
		UserID:  e.UserID,
		ModelID: e.ModelID,
	})
	if err != nil {
		t.Logger.Error("error updating face", zap.Error(err))
		return
	}

	resp := &protobufs.AutosaveEventResponse{
		LastUpdatedVersion: newVersion,
	}
	t.sendAutosaveEventResponse(e, resp)
}

func (t *UpdateEventRoute) handleFaceDelete(
	e *EventContext,
	event *protobufs.AutosaveEvent_FaceDelete,
	inputVersion uint32,
	currentVersion *int32) {
	if inputVersion <= uint32(*currentVersion) {
		return // stale/duplicate
	}

	newVersion, err := t.DB.Queries.UpdateWorkspaceTargetTrapezoids(e.Gin, db_sqlc_gen.UpdateWorkspaceTargetTrapezoidsParams{
		Key:     []string{event.FaceDelete.Id},
		Value:   nil,
		UserID:  e.UserID,
		ModelID: e.ModelID,
	})
	fmt.Println(event.FaceDelete.Id)
	if err != nil {
		t.Logger.Error("error updating face", zap.Error(err))
		return
	}

	resp := &protobufs.AutosaveEventResponse{
		LastUpdatedVersion: newVersion,
	}
	t.sendAutosaveEventResponse(e, resp)
}

func (t *UpdateEventRoute) sendAutosaveEventResponse(e *EventContext, resp *protobufs.AutosaveEventResponse) {
	dataBytes, err := proto.Marshal(&protobufs.WorkspaceEventResponse{
		Resp: &protobufs.WorkspaceEventResponse_Autosave{
			Autosave: resp,
		},
	})
	if err != nil {
		t.Logger.Error("error marshalling response", zap.Error(err))
		return
	}

	e.Conn.WriteMessage(websocket.BinaryMessage, dataBytes)
}

func (t *UpdateEventRoute) handleAutosaveEvent(
	e *EventContext, currentVersion *int32, casted *protobufs.AutosaveEventRequest) {
	if casted.Version <= uint32(*currentVersion) {
		return // stale
	}
	for _, camEvent := range casted.Events {
		switch ce := camEvent.GetEvent().(type) {
		case *protobufs.AutosaveEvent_Delete:
			t.handleEventDelete(e, ce.Delete.Id)
		case *protobufs.AutosaveEvent_Upsert:
			t.handleEventUpsert(e, ce.Upsert.Camera)
		case *protobufs.AutosaveEvent_Calibrate:
			t.handleCalibration(e, ce, casted.Version, currentVersion)
		case *protobufs.AutosaveEvent_FaceDelete:
			t.handleFaceDelete(e, ce, casted.Version, currentVersion)
		case *protobufs.AutosaveEvent_FaceUpsert:
			t.handleFaceUpsert(e, ce, casted.Version, currentVersion)
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

func (t *UpdateEventRoute) sendOptimizeInternalError(conn *websocket.Conn, jobId string) {
	resp := &protobufs.OptimizationEventResp{
		JobId: jobId,
		Payload: &protobufs.OptimizationEventResp_ErrorResp{
			ErrorResp: &protobufs.ErrorOptimizationEventResp{
				Error: "INTERNAL_ERROR",
			},
		},
	}
	t.sendOptimizationEventResp(conn, resp)
}

func (t *UpdateEventRoute) handleOptimizeEvent(projectId uuid.UUID, modelId uuid.UUID, conn *websocket.Conn, casted *protobufs.OptimizationEventReq) {
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
			"amount": c.GetAmount(),
		})
	}

	jobId := uuid.New().String()

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

	pubTopic := strings.NewReplacer("{jobId}", jobId).Replace(t.Env.OptiReqTopicPattern)

	msg, err := t.Nc.Request(pubTopic, jsonData, 2*time.Minute)
	if err != nil {
		t.Logger.Error("Failed to publish to nats stream for optimiz algo",
			zap.Error(err),
			zap.String("job_id", jobId),
		)
		t.sendOptimizeInternalError(conn, jobId)
		return
	}

	optiResp := &protobufs.OptimizationEventResp{}

	if err := protojson.Unmarshal(msg.Data, optiResp); err != nil {
		t.Logger.Error("failed to unmarshal proto-json", zap.Error(err))
		t.sendOptimizeInternalError(conn, jobId)
		return
	}

	t.sendOptimizationEventResp(conn, optiResp)
}

// Have to use websocket instead of SSE because protobuf is binary
func (t *UpdateEventRoute) getLivestream(c *gin.Context) {
	strModelId := c.Param("modelId")
	modelId, err := utils.ParseUuidBase64(strModelId)
	if err != nil {
		t.Logger.Error("error while converting str id to uuid", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model ID"})
		return
	}

	// TODO: Secure sharable link(?)
	strWorkspaceOwnerId := c.Param("workspaceOwnerId")
	workspaceOwnerId, err := utils.ParseUuidBase64(strWorkspaceOwnerId)
	if err != nil {
		t.Logger.Error("error while converting str id to uuid", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model ID"})
		return
	}

	// Check owner
	_, err = t.DB.Queries.GetWorkspaceByID(c, db_sqlc_gen.GetWorkspaceByIDParams{
		UserID:  workspaceOwnerId,
		ModelID: modelId,
	})
	if err != nil {
		// Respond not found for security
		t.Logger.Error("workspace not found", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	conn, err := t.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	sendData := func(data []byte) {
		conn.WriteMessage(websocket.BinaryMessage, data)
	}

	subject := strings.NewReplacer("modelId", modelId.String(), "userId", workspaceOwnerId.String()).Replace(t.Env.LivestreamTopicPattern)

	go func() {
		sub, err := t.Nc.Subscribe(subject, func(msg *nats.Msg) {
			sendData(msg.Data)
		})
		if err != nil {
			// TODO: Error
		}

		<-c.Done()

		sub.Drain()
	}()
}

// Main WebSocket handler
func (t *UpdateEventRoute) getAutosave(c *gin.Context) {
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

	// Check owner
	workspace, err := t.DB.Queries.GetWorkspaceByID(c, db_sqlc_gen.GetWorkspaceByIDParams{
		UserID:  userId,
		ModelID: modelId,
	})
	if err != nil {
		// Respond not found for security
		t.Logger.Error("workspace not found", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	conn, err := t.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	subject := strings.NewReplacer("modelId", modelId.String(), "userId", userId.String()).Replace(t.Env.LivestreamTopicPattern)

	go func() {
		defer conn.Close()

		currentVersion := workspace.Version

		e := &EventContext{
			Gin:     c,
			Conn:    conn,
			ModelID: modelId,
			UserID:  userId,
			Subject: subject,
		}

		// Send initial state on connect — both camera version + calibration values
		initResp := &protobufs.AutosaveEventResponse{
			LastUpdatedVersion: currentVersion,
		}
		t.sendAutosaveEventResponse(e, initResp)

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
				t.handleAutosaveEvent(e, &currentVersion, casted.Autosave)
				t.Nc.Publish(e.Subject, rawMsg)
			case *protobufs.WorkspaceEventRequest_Optimize:
				go func() {
					t.handleOptimizeEvent(projectId, modelId, conn, casted.Optimize)
				}()
			}
		}
	}()
}

func (t *UpdateEventRoute) InitRoute(router gin.IRouter) gin.IRouter {
	router.GET("/projects/:projectId/models/:modelId/autosave", t.getAutosave)
	router.GET("/models/:modelId/livestream/:workspaceOwnerId", t.getLivestream)
	return router
}
