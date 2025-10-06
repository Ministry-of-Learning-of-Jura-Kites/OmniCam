package authentication

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (t *AuthRoute) logout(c *gin.Context) {
	c.SetCookie(
		"auth_token", // cookie name
		"",           // empty value
		-1,           // max age negative = delete
		"/",          // path
		"",           // domain (empty = current domain)
		false,        // secure (set true if using HTTPS)
		true,         // httpOnly
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "logged out successfully",
	})
}

func (t *AuthRoute) InitLogoutRouter(router gin.IRouter) gin.IRouter {
	router.POST("/logout", t.logout)
	return router
}
