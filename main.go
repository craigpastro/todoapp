package main

import (
	"context"
	"fmt"
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
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/kelseyhightower/envconfig"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

type Config struct {
	ServiceName    string `split_words:"true" default:"crudapp"`
	ServiceVersion string `default:"0.1.0"`
	Environment    string `default:"dev"`

	RPCAddr     string `split_words:"true" default:"127.0.0.1:9090"`
	ServerAddr  string `split_words:"true" default:"127.0.0.1:8080"`
	StorageType string `split_words:"true" default:"memory"`

	DynamoDBRegion string `envconfig:"DYNAMODB_REGION" default:"us-west-2"`
	DynamoDBLocal  bool   `envconfig:"DYNAMODB_LOCAL" default:"false"`

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
		panic(fmt.Sprintf("error reading config: %v", err))
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	run(ctx, config)
}

func run(ctx context.Context, config Config) {
	logger, err := instrumentation.NewLogger(instrumentation.LoggerConfig{
		ServiceName:    config.ServiceName,
		ServiceVersion: config.ServiceVersion,
		Environment:    config.Environment,
	})
	if err != nil {
		panic(fmt.Sprintf("error initializing logger: %v", err))
	}

	tracer, err := instrumentation.NewTracer(ctx, instrumentation.TracerConfig{
		Enabled:        config.TraceProviderEnabled,
		ServiceName:    config.ServiceName,
		ServiceVersion: config.ServiceVersion,
		Environment:    config.Environment,
		Endpoint:       config.TraceProviderURL,
	})
	if err != nil {
		logger.Fatal("error initializing tracer", instrumentation.Error(err))
	}

	storage, err := newStorage(ctx, tracer, config)
	if err != nil {
		logger.Fatal("error initializing storage", instrumentation.Error(err))
	}

	s := grpc.NewServer(
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_zap.StreamServerInterceptor(logger),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_zap.UnaryServerInterceptor(logger),
		)),
	)
	pb.RegisterServiceServer(s, server.NewServer(tracer, storage))
	reflection.Register(s)

	lis, err := net.Listen("tcp", config.RPCAddr)
	if err != nil {
		logger.Fatal("failed to listen", instrumentation.Error(err))
	}
	go s.Serve(lis)

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
	}
	if err := pb.RegisterServiceHandlerFromEndpoint(ctx, mux, config.RPCAddr, opts); err != nil {
		logger.Fatal("failed to register service", instrumentation.Error(err))
	}

	logger.Info(fmt.Sprintf("server starting on %s (storage type=%s)", config.ServerAddr, config.StorageType))
	if err := http.ListenAndServe(config.ServerAddr, mux); err != nil {
		logger.Fatal("failed to listen and serve", instrumentation.Error(err))
	}
}

func newStorage(ctx context.Context, tracer trace.Tracer, config Config) (storage.Storage, error) {
	switch config.StorageType {
	case "dynamodb":
		return dynamodb.New(ctx, tracer, config.DynamoDBRegion, config.DynamoDBLocal)
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
