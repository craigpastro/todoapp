package cache

import (
	"context"
)

type Cache interface {
	Add(ctx context.Context, key string, value interface{}) (interface{}, bool)
	Get(ctx context.Context, key string) (interface{}, bool)
	Remove(ctx context.Context, key string) bool
}

type noopCache struct{}

func (n *noopCache) Add(_ context.Context, _ string, _ interface{}) (interface{}, bool) {
	return nil, false
}

func (n *noopCache) Get(_ context.Context, _ string) (interface{}, bool) {
	return nil, false
}

func (n *noopCache) Remove(_ context.Context, _ string) bool {
	return true
}
