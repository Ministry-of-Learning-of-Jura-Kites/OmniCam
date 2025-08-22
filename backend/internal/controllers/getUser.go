package controller_test

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	config_env "omnicam.com/backend/config"
	db_sqlc_gen "omnicam.com/backend/pkg/db/sqlc-gen"
)

type GetUserRoute struct {
	Logger *zap.Logger
	Env    *config_env.AppEnv
	DB     *db_sqlc_gen.Queries
}

func (t *GetUserRoute) Get(c *gin.Context) {
	email := c.Query("email")
	fmt.Print(email)
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Email is Required"})
		return
	}

	user, err := t.DB.GetUserByEmail(context.Background(), email)
	fmt.Println(err.Error())
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"Error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"Data": user})
}
