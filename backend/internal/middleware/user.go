package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time" // Added for jwt.WithLeeway (best practice)

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"
	"omnicam.com/backend/internal/utils"
)

type AuthMiddleware struct {
	Env    *config_env.AppEnv
	Logger *zap.Logger
}

func (t *AuthMiddleware) CreateHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr, err := c.Cookie("auth_token")

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing auth token",
			})
			return
		}

		claims := &utils.UserClaims{}
		keyFunc := func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(t.Env.JWTSecret), nil
		}

		token, err := jwt.ParseWithClaims(tokenStr, claims, keyFunc, jwt.WithLeeway(5*time.Second))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":  "invalid token",
				"detail": err.Error(),
			})
			return
		}

		if !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token expired or invalid"})
			return
		}

		username := claims.Username
		userId, err := uuid.Parse(claims.UserID)
		if err != nil {
			t.Logger.Error("error while converting str id to uuid", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}

		if strings.TrimSpace(username) == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "username not found in token",
			})
			return
		}

		c.Set("username", username)
		c.Set("userId", userId)

		c.Next()
	}
}
