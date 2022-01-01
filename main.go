package main

import (
	"context"
	"io"
	"log"
	"net"
	"net/http"

	"github.com/craigpastro/crudapp/instrumentation/tracer"
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

	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

type Config struct {
	ServiceName    string `split_words:"true" default:"crudapp"`
	ServiceVersion string `default:"0.1.0"`
	Environment    string `default:"dev"`

	RPCAddr     string `split_words:"true" default:"localhost:9090"`
	ServerAddr  string `split_words:"true" default:"localhost:8080"`
	StorageType string `split_words:"true" default:"memory"`

	PostgresURI string `split_words:"true" default:"postgres://postgres:password@127.0.0.1:5432/postgres"`

	RedisAddr     string `split_words:"true" default:"localhost:6379"`
	RedisPassword string `split_words:"true" default:""`

	TraceProviderEnabled bool   `split_words:"true" default:"true"`
	TraceProviderURL     string `split_words:"true" default:"locaalhost:4318"`
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
		log.Fatalf("error initializing storage: %v", err)
	}

	tracer, err := tracer.New(ctx, config.TraceProviderEnabled, tracer.Config{
		ServiceName:    config.ServiceName,
		ServiceVersion: config.ServiceVersion,
		Environment:    config.Environment,
		Endpoint:       config.TraceProviderURL,
	})

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

func newExporter(w io.Writer) (trace.SpanExporter, error) {
	return stdouttrace.New(
		stdouttrace.WithWriter(w),
		stdouttrace.WithPrettyPrint(),
		stdouttrace.WithoutTimestamps(),
	)
}

func newResource() *resource.Resource {
	r, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("crudapp"),
			semconv.ServiceVersionKey.String("v0.1.0"),
		),
	)
	return r
}
