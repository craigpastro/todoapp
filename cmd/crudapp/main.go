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

	Port int `env:"PORT,default=8080"`

	LogFormat string `env:"LOG_FORMAT,default=console"`

	TraceEnabled     bool   `env:"TRACE_ENABLED,default=false"`
	TraceProviderURL string `env:"TRACE_PROVIDER_URL,default=localhost:4317"`

	PostgresConnString  string `env:"POSTGRES_CONN_STRING,default=postgres://postgres:password@127.0.0.1:5432/postgres"`
	PostgresAutoMigrate bool   `env:"POSTGRES_AUTOMIGRATE,default=true"`
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

	db := postgres.MustNew(cfg.PostgresConnString, cfg.PostgresAutoMigrate)

	interceptors := connect.WithInterceptors(
		otelconnect.NewInterceptor(),
		middleware.NewLoggingInterceptor(logr),
	)

	mux := http.NewServeMux()
	reflector := grpcreflect.NewStaticReflector(crudappv1connect.CrudAppServiceName)
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))
	mux.Handle(crudappv1connect.NewCrudAppServiceHandler(
		server.NewServer(db),
		interceptors,
	))

	srv := &http.Server{
		Addr:              fmt.Sprintf("localhost:%d", cfg.Port),
		ReadHeaderTimeout: 3 * time.Second,
		Handler:           h2c.NewHandler(mux, &http2.Server{}),
	}

	go func() {
		logr.Info(fmt.Sprintf("crudapp starting on :%d", cfg.Port))

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("failed to start crudapp", err)
			os.Exit(1)
		}
	}()

	// Wait for shutdown
	ctx, _ = signal.NotifyContext(ctx, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-ctx.Done()

	logr.Info("crudapp attempting to shutdown gracefully")

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logr.Error("crudapp shutdown failed", zap.Error(err))
		os.Exit(1)
	}

	_ = tp.ForceFlush(ctx)
	_ = tp.Shutdown(ctx)

	logr.Info("crudapp shutdown gracefully. bye 👋")
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
