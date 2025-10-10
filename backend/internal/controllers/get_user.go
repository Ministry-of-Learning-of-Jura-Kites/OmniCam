package controller_test

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"
	db_client "omnicam.com/backend/pkg/db"
)

type GetUserRoute struct {
	Logger *zap.Logger
	Env    *config_env.AppEnv
	DB     *db_client.DB
}

func (t *GetUserRoute) Get(c *gin.Context) {
	email := c.Query("email")
	fmt.Print(email)
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Email is Required"})
		return
	}

	// user, err := t.DB.GetUserByEmail(c, email)
	// fmt.Println(err.Error())
	// if err != nil {
	// 	c.JSON(http.StatusNotFound, gin.H{"Error": "User not found"})
	// 	return
	// }

	c.JSON(http.StatusOK, gin.H{"Data": "test"})
}
