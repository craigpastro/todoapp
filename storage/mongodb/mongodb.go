package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/craigpastro/crudapp/storage"
	"github.com/oklog/ulid/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

const (
	userIDField    = "userID"
	postIDField    = "postID"
	dataField      = "data"
	updatedAtField = "updatedAt"
)

var _ storage.Storage = (*MongoDB)(nil)

type MongoDB struct {
	coll   *mongo.Collection
	tracer trace.Tracer
}

func New(coll *mongo.Collection, tracer trace.Tracer) *MongoDB {
	return &MongoDB{
		coll:   coll,
		tracer: tracer,
	}
}

func CreateCollection(ctx context.Context, connString string, logger *zap.Logger) (*mongo.Collection, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connString))
	if err != nil {
		return nil, fmt.Errorf("error connecting to MongoDB: %w", err)
	}

	err = backoff.Retry(func() error {
		err := client.Ping(ctx, readpref.Primary())
		if err != nil {
			logger.Info("waiting for MongoDB")
			return err
		}
		return nil
	}, backoff.NewExponentialBackOff())
	if err != nil {
		return nil, fmt.Errorf("error connecting to MongoDB: %w", err)
	}

	return client.Database("db").Collection("posts"), nil
}

func (m *MongoDB) Create(ctx context.Context, userID, data string) (*storage.Record, error) {
	ctx, span := m.tracer.Start(ctx, "mongodb.Create")
	defer span.End()

	postID := ulid.Make().String()
	now := time.Now()
	record := storage.NewRecord(userID, postID, data, now, now)
	_, err := m.coll.InsertOne(ctx, record)
	if err != nil {
		return nil, fmt.Errorf("error creating: %w", err)
	}

	return record, nil
}

func (m *MongoDB) Read(ctx context.Context, userID, postID string) (*storage.Record, error) {
	ctx, span := m.tracer.Start(ctx, "mongodb.Read")
	defer span.End()

	var record storage.Record
	err := m.coll.FindOne(ctx, bson.M{userIDField: userID, postIDField: postID}).Decode(&record)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, storage.ErrPostDoesNotExist
		}
		return nil, fmt.Errorf("error reading: %w", err)
	}

	return &record, nil
}

func (m *MongoDB) ReadAll(ctx context.Context, userID string) (storage.RecordIterator, error) {
	ctx, span := m.tracer.Start(ctx, "mongodb.ReadAll")
	defer span.End()

	cur, err := m.coll.Find(ctx, bson.M{userIDField: userID})
	if err != nil {
		return nil, fmt.Errorf("error reading all: %w", err)
	}

	return &recordInterator{cur: cur}, nil

}

type recordInterator struct {
	cur *mongo.Cursor
}

func (i *recordInterator) Next(ctx context.Context) bool {
	return i.cur.Next(ctx)
}

func (i *recordInterator) Get(dest *storage.Record) error {
	return i.cur.Decode(dest)
}

func (i *recordInterator) Close(ctx context.Context) {
	i.cur.Close(ctx)
}

func (m *MongoDB) Update(ctx context.Context, userID, postID, data string) (time.Time, error) {
	ctx, span := m.tracer.Start(ctx, "mongodb.Update")
	defer span.End()

	now := time.Now()
	query := bson.M{userIDField: userID, postIDField: postID}
	update := bson.M{"$set": bson.M{dataField: data, updatedAtField: now}}
	res, err := m.coll.UpdateOne(ctx, query, update, options.Update().SetUpsert(false))
	if res.MatchedCount == 0 {
		return time.Time{}, storage.ErrPostDoesNotExist
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("error updating: %w", err)
	}

	return now, nil
}

func (m *MongoDB) Delete(ctx context.Context, userID, postID string) error {
	ctx, span := m.tracer.Start(ctx, "mongodb.Delete")
	defer span.End()

	query := bson.M{userIDField: userID, postIDField: postID}
	_, err := m.coll.DeleteOne(ctx, query)
	if err != nil {
		return fmt.Errorf("error deleting: %w", err)
	}

	return nil
}
