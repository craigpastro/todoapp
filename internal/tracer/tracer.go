package tracer

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

var name string

type TracerConfig struct {
	Enabled        bool
	ServiceName    string
	ServiceVersion string
	Environment    string
	Endpoint       string
}

func MustNewTracerProvider(ctx context.Context, config TracerConfig) *sdktrace.TracerProvider {
	name = config.ServiceName

	client := otlptracegrpc.NewClient(
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(config.Endpoint),
		otlptracegrpc.WithDialOption(grpc.WithBlock()),
	)

	exp, err := otlptrace.New(ctx, client)
	if err != nil {
		panic(err)
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
		panic(err)
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(0.5)),
		sdktrace.WithResource(resource),
		sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(exp)),
	)
}

func Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return otel.Tracer(name).Start(ctx, spanName, opts...)
}

func TraceError(span trace.Span, err error) {
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}
