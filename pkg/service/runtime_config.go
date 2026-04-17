package service

import (
	"sync"

	"github.com/bsonger/devflow-release-service/pkg/model"
)

type RuntimeConfig struct {
	ImageRegistry           model.ImageRegistryConfig
	ManifestRegistry        model.ManifestRegistryConfig
	ManifestRegistryEnabled bool
	Downstream              model.DownstreamConfig
}

var (
	runtimeConfigMu sync.RWMutex
	runtimeConfig   RuntimeConfig
)

func ConfigureRuntimeConfig(cfg RuntimeConfig) {
	runtimeConfigMu.Lock()
	defer runtimeConfigMu.Unlock()
	runtimeConfig = cfg
}

func CurrentRuntimeConfig() RuntimeConfig {
	runtimeConfigMu.RLock()
	defer runtimeConfigMu.RUnlock()
	return runtimeConfig
}
