package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/bufbuild/connect-go"
	"github.com/craigpastro/crudapp/cache"
	"github.com/craigpastro/crudapp/cache/memcached"
	cachememory "github.com/craigpastro/crudapp/cache/memory"
	"github.com/craigpastro/crudapp/errors"
	"github.com/craigpastro/crudapp/internal/gen/crudapp/v1/crudappv1connect"
	"github.com/craigpastro/crudapp/internal/middleware"
	"github.com/craigpastro/crudapp/server"
	"github.com/craigpastro/crudapp/storage"
	"github.com/craigpastro/crudapp/storage/dynamodb"
	"github.com/craigpastro/crudapp/storage/memory"
	"github.com/craigpastro/crudapp/storage/mongodb"
	"github.com/craigpastro/crudapp/storage/postgres"
	"github.com/craigpastro/crudapp/storage/redis"
	"github.com/craigpastro/crudapp/telemetry"
	"github.com/kelseyhightower/envconfig"
	"go.opentelemetry.io/otel/trace"
)

type Config struct {
	ServiceName    string `split_words:"true" default:"crudapp"`
	ServiceVersion string `default:"0.1.0"`
	Environment    string `default:"dev"`

	Addr        string `split_words:"true" default:"127.0.0.1:8080"`
	StorageType string `split_words:"true" default:"memory"`
	CacheType   string `split_words:"true" default:"memory"`

	CacheSize int `split_words:"true" default:"10000"`

	DynamoDBRegion string `envconfig:"DYNAMODB_REGION" default:"us-west-2"`
	DynamoDBLocal  bool   `envconfig:"DYNAMODB_LOCAL" default:"false"`

	MongoDBURI string `envconfig:"MONGODB_URI" default:"mongodb://mongodb:password@127.0.0.1:27017"`

	MemcachedServers string `split_words:"true" default:"localhost:11211"`

	PostgresURI string `split_words:"true" default:"postgres://postgres:password@127.0.0.1:5432/postgres"`

	RedisAddr     string `split_words:"true" default:"localhost:6379"`
	RedisPassword string `split_words:"true" default:""`

	TraceProviderEnabled bool   `split_words:"true" default:"false"`
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

	if err := run(ctx, config); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, config Config) error {
	logger, err := telemetry.NewLogger(telemetry.LoggerConfig{
		ServiceName:    config.ServiceName,
		ServiceVersion: config.ServiceVersion,
		Environment:    config.Environment,
	})
	if err != nil {
		log.Fatal(fmt.Errorf("error initializing logger: %w", err))
	}

	tracer, err := telemetry.NewTracer(ctx, telemetry.TracerConfig{
		Enabled:        config.TraceProviderEnabled,
		ServiceName:    config.ServiceName,
		ServiceVersion: config.ServiceVersion,
		Environment:    config.Environment,
		Endpoint:       config.TraceProviderURL,
	})
	if err != nil {
		logger.Fatal("error initializing tracer", telemetry.Error(err))
	}

	storage, err := newStorage(ctx, tracer, config)
	if err != nil {
		logger.Fatal("error initializing storage", telemetry.Error(err))
	}

	cache, err := newCache(tracer, config)
	if err != nil {
		logger.Fatal("error initializing cache", telemetry.Error(err))
	}

	interceptors := connect.WithInterceptors(
		middleware.NewLoggingInterceptor(logger),
	)

	mux := http.NewServeMux()
	mux.Handle(crudappv1connect.NewCrudAppServiceHandler(
		server.NewServer(cache, storage, tracer),
		interceptors,
	))

	logger.Info(fmt.Sprintf("server starting on %s (storage type=%s)", config.Addr, config.StorageType))
	if err := http.ListenAndServe(config.Addr, mux); err != nil {
		logger.Fatal("failed to listen and serve", telemetry.Error(err))
	}

	return nil
}

func newCache(tracer trace.Tracer, config Config) (cache.Cache, error) {
	switch config.CacheType {
	case "memcached":
		client, err := memcached.CreateClient(config.MemcachedServers)
		if err != nil {
			return nil, err
		}
		return memcached.New(client, tracer), nil
	case "memory":
		return cachememory.New(tracer, config.CacheSize)
	case "noop":
		return cache.NewNoopCache(), nil
	default:
		return nil, errors.ErrUndefinedCacheType
	}
}

func newStorage(ctx context.Context, tracer trace.Tracer, config Config) (storage.Storage, error) {
	switch config.StorageType {
	case "dynamodb":
		client, err := dynamodb.CreateClient(ctx, config.DynamoDBRegion, config.DynamoDBLocal)
		if err != nil {
			return nil, err
		}
		return dynamodb.New(client, tracer), nil
	case "memory":
		return memory.New(tracer), nil
	case "mongodb":
		coll, err := mongodb.CreateCollection(ctx, config.MongoDBURI)
		if err != nil {
			return nil, err
		}
		return mongodb.New(coll, tracer), nil
	case "postgres":
		pool, err := postgres.CreatePool(ctx, config.PostgresURI)
		if err != nil {
			return nil, err
		}
		return postgres.New(pool, tracer), nil
	case "redis":
		client, err := redis.CreateClient(ctx, config.RedisAddr, config.RedisPassword)
		if err != nil {
			return nil, err
		}
		return redis.New(client, tracer), nil
	default:
		return nil, errors.ErrUndefinedStorageType
	}
}
