package main

import (
	"context"
	"log"
	"net"
	"net/http"

	pb "github.com/craigpastro/crudapp/protos/api/v1"
	"github.com/craigpastro/crudapp/server"
	"github.com/craigpastro/crudapp/storage"
	"github.com/craigpastro/crudapp/storage/memory"
	"github.com/craigpastro/crudapp/storage/postgres"
	"github.com/craigpastro/crudapp/storage/redis"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/kelseyhightower/envconfig"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Config struct {
	RPCAddr     string `split_words:"true" default:"localhost:9090"`
	ServerAddr  string `split_words:"true" default:"localhost:8080"`
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

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	run(ctx, config)
}

func run(ctx context.Context, config Config) {
	storage, err := newStorage(ctx, config)
	if err != nil {
		log.Fatalf("error initializing storage: %s", err)
	}

	s := grpc.NewServer()
	pb.RegisterServiceServer(s, server.NewServer(storage))

	lis, err := net.Listen("tcp", config.RPCAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	go s.Serve(lis)

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	if err := pb.RegisterServiceHandlerFromEndpoint(ctx, mux, config.RPCAddr, opts); err != nil {
		log.Fatalf("failed to register service: %v", err)
	}

	if err := http.ListenAndServe(config.ServerAddr, mux); err != nil {
		log.Fatalf("failed to listen and serve: %v", err)
	}
}

func newStorage(ctx context.Context, config Config) (storage.Storage, error) {
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
