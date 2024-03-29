package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/bufbuild/connect-go"
	grpcreflect "github.com/bufbuild/connect-grpcreflect-go"
	otelconnect "github.com/bufbuild/connect-opentelemetry-go"
	"github.com/craigpastro/todoapp/internal/gen/sqlc"
	"github.com/craigpastro/todoapp/internal/gen/todoapp/v1/todoappv1connect"
	"github.com/craigpastro/todoapp/internal/instrumentation"
	"github.com/craigpastro/todoapp/internal/middleware"
	"github.com/craigpastro/todoapp/internal/postgres"
	"github.com/craigpastro/todoapp/internal/server"
	"github.com/sethvargo/go-envconfig"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type config struct {
	ServiceName        string `env:"SERVICE_NAME,default=crupapp"`
	ServiceVersion     string `env:"SERVICE_VERSION,default=0.1.0"`
	ServiceEnvironment string `env:"SERVICE_ENVIRONMENT,default=dev"`

	Port int `env:"PORT,default=8080"`

	JWTSecret string `env:"JWT_SECRET,default=PMBrjiOH5RMo6nQHidA62XctWGxDG0rw"`

	LogFormat string `env:"LOG_FORMAT,default=console"`

	TraceEnabled     bool    `env:"TRACE_ENABLED,default=false"`
	TraceProviderURL string  `env:"TRACE_PROVIDER_URL,default=localhost:4317"`
	TraceSampleRatio float64 `env:"TRACE_SAMPLE_RATIO,default=1"`

	PostgresConnString        string `env:"POSTGRES_CONN_STRING,default=postgres://authenticator:password@127.0.0.1:5432/postgres"`
	PostgresAutoMigrate       bool   `env:"POSTGRES_AUTOMIGRATE,default=true"`
	PostgresMigrateConnString string `env:"POSTGRES_MIGRATE_CONN_STRING,default=postgres://postgres:password@127.0.0.1:5432/postgres"`
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
	handler := slog.NewTextHandler(os.Stdout, nil)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	tpShutdown := instrumentation.MustNewTracerProvider(
		cfg.TraceEnabled,
		instrumentation.TracerConfig{
			ServiceName:    cfg.ServiceName,
			ServiceVersion: cfg.ServiceVersion,
			Environment:    cfg.ServiceEnvironment,
			Endpoint:       cfg.TraceProviderURL,
			SampleRatio:    cfg.TraceSampleRatio,
		},
	)

	pool := postgres.MustNew(&postgres.Config{
		ConnString:        cfg.PostgresConnString,
		Migrate:           cfg.PostgresAutoMigrate,
		MigrateConnString: cfg.PostgresMigrateConnString,
	})
	queries := sqlc.New(pool)

	interceptors := connect.WithInterceptors(
		middleware.NewLoggingInterceptor(),
		otelconnect.NewInterceptor(),
		middleware.NewValidatorInterceptor(),
		middleware.NewAuthenticationInterceptor(pool, cfg.JWTSecret),
	)

	mux := http.NewServeMux()
	reflector := grpcreflect.NewStaticReflector(todoappv1connect.TodoAppServiceName)
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))
	mux.Handle(todoappv1connect.NewTodoAppServiceHandler(
		server.NewServer(queries),
		interceptors,
	))

	srv := &http.Server{
		Addr:              fmt.Sprintf("0.0.0.0:%d", cfg.Port),
		ReadHeaderTimeout: 3 * time.Second,
		Handler:           h2c.NewHandler(mux, &http2.Server{}),
	}

	go func() {
		slog.Info(fmt.Sprintf("todoapp starting on ':%d'", cfg.Port))

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("failed to start todoapp", err)
		}
	}()

	// Wait for shutdown
	ctx, _ = signal.NotifyContext(ctx, os.Interrupt)
	<-ctx.Done()

	slog.Info("todoapp attempting to shutdown gracefully")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("todoapp shutdown failed", err)
	}

	_ = tpShutdown(ctx)
	pool.Close()

	slog.Info("todoapp shutdown gracefully. bye 👋")
}
