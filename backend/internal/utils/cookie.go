package utils

import (
	"github.com/gin-gonic/gin"
	config_env "omnicam.com/backend/config"
)

func SetCookie(c *gin.Context, jwtToken string, env *config_env.AppEnv) {
	c.SetCookie(
		env.CookieName,
		jwtToken,
		int(env.JWTExpireTime.Seconds()),
		"/",
		env.CookieHost,
		env.Secure,
		true,
	)
}

func DeleteCookie(c *gin.Context, env *config_env.AppEnv) {
	c.SetCookie(
		env.CookieName,
		"",
		-1,
		"/",
		env.CookieHost,
		env.Secure,
		true,
	)
}
