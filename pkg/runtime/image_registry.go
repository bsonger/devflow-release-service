package runtime

import (
	"fmt"
	"os"
	"strings"

	"github.com/bsonger/devflow-release-service/pkg/model"
)

func ImageRegistryConfigFromEnv() (model.ImageRegistryConfig, error) {
	cfg := model.ImageRegistryConfig{
		Registry:  strings.TrimSpace(os.Getenv("IMAGE_REGISTRY")),
		Namespace: strings.TrimSpace(os.Getenv("IMAGE_REGISTRY_NAMESPACE")),
		Username:  strings.TrimSpace(os.Getenv("IMAGE_REGISTRY_USERNAME")),
		Password:  strings.TrimSpace(os.Getenv("IMAGE_REGISTRY_PASSWORD")),
	}
	if cfg.Registry == "" {
		return model.ImageRegistryConfig{}, fmt.Errorf("IMAGE_REGISTRY is required")
	}
	if cfg.Namespace == "" {
		return model.ImageRegistryConfig{}, fmt.Errorf("IMAGE_REGISTRY_NAMESPACE is required")
	}
	return cfg, nil
}
