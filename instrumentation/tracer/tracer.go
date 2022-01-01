package tracer

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

func New() trace.Tracer {
	return otel.Tracer("crudapp")
}
