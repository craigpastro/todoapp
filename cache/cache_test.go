package cache_test

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/craigpastro/crudapp/cache"
	"github.com/craigpastro/crudapp/cache/memcached"
	"github.com/craigpastro/crudapp/cache/memory"
	"github.com/craigpastro/crudapp/myid"
	"github.com/craigpastro/crudapp/storage"
	"github.com/craigpastro/crudapp/telemetry"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
)

type cacheTest struct {
	name      string
	cache     cache.Cache
	container testcontainers.Container
}

const data = "some data"

func TestCache(t *testing.T) {
	cacheTests := []cacheTest{
		newMemcached(t),
		newMemory(t),
	}

	for _, test := range cacheTests {
		t.Run(test.name, func(t *testing.T) {
			testGet(t, test.cache)
			testRemove(t, test.cache)

			if test.container != nil {
				if err := test.container.Terminate(context.Background()); err != nil {
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

func newMemcached(t *testing.T) cacheTest {
	ctx := context.Background()
	tracer := telemetry.NewNoopTracer()

	req := testcontainers.ContainerRequest{
		Image:        "memcached:latest",
		ExposedPorts: []string{"11211/tcp"},
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "11211/tcp")
	require.NoError(t, err)

	client, err := memcached.CreateClient(memcached.Config{Servers: fmt.Sprintf("localhost:%s", port.Port())})
	require.NoError(t, err)

	return cacheTest{
		name:      "memcached",
		cache:     memcached.New(client, tracer),
		container: container,
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
