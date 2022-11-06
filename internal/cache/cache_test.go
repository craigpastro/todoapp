package cache_test

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/craigpastro/crudapp/internal/cache"
	"github.com/craigpastro/crudapp/internal/cache/memcached"
	"github.com/craigpastro/crudapp/internal/cache/memory"
	"github.com/craigpastro/crudapp/internal/cache/redis"
	"github.com/craigpastro/crudapp/internal/storage"
	"github.com/craigpastro/crudapp/internal/telemetry"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"go.uber.org/zap"
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
		newRedis(t),
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
	logger := zap.NewNop()
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

	client, err := memcached.CreateClient(fmt.Sprintf("localhost:%s", port.Port()), logger)
	require.NoError(t, err)

	return cacheTest{
		name:      "memcached",
		cache:     memcached.New(client, tracer),
		container: container,
	}
}

func newRedis(t *testing.T) cacheTest {
	ctx := context.Background()
	logger := zap.NewNop()
	tracer := telemetry.NewNoopTracer()

	req := testcontainers.ContainerRequest{
		Image:        "redis:latest",
		ExposedPorts: []string{"6379/tcp"},
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "6379/tcp")
	require.NoError(t, err)

	client, err := redis.CreateClient(ctx, fmt.Sprintf("localhost:%s", port.Port()), "", logger)
	require.NoError(t, err)

	return cacheTest{
		name:      "redis",
		cache:     redis.New(client, tracer),
		container: container,
	}
}

func testGet(t *testing.T, cache cache.Cache) {
	ctx := context.Background()
	userID := ulid.Make().String()
	postID := ulid.Make().String()
	now := time.Now()
	record := storage.NewRecord(userID, postID, data, now, now)

	err := cache.Add(ctx, userID, postID, record)
	require.NoError(t, err)

	gotRecord, ok := cache.Get(ctx, userID, postID)
	require.True(t, ok)

	// Monotonic clock issues: see https://github.com/stretchr/testify/issues/502
	require.True(t, cmp.Equal(record, gotRecord, cmpopts.IgnoreFields(storage.Record{}, "CreatedAt", "UpdatedAt")))
}

func testRemove(t *testing.T, cache cache.Cache) {
	ctx := context.Background()
	userID := ulid.Make().String()
	postID := ulid.Make().String()
	now := time.Now()
	record := storage.NewRecord(userID, postID, data, now, now)

	err := cache.Add(ctx, userID, postID, record)
	require.NoError(t, err)

	_, ok := cache.Get(ctx, userID, postID)
	require.True(t, ok)

	err = cache.Remove(ctx, userID, postID)
	require.NoError(t, err)

	_, ok = cache.Get(ctx, userID, postID)
	require.False(t, ok)
}
