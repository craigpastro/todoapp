package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

type TracerConfig struct {
	Enabled        bool
	ServiceName    string
	ServiceVersion string
	Environment    string
	Endpoint       string
}

func NewNoopTracer() trace.Tracer {
	return trace.NewNoopTracerProvider().Tracer("noop")
}

func NewTracer(ctx context.Context, config TracerConfig) (trace.Tracer, error) {
	if !config.Enabled {
		return NewNoopTracer(), nil
	}

	client := otlptracegrpc.NewClient(
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(config.Endpoint),
		otlptracegrpc.WithDialOption(grpc.WithBlock()),
	)

	exp, err := otlptrace.New(ctx, client)
	if err != nil {
		return nil, err
	}

	resource, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(config.ServiceName),
			semconv.ServiceVersionKey.String(config.ServiceVersion),
			semconv.DeploymentEnvironmentKey.String(config.Environment),
		),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(resource),
		sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(exp)),
	)
	otel.SetTracerProvider(tp)

	return tp.Tracer(config.ServiceName), nil
}

func TraceError(span trace.Span, err error) {
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}
