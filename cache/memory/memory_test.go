package memory

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/craigpastro/crudapp/cache"
	"github.com/craigpastro/crudapp/instrumentation"
	"github.com/craigpastro/crudapp/myid"
	"github.com/craigpastro/crudapp/storage"
	lru "github.com/hashicorp/golang-lru"
	"github.com/kelseyhightower/envconfig"
)

const data = "some data"

var (
	ctx context.Context
	c   cache.Cache
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

	store, err := lru.New(10)
	if err != nil {
		fmt.Printf("error creating cache: %v\n", err)
		os.Exit(1)
	}

	c = &Memory{
		store:  store,
		tracer: instrumentation.NewNoopTracer(),
	}
}

func TestGet(t *testing.T) {
	userID := myid.New()
	postID := myid.New()
	now := time.Now()
	record := storage.NewRecord(userID, postID, data, now, now)
	c.Add(ctx, userID, postID, record)
	gotRecord, ok := c.Get(ctx, userID, postID)

	if ok != true {
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
	if _, ok := c.Get(ctx, userID, postID); ok == false {
		t.Error("error inserting record")
	}

	c.Remove(ctx, userID, postID)
	if _, ok := c.Get(ctx, userID, postID); ok == true {
		t.Error("error removing record")
	}
}
