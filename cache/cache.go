package cache

import (
	"context"

	"github.com/craigpastro/crudapp/storage"
)

type Cache interface {
	Add(ctx context.Context, userID, postID string, record *storage.Record)
	Get(ctx context.Context, userID, postID string) (*storage.Record, bool)
	Remove(ctx context.Context, userID, postID string) bool
}

type noopCache struct{}

func NewNoopCache() *noopCache {
	return &noopCache{}
}

func (n *noopCache) Add(_ context.Context, _, _ string, _ *storage.Record) {}

func (n *noopCache) Get(_ context.Context, _, _ string) (*storage.Record, bool) {
	return nil, false
}

func (n *noopCache) Remove(_ context.Context, _, _ string) bool {
	return true
}
