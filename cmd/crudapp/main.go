package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/bufbuild/connect-go"
	grpcreflect "github.com/bufbuild/connect-grpcreflect-go"
	"github.com/craigpastro/crudapp/internal/cache"
	"github.com/craigpastro/crudapp/internal/cache/memcached"
	cachememory "github.com/craigpastro/crudapp/internal/cache/memory"
	"github.com/craigpastro/crudapp/internal/cache/redis"
	"github.com/craigpastro/crudapp/internal/gen/crudapp/v1/crudappv1connect"
	"github.com/craigpastro/crudapp/internal/middleware"
	"github.com/craigpastro/crudapp/internal/storage"
	"github.com/craigpastro/crudapp/internal/storage/memory"
	"github.com/craigpastro/crudapp/internal/storage/mongodb"
	"github.com/craigpastro/crudapp/internal/storage/postgres"
	"github.com/craigpastro/crudapp/internal/telemetry"
	"github.com/craigpastro/crudapp/server"
	"github.com/sethvargo/go-envconfig"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type config struct {
	ServiceName        string `env:"SERVICE_NAME,default=crupapp"`
	ServiceVersion     string `env:"SERVICE_VERSION,default=0.1.0"`
	ServiceEnvironment string `env:"SERVICE_ENVIRONMENT,default=dev"`

	Addr string `env:"ADDR,default=localhost:8080"`

	LogFormat string `env:"LOG_FORMAT,default=console"`

	TraceEnabled     bool   `env:"TRACE_ENABLED,default=false"`
	TraceProviderURL string `env:"TRACE_PROVIDER_URL,default=localhost:4317"`

	StorageType string `env:"STORAGE_TYPE,default=memory"` // memory, mongodb, postgres
	MongoDBURL  string `env:"MONGODB_URL,default=mongodb://mongodb:password@127.0.0.1:27017"`
	PostgresURL string `env:"POSTGRES_URL,default=postgres://postgres:password@127.0.0.1:5432/postgres"`

	CacheType        string `env:"CACHE_TYPE,default=memory"` // memory, memcached, redis
	CacheSize        int    `env:"CACHE_SIZE,default=10000"`
	MemcachedServers string `env:"MEMCACHED_SERVERS,default=localhost:11211"`
	RedisAddr        string `env:"REDIS_ADDR,default=localhost:6379"`
	RedisPassword    string `env:"REDIS_PASSWORD,default="`
}

func main() {
	var cfg config
	if err := envconfig.Process(context.Background(), &cfg); err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if err := run(ctx, &cfg); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, cfg *config) error {
	logger := newLogger(cfg)

	tracer, err := telemetry.NewTracer(ctx, telemetry.TracerConfig{
		Enabled:        cfg.TraceEnabled,
		ServiceName:    cfg.ServiceName,
		ServiceVersion: cfg.ServiceVersion,
		Environment:    cfg.ServiceEnvironment,
		Endpoint:       cfg.TraceProviderURL,
	})
	if err != nil {
		logger.Fatal("error initializing tracer", zap.Error(err))
	}

	cache, err := newCache(ctx, logger, tracer, cfg)
	if err != nil {
		logger.Fatal("error initializing cache", zap.Error(err))
	}

	storage, err := newStorage(ctx, logger, tracer, cfg)
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

	logger.Info(fmt.Sprintf("server starting on %s (storage type=%s)", cfg.Addr, cfg.StorageType))

	return http.ListenAndServe(
		":8080",
		h2c.NewHandler(mux, &http2.Server{}),
	)
}

func newLogger(cfg *config) *zap.Logger {
	zapCfg := zap.NewProductionConfig()
	if cfg.LogFormat == "console" {
		zapCfg.Encoding = "console"
	}
	return zap.Must(zapCfg.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.PanicLevel),
		zap.Fields(
			zap.String("serviceName", cfg.ServiceName),
			zap.String("serviceVersion", cfg.ServiceVersion),
			zap.String("environment", cfg.ServiceEnvironment),
		),
	))
}

func newCache(ctx context.Context, logger *zap.Logger, tracer trace.Tracer, cfg *config) (cache.Cache, error) {
	switch cfg.CacheType {
	case "memcached":
		client, err := memcached.CreateClient(cfg.MemcachedServers, logger)
		if err != nil {
			return nil, err
		}
		return memcached.New(client, tracer), nil
	case "memory":
		return cachememory.New(cfg.CacheSize, tracer)
	case "redis":
		client, err := redis.CreateClient(ctx, cfg.RedisAddr, cfg.RedisPassword, logger)
		if err != nil {
			return nil, err
		}
		return redis.New(client, tracer), nil
	case "noop":
		return cache.NewNoopCache(), nil
	default:
		return nil, fmt.Errorf("undefined cache kind: %s", cfg.CacheType)
	}
}

func newStorage(ctx context.Context, logger *zap.Logger, tracer trace.Tracer, cfg *config) (storage.Storage, error) {
	switch cfg.StorageType {
	case "memory":
		return memory.New(tracer), nil
	case "mongodb":
		coll, err := mongodb.CreateCollection(ctx, cfg.MongoDBURL, logger)
		if err != nil {
			return nil, err
		}
		return mongodb.New(coll, tracer), nil
	case "postgres":
		pool, err := postgres.CreatePool(ctx, cfg.PostgresURL, logger)
		if err != nil {
			return nil, err
		}
		return postgres.New(pool, tracer), nil
	default:
		return nil, fmt.Errorf("undefined storage kind: %s", cfg.StorageType)
	}
}
