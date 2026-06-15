package controller_camera

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	config_env "omnicam.com/backend/config"
	"omnicam.com/backend/internal/utils"
	db_client "omnicam.com/backend/pkg/db"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
	"omnicam.com/backend/pkg/logger"
	messages_cameras "omnicam.com/backend/pkg/messages/cameras"
	messages_errors "omnicam.com/backend/pkg/messages/errors"
	"omnicam.com/backend/pkg/messages/protobufs"
	messsages_trapezoids "omnicam.com/backend/pkg/messages/trapezoids"
)

type UpdateEventRoute struct {
	Logger   *zap.Logger
	Env      *config_env.AppEnv
	DB       *db_client.DB
	Nc       *nats.Conn
	Upgrader websocket.Upgrader
	writeMu  sync.Mutex
}

type EventContext struct {
	Gin     *gin.Context
	Conn    *websocket.Conn
	ModelID uuid.UUID
	UserID  uuid.UUID
	Subject string
	Scale   float64
	Height  float64
	Ctx     context.Context
}

func (r *UpdateEventRoute) writeMessage(conn *websocket.Conn, messageType int, data []byte) error {
	r.writeMu.Lock()
	defer r.writeMu.Unlock()
	return conn.WriteMessage(messageType, data)
}

// Camera handlers
func (t *UpdateEventRoute) handleEventDelete(
	e *EventContext, deleteId string,
) {
	_, err := uuid.Parse(deleteId)
	if err != nil {
		return
	}

	newVersion, err := t.DB.Queries.UpdateWorkspaceCams(e.Ctx, db_sqlc_gen.UpdateWorkspaceCamsParams{
		Key:     []string{deleteId},
		UserID:  e.UserID,
		ModelID: e.ModelID,
	})
	if err != nil {
		logger.WithTraceID(e.Ctx, t.Logger).
			Error("error while updating workspace", zap.Error(err))
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
		logger.WithTraceID(e.Ctx, t.Logger).
			Error("error while marshaling camera", zap.Error(err))
		return
	}
	newVersion, err := t.DB.Queries.UpdateWorkspaceCams(e.Ctx, db_sqlc_gen.UpdateWorkspaceCamsParams{
		Key:     []string{upsert.Id},
		Value:   marshalled,
		UserID:  e.UserID,
		ModelID: e.ModelID,
	})
	if err != nil {
		logger.WithTraceID(e.Ctx, t.Logger).
			Error("error while updating workspace cameras", zap.Error(err))
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
		logger.WithTraceID(e.Gin.Request.Context(), t.Logger).Warn("version mismatch", zap.Uint32("inputVersion", inputVersion), zap.Uint32("currentVersion", uint32(*currentVersion)))
		return // stale/duplicate
	}

	row, err := t.DB.Queries.UpdateWorkspaceCalibration(e.Ctx, db_sqlc_gen.UpdateWorkspaceCalibrationParams{
		UserID:      e.UserID,
		ModelID:     e.ModelID,
		ScaleFactor: event.Calibrate.ScaleFactor,
		ModelHeight: event.Calibrate.ModelHeight,
	})
	if err != nil {
		logger.WithTraceID(e.Ctx, t.Logger).
			Error("error updating calibration", zap.Error(err))
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
		logger.WithTraceID(e.Gin.Request.Context(), t.Logger).Warn("version mismatch", zap.Uint32("inputVersion", inputVersion), zap.Uint32("currentVersion", uint32(*currentVersion)))
		return // stale/duplicate
	}

	trapezoid := messsages_trapezoids.ProtoTrapezoidToTrapezoid(event.FaceUpsert.CoverageFace)
	marshalled, err := json.Marshal(trapezoid)
	if err != nil {
		logger.WithTraceID(e.Ctx, t.Logger).
			Error("error while marshaling camera", zap.Error(err))
		return
	}

	newVersion, err := t.DB.Queries.UpdateWorkspaceTargetTrapezoids(e.Ctx, db_sqlc_gen.UpdateWorkspaceTargetTrapezoidsParams{
		Key:     []string{event.FaceUpsert.CoverageFace.Id},
		Value:   marshalled,
		UserID:  e.UserID,
		ModelID: e.ModelID,
	})
	if err != nil {
		logger.WithTraceID(e.Ctx, t.Logger).
			Error("error updating face", zap.Error(err))
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
		logger.WithTraceID(e.Gin.Request.Context(), t.Logger).Warn("version mismatch", zap.Uint32("inputVersion", inputVersion), zap.Uint32("currentVersion", uint32(*currentVersion)))
		return // stale/duplicate
	}

	newVersion, err := t.DB.Queries.UpdateWorkspaceTargetTrapezoids(e.Ctx, db_sqlc_gen.UpdateWorkspaceTargetTrapezoidsParams{
		Key:     []string{event.FaceDelete.Id},
		Value:   nil,
		UserID:  e.UserID,
		ModelID: e.ModelID,
	})
	if err != nil {
		logger.WithTraceID(e.Ctx, t.Logger).
			Error("error updating face", zap.Error(err))
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
		logger.WithTraceID(e.Ctx, t.Logger).
			Error("error marshalling response", zap.Error(err))
		return
	}

	e.Conn.WriteMessage(websocket.BinaryMessage, dataBytes)
}

func (t *UpdateEventRoute) handleAutosaveEvent(
	e *EventContext, currentVersion *int32, casted *protobufs.AutosaveEventRequest) {
	if casted.Version <= uint32(*currentVersion) {
		return // stale
	}

	camUpserts := make(map[string]interface{})
	camDeletes := []string{}
	trapUpserts := make(map[string]interface{})
	trapDeletes := []string{}

	scale := e.Scale
	height := e.Height

	for _, camEvent := range casted.Events {
		switch ce := camEvent.GetEvent().(type) {
		case *protobufs.AutosaveEvent_Upsert:
			camUpserts[ce.Upsert.Camera.Id] = messages_cameras.ProtoCamToCam(ce.Upsert.Camera)
		case *protobufs.AutosaveEvent_Delete:
			camDeletes = append(camDeletes, ce.Delete.Id)
		case *protobufs.AutosaveEvent_FaceUpsert:
			trapUpserts[ce.FaceUpsert.CoverageFace.Id] = messsages_trapezoids.ProtoTrapezoidToTrapezoid(ce.FaceUpsert.CoverageFace)
		case *protobufs.AutosaveEvent_FaceDelete:
			trapDeletes = append(trapDeletes, ce.FaceDelete.Id)
		case *protobufs.AutosaveEvent_Calibrate:
			scale = ce.Calibrate.ScaleFactor
			height = ce.Calibrate.ModelHeight
		}
	}

	camJSON, _ := json.Marshal(camUpserts)
	trapJSON, _ := json.Marshal(trapUpserts)

	// One DB call, one version increment
	newVersion, err := t.DB.Queries.BulkUpdateWorkspace(e.Ctx, db_sqlc_gen.BulkUpdateWorkspaceParams{
		UserID:      e.UserID,
		ModelID:     e.ModelID,
		CamUpserts:  camJSON,
		CamDeletes:  camDeletes,
		TrapUpserts: trapJSON,
		TrapDeletes: trapDeletes,
		ScaleFactor: scale,
		ModelHeight: height,
	})

	if err != nil {
		logger.WithTraceID(e.Ctx, t.Logger).
			Error("batch update failed", zap.Error(err))
		return
	}

	e.Scale = scale
	e.Height = height

	*currentVersion = newVersion
	t.sendAutosaveEventResponse(e, &protobufs.AutosaveEventResponse{
		LastUpdatedVersion: newVersion,
	})
}

func (t *UpdateEventRoute) sendOptimizationEventResp(conn *websocket.Conn, ctx context.Context, optiResp *protobufs.OptimizationEventResp) {
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)

	eventResp := &protobufs.WorkspaceEventResponse{
		Trace: &protobufs.TraceContext{
			Traceparent: carrier["traceparent"],
			Tracestate:  carrier["tracestate"],
		},
		Resp: &protobufs.WorkspaceEventResponse_Optimize{
			Optimize: optiResp,
		},
	}

	bytes, err := proto.Marshal(eventResp)

	if err != nil {
		logger.WithTraceID(ctx, t.Logger).
			Error("error marshalling response", zap.Error(err))
		return
	}
	t.writeMessage(conn, websocket.BinaryMessage, bytes)
	// conn.WriteMessage(websocket.BinaryMessage, bytes)
}

func (t *UpdateEventRoute) sendOptimizeInternalError(conn *websocket.Conn, ctx context.Context, jobId string) {
	resp := &protobufs.OptimizationEventResp{
		JobId: jobId,
		Payload: &protobufs.OptimizationEventResp_ErrorResp{
			ErrorResp: &protobufs.EventError{
				Error: &protobufs.EventError_InternalError{
					InternalError: &protobufs.SimpleError{},
				},
			},
		},
	}
	t.sendOptimizationEventResp(conn, ctx, resp)
}

func (t *UpdateEventRoute) handleOptimizeEvent(ctx context.Context, projectId uuid.UUID, modelId uuid.UUID, conn *websocket.Conn, casted *protobufs.OptimizationEventReq) {
	if len(casted.GetCoverageFace()) == 0 {
		logger.WithTraceID(ctx, t.Logger).
			Warn("optimization aborted: no coverage faces provided")
		return
	}

	tr := otel.Tracer("omnicam-backend")
	ctx, span := tr.Start(ctx, "optimization.request",
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(
			attribute.String("model.id", modelId.String()),
			attribute.String("project.id", projectId.String()),
		),
	)

	defer span.End()

	faces := make([][][]float64, 0, len(casted.GetCoverageFace()))
	for _, face := range casted.GetCoverageFace() {
		var points [][]float64
		for _, p := range face.GetPoints() {
			points = append(points, []float64{p.GetX(), p.GetY(), p.GetZ()})
		}
		faces = append(faces, points)
	}

	camConfigs := make([]map[string]interface{}, 0, len(casted.GetCameraConfig()))
	for _, c := range casted.GetCameraConfig() {
		camConfigs = append(camConfigs, map[string]interface{}{
			"name":   c.GetName(),
			"vfov":   c.GetFov(),
			"pixels": []float64{c.GetWidthRes(), c.GetHeightRes()},
			"amount": c.GetAmount(),
		})
	}

	jobId := uuid.New().String()
	span.SetAttributes(attribute.String("job.id", jobId))

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
		logger.WithTraceID(ctx, t.Logger).
			Error("failed to searilize optimize request", zap.Error(err), zap.Int("face_count", len(faces)))
		span.RecordError(err)
		return
	}

	pubTopic := strings.NewReplacer("{jobId}", jobId).Replace(t.Env.OptiReqTopicPattern)

	natsMsg := nats.NewMsg(pubTopic)
	natsMsg.Data = jsonData

	// inject trace context into NATS headers so algo continues the same trace
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	for k, v := range carrier {
		natsMsg.Header.Set(k, v)
	}

	logger.WithTraceID(ctx, t.Logger).Info("optimization started",
		zap.String("jobId", jobId),
		zap.String("modelId", modelId.String()),
		zap.Int("faceCount", len(faces)),
		zap.Int("camCount", len(camConfigs)),
	)

	msg, err := t.Nc.RequestMsg(natsMsg, 2*time.Minute)
	if err != nil {
		logger.WithTraceID(ctx, t.Logger).
			Error("failed to publish to nats stream for optimization", zap.Error(err), zap.String("job_id", jobId))
		span.RecordError(err)
		t.sendOptimizeInternalError(conn, ctx, jobId)
		return
	}

	optiResp := &protobufs.OptimizationEventResp{}
	if err := protojson.Unmarshal(msg.Data, optiResp); err != nil {
		logger.WithTraceID(ctx, t.Logger).Error("failed to unmarshal proto-json", zap.Error(err))
		span.RecordError(err)
		t.sendOptimizeInternalError(conn, ctx, jobId)
		return
	}

	t.sendOptimizationEventResp(conn, ctx, optiResp)
	logger.WithTraceID(ctx, t.Logger).Info("optimization completed",
		zap.String("jobId", jobId),
		zap.String("modelId", modelId.String()),
	)
}

// Have to use websocket instead of SSE because protobuf is binary
func (t *UpdateEventRoute) getLivestream(c *gin.Context) {
	strModelId := c.Param("modelId")
	modelId, err := utils.ParseUuidBase64(strModelId)
	if err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).
			Error(messages_errors.ErrorWhileConvertToUUID, zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": messages_errors.InvalidModelID})
		return
	}

	strWorkspaceOwnerId := c.Param("workspaceOwnerId")
	workspaceOwnerId, err := utils.ParseUuidBase64(strWorkspaceOwnerId)
	if err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).
			Error(messages_errors.ErrorWhileConvertToUUID, zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": messages_errors.InvalidModelID})
		return
	}

	username := c.GetString("username")

	_, err = t.DB.Queries.GetUserWithModel(c.Request.Context(), db_sqlc_gen.GetUserWithModelParams{
		Username: pgtype.Text{Valid: true, String: username},
		ModelID:  modelId,
	})
	if err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).
			Warn("not a member of project",
				zap.String("modelId", modelId.String()),
				zap.String("username", username),
				zap.Error(err),
			)
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	_, err = t.DB.Queries.GetWorkspaceByID(c.Request.Context(), db_sqlc_gen.GetWorkspaceByIDParams{
		UserID:  workspaceOwnerId,
		ModelID: modelId,
	})
	if err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).
			Error("workspace not found",
				zap.String("modelId", modelId.String()),
				zap.String("workspaceOwnerId", workspaceOwnerId.String()),
				zap.Error(err),
			)
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	conn, err := t.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).
			Error("websocket upgrade failed", zap.Error(err))
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
			logger.WithTraceID(c.Request.Context(), t.Logger).
				Error("failed to subscribe to livestream topic",
					zap.Error(err),
					zap.String("subject", subject),
				)
			resp := &protobufs.LivestreamBroadcast{
				Event: &protobufs.LivestreamBroadcast_Error{
					Error: &protobufs.EventError{
						Error: &protobufs.EventError_InternalError{
							InternalError: &protobufs.SimpleError{},
						},
					},
				},
			}

			dataBytes, err := proto.Marshal(resp)
			if err != nil {
				logger.WithTraceID(c.Request.Context(), t.Logger).
					Error("failed to marshal livestream error response", zap.Error(err))
				return
			}
			conn.WriteMessage(websocket.BinaryMessage, dataBytes)
			return
		}

		<-c.Done()

		if err := sub.Drain(); err != nil {
			logger.WithTraceID(c.Request.Context(), t.Logger).
				Error("error draining NATS subscription",
					zap.Error(err),
					zap.String("subject", subject),
				)
		}
	}()
}

// Main WebSocket handler
func (t *UpdateEventRoute) getAutosave(c *gin.Context) {
	strProjectId := c.Param("projectId")
	projectId, err := utils.ParseUuidBase64(strProjectId)
	if err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).Error(
			messages_errors.ErrorWhileConvertToUUID,
			zap.String("projectId", strProjectId),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": messages_errors.InvalidModelID})
		return
	}

	strModelId := c.Param("modelId")
	modelId, err := utils.ParseUuidBase64(strModelId)
	if err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).Error(
			messages_errors.ErrorWhileConvertToUUID,
			zap.String("modelId", strModelId),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": messages_errors.InvalidModelID})
		return
	}

	userId, err := utils.GetUuidFromCtx(c, "userId")
	if err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).Error(
			"error while getting userId",
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	workspace, err := t.DB.Queries.GetWorkspaceByID(c.Request.Context(), db_sqlc_gen.GetWorkspaceByIDParams{
		UserID:  userId,
		ModelID: modelId,
	})
	if err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).Error(
			"workspace not found",
			zap.String("modelId", modelId.String()),
			zap.String("userId", userId.String()),
			zap.Error(err),
		)
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	conn, err := t.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	subject := strings.NewReplacer("modelId", modelId.String(), "userId", userId.String()).Replace(t.Env.LivestreamTopicPattern)

	tr := otel.Tracer("omnicam-backend")

	// Start websocket session span — carries the trace for the entire session lifetime
	wsCtx, span := tr.Start(context.Background(), "autosave.websocket.session",
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(
			attribute.String("model.id", modelId.String()),
			attribute.String("user.id", userId.String()),
		),
	)

	t.Logger.Info("autosave websocket connected",
		zap.String("modelId", modelId.String()),
		zap.String("userId", userId.String()),
	)

	go func() {
		defer conn.Close()
		defer span.End() // ← ends when goroutine exits, not when getAutosave returns
		t.Logger.Info("autosave websocket disconnected",
			zap.String("modelId", modelId.String()),
			zap.String("userId", userId.String()),
		)

		currentVersion := workspace.Version

		e := &EventContext{
			Gin:     c,
			Conn:    conn,
			ModelID: modelId,
			UserID:  userId,
			Subject: subject,
			Scale:   workspace.ScaleFactor,
			Height:  workspace.ModelHeight,
			Ctx:     wsCtx, // ← websocket session context with span
		}

		initResp := &protobufs.AutosaveEventResponse{
			LastUpdatedVersion: currentVersion,
		}
		t.sendAutosaveEventResponse(e, initResp)

		for {
			_, rawMsg, err := conn.ReadMessage()
			if err != nil {
				logger.WithTraceID(c.Request.Context(), t.Logger).Error(
					"error reading websocket message",
					zap.Error(err),
				)
				break
			}

			msg := &protobufs.WorkspaceEventRequest{}
			if err := proto.Unmarshal(rawMsg, msg); err != nil {
				logger.WithTraceID(c.Request.Context(), t.Logger).Error(
					"error unmarshalling event",
					zap.Error(err),
				)
				continue
			}

			carrier := propagation.MapCarrier{}

			if msg.Trace != nil {
				carrier["traceparent"] = msg.Trace.Traceparent
				carrier["tracestate"] = msg.Trace.Tracestate
			}

			ctx := context.Background()

			if msg.Trace != nil {
				carrier := propagation.MapCarrier{
					"traceparent": msg.Trace.Traceparent,
					"tracestate":  msg.Trace.Tracestate,
				}

				ctx = otel.GetTextMapPropagator().
					Extract(context.Background(), carrier)
			}

			switch casted := msg.Event.(type) {
			case *protobufs.WorkspaceEventRequest_Autosave:
				// child span per autosave batch
				_, autosaveSpan := tr.Start(
					ctx,
					"autosave.event.batch",
					trace.WithAttributes(
						attribute.Int("event.count", len(casted.Autosave.Events)),
					),
				)
				t.handleAutosaveEvent(e, &currentVersion, casted.Autosave)
				autosaveSpan.End()

				broadcast, _ := proto.Marshal(&protobufs.LivestreamBroadcast{
					Event: &protobufs.LivestreamBroadcast_Autosave{
						Autosave: casted.Autosave,
					},
				})
				t.Nc.Publish(e.Subject, broadcast)

			case *protobufs.WorkspaceEventRequest_Optimize:
				go func() {
					t.handleOptimizeEvent(ctx, projectId, modelId, conn, casted.Optimize)
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
