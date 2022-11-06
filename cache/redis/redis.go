package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cenkalti/backoff/v4"
	"github.com/craigpastro/crudapp/cache"
	"github.com/craigpastro/crudapp/storage"
	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var _ cache.Cache = (*Redis)(nil)

type Redis struct {
	client *redis.Client
	tracer trace.Tracer
}

type Config struct {
	Addr     string
	Password string
}

func New(client *redis.Client, tracer trace.Tracer) *Redis {
	return &Redis{
		client: client,
		tracer: tracer,
	}
}

func CreateClient(ctx context.Context, config Config, logger *zap.Logger) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
	})

	err := backoff.Retry(func() error {
		_, err := client.Ping(ctx).Result()
		if err != nil {
			logger.Info("waiting for Redis")
			return err
		}
		return nil
	}, backoff.NewExponentialBackOff())
	if err != nil {
		return nil, fmt.Errorf("error connecting to Redis: %w", err)
	}

	return client, nil
}

func (r *Redis) Add(ctx context.Context, userID, postID string, record *storage.Record) error {
	_, span := r.tracer.Start(ctx, "redis.Add")
	defer span.End()

	b, err := json.Marshal(record)
	if err != nil {
		return err
	}

	_, err = r.client.Set(ctx, cache.CreateKey(userID, postID), b, 0).Result()
	if err != nil {
		return err
	}

	return nil
}

func (r *Redis) Get(ctx context.Context, userID, postID string) (*storage.Record, bool) {
	_, span := r.tracer.Start(ctx, "redis.Get")
	defer span.End()

	s, err := r.client.Get(ctx, cache.CreateKey(userID, postID)).Result()
	if err != nil {
		return nil, false
	}

	var record storage.Record
	if err := json.Unmarshal([]byte(s), &record); err != nil {
		return nil, false
	}

	return &record, true
}

func (r *Redis) Remove(ctx context.Context, userID, postID string) error {
	_, span := r.tracer.Start(ctx, "redis.Remove")
	defer span.End()

	_, err := r.client.Del(ctx, cache.CreateKey(userID, postID)).Result()
	return err
}
