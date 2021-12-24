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

type Record struct {
	UserID    string
	PostID    string
	Data      string
	CreatedAt time.Time
}

func NewRecord(userID, postID, data string, createdAt time.Time) *Record {
	return &Record{
		UserID:    userID,
		PostID:    postID,
		Data:      data,
		CreatedAt: createdAt,
	}
}

type Storage interface {
	Create(userID, data string) (string, time.Time, error)
	Read(userID, postID string) (*Record, error)
	ReadAll(userID string) ([]*Record, error)
	Update(userID, postID, data string) error
	Delete(userID, postID string) error
}
