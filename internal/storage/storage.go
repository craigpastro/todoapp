package storage

import (
	"context"
	"errors"
	"time"
)

var ErrPostDoesNotExist = errors.New("post does not exist")

// Record is a storage record.
// It is tagged with json for storing in Redis
type Record struct {
	UserID    string    `json:"userID"`
	PostID    string    `json:"postID"`
	Data      string    `json:"data"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func NewRecord(userID, postID, data string, createdAt time.Time, updatedAt time.Time) *Record {
	return &Record{
		UserID:    userID,
		PostID:    postID,
		Data:      data,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}

type RecordIterator interface {
	Next(context.Context) bool
	Get(dest *Record) error
	Close(context.Context)
}

type Storage interface {
	Create(ctx context.Context, userID, data string) (*Record, error)
	Read(ctx context.Context, userID, postID string) (*Record, error)
	ReadAll(ctx context.Context, userID string) (RecordIterator, error)
	Update(ctx context.Context, userID, postID, data string) (*Record, error)
	Delete(ctx context.Context, userID, postID string) error
}
