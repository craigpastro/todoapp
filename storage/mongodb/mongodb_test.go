package mongodb

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/craigpastro/crudapp/instrumentation"
	"github.com/craigpastro/crudapp/myid"
	"github.com/craigpastro/crudapp/storage"
	"github.com/kelseyhightower/envconfig"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const data = "some data"

var (
	ctx context.Context
	db  storage.Storage
)

type Config struct {
	MongoDBURI string `split_words:"true" default:"mongodb://mongodb:password@127.0.0.1:27017"`
}

func TestMain(m *testing.M) {
	var config Config
	if err := envconfig.Process("", &config); err != nil {
		fmt.Printf("error reading config: %v\n", err)
		os.Exit(1)
	}

	ctx = context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(config.MongoDBURI))
	if err != nil {
		fmt.Printf("error initializing MongoDB: %v\n", err)
		os.Exit(1)
	}
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		fmt.Printf("error connecting to MongoDB: %v\n", err)
		os.Exit(1)
	}

	tracer, _ := instrumentation.NewTracer(ctx, instrumentation.TracerConfig{Enabled: false})
	db = &MongoDB{
		coll:   client.Database("db").Collection("posts"),
		tracer: tracer,
	}

	os.Exit(m.Run())
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
	records, err := db.ReadAll(ctx, userID)

	if err != nil {
		t.Errorf("error not nil: %v", err)
	}

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
