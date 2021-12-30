package main

import (
	"context"
	"log"
	"net"

	pb "github.com/craigpastro/crudapp/api/proto/v1"
	"github.com/craigpastro/crudapp/server"
	"github.com/craigpastro/crudapp/storage"
	"github.com/craigpastro/crudapp/storage/memory"
	"github.com/craigpastro/crudapp/storage/postgres"
	"github.com/craigpastro/crudapp/storage/redis"
	"github.com/kelseyhightower/envconfig"
	"google.golang.org/grpc"
)

type Config struct {
	ServerAddr  string `split_words:"true" default:"127.0.0.1:8080"`
	StorageType string `split_words:"true" default:"memory"`

	PostgresURI string `split_words:"true" default:"postgres://postgres:password@127.0.0.1:5432/postgres"`

	RedisAddr     string `split_words:"true" default:"localhost:6379"`
	RedisPassword string `split_words:"true" default:""`
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
		log.Fatalf("error initializing storage: %s", err)
	}

	s := grpc.NewServer()
	pb.RegisterServiceServer(s, server.NewServer(storage))

	lis, err := net.Listen("tcp", config.ServerAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func NewStorage(ctx context.Context, config Config) (storage.Storage, error) {
	switch config.StorageType {
	case "memory":
		return memory.New(), nil
	case "postgres":
		return postgres.New(ctx, config.PostgresURI)
	case "redis":
		return redis.New(ctx, config.RedisAddr, config.RedisPassword)
	default:
		return nil, storage.ErrUndefinedStorageType
	}
}
