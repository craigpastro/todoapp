package memory

import (
	"context"

	"github.com/craigpastro/crudapp/cache"
	"github.com/craigpastro/crudapp/storage"
	lru "github.com/hashicorp/golang-lru"
	"go.opentelemetry.io/otel/trace"
)

var _ cache.Cache = (*Memory)(nil)

type Memory struct {
	cache  *lru.Cache
	tracer trace.Tracer
}

func New(size int, tracer trace.Tracer) (*Memory, error) {
	cache, err := lru.New(size)
	if err != nil {
		return nil, err
	}
	return &Memory{
		cache:  cache,
		tracer: tracer,
	}, nil
}

func (m *Memory) Add(ctx context.Context, userID, postID string, record *storage.Record) error {
	_, span := m.tracer.Start(ctx, "memory.Add")
	defer span.End()

	m.cache.Add(cache.CreateKey(userID, postID), record)

	return nil
}

func (m *Memory) Get(ctx context.Context, userID, postID string) (*storage.Record, bool) {
	_, span := m.tracer.Start(ctx, "memory.Get")
	defer span.End()

	value, ok := m.cache.Get(cache.CreateKey(userID, postID))
	if !ok {
		return nil, false
	}

	return value.(*storage.Record), true
}

func (m *Memory) Remove(ctx context.Context, userID, postID string) error {
	_, span := m.tracer.Start(ctx, "memory.Remove")
	defer span.End()

	m.cache.Remove(cache.CreateKey(userID, postID))

	return nil
}
