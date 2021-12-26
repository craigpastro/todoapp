package main

import (
	"context"
	"log"

	"github.com/craigpastro/crudapp/router"
	"github.com/craigpastro/crudapp/storage"
	"github.com/craigpastro/crudapp/storage/memory"
	"github.com/craigpastro/crudapp/storage/postgres"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	ServerAddr  string `split_words:"true" default:"127.0.0.1:8080"`
	StorageType string `split_words:"true" default:"memory"`

	PostgresURI string `split_words:"true" default:"postgres://postgres:password@127.0.0.1:5432/postgres"`
}

func main() {
	var config Config
	if err := envconfig.Process("", &config); err != nil {
		log.Fatal("error reading config", err)
	}

	Run(context.Background(), config)
}

func Run(ctx context.Context, config Config) {
	storage, err := NewStorage(ctx, config)
	if err != nil {
		log.Fatal("error initializing storage", err)
	}

	if err := router.Run(config.ServerAddr, storage); err != nil {
		log.Fatal("error starting the server", err)
	}
}

func NewStorage(ctx context.Context, config Config) (storage.Storage, error) {
	switch config.StorageType {
	case "memory":
		return memory.New(), nil
	case "postgres":
		return postgres.New(ctx, config.PostgresURI)
	default:
		return nil, storage.ErrUndefinedStorageType
	}
}
