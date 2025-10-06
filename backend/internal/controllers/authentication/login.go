package authentication

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"
	"omnicam.com/backend/internal/utils"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
)

type AuthRoute struct {
	Logger *zap.Logger
	Env    *config_env.AppEnv
	DB     *db_sqlc_gen.Queries
}

type LoginRequest struct {
	Identifier string `json:"identifier" binding:"required"`
	Password   string `json:"password" binding:"required"`
}

func (t *AuthRoute) login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		t.Logger.Debug("invalid login data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid login data"})
		return
	}

	user, err := t.DB.GetUserByIdentifier(c, req.Identifier)
	if err != nil {
		t.Logger.Error("user not found", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	isSuccess := utils.CheckPassword(string(user.Password), req.Password)
	if !isSuccess {
		t.Logger.Error("password is incorrect", zap.Error((err)))
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	jwtToken, err := utils.GenerateJWT(user.FirstName, user.LastName, user.ID.String(), user.Username, t.Env.JWTSecret, int32(t.Env.JWTExpireTime))
	if err != nil {
		t.Logger.Error("failed to generate JWT token", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"message": "gen jwt token failed"})
		return
	}

	c.SetCookie(
		"auth_token",             // cookie name
		jwtToken,                 // value
		int(t.Env.JWTExpireTime), // max age in seconds
		"/",                      // path
		"",                       // domain (empty = current domain)
		false,                    // secure (set true if using HTTPS)
		true,                     // httpOnly
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
