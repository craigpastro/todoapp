package main

import (
	"log"

	"github.com/craigpastro/crudapp/router"
	"github.com/craigpastro/crudapp/storage"
	"github.com/craigpastro/crudapp/storage/memory"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	ServerAddr  string `split_words:"true" default:"127.0.0.1"`
	StorageType string `split_words:"true" default:"memory"`
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("error loading '.env'", err)
	}

	var config Config
	if err := envconfig.Process("", &config); err != nil {
		log.Fatal("error reading config", err)
	}

	Run(config)
}

func Run(config Config) {
	storage, err := NewStorage(config.StorageType)
	if err != nil {
		log.Fatal("error initializing storage")
	}

	if err := router.Run(config.ServerAddr, storage); err != nil {
		log.Fatal("error starting the server", err)
	}
}

func NewStorage(storageType string) (storage.Storage, error) {
	switch storageType {
	case "memory":
		return memory.New(), nil
	default:
		return nil, storage.ErrUndefinedStorageType
	}
}
