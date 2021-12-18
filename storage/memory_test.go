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

	postID, _, _ := db.Create(userID, data)
	record, _ := db.Read(userID, postID)

	if record.UserID != userID {
		t.Errorf("wrong userID. got '%s', want '%s'", record.UserID, userID)
	}

	if record.PostID != postID {
		t.Errorf("wrong postID. got '%s', want '%s'", record.PostID, postID)
	}

	if record.Data != data {
		t.Errorf("wrong data. got '%s', want '%s'", record.Data, data)
	}
}

func TestReadAll(t *testing.T) {
	db := NewMemoryDB()

	db.Create(userID, "data 1")
	db.Create(userID, "data 2")
	records, _ := db.ReadAll(userID)

	if len(records) != 2 {
		t.Errorf("wrong number of records. got '%d', want '%d'", len(records), 2)
	}
}

func TestUpdate(t *testing.T) {
	db := NewMemoryDB()

	postID, _, _ := db.Create(userID, data)
	newData := "new data"
	db.Update(userID, postID, newData)
	record, _ := db.Read(userID, postID)

	if record.Data != "new data" {
		t.Errorf("wrong data. got '%s', want '%s'", record.Data, newData)
	}
}

func TestDelete(t *testing.T) {
	db := NewMemoryDB()

	postID, _, _ := db.Create(userID, data)
	db.Delete(userID, postID)
	_, err := db.Read(userID, postID)

	if errors.Is(err, PostDoesNotExist(postID)) {
		t.Errorf("unexpected error. got '%v', want '%v'", err, PostDoesNotExist(postID))
	}
}
