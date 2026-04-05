package telemetry

import (
	"context"
	"testing"

	"github.com/bsonger/devflow-release-service/pkg/model"
	"github.com/bsonger/devflow-service-common/loggingx"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestResolveServiceNamePrefersOverride(t *testing.T) {
	name := ResolveServiceName(&model.OtelConfig{ServiceName: "from-config"}, "release-service")
	if name != "release-service" {
		t.Fatalf("unexpected service name: got %q want %q", name, "release-service")
	}
}

func TestResolveServiceNameFallsBackToConfig(t *testing.T) {
	name := ResolveServiceName(&model.OtelConfig{ServiceName: "verify-service"}, "")
	if name != "verify-service" {
		t.Fatalf("unexpected service name: got %q want %q", name, "verify-service")
	}
}

func TestResolveServiceNameDefaultsToDevflow(t *testing.T) {
	name := ResolveServiceName(nil, "")
	if name != "devflow" {
		t.Fatalf("unexpected service name: got %q want %q", name, "devflow")
	}
}

func TestStartSpanReinjectsLoggerWithTraceFields(t *testing.T) {
	core, observed := observer.New(zapcore.InfoLevel)
	base := zap.New(core)

	ctx := loggingx.InjectLogger(context.Background(), base)

	tp := sdktrace.NewTracerProvider()
	t.Cleanup(func() {
		_ = tp.Shutdown(context.Background())
	})

	ctx, span := StartSpan(ctx, tp.Tracer("test"), "test-span")
	defer span.End()

	loggingx.LoggerFromContext(ctx).Info("inside-span")

	entries := observed.AllUntimed()
	if len(entries) != 1 {
		t.Fatalf("unexpected log count: got %d want %d", len(entries), 1)
	}

	fields := entries[0].ContextMap()
	traceID, ok := fields["trace_id"].(string)
	if !ok || traceID == "" {
		t.Fatal("trace_id was not injected into logger")
	}
	spanID, ok := fields["span_id"].(string)
	if !ok || spanID == "" {
		t.Fatal("span_id was not injected into logger")
	}

	if traceID == "00000000000000000000000000000000" {
		t.Fatal("trace_id should not be zero")
	}
}
