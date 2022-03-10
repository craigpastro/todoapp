package memory

import (
	"context"

	"github.com/craigpastro/crudapp/cache"
	"github.com/craigpastro/crudapp/storage"
	lru "github.com/hashicorp/golang-lru"
	"go.opentelemetry.io/otel/trace"
)

type Memory struct {
	store  *lru.Cache
	tracer trace.Tracer
}

func New(tracer trace.Tracer, size int) (cache.Cache, error) {
	store, err := lru.New(size)
	if err != nil {
		return nil, err
	}
	return &Memory{
		store:  store,
		tracer: tracer,
	}, nil
}

func (m *Memory) Add(ctx context.Context, userID, postID string, record *storage.Record) {
	_, span := m.tracer.Start(ctx, "memory.Add")
	defer span.End()

	m.store.Add(cache.CreateKey(userID, postID), record)
}

func (m *Memory) Get(ctx context.Context, userID, postID string) (*storage.Record, bool) {
	_, span := m.tracer.Start(ctx, "memory.Get")
	defer span.End()

	value, ok := m.store.Get(cache.CreateKey(userID, postID))
	if !ok {
		return nil, false
	}

	record := value.(*storage.Record)

	return record, true
}

func (m *Memory) Remove(ctx context.Context, userID, postID string) {
	_, span := m.tracer.Start(ctx, "memory.Removed")
	defer span.End()

	m.store.Remove(cache.CreateKey(userID, postID))
}
