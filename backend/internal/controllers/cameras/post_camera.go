package controller_camera

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"
	db_client "omnicam.com/backend/pkg/db"
)

type PostCameraRoute struct {
	Logger *zap.Logger
	Env    *config_env.AppEnv
	DB     *db_client.DB
}

type CreateCameraRequest struct {
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

func (t *PostCameraRoute) post(c *gin.Context) {
	// var req CreateCameraRequest
	// if err := c.ShouldBindJSON(&req); err != nil {
	// 	t.Logger.Debug("error while validating body", zap.Error(err))
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
	// 	return
	// }

	// // parse model_id
	// modelID, err := uuid.Parse(req.ModelID)
	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model_id"})
	// 	return
	// }

	// // parse user_id if provided
	// var userID *uuid.UUID
	// if req.UserID != nil {
	// 	uid, err := uuid.Parse(*req.UserID)
	// 	if err != nil {
	// 		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
	// 		return
	// 	}
	// 	userID = &uid
	// }
	// var pgUserID pgtype.UUID
	// if userID != nil {
	// 	var tmp pgtype.UUID
	// 	tmp.Scan(userID.String())
	// 	pgUserID = tmp
	// } else {
	// 	pgUserID = pgtype.UUID{Valid: false}
	// }

	// // insert into DB
	// camera, err := t.DB.CreateCamera(c, db_sqlc_gen.CreateCameraParams{
	// 	Name:    req.Name,
	// 	AngleX:  req.AngleX,
	// 	AngleY:  req.AngleY,
	// 	AngleZ:  req.AngleZ,
	// 	AngleW:  req.AngleW,
	// 	PosX:    req.PosX,
	// 	PosY:    req.PosY,
	// 	PosZ:    req.PosZ,
	// 	ModelID: modelID,
	// 	UserID:  pgUserID,
	// })
	// if err != nil {
	// 	t.Logger.Error("error while creating camera", zap.Error(err))
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create camera"})
	// 	return
	// }

	// // response
	// c.JSON(http.StatusOK, gin.H{"data": gin.H{
	// 	"id":          camera.ID,
	// 	"name":        camera.Name,
	// 	"angle_x":     camera.AngleX,
	// 	"angle_y":     camera.AngleY,
	// 	"angle_z":     camera.AngleZ,
	// 	"angle_w":     camera.AngleW,
	// 	"pos_x":       camera.PosX,
	// 	"pos_y":       camera.PosY,
	// 	"pos_z":       camera.PosZ,
	// 	"model_id":    camera.ModelID,
	// 	"user_id":     camera.UserID,
	// 	"is_snapshot": camera.IsSnapshot,
	// }})
}

func (t *PostCameraRoute) InitCreateCameraRoute(router gin.IRouter) gin.IRouter {
	router.POST("/projects/:projectId/models/:modelId/camera", t.post)
	return router
}
