package config

import (
	"testing"

	"github.com/bsonger/devflow-release-service/pkg/model"
)

func TestResolveObservabilityServiceName(t *testing.T) {
	cfg := &model.OtelConfig{ServiceName: "otel-service"}

	if got := resolveObservabilityServiceName(cfg, "runtime-override"); got != "runtime-override" {
		t.Fatalf("got %q want runtime-override", got)
	}
	if got := resolveObservabilityServiceName(cfg, ""); got != "otel-service" {
		t.Fatalf("got %q want otel-service", got)
	}
	if got := resolveObservabilityServiceName(nil, ""); got != "devflow" {
		t.Fatalf("got %q want devflow", got)
	}
}
