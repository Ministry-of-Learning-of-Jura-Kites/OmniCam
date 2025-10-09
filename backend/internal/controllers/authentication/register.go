package authentication

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"omnicam.com/backend/internal/utils"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
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
		t.Logger.Debug("invalid form data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid form data"})
		return
	}

	if !utils.IsValidUsername(req.Username) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid username"})
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		t.Logger.Error("failed to hash password", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	user, err := t.DB.CreateUser(c, db_sqlc_gen.CreateUserParams{
		FirstName: req.FirstName,
		Email:     req.Email,
		LastName:  req.LastName,
		Username:  req.Username,
		Password:  []byte(hashedPassword),
	})
	if err != nil {
		t.Logger.Error("failed to create user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	jwtToken, err := utils.GenerateJWT(user.FirstName, user.LastName, user.ID.String(), user.Username, t.Env.JWTSecret, t.Env.JWTExpireTime)
	if err != nil {
		t.Logger.Error("failed to generate JWT", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to login"})
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

func (t *AuthRoute) InitRegisterRouter(router gin.IRouter) gin.IRouter {
	router.POST("/register", t.register)
	return router
}
