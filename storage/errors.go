package storage

import (
	"errors"
	"fmt"
)

func PostDoesNotExist(postID string) error {
	return errors.New(fmt.Sprintf("post '%s' does not exist", postID))
}

func UserDoesNotExist(userID string) error {
	return errors.New(fmt.Sprintf("user '%s' does not exist", userID))
}

func UndefinedStorageType(storage string) error {
	return errors.New(fmt.Sprintf("storage type '%s' is undefined", storage))
}
