package authentication

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"omnicam.com/backend/internal/utils"
)

func (t *AuthRoute) logout(c *gin.Context) {
	utils.DeleteCookie(c, t.Env)

	c.JSON(http.StatusOK, gin.H{
		"message": "logged out successfully",
	})
}

func (t *AuthRoute) InitLogoutRouter(router gin.IRouter) gin.IRouter {
	router.POST("/logout", t.logout)
	return router
}
