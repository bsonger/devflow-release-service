package telemetry

import (
	"context"

	"github.com/bsonger/devflow-release-service/pkg/model"
	"github.com/bsonger/devflow-service-common/observability"
	"go.opentelemetry.io/otel/trace"
)

func Init(ctx context.Context, logCfg *model.LogConfig, otelCfg *model.OtelConfig, pyroscopeAddr, serviceName string) (func(context.Context) error, error) {
	opts := observability.RuntimeOptions{
		LogLevel:        "",
		LogFormat:       "",
		OtelEndpoint:    "",
		OtelService:     ResolveServiceName(otelCfg, serviceName),
		PyroscopeAddr:   pyroscopeAddr,
		ServiceOverride: serviceName,
	}
	if logCfg != nil {
		opts.LogLevel = logCfg.Level
		opts.LogFormat = logCfg.Format
	}
	if otelCfg != nil {
		opts.OtelEndpoint = otelCfg.Endpoint
	}
	return observability.Init(ctx, opts)
}

func ReinjectLogger(ctx context.Context) context.Context {
	return observability.ReinjectLogger(ctx)
}

func StartSpan(ctx context.Context, tracer trace.Tracer, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return observability.StartSpan(ctx, tracer, spanName, opts...)
}

func ResolveServiceName(otelCfg *model.OtelConfig, override string) string {
	if override != "" {
		return override
	}
	if otelCfg != nil && otelCfg.ServiceName != "" {
		return otelCfg.ServiceName
	}
	return "devflow"
}
