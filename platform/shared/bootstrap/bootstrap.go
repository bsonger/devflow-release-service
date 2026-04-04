package bootstrap

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/bsonger/devflow-common/client/logging"
	"github.com/bsonger/devflow-release-service/pkg/config"
	"github.com/bsonger/devflow-release-service/pkg/router"
	"github.com/bsonger/devflow-release-service/pkg/runtime"
	"github.com/bsonger/devflow-release-service/pkg/telemetry"
	"go.uber.org/zap"
)

type Options struct {
	Name           string
	RouteOptions   router.Options
	ExecutionMode  runtime.ExecutionMode
	PortEnv        string
	DefaultPort    int
	MetricsPortEnv string
	DefaultMetrics int
	PprofPortEnv   string
	DefaultPprof   int
}

func Run(opts Options) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	shutdown, err := config.InitRuntime(context.Background(), cfg, opts.Name)
	if err != nil {
		return err
	}
	defer func() {
		_ = shutdown(context.Background())
	}()

	if opts.ExecutionMode != "" {
		runtime.SetExecutionMode(opts.ExecutionMode)
	}

	metricsPort := resolvePort(opts.DefaultMetrics, opts.MetricsPortEnv)
	if metricsPort > 0 {
		telemetry.StartMetricsServer(fmt.Sprintf(":%d", metricsPort))
	}

	pprofPort := resolvePort(opts.DefaultPprof, opts.PprofPortEnv)
	if pprofPort > 0 {
		telemetry.StartPprofServer(fmt.Sprintf(":%d", pprofPort))
	}

	r := router.NewRouterWithOptions(opts.RouteOptions)
	port := resolveConfiguredPort(cfg, opts.DefaultPort, opts.PortEnv)

	logging.Logger.Info("starting service",
		zap.String("service", opts.Name),
		zap.Int("port", port),
		zap.Int("metrics_port", metricsPort),
		zap.Int("pprof_port", pprofPort),
	)

	return r.Run(fmt.Sprintf(":%d", port))
}

func resolveConfiguredPort(cfg *config.Config, defaultPort int, envKey string) int {
	port := defaultPort
	if cfg != nil && cfg.Server != nil && cfg.Server.Port > 0 {
		port = cfg.Server.Port
	}
	override := resolvePort(0, envKey)
	if override > 0 {
		port = override
	}
	return port
}

func resolvePort(defaultPort int, envKey string) int {
	if envKey == "" {
		return defaultPort
	}
	if value := os.Getenv(envKey); value != "" {
		port, err := strconv.Atoi(value)
		if err == nil && port > 0 {
			return port
		}
	}
	return defaultPort
}
