package main

import (
	"context"
	"log"
	"net"
	"net/http"

	"github.com/craigpastro/crudapp/instrumentation"
	pb "github.com/craigpastro/crudapp/protos/api/v1"
	"github.com/craigpastro/crudapp/server"
	"github.com/craigpastro/crudapp/storage"
	"github.com/craigpastro/crudapp/storage/dynamodb"
	"github.com/craigpastro/crudapp/storage/memory"
	"github.com/craigpastro/crudapp/storage/mongodb"
	"github.com/craigpastro/crudapp/storage/postgres"
	"github.com/craigpastro/crudapp/storage/redis"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/kelseyhightower/envconfig"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Config struct {
	ServiceName    string `split_words:"true" default:"crudapp"`
	ServiceVersion string `default:"0.1.0"`
	Environment    string `default:"dev"`

	RPCAddr     string `split_words:"true" default:"localhost:9090"`
	ServerAddr  string `split_words:"true" default:"localhost:8080"`
	StorageType string `split_words:"true" default:"memory"`

	DynamoDBRegion   string `envconfig:"DYNAMODB_REGION" default:"us-west-2"`
	DynamoDBEndpoint string `envconfig:"DYNAMODB_ENDPOINT" default:"http://localhost:8000"`

	MongoDBURI string `envconfig:"MONGODB_URI" default:"mongodb://mongodb:password@127.0.0.1:27017"`

	PostgresURI string `split_words:"true" default:"postgres://postgres:password@127.0.0.1:5432/postgres"`

	RedisAddr     string `split_words:"true" default:"localhost:6379"`
	RedisPassword string `split_words:"true" default:""`

	TraceProviderEnabled bool   `split_words:"true" default:"true"`
	TraceProviderURL     string `split_words:"true" default:"localhost:4317"`
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
	tracer, err := instrumentation.NewTracer(ctx, instrumentation.TracerConfig{
		Enabled:        config.TraceProviderEnabled,
		ServiceName:    config.ServiceName,
		ServiceVersion: config.ServiceVersion,
		Environment:    config.Environment,
		Endpoint:       config.TraceProviderURL,
	})
	if err != nil {
		log.Fatalf("error initializing tracer: %v", err)
	}

	storage, err := newStorage(ctx, tracer, config)
	if err != nil {
		log.Fatalf("error initializing storage: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterServiceServer(s, server.NewServer(tracer, storage))

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

	log.Printf("starting server on %s", config.ServerAddr)
	if err := http.ListenAndServe(config.ServerAddr, mux); err != nil {
		log.Fatalf("failed to listen and serve: %v", err)
	}
}

func newStorage(ctx context.Context, tracer trace.Tracer, config Config) (storage.Storage, error) {
	switch config.StorageType {
	case "dynamodb":
		return dynamodb.New(ctx, tracer, config.DynamoDBRegion, config.DynamoDBEndpoint)
	case "memory":
		return memory.New(tracer), nil
	case "mongodb":
		return mongodb.New(ctx, tracer, config.MongoDBURI)
	case "postgres":
		return postgres.New(ctx, tracer, config.PostgresURI)
	case "redis":
		return redis.New(ctx, tracer, config.RedisAddr, config.RedisPassword)
	default:
		return nil, storage.ErrUndefinedStorageType
	}
}
