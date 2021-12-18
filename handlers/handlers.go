package handlers

import (
	"net/http"
	"time"

	"github.com/craigpastro/crudapp/storage"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	Storage storage.Storage
}

type CreateRequest struct {
	Data string `json:"data"`
}

type CreateResponse struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
}

type ReadResponse struct {
	PostID    string    `json:"postID"`
	Data      string    `json:"data"`
	CreatedAt time.Time `json:"createdAt"`
}

func (e *Handler) CreateHandler(c *gin.Context) {
	userID := c.Param("userid")
	var req CreateRequest
	if err := c.BindJSON(&req); err != nil {
		return
	}

	postID, createdAt, err := e.Storage.Create(userID, req.Data)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusCreated, CreateResponse{
		ID:        postID,
		CreatedAt: createdAt,
	})
}

func (e *Handler) ReadHandler(c *gin.Context) {
	userID := c.Param("userid")
	postID := c.Param("postid")

	record, err := e.Storage.Read(userID, postID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, ReadResponse{
		PostID:    record.PostID,
		Data:      record.Data,
		CreatedAt: record.CreatedAt,
	})
}

func (e *Handler) UpdateHandler(c *gin.Context) {
	userID := c.Param("userid")
	postID := c.Param("postid")

	if err := e.Storage.Update(userID, postID, ""); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

func (e *Handler) DeleteHandler(c *gin.Context) {
	userID := c.Param("userid")
	postID := c.Param("postid")

	if err := e.Storage.Delete(userID, postID); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}
