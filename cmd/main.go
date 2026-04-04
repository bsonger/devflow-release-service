package main

import (
	"github.com/bsonger/devflow-release-service/pkg/router"
	"github.com/bsonger/devflow-release-service/pkg/runtime"
	"github.com/bsonger/devflow-release-service/platform/shared/bootstrap"
)

func main() {
	err := bootstrap.Run(bootstrap.Options{
		Name:          "release-service",
		ExecutionMode: runtime.ExecutionModeIntent,
		RouteOptions: router.Options{
			ServiceName:   "release-service",
			EnableSwagger: true,
			Modules: []router.Module{
				router.ModuleManifest,
				router.ModuleJob,
				router.ModuleIntent,
			},
		},
		PortEnv:        "RELEASE_SERVICE_PORT",
		DefaultPort:    8083,
		MetricsPortEnv: "RELEASE_SERVICE_METRICS_PORT",
		PprofPortEnv:   "RELEASE_SERVICE_PPROF_PORT",
	})
	if err != nil {
		panic(err)
	}
}
