package redis

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/craigpastro/crudapp/myid"
	"github.com/craigpastro/crudapp/storage"
	"github.com/go-redis/redis/v8"
	"github.com/kelseyhightower/envconfig"
)

var (
	ctx context.Context
	db  *Redis
)

type Config struct {
	RedisAddr     string `split_words:"true" default:"localhost:6379"`
	RedisPassword string `split_words:"true" default:""`
}

func TestMain(m *testing.M) {
	var config Config
	if err := envconfig.Process("", &config); err != nil {
		fmt.Println("error reading config", err)
		os.Exit(1)
	}

	client := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
	})

	ctx = context.Background()
	db = &Redis{client: client}

	os.Exit(m.Run())
}

func TestReadNotExists(t *testing.T) {
	userID := myid.New()
	_, err := db.Read(ctx, userID, "1")
	if err != storage.ErrPostDoesNotExist {
		t.Errorf("wanted ErrPostDoesNotExist, but got: %s", err)
	}
}

func TestUpdateNotExists(t *testing.T) {
	userID := myid.New()
	_, err := db.Update(ctx, userID, "1", "new data")
	if err != storage.ErrPostDoesNotExist {
		t.Errorf("wanted ErrPostDoesNotExist, but got: %s", err)
	}
}
