package router

import (
	"fmt"
	"time"

	"github.com/craigpastro/crudapp/handlers"
	"github.com/craigpastro/crudapp/storage"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Run(serverAddr string, storage storage.Storage) error {
	r := gin.New()
	r.SetTrustedProxies(nil)

	logger, err := zap.NewProduction()
	if err != nil {
		return fmt.Errorf("error initializing logger: %w", err)
	}
	r.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	r.Use(ginzap.RecoveryWithZap(logger, true))

	h := handlers.New(storage)

	r.POST("/v1/users/:userid/posts", h.CreateHandler)
	r.GET("/v1/users/:userid/posts/:postid", h.ReadHandler)
	r.GET("/v1/users/:userid/posts", h.ReadAllHandler)
	r.PATCH("/v1/users/:userid/posts/:postid", h.UpdateHandler)
	r.DELETE("/v1/users/:userid/posts/:postid", h.DeleteHandler)

	return r.Run(serverAddr)
}
