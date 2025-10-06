package authentication

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"omnicam.com/backend/internal/utils"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
)

type RegisterRequest struct {
	Name     string `json:"name" binding:"required,utf8only"`
	Surname  string `json:"surname" binding:"required,utf8only"`
	Username string `json:"username" binding:"required,base64"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (t *AuthRoute) register(c *gin.Context) {
	utils.RegisterCustomValidations()
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		t.Logger.Debug("invalid form data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid form data"})
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		t.Logger.Error("failed to hash password", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	user, err := t.DB.CreateUser(c, db_sqlc_gen.CreateUserParams{
		Name:     req.Name,
		Email:    req.Email,
		Username: req.Username,
		Surname:  req.Surname,
		Password: []byte(hashedPassword),
	})
	if err != nil {
		t.Logger.Error("failed to create user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	jwtToken, err := utils.GenerateJWT(user.Name, user.Surname, user.ID.String(), user.Username, t.Env.JWTSecret, int32(t.Env.JWTExpireTime))
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
			"name":       user.Name,
			"email":      user.Email,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		},
	})
}

func (t *AuthRoute) InitRegisterRouter(router gin.IRouter) gin.IRouter {
	router.POST("/register", t.register)
	return router
}
