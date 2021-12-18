package storage

import (
	"fmt"
	"time"
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

	return nil, UndefinedStorageType(storageType)
}

func PostDoesNotExist(postID string) error {
	return fmt.Errorf("post '%s' does not exist", postID)
}

func UserDoesNotExist(userID string) error {
	return fmt.Errorf(fmt.Sprintf("user '%s' does not exist", userID))
}

func UndefinedStorageType(storage string) error {
	return fmt.Errorf(fmt.Sprintf("storage type '%s' is undefined", storage))
}
