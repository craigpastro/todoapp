package storage

import (
	"errors"
	"testing"
)

const (
	userID = "123"
	data   = "some data"
)

func TestRead(t *testing.T) {
	db := NewMemoryDB()

	postID, _ := db.Create(userID, data)
	record, _ := db.Read(userID, postID)

	if record.userID != userID {
		t.Errorf("wrong userID. got '%s', want '%s'", record.userID, userID)
	}

	if record.postID != postID {
		t.Errorf("wrong postID. got '%s', want '%s'", record.postID, postID)
	}

	if record.data != data {
		t.Errorf("wrong data. got '%s', want '%s'", record.data, data)
	}
}

func TestUpdate(t *testing.T) {
	db := NewMemoryDB()

	postID, _ := db.Create(userID, data)
	newData := "new data"
	db.Update(userID, postID, newData)
	record, _ := db.Read(userID, postID)

	if record.data != "new data" {
		t.Errorf("wrong data. got '%s', want '%s'", record.data, newData)
	}
}

func TestDelete(t *testing.T) {
	db := NewMemoryDB()

	postID, _ := db.Create(userID, data)
	db.Delete(userID, postID)
	_, err := db.Read(userID, postID)

	if errors.Is(err, PostDoesNotExist(postID)) {
		t.Errorf("unexpected error. got '%v', want '%v'", err, PostDoesNotExist(postID))
	}
}
