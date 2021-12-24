package storage

import (
	"errors"
	"time"
)

var (
	ErrPostDoesNotExist     = errors.New("post does not exist")
	ErrUndefinedStorageType = errors.New("undefined storage type")
	ErrUserDoesNotExist     = errors.New("user does not exist")
)

type record struct {
	UserID    string
	PostID    string
	Data      string
	CreatedAt time.Time
}

func NewRecord(userID, postID, data string, createdAt time.Time) *record {
	return &record{
		UserID:    userID,
		PostID:    postID,
		Data:      data,
		CreatedAt: createdAt,
	}
}

type Storage interface {
	Create(userID, data string) (string, time.Time, error)
	Read(userID, postID string) (*record, error)
	ReadAll(userID string) ([]*record, error)
	Update(userID, postID, data string) error
	Delete(userID, postID string) error
}

func New(storageType string) (Storage, error) {
	if storageType == "memory" {
		return NewMemoryDB(), nil
	}

	return nil, ErrUndefinedStorageType
}
