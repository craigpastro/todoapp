package cache_test

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/craigpastro/crudapp/cache"
	"github.com/craigpastro/crudapp/cache/memcached"
	"github.com/craigpastro/crudapp/cache/memory"
	"github.com/craigpastro/crudapp/myid"
	"github.com/craigpastro/crudapp/storage"
	"github.com/craigpastro/crudapp/telemetry"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/require"
)

type cacheTest struct {
	name     string
	cache    cache.Cache
	resource *dockertest.Resource
}

const data = "some data"

func TestCache(t *testing.T) {
	dockerpool, err := dockertest.NewPool("")
	require.NoError(t, err)

	cacheTests := []cacheTest{
		newMemcached(t, dockerpool),
		newMemory(t),
	}

	for _, test := range cacheTests {
		t.Run(test.name, func(t *testing.T) {
			testGet(t, test.cache)
			testRemove(t, test.cache)

			if test.resource != nil {
				err := test.resource.Close()
				if err != nil {
					log.Println(err)
				}
			}
		})
	}
}

func newMemory(t *testing.T) cacheTest {
	tracer := telemetry.NewNoopTracer()

	cache, err := memory.New(10, tracer)
	require.NoError(t, err)
	return cacheTest{
		name:  "memory",
		cache: cache,
	}
}

func newMemcached(t *testing.T, dockerpool *dockertest.Pool) cacheTest {
	tracer := telemetry.NewNoopTracer()

	resource, err := dockerpool.Run("memcached", "latest", nil)
	require.NoError(t, err)

	var client *memcache.Client
	err = dockerpool.Retry(func() error {
		var err error
		client, err = memcached.CreateClient(memcached.Config{Servers: fmt.Sprintf("localhost:%s", resource.GetPort("11211/tcp"))})
		if err != nil {
			return err
		}
		return client.Ping()
	})
	require.NoError(t, err)

	return cacheTest{
		name:     "memcached",
		cache:    memcached.New(client, tracer),
		resource: resource,
	}
}

func testGet(t *testing.T, cache cache.Cache) {
	ctx := context.Background()
	userID := myid.New()
	postID := myid.New()
	now := time.Now()
	record := storage.NewRecord(userID, postID, data, now, now)

	cache.Add(ctx, userID, postID, record)
	gotRecord, ok := cache.Get(ctx, userID, postID)
	require.True(t, ok)

	require.Equal(t, record.UserID, gotRecord.UserID)
	require.Equal(t, record.PostID, gotRecord.PostID)
	require.Equal(t, record.Data, gotRecord.Data)
	require.True(t, record.CreatedAt.Equal(gotRecord.CreatedAt))
	require.True(t, record.UpdatedAt.Equal(gotRecord.UpdatedAt))
}

func testRemove(t *testing.T, cache cache.Cache) {
	ctx := context.Background()
	userID := myid.New()
	postID := myid.New()
	now := time.Now()
	record := storage.NewRecord(userID, postID, data, now, now)

	cache.Add(ctx, userID, postID, record)
	_, ok := cache.Get(ctx, userID, postID)
	require.True(t, ok)

	cache.Remove(ctx, userID, postID)
	_, ok = cache.Get(ctx, userID, postID)
	require.False(t, ok)
}
