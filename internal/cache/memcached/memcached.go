package memcached

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/cenkalti/backoff/v4"
	"github.com/craigpastro/crudapp/internal/cache"
	"github.com/craigpastro/crudapp/internal/storage"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var _ cache.Cache = (*Memcached)(nil)

type Memcached struct {
	client *memcache.Client
	tracer trace.Tracer
}

func New(client *memcache.Client, tracer trace.Tracer) *Memcached {
	return &Memcached{
		client: client,
		tracer: tracer,
	}
}

func CreateClient(servers string, logger *zap.Logger) (*memcache.Client, error) {
	client := memcache.New(strings.Split(servers, ",")...)

	err := backoff.Retry(func() error {
		err := client.Ping()
		if err != nil {
			logger.Info("waiting for Memcached")
			return err
		}
		return nil
	}, backoff.NewExponentialBackOff())
	if err != nil {
		return nil, fmt.Errorf("error connecting to Memcached: %w", err)
	}

	return client, nil
}

func (m *Memcached) Add(ctx context.Context, userID, postID string, record *storage.Record) error {
	_, span := m.tracer.Start(ctx, "memcached.Add")
	defer span.End()

	b, err := json.Marshal(record)
	if err != nil {
		return err
	}

	return m.client.Set(&memcache.Item{
		Key:   cache.CreateKey(userID, postID),
		Value: b,
	})
}

func (m *Memcached) Get(ctx context.Context, userID, postID string) (*storage.Record, bool) {
	_, span := m.tracer.Start(ctx, "memcached.Get")
	defer span.End()

	item, err := m.client.Get(cache.CreateKey(userID, postID))
	if err != nil {
		return nil, false
	}

	var record storage.Record
	if err := json.Unmarshal(item.Value, &record); err != nil {
		return nil, false
	}

	return &record, true
}

func (m *Memcached) Remove(ctx context.Context, userID, postID string) error {
	_, span := m.tracer.Start(ctx, "memcached.Remove")
	defer span.End()

	return m.client.Delete(cache.CreateKey(userID, postID))
}
