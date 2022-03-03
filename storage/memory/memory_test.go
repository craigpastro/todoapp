package memory

import (
	"context"
	"errors"
	"testing"

	"github.com/craigpastro/crudapp/instrumentation"
	"github.com/craigpastro/crudapp/myid"
	"github.com/craigpastro/crudapp/storage"
	"go.opentelemetry.io/otel/trace"
)

const (
	userID = "123"
	data   = "some data"
)

var (
	ctx    context.Context
	tracer trace.Tracer
	db     storage.Storage
)

func TestMain(m *testing.M) {
	ctx = context.Background()
	tracer, _ = instrumentation.NewTracer(ctx, instrumentation.TracerConfig{Enabled: false})
	db = New(tracer)
}

func TestRead(t *testing.T) {
	userID := myid.New()
	created, _ := db.Create(ctx, userID, data)
	record, err := db.Read(ctx, created.UserID, created.PostID)

	if err != nil {
		t.Errorf("error not nil: %v", err)
	}

	if record.UserID != created.UserID {
		t.Errorf("wrong userID: got '%s', but wanted '%s'", record.UserID, userID)
	}

	if record.PostID != created.PostID {
		t.Errorf("wrong postID: got '%s', but wanted '%s'", record.PostID, created.PostID)
	}

	if record.Data != data {
		t.Errorf("wrong data: got '%s', but wanted '%s'", record.Data, data)
	}
}

func TestReadNotExists(t *testing.T) {
	userID := myid.New()
	if _, err := db.Read(ctx, userID, "1"); err != storage.ErrPostDoesNotExist {
		t.Errorf("unexpected error: got '%v', but wanted '%v'", err, storage.ErrPostDoesNotExist)
	}
}

func TestReadAll(t *testing.T) {
	userID := myid.New()
	db.Create(ctx, userID, "data 1")
	db.Create(ctx, userID, "data 2")
	records, _ := db.ReadAll(ctx, userID)

	if len(records) != 2 {
		t.Errorf("wrong number of records: got '%d', but wanted '%d'", len(records), 2)
	}
}

func TestUpdate(t *testing.T) {
	userID := myid.New()
	created, _ := db.Create(ctx, userID, data)
	newData := "new data"
	db.Update(ctx, userID, created.PostID, newData)
	record, _ := db.Read(ctx, created.UserID, created.PostID)

	if record.Data != newData {
		t.Errorf("wrong data: got '%s', but wanted '%s'", record.Data, newData)
	}

	if record.CreatedAt.After(record.UpdatedAt) {
		t.Errorf("createdAt is after updatedAt")
	}
}

func TestUpdateNotExists(t *testing.T) {
	userID := myid.New()
	if _, err := db.Update(ctx, userID, "1", "new data"); err != storage.ErrPostDoesNotExist {
		t.Errorf("unexpected error: got '%v', but wanted '%v'", err, storage.ErrPostDoesNotExist)
	}
}

func TestDelete(t *testing.T) {
	userID := myid.New()
	created, _ := db.Create(ctx, userID, data)

	if err := db.Delete(ctx, userID, created.PostID); err != nil {
		t.Errorf("error not nil: %v", err)
	}

	// Now try to read the deleted record; it should not exist.
	if _, err := db.Read(ctx, userID, created.PostID); !errors.Is(err, storage.ErrPostDoesNotExist) {
		t.Errorf("unexpected error: got '%v', but wanted '%v'", err, storage.ErrPostDoesNotExist)
	}
}

func TestDeleteNotExists(t *testing.T) {
	userID := myid.New()
	postID := myid.New()

	if err := db.Delete(ctx, userID, postID); err != nil {
		t.Errorf("error not nil: %v", err)
	}
}
