package main

import (
	_ "github.com/bsonger/devflow-release-service/docs/generated/swagger"
	"github.com/bsonger/devflow-release-service/pkg/config"
	"github.com/bsonger/devflow-release-service/pkg/router"
	"github.com/bsonger/devflow-release-service/pkg/runtime"
	"github.com/bsonger/devflow-service-common/bootstrap"
	"github.com/bsonger/devflow-service-common/observability"
)

func main() {
	err := bootstrap.Run(bootstrap.Options[config.Config, router.Options, runtime.ExecutionMode]{
		Name:        "release-service",
		Load:        config.Load,
		InitRuntime: config.InitRuntime,
		NewRouter: func(opts router.Options) bootstrap.Runner {
			return router.NewRouterWithOptions(opts)
		},
		SetExecutionMode: runtime.SetExecutionMode,
		ResolveConfigPort: func(cfg *config.Config) int {
			if cfg != nil && cfg.Server != nil {
				return cfg.Server.Port
			}
			return 0
		},
		StartMetricsServer: observability.StartMetricsServer,
		StartPprofServer:   observability.StartPprofServer,
		ExecutionMode:      runtime.ExecutionModeDirect,
		RouteOptions: router.Options{
			ServiceName:   "release-service",
			EnableSwagger: true,
			Modules: []router.Module{
				router.ModuleManifest,
				router.ModuleImage,
				router.ModuleRelease,
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
