package memcached

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/craigpastro/crudapp/cache"
	"github.com/craigpastro/crudapp/storage"
	"go.opentelemetry.io/otel/trace"
)

type Memcached struct {
	client *memcache.Client
	tracer trace.Tracer
}

type Config struct {
	Servers string
}

func New(client *memcache.Client, tracer trace.Tracer) *Memcached {
	return &Memcached{
		client: client,
		tracer: tracer,
	}
}

func CreateClient(config Config) (*memcache.Client, error) {
	client := memcache.New(strings.Split(config.Servers, ",")...)
	if err := client.Ping(); err != nil {
		return nil, err
	}
	return client, nil
}

func (m *Memcached) Add(ctx context.Context, userID, postID string, record *storage.Record) {
	_, span := m.tracer.Start(ctx, "memcached.Add")
	defer span.End()

	b, err := json.Marshal(record)
	if err != nil {
		return
	}

	_ = m.client.Set(&memcache.Item{
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

func (m *Memcached) Remove(ctx context.Context, userID, postID string) {
	_, span := m.tracer.Start(ctx, "memcached.Removed")
	defer span.End()

	_ = m.client.Delete(cache.CreateKey(userID, postID))
}
