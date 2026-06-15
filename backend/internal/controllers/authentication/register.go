package authentication

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"omnicam.com/backend/internal/utils"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
	"omnicam.com/backend/pkg/logger"
	messages_errors "omnicam.com/backend/pkg/messages/errors"
)

type RegisterRequest struct {
	FirstName string `json:"firstName" binding:"required,utf8only"`
	LastName  string `json:"lastName" binding:"required,utf8only"`
	Username  string `json:"username" binding:"required,max=255"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required"`
}

func (t *AuthRoute) register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).Debug("invalid form data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid form data"})
		return
	}

	if !utils.IsValidUsername(req.Username) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid username"})
		return
	}

	if !utils.CheckPasswordFormat(req.Password) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid password"})
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).Error("failed to hash password", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages_errors.FailedToCreateUser})
		return
	}

	user, err := t.DB.Queries.CreateUser(c.Request.Context(), db_sqlc_gen.CreateUserParams{
		FirstName: req.FirstName,
		Email:     req.Email,
		LastName:  req.LastName,
		Username:  req.Username,
		Password:  []byte(hashedPassword),
	})
	if err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).Error(messages_errors.FailedToCreateUser, zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages_errors.FailedToCreateUser})
		return
	}

	jwtToken, err := utils.GenerateJWT(user.FirstName, user.LastName, user.ID.String(), user.Username, t.Env.JWTSecret, t.Env.JWTExpireTime)
	if err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).Error("failed to generate JWT", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to login"})
		return
	}

	utils.SetCookie(c, jwtToken, t.Env)

	logger.WithTraceID(c.Request.Context(), t.Logger).Info("Successfully register",
		zap.String("userId", user.ID.String()),
	)
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"id":         user.ID,
			"firstName":  user.FirstName,
			"lastName":   user.LastName,
			"username":   user.Username,
			"email":      user.Email,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		},
		"token": jwtToken,
	})
}

func (t *AuthRoute) InitRegisterRouter(router gin.IRouter) gin.IRouter {
	router.POST("/register", t.register)
	return router
}
