package redis

import (
	"context"
	"fmt"

	"github.com/cenkalti/backoff/v4"
	"github.com/craigpastro/crudapp/cache"
	"github.com/craigpastro/crudapp/storage"
	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel/trace"
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

func CreateClient(ctx context.Context, config Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
	})

	err := backoff.Retry(func() error {
		_, err := client.Ping(ctx).Result()
		return err
	}, backoff.NewExponentialBackOff())
	if err != nil {
		return nil, fmt.Errorf("error connecting to Redis: %w", err)
	}

	return client, nil
}

func (r *Redis) Add(ctx context.Context, userID, postID string, record *storage.Record) {
	_, span := r.tracer.Start(ctx, "redis.Add")
	defer span.End()

	panic("not implemented yet")
}

func (r *Redis) Get(ctx context.Context, userID, postID string) (*storage.Record, bool) {
	_, span := r.tracer.Start(ctx, "redis.Get")
	defer span.End()

	panic("not implemented yet")
}

func (r *Redis) Remove(ctx context.Context, userID, postID string) {
	_, span := r.tracer.Start(ctx, "redis.Remove")
	defer span.End()

	panic("not implemented yet")
}
