package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/bufbuild/connect-go"
	grpcreflect "github.com/bufbuild/connect-grpcreflect-go"
	"github.com/craigpastro/crudapp/cache"
	"github.com/craigpastro/crudapp/cache/memcached"
	cachememory "github.com/craigpastro/crudapp/cache/memory"
	"github.com/craigpastro/crudapp/cache/redis"
	"github.com/craigpastro/crudapp/internal/gen/crudapp/v1/crudappv1connect"
	"github.com/craigpastro/crudapp/internal/middleware"
	"github.com/craigpastro/crudapp/server"
	"github.com/craigpastro/crudapp/storage"
	"github.com/craigpastro/crudapp/storage/memory"
	"github.com/craigpastro/crudapp/storage/mongodb"
	"github.com/craigpastro/crudapp/storage/postgres"
	"github.com/craigpastro/crudapp/telemetry"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type Config struct {
	Service   ServiceConfig
	LogFormat string
	Trace     TraceConfig
	Storage   StorageConfig
	Cache     CacheConfig
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
	Type     string // memory, mongodb, postgres
	MongoDB  mongodb.Config
	Postgres postgres.Config
}

type CacheConfig struct {
	Type      string // memory, memcached, redis
	Size      int
	Memcached memcached.Config
	Redis     redis.Config
}

func main() {
	cfg, err := readConfig()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if err := run(ctx, cfg); err != nil {
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

func run(ctx context.Context, cfg *Config) error {
	logger := newLogger(cfg)

	tracer, err := telemetry.NewTracer(ctx, telemetry.TracerConfig{
		Enabled:        cfg.Trace.Enabled,
		ServiceName:    cfg.Service.Name,
		ServiceVersion: cfg.Service.Version,
		Environment:    cfg.Service.Environment,
		Endpoint:       cfg.Trace.ProviderURL,
	})
	if err != nil {
		logger.Fatal("error initializing tracer", zap.Error(err))
	}

	cache, err := newCache(ctx, logger, tracer, &cfg.Cache)
	if err != nil {
		logger.Fatal("error initializing cache", zap.Error(err))
	}

	storage, err := newStorage(ctx, logger, tracer, &cfg.Storage)
	if err != nil {
		logger.Fatal("error initializing storage", zap.Error(err))
	}

	interceptors := connect.WithInterceptors(
		middleware.NewLoggingInterceptor(logger),
	)

	mux := http.NewServeMux()
	reflector := grpcreflect.NewStaticReflector(crudappv1connect.CrudAppServiceName)
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))
	mux.Handle(crudappv1connect.NewCrudAppServiceHandler(
		server.NewServer(cache, storage, tracer),
		interceptors,
	))

	logger.Info(fmt.Sprintf("server starting on %s (storage type=%s)", cfg.Service.Addr, cfg.Storage.Type))

	return http.ListenAndServe(
		":8080",
		h2c.NewHandler(mux, &http2.Server{}),
	)
}

func newLogger(cfg *Config) *zap.Logger {
	zapCfg := zap.NewProductionConfig()
	if cfg.LogFormat == "console" {
		zapCfg.Encoding = "console"
	}
	return zap.Must(zapCfg.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.PanicLevel),
		zap.Fields(
			zap.String("serviceName", cfg.Service.Name),
			zap.String("serviceVersion", cfg.Service.Version),
			zap.String("environment", cfg.Service.Environment),
		),
	))
}

func newCache(ctx context.Context, logger *zap.Logger, tracer trace.Tracer, cfg *CacheConfig) (cache.Cache, error) {
	switch cfg.Type {
	case "memcached":
		client, err := memcached.CreateClient(cfg.Memcached, logger)
		if err != nil {
			return nil, err
		}
		return memcached.New(client, tracer), nil
	case "memory":
		return cachememory.New(cfg.Size, tracer)
	case "redis":
		client, err := redis.CreateClient(ctx, cfg.Redis, logger)
		if err != nil {
			return nil, err
		}
		return redis.New(client, tracer), nil
	case "noop":
		return cache.NewNoopCache(), nil
	default:
		return nil, fmt.Errorf("undefined cache kind: %s", cfg.Type)
	}
}

func newStorage(ctx context.Context, logger *zap.Logger, tracer trace.Tracer, cfg *StorageConfig) (storage.Storage, error) {
	switch cfg.Type {
	case "memory":
		return memory.New(tracer), nil
	case "mongodb":
		coll, err := mongodb.CreateCollection(ctx, cfg.MongoDB, logger)
		if err != nil {
			return nil, err
		}
		return mongodb.New(coll, tracer), nil
	case "postgres":
		pool, err := postgres.CreatePool(ctx, cfg.Postgres, logger)
		if err != nil {
			return nil, err
		}
		return postgres.New(pool, tracer), nil
	default:
		return nil, fmt.Errorf("undefined storage kind: %s", cfg.Type)
	}
}
