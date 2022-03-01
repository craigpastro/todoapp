package storage

import (
	"context"
	"errors"
	"time"
)

var (
	ErrPostDoesNotExist     = errors.New("post does not exist")
	ErrUndefinedStorageType = errors.New("undefined storage type")
)

// Record is a storage record.
// It is tagged with json for storing in Redis and bson for storing in MongoDB.
type Record struct {
	UserID    string    `json:"userID" bson:"userID" dynamodbav:"UserID"`
	PostID    string    `json:"postID" bson:"postID" dynamodbav:"PostID"`
	Data      string    `json:"data" bson:"data" dynamodbav:"Data"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt" dynamodbav:"CreatedAt"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt" dynamodbav:"UpdatedAt"`
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

type Storage interface {
	Create(ctx context.Context, userID, data string) (*Record, error)
	Read(ctx context.Context, userID, postID string) (*Record, error)
	ReadAll(ctx context.Context, userID string) ([]*Record, error)
	Update(ctx context.Context, userID, postID, data string) (time.Time, error)
	Delete(ctx context.Context, userID, postID string) error
}
