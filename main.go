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
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel/trace"
)

type Config struct {
	Service ServiceConfig
	Trace   TraceConfig
	Storage StorageConfig
	Cache   CacheConfig
}

type ServiceConfig struct {
	Name        string
	Version     string
	Environment string
	Addr        string
}

type TraceConfig struct {
	Enabled     bool
	ProviderURL string
}

type StorageConfig struct {
	Type     string // memory, dynamodb, mongodb, postgres, redis
	DynamoDB dynamodb.Config
	MongoDB  mongodb.Config
	Postgres postgres.Config
	Redis    redis.Config
}

type CacheConfig struct {
	Type      string // memory, memcached
	Size      int
	Memcached memcached.Config
}

func main() {
	config, err := readConfig()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if err := run(ctx, config); err != nil {
		log.Fatal(err)
	}
}

func readConfig() (*Config, error) {
	viper.SetConfigFile("config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error reading config: %w", err)
	}

	return &config, nil
}

func run(ctx context.Context, config *Config) error {
	logger, err := telemetry.NewLogger(telemetry.LoggerConfig{
		ServiceName:    config.Service.Name,
		ServiceVersion: config.Service.Version,
		Environment:    config.Service.Environment,
	})
	if err != nil {
		log.Fatal(fmt.Errorf("error initializing logger: %w", err))
	}

	tracer, err := telemetry.NewTracer(ctx, telemetry.TracerConfig{
		Enabled:        config.Trace.Enabled,
		ServiceName:    config.Service.Name,
		ServiceVersion: config.Service.Version,
		Environment:    config.Service.Environment,
		Endpoint:       config.Trace.ProviderURL,
	})
	if err != nil {
		logger.Fatal("error initializing tracer", telemetry.Error(err))
	}

	storage, err := newStorage(ctx, tracer, &config.Storage)
	if err != nil {
		logger.Fatal("error initializing storage", telemetry.Error(err))
	}

	cache, err := newCache(tracer, &config.Cache)
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

	logger.Info(fmt.Sprintf("server starting on %s (storage type=%s)", config.Service.Addr, config.Storage.Type))
	if err := http.ListenAndServe(config.Service.Addr, mux); err != nil {
		logger.Fatal("failed to listen and serve", telemetry.Error(err))
	}

	return nil
}

func newCache(tracer trace.Tracer, config *CacheConfig) (cache.Cache, error) {
	switch config.Type {
	case "memcached":
		client, err := memcached.CreateClient(config.Memcached)
		if err != nil {
			return nil, err
		}
		return memcached.New(client, tracer), nil
	case "memory":
		return cachememory.New(tracer, config.Size)
	case "noop":
		return cache.NewNoopCache(), nil
	default:
		return nil, fmt.Errorf("undefined cache kind: %s", config.Type)
	}
}

func newStorage(ctx context.Context, tracer trace.Tracer, config *StorageConfig) (storage.Storage, error) {
	switch config.Type {
	case "dynamodb":
		client, err := dynamodb.CreateClient(ctx, config.DynamoDB)
		if err != nil {
			return nil, err
		}
		return dynamodb.New(client, tracer), nil
	case "memory":
		return memory.New(tracer), nil
	case "mongodb":
		coll, err := mongodb.CreateCollection(ctx, config.MongoDB)
		if err != nil {
			return nil, err
		}
		return mongodb.New(coll, tracer), nil
	case "postgres":
		pool, err := postgres.CreatePool(ctx, config.Postgres)
		if err != nil {
			return nil, err
		}
		return postgres.New(pool, tracer), nil
	case "redis":
		client, err := redis.CreateClient(ctx, config.Redis)
		if err != nil {
			return nil, err
		}
		return redis.New(client, tracer), nil
	default:
		return nil, fmt.Errorf("undefined storage kind: %s", config.Type)
	}
}
