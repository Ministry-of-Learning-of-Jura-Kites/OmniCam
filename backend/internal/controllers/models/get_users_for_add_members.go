package controller_model // สร้าง package ใหม่สำหรับ user controller

import (
	"encoding/base64"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
)

type UserResponse struct {
	Id        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Email     *string   `json:"email"`
	FirstName *string   `json:"first_name"`
	LastName  *string   `json:"last_name"`
	CreatedAt string    `json:"createdAt"`
	UpdatedAt string    `json:"updatedAt"`
}

type UsersForAddMembersRoute struct {
	Logger *zap.Logger
	Env    *config_env.AppEnv
	DB     *db_sqlc_gen.Queries
}

func (t *UsersForAddMembersRoute) getAllUsersForAddMembers(c *gin.Context) {

	strId := c.Param("projectId")
	decodedBytes, err := base64.RawURLEncoding.DecodeString(strId)
	if err != nil {
		t.Logger.Error("error decoding Base64", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}
	projectID, err := uuid.FromBytes(decodedBytes)
	if err != nil {
		t.Logger.Error("error converting to uuid", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page number"})
		return
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if err != nil || pageSize < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page size"})
		return
	}

	search := c.DefaultQuery("search", "")
	pageOffset := (page - 1) * pageSize

	users, err := t.DB.GetUsersForAddMembers(c, db_sqlc_gen.GetUsersForAddMembersParams{
		PageSize:   int32(pageSize),
		PageOffset: int32(pageOffset),
		Search:     search,
		ProjectID:  projectID,
	})
	if err != nil {
		t.Logger.Error("failed to get all users", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve users"})
		return
	}

	total, err := t.DB.CountUsersForAddMembers(c, db_sqlc_gen.CountUsersForAddMembersParams{
		Search:    search,
		ProjectID: projectID,
	})
	if err != nil {
		t.Logger.Error("failed to count all users", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count users"})
		return
	}

	var userList []UserResponse
	for _, u := range users {
		var email, firstName, lastName *string
		if u.Email != "" {
			email = &u.Email
		}
		if u.FirstName != "" {
			firstName = &u.FirstName
		}
		if u.LastName != "" {
			lastName = &u.LastName
		}

		userList = append(userList, UserResponse{
			Id:        u.ID,
			Username:  u.Username,
			Email:     email,
			FirstName: firstName,
			LastName:  lastName,
			CreatedAt: u.CreatedAt.Time.Format(time.RFC3339),
			UpdatedAt: u.UpdatedAt.Time.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  userList,
		"count": total,
	})
}

// InitUserRouter sets up the routes for users
func (t *UsersForAddMembersRoute) InitUserRouter(router gin.IRouter) gin.IRouter {
	router.GET("/projects/:projectId/userForAddMembers", t.getAllUsersForAddMembers)
	return router
}
