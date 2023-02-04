package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bufbuild/connect-go"
	grpcreflect "github.com/bufbuild/connect-grpcreflect-go"
	otelconnect "github.com/bufbuild/connect-opentelemetry-go"
	"github.com/craigpastro/crudapp/internal/gen/crudapp/v1/crudappv1connect"
	"github.com/craigpastro/crudapp/internal/instrumentation"
	"github.com/craigpastro/crudapp/internal/middleware"
	"github.com/craigpastro/crudapp/internal/server"
	"github.com/craigpastro/crudapp/internal/storage"
	"github.com/craigpastro/crudapp/internal/storage/memory"
	"github.com/craigpastro/crudapp/internal/storage/postgres"
	"github.com/sethvargo/go-envconfig"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
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

	run(context.Background(), &cfg)
}

// run runs the server. It takes a context. You may cancel the context to
// gracefully shutdown the server.
func run(ctx context.Context, cfg *config) {
	logr := newLogger(cfg)

	tp := sdktrace.NewTracerProvider()
	if cfg.TraceEnabled {
		tp = instrumentation.MustNewTracerProvider(instrumentation.TracerConfig{
			ServiceName:    cfg.ServiceName,
			ServiceVersion: cfg.ServiceVersion,
			Environment:    cfg.ServiceEnvironment,
			Endpoint:       cfg.TraceProviderURL,
		})
	}

	storage := mustNewStorage(ctx, logr, cfg)

	interceptors := connect.WithInterceptors(
		otelconnect.NewInterceptor(),
		middleware.NewLoggingInterceptor(logr),
	)

	mux := http.NewServeMux()
	reflector := grpcreflect.NewStaticReflector(crudappv1connect.CrudAppServiceName)
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))
	mux.Handle(crudappv1connect.NewCrudAppServiceHandler(
		server.NewServer(storage),
		interceptors,
	))

	srv := &http.Server{
		Addr:              cfg.Addr,
		ReadHeaderTimeout: 3 * time.Second,
		Handler:           h2c.NewHandler(mux, &http2.Server{}),
	}

	go func() {
		logr.Info(fmt.Sprintf("app starting on %s", cfg.Addr))

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("failed to start app", err)
			os.Exit(1)
		}
	}()

	// Shutdown stuff
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-done:
	case <-ctx.Done():
	}
	logr.Info("app attempting to shutdown gracefully")

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logr.Error("app shutdown failed", zap.Error(err))
		os.Exit(1)
	}

	_ = tp.ForceFlush(ctx)
	_ = tp.Shutdown(ctx)

	logr.Info("app shutdown gracefully. bye 👋")
}

func newLogger(cfg *config) *zap.Logger {
	zapCfg := zap.NewDevelopmentConfig()
	if cfg.LogFormat == "json" {
		zapCfg.Encoding = "json"
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

func mustNewStorage(ctx context.Context, logger *zap.Logger, cfg *config) storage.Storage {
	switch cfg.StorageType {
	case "memory":
		return memory.New()
	case "postgres":
		return postgres.MustNew(ctx, cfg.PostgresURL, logger)
	default:
		panic(fmt.Sprintf("undefined storage type: %s", cfg.StorageType))
	}
}
