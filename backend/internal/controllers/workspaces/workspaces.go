package controller_workspaces

import (
	"cmp"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"slices"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/r3labs/diff/v3"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"
	db_client "omnicam.com/backend/pkg/db"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
	messages_cameras "omnicam.com/backend/pkg/messages/cameras"
)

type Workspace struct {
	Id          uuid.UUID                 `json:"id"`
	ModelId     uuid.UUID                 `json:"modelId"`
	Name        string                    `json:"name"`
	Description string                    `json:"description"`
	Version     int32                     `json:"version"`
	CreatedAt   string                    `json:"createdAt"`
	UpdatedAt   string                    `json:"updatedAt"`
	Cameras     *messages_cameras.Cameras `json:"cameras"`
}

type WorkspaceRoute struct {
	Logger *zap.Logger
	Env    *config_env.AppEnv
	DB     *db_client.DB
}

type FieldConflict struct {
	Base      messages_cameras.Cameras `json:"base"`
	Main      messages_cameras.Cameras `json:"main"`
	Workspace messages_cameras.Cameras `json:"workspace"`
}

func applyChangeReflect(cam *messages_cameras.CameraStruct, c *diff.Change) error {
	if len(c.Path) == 0 {
		return nil
	}
	field := c.Path[0]

	v := reflect.ValueOf(cam).Elem()
	f := v.FieldByNameFunc(func(name string) bool {
		// match against json or diff tag
		t, _ := v.Type().FieldByName(name)
		tag := t.Tag.Get("diff")
		return tag == field || t.Tag.Get("json") == field
	})

	if !f.IsValid() {
		return fmt.Errorf("unknown field: %s", field)
	}
	if !f.CanSet() {
		return fmt.Errorf("cannot set field: %s", field)
	}

	val := reflect.ValueOf(c.To)
	if val.Type().ConvertibleTo(f.Type()) {
		f.Set(val.Convert(f.Type()))
	} else {
		return fmt.Errorf("type mismatch for field %s", field)
	}
	return nil
}

func threeWayMerge(id string, base, main, workspace messages_cameras.CameraStruct) (messages_cameras.CameraStruct, map[string]FieldConflict) {
	merged := base
	conflicts := map[string]FieldConflict{}

	changesMain, _ := diff.Diff(base, main)
	changesWork, _ := diff.Diff(base, workspace)

	changeMap := map[string][2]*diff.Change{}

	for _, c := range changesMain {
		field := c.Path[0]
		changeMap[field] = [2]*diff.Change{&c, nil}
	}
	for _, c := range changesWork {
		field := c.Path[0]
		if prev, ok := changeMap[field]; ok {
			prev[1] = &c
			changeMap[field] = prev
		} else {
			changeMap[field] = [2]*diff.Change{nil, &c}
		}
	}

	for field, pair := range changeMap {
		mainChange, workChange := pair[0], pair[1]

		switch {
		case mainChange != nil && workChange == nil:
			applyChangeReflect(&merged, mainChange)
		case workChange != nil && mainChange == nil:
			applyChangeReflect(&merged, workChange)
		case mainChange != nil && workChange != nil:
			if mainChange.To != workChange.To {
				conflicts[field] = FieldConflict{
					Base:      mainChange.From.(messages_cameras.Cameras),
					Main:      mainChange.To.(messages_cameras.Cameras),
					Workspace: workChange.To.(messages_cameras.Cameras),
				}
			} else {
				applyChangeReflect(&merged, mainChange) // both same
			}
		}
	}

	return merged, conflicts
}

func mergeAllCameras(base, main, workspace messages_cameras.Cameras) (messages_cameras.Cameras, map[string]map[string]FieldConflict) {
	result := make(messages_cameras.Cameras)
	conflicts := map[string]map[string]FieldConflict{}

	keys := map[string]struct{}{}
	for k := range base {
		keys[k] = struct{}{}
	}
	for k := range main {
		keys[k] = struct{}{}
	}
	for k := range workspace {
		keys[k] = struct{}{}
	}

	for id := range keys {
		b, bok := base[id]
		m, mok := main[id]
		w, wok := workspace[id]

		// new camera in workspace
		if !bok && !mok && wok {
			result[id] = w
			continue
		}
		// new camera in main
		if !bok && mok && !wok {
			result[id] = m
			continue
		}
		// deleted in both
		if !mok && !wok {
			continue
		}

		// zero values if missing
		if !bok {
			b = messages_cameras.CameraStruct{}
		}
		if !mok {
			m = messages_cameras.CameraStruct{}
		}
		if !wok {
			w = messages_cameras.CameraStruct{}
		}

		merged, camsConflicts := threeWayMerge(id, b, m, w)
		result[id] = merged
		conflicts[id] = camsConflicts
	}

	return result, conflicts
}

type ResolveRequest struct {
}

// func (t *WorkspaceRoute) postResolveWorkspaceMe(c *gin.Context) {
// 	strProjectId := c.Param("modelId")
// 	userId := uuid.Nil

// 	decodedBytes, err := base64.RawURLEncoding.DecodeString(strProjectId)
// 	if err != nil {
// 		t.Logger.Error("error decoding Base64", zap.Error(err))
// 		return
// 	}
// 	modelId, err := uuid.FromBytes(decodedBytes)
// 	if err != nil {
// 		t.Logger.Error("error while converting str id to uuid", zap.Error(err))
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model ID"})
// 		return
// 	}

// 	workspaceData, err := t.DB.GetWorkspaceByID(c, db_sqlc_gen.GetWorkspaceByIDParams{
// 		Fields:  []string{"cameras"},
// 		UserID:  userId,
// 		ModelID: modelId,
// 	})
// 	if err != nil {
// 		t.Logger.Error("model not found", zap.Error(err))
// 		c.JSON(http.StatusNotFound, gin.H{})
// 		return
// 	}

// 	if workspaceData.Version == workspaceData.BaseVersion {
// 		t.Logger.Error("model not found", zap.Error(err))
// 		c.JSON(http.StatusNotFound, gin.H{"noChanges": true})
// 		return
// 	}

// 	modelData, err := t.DB.GetModelByID(c, db_sqlc_gen.GetModelByIDParams{
// 		Fields: []string{"cameras"},
// 		ID:     uuid.Nil,
// 	})
// 	if err != nil {
// 		t.Logger.Error("model not found", zap.Error(err))
// 		c.JSON(http.StatusNotFound, gin.H{})
// 		return
// 	}

// 	var resolveRequest ResolveRequest
// 	if err := c.ShouldBindJSON(&resolveRequest); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// }

func (t *WorkspaceRoute) postMergeWorkspaceMe(c *gin.Context) {
	strProjectId := c.Param("modelId")
	userId := uuid.Nil

	decodedBytes, err := base64.RawURLEncoding.DecodeString(strProjectId)
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

	workspaceData, err := t.DB.Queries.GetWorkspaceByID(c, db_sqlc_gen.GetWorkspaceByIDParams{
		Fields:  []string{"cameras", "base_cameras"},
		UserID:  userId,
		ModelID: modelId,
	})
	if err != nil {
		t.Logger.Error("model not found", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	if workspaceData.Version == workspaceData.BaseVersion {
		c.JSON(http.StatusNotFound, gin.H{"noChanges": true})
		return
	}

	modelData, err := t.DB.Queries.GetModelByID(c, db_sqlc_gen.GetModelByIDParams{
		Fields: []string{"cameras"},
		ID:     modelId,
	})
	if err != nil {
		t.Logger.Error("model not found", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	switch cmp.Compare(modelData.Version, workspaceData.BaseVersion) {
	// equal
	case 0:
		t.DB.Queries.UpdateModelCams(c, db_sqlc_gen.UpdateModelCamsParams{
			Value:   workspaceData.Cameras,
			ModelID: modelId,
		})
		c.JSON(http.StatusOK, gin.H{})
		return
	// model is grater than workspace
	case 1:
		var workspaceCameras messages_cameras.Cameras
		var baseCameras messages_cameras.Cameras
		var mainCameras messages_cameras.Cameras

		if err := json.Unmarshal(workspaceData.Cameras, &workspaceCameras); err != nil {
			t.Logger.Error("error while unmarshalling workspace cams", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}

		if err := json.Unmarshal(workspaceData.BaseCameras, &baseCameras); err != nil {
			t.Logger.Error("error while unmarshalling workspace base cams", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}

		if err := json.Unmarshal(modelData.Cameras, &mainCameras); err != nil {
			t.Logger.Error("error while unmarshalling model cams", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}
		merged, conflicts := mergeAllCameras(baseCameras, mainCameras, workspaceCameras)

		mergedEncoded, err := json.Marshal(merged)
		if err != nil {
			t.Logger.Error("error while marshalling merged cameras", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}

		if len(conflicts) == 0 {
			base_version, err := t.DB.Queries.UpdateModelCams(c, db_sqlc_gen.UpdateModelCamsParams{
				Value:   mergedEncoded,
				ModelID: modelId,
			})

			if err != nil {
				t.Logger.Error("error while saving merged workspace into model", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{})
				return
			}

			err = t.DB.Queries.UpdateSetWorkspaceCams(c, db_sqlc_gen.UpdateSetWorkspaceCamsParams{
				Cameras:     mergedEncoded,
				BaseCameras: mergedEncoded,
				BaseVersion: base_version,
				UserID:      uuid.Nil,
				ModelID:     modelId,
			})

			if err != nil {
				t.Logger.Error("error while saving merged workspace into model", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"noChanges": false,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"merged":    merged,
			"conflicts": conflicts,
		})
		return
	// workspace is grater than model
	case -1:
		t.Logger.Error("workspace version is ahead of main", zap.String("modelId", modelId.String()), zap.String("userId", userId.String()))
		c.JSON(http.StatusInternalServerError, gin.H{})
	}
}

func (t *WorkspaceRoute) getWorkspaceMe(c *gin.Context) {
	strProjectId := c.Param("modelId")

	decodedBytes, err := base64.RawURLEncoding.DecodeString(strProjectId)
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

	includedFields := c.QueryArray("fields")

	data, err := t.DB.Queries.GetWorkspaceByID(c, db_sqlc_gen.GetWorkspaceByIDParams{
		Fields:  includedFields,
		UserID:  uuid.Nil,
		ModelID: modelId,
	})
	if err != nil {
		t.Logger.Error("model not found", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	var cameras messages_cameras.Cameras
	if slices.Contains(includedFields, "cameras") {
		err = json.Unmarshal(data.Cameras, &cameras)
		if err != nil {
			t.Logger.Error("cameras jsonb are invalid", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": Workspace{
		ModelId:     modelId,
		Name:        data.Name,
		Description: data.Description,
		Version:     data.Version,
		CreatedAt:   data.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:   data.UpdatedAt.Time.Format(time.RFC3339),
		Cameras:     &cameras,
	}})
}

func (t *WorkspaceRoute) InitRoute(router gin.IRouter) gin.IRouter {
	router.POST("/projects/:projectId/models/:modelId/workspaces/merge", t.postMergeWorkspaceMe)
	router.GET("/projects/:projectId/models/:modelId/workspaces/me", t.getWorkspaceMe)
	return router
}
