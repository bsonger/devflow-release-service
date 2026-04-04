package telemetry

import (
	"context"
	"os"

	"github.com/bsonger/devflow-common/client/logging"
	devflowOtel "github.com/bsonger/devflow-common/client/otel"
	"github.com/bsonger/devflow-common/client/pyroscope"
	commonModel "github.com/bsonger/devflow-common/model"
	"github.com/bsonger/devflow-release-service/pkg/model"
	"go.opentelemetry.io/otel/trace"
)

func Init(ctx context.Context, logCfg *model.LogConfig, otelCfg *model.OtelConfig, pyroscopeAddr, serviceName string) (func(context.Context) error, error) {
	logging.InitZapLogger(toCommonLogConfig(logCfg))

	resolvedServiceName := resolveServiceName(otelCfg, serviceName)
	if resolvedServiceName != "" {
		_ = os.Setenv("SERVICE_NAME", resolvedServiceName)
	}

	shutdown := func(context.Context) error { return nil }
	if commonOtelCfg := toCommonOtelConfig(otelCfg, resolvedServiceName); commonOtelCfg != nil {
		tpShutdown, err := devflowOtel.InitOtel(ctx, commonOtelCfg)
		if err != nil {
			return nil, err
		}
		shutdown = tpShutdown
	}

	if pyroscopeAddr != "" {
		pyroscope.InitPyroscope(resolvedServiceName, pyroscopeAddr)
	}

	if err := devflowOtel.InitMetricProvider(); err != nil {
		return shutdown, err
	}

	return shutdown, nil
}

func ReinjectLogger(ctx context.Context) context.Context {
	return logging.InjectLogger(ctx, logging.LoggerFromContext(ctx))
}

func StartSpan(ctx context.Context, tracer trace.Tracer, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	ctx, span := tracer.Start(ctx, spanName, opts...)
	return ReinjectLogger(ctx), span
}

func ResolveServiceName(otelCfg *model.OtelConfig, override string) string {
	return resolveServiceName(otelCfg, override)
}

func resolveServiceName(otelCfg *model.OtelConfig, override string) string {
	if override != "" {
		return override
	}
	if otelCfg != nil && otelCfg.ServiceName != "" {
		return otelCfg.ServiceName
	}
	return "devflow"
}

func toCommonLogConfig(cfg *model.LogConfig) *commonModel.LogConfig {
	if cfg == nil {
		return nil
	}
	return &commonModel.LogConfig{
		Level:  cfg.Level,
		Format: cfg.Format,
	}
}

func toCommonOtelConfig(cfg *model.OtelConfig, serviceName string) *commonModel.OtelConfig {
	if cfg == nil {
		return nil
	}
	return &commonModel.OtelConfig{
		Endpoint:    cfg.Endpoint,
		ServiceName: resolveServiceName(cfg, serviceName),
	}
}
