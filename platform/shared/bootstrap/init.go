package bootstrap

import (
	"context"

	"github.com/bsonger/devflow-release-service/pkg/config"
	"github.com/bsonger/devflow-release-service/pkg/runtime"
)

func Init(serviceName string, executionMode runtime.ExecutionMode) (*config.Config, func(context.Context) error, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, err
	}

	shutdown, err := config.InitRuntime(context.Background(), cfg, serviceName)
	if err != nil {
		return nil, nil, err
	}

	if executionMode != "" {
		runtime.SetExecutionMode(executionMode)
	}

	return cfg, shutdown, nil
}
