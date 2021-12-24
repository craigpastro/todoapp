package memory

import (
	"context"
	"errors"
	"testing"

	"github.com/craigpastro/crudapp/storage"
)

const (
	userID = "123"
	data   = "some data"
)

func TestRead(t *testing.T) {
	db := New()
	ctx := context.Background()

	postID, _, _ := db.Create(ctx, userID, data)
	record, _ := db.Read(ctx, userID, postID)

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
	db := New()
	ctx := context.Background()

	db.Create(ctx, userID, "data 1")
	db.Create(ctx, userID, "data 2")
	records, _ := db.ReadAll(ctx, userID)

	if len(records) != 2 {
		t.Errorf("wrong number of records. got '%d', want '%d'", len(records), 2)
	}
}

func TestUpdate(t *testing.T) {
	db := New()
	ctx := context.Background()

	postID, _, _ := db.Create(ctx, userID, data)
	newData := "new data"
	db.Update(ctx, userID, postID, newData)
	record, _ := db.Read(ctx, userID, postID)

	if record.Data != "new data" {
		t.Errorf("wrong data. got '%s', want '%s'", record.Data, newData)
	}
}

func TestDelete(t *testing.T) {
	db := New()
	ctx := context.Background()

	postID, _, _ := db.Create(ctx, userID, data)
	db.Delete(ctx, userID, postID)
	_, err := db.Read(ctx, userID, postID)

	if !errors.Is(err, storage.ErrPostDoesNotExist) {
		t.Errorf("unexpected error. got '%v', want '%v'", err, storage.ErrPostDoesNotExist)
	}
}
