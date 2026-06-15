package authentication

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"
	"omnicam.com/backend/internal/utils"
	db_client "omnicam.com/backend/pkg/db"
	"omnicam.com/backend/pkg/logger"
)

type AuthRoute struct {
	Logger *zap.Logger
	Env    *config_env.AppEnv
	DB     *db_client.DB
}

type LoginRequest struct {
	Identifier string `json:"identifier" binding:"required"`
	Password   string `json:"password" binding:"required"`
}

func (t *AuthRoute) login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).Debug("invalid login data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid login data"})
		return
	}

	user, err := t.DB.Queries.GetUserByIdentifier(c.Request.Context(), req.Identifier)
	if err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).Debug("user not found", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	isSuccess := utils.CheckPassword(string(user.Password), req.Password)
	if !isSuccess {
		logger.WithTraceID(c.Request.Context(), t.Logger).Debug("password is incorrect")
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	jwtToken, err := utils.GenerateJWT(user.FirstName, user.LastName, user.ID.String(), user.Username, t.Env.JWTSecret, t.Env.JWTExpireTime)
	if err != nil {
		logger.WithTraceID(c.Request.Context(), t.Logger).Error("failed to generate JWT token", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"message": "gen jwt token failed"})
		return
	}

	utils.SetCookie(c, jwtToken, t.Env)

	logger.WithTraceID(c.Request.Context(), t.Logger).Info("Successfully login",
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

func (t *AuthRoute) InitLoginRouter(router gin.IRouter) gin.IRouter {
	router.POST("/login", t.login)
	return router
}
