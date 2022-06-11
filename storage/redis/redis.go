package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/craigpastro/crudapp/myid"
	"github.com/craigpastro/crudapp/storage"
	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel/trace"
)

type Redis struct {
	client *redis.Client
	tracer trace.Tracer
}

func New(client *redis.Client, tracer trace.Tracer) *Redis {
	return &Redis{
		client: client,
		tracer: tracer,
	}
}

func CreateClient(ctx context.Context, addr, password string) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
	})

	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("unable to connect to Redis: %w", err)
	}

	return client, nil
}

func (r *Redis) Create(ctx context.Context, userID, data string) (*storage.Record, error) {
	ctx, span := r.tracer.Start(ctx, "redis.Create")
	defer span.End()

	postID := myid.New()
	now := time.Now()
	record := storage.NewRecord(userID, postID, data, now, now)
	marshaledRecord, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("error marhalling record: %w", err)
	}

	_, err = r.client.HSet(ctx, userID, postID, string(marshaledRecord)).Result()
	if err != nil {
		return nil, fmt.Errorf("error creating: %w", err)
	}

	return record, nil
}

func (r *Redis) Read(ctx context.Context, userID, postID string) (*storage.Record, error) {
	ctx, span := r.tracer.Start(ctx, "redis.Read")
	defer span.End()

	record, err := r.client.HGet(ctx, userID, postID).Result()
	if errors.Is(err, redis.Nil) {
		return nil, storage.ErrPostDoesNotExist
	} else if err != nil {
		return nil, fmt.Errorf("error reading: %w", err)
	}

	res := &storage.Record{}
	if err := json.Unmarshal([]byte(record), &res); err != nil {
		return nil, fmt.Errorf("error unmarhalling record: %w", err)
	}

	return res, nil
}

func (r *Redis) ReadAll(ctx context.Context, userID string) ([]*storage.Record, error) {
	ctx, span := r.tracer.Start(ctx, "redis.ReadAll")
	defer span.End()

	records, err := r.client.HGetAll(ctx, userID).Result()
	if err != nil {
		return nil, fmt.Errorf("error reading all: %w", err)
	}

	res := []*storage.Record{}
	for _, record := range records {
		r := &storage.Record{}
		if err := json.Unmarshal([]byte(record), &r); err != nil {
			return nil, fmt.Errorf("error unmarhalling record: %w", err)
		}
		res = append(res, r)
	}

	return res, nil
}

func (r *Redis) Update(ctx context.Context, userID, postID, data string) (time.Time, error) {
	ctx, span := r.tracer.Start(ctx, "redis.Update")
	defer span.End()

	record, err := r.Read(ctx, userID, postID)
	if errors.Is(err, storage.ErrPostDoesNotExist) {
		return time.Time{}, err
	} else if err != nil {
		return time.Time{}, fmt.Errorf("error updating: %w", err)
	}

	now := time.Now()
	newRecord, err := json.Marshal(storage.NewRecord(record.UserID, record.PostID, data, record.CreatedAt, now))
	if err != nil {
		return time.Time{}, fmt.Errorf("error marhalling record: %w", err)
	}

	_, err = r.client.HSet(ctx, userID, postID, string(newRecord)).Result()
	if err != nil {
		return time.Time{}, fmt.Errorf("error creating: %w", err)
	}

	return now, nil
}

func (r *Redis) Delete(ctx context.Context, userID, postID string) error {
	ctx, span := r.tracer.Start(ctx, "redis.Delete")
	defer span.End()

	if _, err := r.client.HDel(ctx, userID, postID).Result(); err != nil {
		return fmt.Errorf("error deleting: %w", err)
	}

	return nil
}
