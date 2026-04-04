package service

import (
	"context"

	"github.com/bsonger/devflow-service-common/observability"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

const (
	ServiceName = "devflow"
)

var devflowTracer = otel.Tracer(ServiceName)

func StartServiceSpan(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return observability.StartSpan(ctx, devflowTracer, spanName, opts...)
}

func StartWorkerSpan(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return observability.StartSpan(ctx, otel.Tracer("release-worker"), spanName, opts...)
}
