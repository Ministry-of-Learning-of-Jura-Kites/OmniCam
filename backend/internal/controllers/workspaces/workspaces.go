package controller_workspaces

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/r3labs/diff/v3"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"
	"omnicam.com/backend/internal/utils"
	db_client "omnicam.com/backend/pkg/db"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
	messages_cameras "omnicam.com/backend/pkg/messages/cameras"
	messages_model_workspace "omnicam.com/backend/pkg/messages/model_workspace"
	messages_workspace "omnicam.com/backend/pkg/messages/workspace"
)

type WorkspaceRoute struct {
	Logger *zap.Logger
	Env    *config_env.AppEnv
	DB     *db_client.DB
}

type FieldConflict struct {
	Base      any `json:"base"`
	Main      any `json:"main"`
	Workspace any `json:"workspace"`
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

func threeWayMerge(base, main, workspace messages_cameras.CameraStruct) (messages_cameras.CameraStruct, map[string]interface{}) {
	merged := base

	conflicts := make(map[string]interface{})

	changesMain, _ := diff.Diff(base, main)
	changesWork, _ := diff.Diff(base, workspace)

	changeMap := map[string][2]*diff.Change{}

	for _, c := range changesMain {
		field := strings.Join(c.Path, ".") // join full path, e.g. "Transform.AngleX"
		changeMap[field] = [2]*diff.Change{&c, nil}
	}

	for _, c := range changesWork {
		field := strings.Join(c.Path, ".")
		if prev, ok := changeMap[field]; ok {
			prev[1] = &c
			changeMap[field] = prev
		} else {
			changeMap[field] = [2]*diff.Change{nil, &c}
		}
	}

	for _, pair := range changeMap {
		mainChange, workChange := pair[0], pair[1]
		if mainChange == nil && workChange == nil {
			continue
		}
		// If to is not equal
		if mainChange == nil || workChange == nil || mainChange.To != workChange.To {
			var base, mainTo, workTo any
			var path []string

			if mainChange != nil {
				base = mainChange.From
				mainTo = mainChange.To
				// If workChange is nil, it doesn't change from base
				workTo = mainChange.From
				path = mainChange.Path
			} else {
				base = workChange.From
				mainTo = workChange.From
			}

			if workChange != nil {
				// If workChange is not nil, it changes from base
				workTo = workChange.To
				path = workChange.Path
			}

			traverse := &conflicts
			for i, pathEle := range path {
				if i == len(path)-1 {
					(*traverse)[pathEle] = FieldConflict{
						Base:      base,
						Main:      mainTo,
						Workspace: workTo,
					}
					continue
				}
				if _, ok := (*traverse)[pathEle]; !ok {
					(*traverse)[pathEle] = make(map[string]interface{})
				}
				leaf := (*traverse)[pathEle]
				casted := leaf.(map[string]interface{})
				traverse = &casted
			}
		} else {
			applyChangeReflect(&merged, mainChange) // both same
		}
	}

	return merged, conflicts
}

func mergeAllCameras(base, main, workspace messages_cameras.Cameras) (messages_cameras.Cameras, map[string]interface{}) {
	result := make(messages_cameras.Cameras)
	conflicts := make(map[string]interface{})

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

		merged, conflictOfId := threeWayMerge(b, m, w)
		if len(conflictOfId) != 0 {
			conflicts[id] = conflictOfId
		}
		result[id] = merged
	}

	return result, conflicts
}

func getAllPathsFromConflicts(conflicts map[string]interface{}) []string {
	
}

type ResolveRequest struct {
	Merged map[string]map[string]any `json:"merged"`
}

func (t *WorkspaceRoute) postResolveWorkspaceMe(c *gin.Context) {
	strModelId := c.Param("modelId")
	modelId, err := utils.ParseUuidBase64(strModelId)
	if err != nil {
		t.Logger.Error("error while converting str id to uuid", zap.Error(err), zap.String("modelId", strModelId))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model ID"})
		return
	}

	userId, err := utils.GetUuidFromCtx(c, "userId")
	if err != nil {
		t.Logger.Error("error while getting userId form", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	// Get workspace cams
	workspaceData, err := t.DB.Queries.GetWorkspaceByID(c, db_sqlc_gen.GetWorkspaceByIDParams{
		Fields:  []string{"cameras"},
		UserID:  userId,
		ModelID: modelId,
	})
	if err != nil {
		t.Logger.Error("model not found", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	if workspaceData.Version == workspaceData.BaseVersion {
		c.Status(http.StatusOK)
		return
	}

	var workspaceCameras messages_cameras.Cameras
	if err := json.Unmarshal(workspaceData.Cameras, &workspaceCameras); err != nil {
		t.Logger.Error("error while unmarshalling workspace cams", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	var baseCameras messages_cameras.Cameras
	if err := json.Unmarshal(workspaceData.BaseCameras, &workspaceCameras); err != nil {
		t.Logger.Error("error while unmarshalling workspace cams", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	// Get  model cams
	modelData, err := t.DB.Queries.GetModelByID(c, db_sqlc_gen.GetModelByIDParams{
		Fields: []string{"cameras"},
		ID:     modelId,
	})
	if err != nil {
		t.Logger.Error("model not found", zap.Error(err), zap.String("modelId", modelId.String()))
		c.Status(http.StatusNotFound)
		return
	}

	var modelCameras messages_cameras.Cameras
	if err := json.Unmarshal(modelData.Cameras, &modelCameras); err != nil {
		t.Logger.Error("error while unmarshalling workspace base cams", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	// Validate request
	var resolveRequest ResolveRequest
	if err := c.ShouldBindJSON(&resolveRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	merged,conflicts := mergeAllCameras(baseCameras,modelCameras,workspaceCameras)
	
	actualConflictsPaths := getAllPathsFromConflicts(conflicts)
	inputConflictsPaths := getAllPathsFromConflicts(resolvedRequest.Merged)

	if !slices.Equal(actualConflictsPaths,inputConflictsPaths){
		c.JSON(http.StatusBadRequest, gin.H{"error": "Not all conflicts are resolved"})
		return
	}

	// Apply merge
	for camId, mergedCam := range resolveRequest.Merged {
		cam := merged[camId]

		for key, value := range mergedCam {
			log.Println("ggg", key, value)
			// Use reflection here to set the field INSIDE the *Camera struct
			val := reflect.ValueOf(value)
			if val.Kind() == reflect.Map || val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
				// skip nested types for now
				continue
			}

			// Also apply to workspace cam
			if !utils.SetFieldByJSONTag(&cam, key, val) {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid field",
				})
				return
			}
		}
		merged[camId] = cam
	}

	log.Println("ggg", modelCameras, workspaceCameras)
	// Check if applied workspace and model are the same
	if reflect.DeepEqual(workspaceCameras, modelCameras) {
		t.DB.Queries.UpdateModelCams(c, db_sqlc_gen.UpdateModelCamsParams{
			Value:   workspaceData.Cameras,
			ModelID: modelId,
		})
		c.Status(http.StatusOK)
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{
		"error": "Not all conflicts are resolved",
	})
}

func (t *WorkspaceRoute) postMergeWorkspace(c *gin.Context) {
	userId, err := utils.GetUuidFromCtx(c, "userId")
	if err != nil {
		t.Logger.Error("error while getting userId form", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	strModelId := c.Param("modelId")
	modelId, err := utils.ParseUuidBase64(strModelId)
	if err != nil {
		t.Logger.Error("error while converting str id to uuid", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
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
		c.JSON(http.StatusOK, gin.H{"noChanges": true})
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
		c.JSON(http.StatusOK, gin.H{"noChanges": false})
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
				UserID:      userId,
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

func (t *WorkspaceRoute) deleteWorkspaceMe(c *gin.Context) {
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

	err = t.DB.Queries.DeleteWorkspace(c, db_sqlc_gen.DeleteWorkspaceParams{
		UserID:  userId,
		ModelID: modelId,
	})

	if err != nil {
		t.Logger.Error("error while deleting workspace", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.Status(http.StatusNoContent)
}

func (t *WorkspaceRoute) postWorkspaceMe(c *gin.Context) {
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

	workspace, err := t.DB.Queries.CreateWorkspace(c, db_sqlc_gen.CreateWorkspaceParams{
		UserID:  userId,
		ModelID: modelId,
	})

	if err != nil {
		t.Logger.Error("error while creating workspace", zap.Error(err))
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "workspace already exists"})
			return
		}
		var e *pgconn.PgError
		if errors.As(err, &e) && e.Code == pgerrcode.UniqueViolation {
			c.JSON(http.StatusBadRequest, gin.H{"error": "workspace already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": messages_workspace.WorkspaceRawCams{
			WorkspaceNoCams: messages_workspace.WorkspaceNoCams{
				ModelId:     workspace.ModelID,
				UserId:      workspace.UserID,
				Version:     workspace.Version,
				BaseVersion: workspace.BaseVersion,
				CreatedAt:   workspace.CreatedAt.Time.Format(time.RFC3339),
				UpdatedAt:   workspace.UpdatedAt.Time.Format(time.RFC3339),
			},
			Cameras:     json.RawMessage(workspace.Cameras),
			BaseCameras: json.RawMessage(workspace.BaseCameras),
		},
	})
}

func (t *WorkspaceRoute) getWorkspaceMe(c *gin.Context) {
	strModelId := c.Param("modelId")
	modelId, err := utils.ParseUuidBase64(strModelId)
	if err != nil {
		t.Logger.Error("error while converting str id to uuid", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model ID"})
		return
	}

	strProjectId := c.Param("projectId")
	projectId, err := utils.ParseUuidBase64(strProjectId)
	if err != nil {
		t.Logger.Error("error while converting str id to uuid", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	username := c.GetString("username")

	userInfo, err := t.DB.Queries.GetUserOfProject(c, db_sqlc_gen.GetUserOfProjectParams{
		Username: pgtype.Text{
			String: username,
			Valid:  true,
		},
		Projectid: projectId,
	})
	if err != nil {
		t.Logger.Error("user of project not found", zap.String("projectId", strProjectId), zap.String("username", username), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	includedFields := c.QueryArray("fields")

	data, err := t.DB.Queries.GetWorkspaceByID(c, db_sqlc_gen.GetWorkspaceByIDParams{
		Fields:  includedFields,
		UserID:  userInfo.ID,
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

	c.JSON(http.StatusOK, gin.H{"data": messages_model_workspace.ModelWorkspace{
		ModelId:     modelId,
		Name:        data.Model.Name,
		Description: data.Model.Description,
		ProjectId:   data.Model.ProjectID,
		FilePath:    data.Model.FilePath,
		ImagePath:   data.Model.ImagePath,
		Version:     data.Version,
		CreatedAt:   data.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:   data.UpdatedAt.Time.Format(time.RFC3339),
		Cameras:     &cameras,
	}})
}

func (t *WorkspaceRoute) InitRoute(router gin.IRouter) gin.IRouter {
	router.GET("/projects/:projectId/models/:modelId/workspaces/me", t.getWorkspaceMe)
	router.POST("/projects/:projectId/models/:modelId/workspaces/me", t.postWorkspaceMe)
	router.DELETE("/projects/:projectId/models/:modelId/workspaces/me", t.deleteWorkspaceMe)

	router.POST("/projects/:projectId/models/:modelId/workspaces/me/resolve", t.postResolveWorkspaceMe)
	router.POST("/projects/:projectId/models/:modelId/workspaces/me/merge", t.postMergeWorkspace)
	return router
}
