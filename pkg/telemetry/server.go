package telemetry

import "github.com/bsonger/devflow-service-common/observability"

func StartMetricsServer(addr string) {
	observability.StartMetricsServer(addr)
}

func StartPprofServer(addr string) {
	observability.StartPprofServer(addr)
}
