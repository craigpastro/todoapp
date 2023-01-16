package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/bufbuild/connect-go"
	grpcreflect "github.com/bufbuild/connect-grpcreflect-go"
	"github.com/craigpastro/crudapp/internal/gen/crudapp/v1/crudappv1connect"
	"github.com/craigpastro/crudapp/internal/middleware"
	"github.com/craigpastro/crudapp/internal/server"
	"github.com/craigpastro/crudapp/internal/storage"
	"github.com/craigpastro/crudapp/internal/storage/memory"
	"github.com/craigpastro/crudapp/internal/storage/postgres"
	"github.com/craigpastro/crudapp/internal/tracer"
	"github.com/sethvargo/go-envconfig"
	"go.opentelemetry.io/otel"
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

	StorageType string `env:"STORAGE_TYPE,default=memory"` // memory, postgres
	PostgresURL string `env:"POSTGRES_URL,default=postgres://postgres:password@127.0.0.1:5432/postgres"`
	CacheSize   int    `env:"CACHE_SIZE,default=10000"`
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

	var tp trace.TracerProvider
	if cfg.TraceEnabled {
		sdktp := tracer.MustNewTracerProvider(ctx, tracer.TracerConfig{
			ServiceName:    cfg.ServiceName,
			ServiceVersion: cfg.ServiceVersion,
			Environment:    cfg.ServiceEnvironment,
			Endpoint:       cfg.TraceProviderURL,
		})
		defer sdktp.Shutdown(context.Background())
		tp = sdktp
	} else {
		tp = trace.NewNoopTracerProvider()
	}
	otel.SetTracerProvider(tp)

	storage, err := newStorage(ctx, logger, cfg)
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
		server.NewServer(storage),
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

func newStorage(ctx context.Context, logger *zap.Logger, cfg *config) (storage.Storage, error) {
	var s storage.Storage
	switch cfg.StorageType {
	case "memory":
		s = memory.New()
	case "postgres":
		db, err := postgres.CreateDB(ctx, cfg.PostgresURL, logger)
		if err != nil {
			return nil, err
		}
		s = postgres.New(db)
	default:
		return nil, fmt.Errorf("undefined storage kind: %s", cfg.StorageType)
	}

	return storage.NewCachingStorage(s, cfg.CacheSize)
}
