package memcached

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/craigpastro/crudapp/cache"
	"github.com/craigpastro/crudapp/instrumentation"
	"github.com/craigpastro/crudapp/myid"
	"github.com/craigpastro/crudapp/storage"
	"github.com/kelseyhightower/envconfig"
	"go.opentelemetry.io/otel/trace"
)

const data = "some data"

var (
	ctx    context.Context
	tracer trace.Tracer
	c      cache.Cache
)

type Config struct {
	MemcachedServers string `split_words:"true" default:"localhost:11211"`
}

func TestMain(m *testing.M) {
	var config Config
	if err := envconfig.Process("", &config); err != nil {
		fmt.Printf("error reading config: %v\n", err)
		os.Exit(1)
	}

	client := memcache.New(config.MemcachedServers)
	if err := client.Ping(); err != nil {
		fmt.Printf("error connecting to Memcached: %v\n", err)
		os.Exit(1)
	}

	ctx = context.Background()
	tracer, _ = instrumentation.NewTracer(ctx, instrumentation.TracerConfig{Enabled: false})
	c = &Memcached{
		client: client,
		tracer: tracer,
	}
}

func TestGet(t *testing.T) {
	userID := myid.New()
	postID := myid.New()
	now := time.Now()
	record := storage.NewRecord(userID, postID, data, now, now)
	c.Add(ctx, userID, postID, record)
	gotRecord, ok := c.Get(ctx, userID, postID)

	if !ok {
		t.Error("did not get record")
	}

	if !reflect.DeepEqual(gotRecord, record) {
		t.Errorf("gotRecord is not the same as added record: got=%v, added=%v", gotRecord, record)
	}
}

func TestRemove(t *testing.T) {
	userID := myid.New()
	postID := myid.New()
	now := time.Now()
	record := storage.NewRecord(userID, postID, data, now, now)
	c.Add(ctx, userID, postID, record)
	if _, ok := c.Get(ctx, userID, postID); !ok {
		t.Error("error inserting record")
	}

	c.Remove(ctx, userID, postID)
	if _, ok := c.Get(ctx, userID, postID); ok {
		t.Error("error removing record")
	}
}
