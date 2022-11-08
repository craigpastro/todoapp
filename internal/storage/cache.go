package storage

import (
	"context"
	"fmt"

	lru "github.com/hashicorp/golang-lru"
)

type CachingStorage struct {
	cache   *lru.Cache
	storage Storage
}

func NewCachingStorage(storage Storage, size int) (*CachingStorage, error) {
	cache, err := lru.New(size)
	if err != nil {
		return nil, err
	}

	return &CachingStorage{
		cache:   cache,
		storage: storage,
	}, nil
}

func (c *CachingStorage) Create(ctx context.Context, userID, data string) (*Record, error) {
	record, err := c.storage.Create(ctx, userID, data)
	if err != nil {
		return nil, err
	}

	c.cache.Add(createKey(userID, record.PostID), record)

	return record, nil
}

func (c *CachingStorage) Read(ctx context.Context, userID, postID string) (*Record, error) {
	if record, ok := c.cache.Get(createKey(userID, postID)); ok {
		return record.(*Record), nil
	}

	return c.storage.Read(ctx, userID, postID)
}

func (c *CachingStorage) ReadAll(ctx context.Context, userID string) (RecordIterator, error) {
	return c.storage.ReadAll(ctx, userID)
}

func (c *CachingStorage) Upsert(ctx context.Context, userID, postID, data string) (*Record, error) {
	record, err := c.storage.Upsert(ctx, userID, postID, userID)
	if err != nil {
		return nil, err
	}

	c.cache.Add(createKey(userID, postID), record)

	return record, nil
}

func (c *CachingStorage) Delete(ctx context.Context, userID, postID string) error {
	c.cache.Remove(createKey(userID, postID))

	return c.storage.Delete(ctx, userID, postID)
}

func createKey(userID, postID string) string {
	return fmt.Sprintf("%s#%s", userID, postID)
}
