package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/craigpastro/crudapp/storage"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	Storage storage.Storage
}

func New(storage storage.Storage) Handler {
	return Handler{
		Storage: storage,
	}
}

type Request struct {
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

type ReadAllResponse struct {
	Posts []ReadResponse `json:"posts"`
}

func (h *Handler) CreateHandler(c *gin.Context) {
	userID := c.Param("userid")
	var req Request
	if err := c.BindJSON(&req); err != nil {
		return
	}

	postID, createdAt, err := h.Storage.Create(userID, req.Data)
	if err != nil {
		handleStorageError(c, err)
		return
	}

	c.JSON(http.StatusCreated, CreateResponse{
		ID:        postID,
		CreatedAt: createdAt,
	})
}

func (h *Handler) ReadHandler(c *gin.Context) {
	userID := c.Param("userid")
	postID := c.Param("postid")

	record, err := h.Storage.Read(userID, postID)
	if err != nil {
		handleStorageError(c, err)
		return
	}

	c.JSON(http.StatusOK, ReadResponse{
		PostID:    record.PostID,
		Data:      record.Data,
		CreatedAt: record.CreatedAt,
	})
}

func (h *Handler) ReadAllHandler(c *gin.Context) {
	userID := c.Param("userid")

	records, err := h.Storage.ReadAll(userID)
	if err != nil {
		handleStorageError(c, err)
		return
	}

	res := []ReadResponse{}
	for _, record := range records {
		res = append(res, ReadResponse{
			PostID:    record.PostID,
			Data:      record.Data,
			CreatedAt: record.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, ReadAllResponse{Posts: res})
}

func (h *Handler) UpdateHandler(c *gin.Context) {
	userID := c.Param("userid")
	postID := c.Param("postid")
	var req Request
	if err := c.BindJSON(&req); err != nil {
		return
	}

	if err := h.Storage.Update(userID, postID, req.Data); err != nil {
		handleStorageError(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (h *Handler) DeleteHandler(c *gin.Context) {
	userID := c.Param("userid")
	postID := c.Param("postid")

	if err := h.Storage.Delete(userID, postID); err != nil {
		handleStorageError(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func handleStorageError(c *gin.Context, err error) {
	if errors.Is(err, storage.ErrPostDoesNotExist) {
		c.Status(http.StatusBadRequest)
	} else if errors.Is(err, storage.ErrUserDoesNotExist) {
		c.Status(http.StatusBadRequest)
	} else {
		c.Status(http.StatusInternalServerError)
	}
}
