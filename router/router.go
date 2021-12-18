package router

import (
	"github.com/craigpastro/crudapp/handlers"
	"github.com/craigpastro/crudapp/storage"
	"github.com/gin-gonic/gin"
)

func Run(serverAddr string, storage storage.Storage) error {
	r := gin.Default()
	r.SetTrustedProxies(nil)

	h := handlers.Handler{Storage: storage}

	r.POST("/v1/users/:userid/posts", h.CreateHandler)
	r.GET("/v1/users/:userid/posts/:postid", h.ReadHandler)
	r.PATCH("/v1/users/:userid/posts/:postid", h.UpdateHandler)
	r.DELETE("/v1/users/:userid/posts/:postid", h.DeleteHandler)

	return r.Run(serverAddr)
}
