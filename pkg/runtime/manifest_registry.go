package runtime

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/bsonger/devflow-release-service/pkg/model"
)

func ManifestRegistryConfigFromEnv() (model.ManifestRegistryConfig, bool, error) {
	cfg := model.ManifestRegistryConfig{
		Registry:   firstNonEmptyEnv("MANIFEST_REGISTRY", "IMAGE_REGISTRY"),
		Namespace:  firstNonEmptyEnv("MANIFEST_REGISTRY_NAMESPACE", "IMAGE_REGISTRY_NAMESPACE"),
		Repository: firstNonEmptyEnv("MANIFEST_REGISTRY_REPOSITORY", "MANIFEST_OCI_REPOSITORY"),
		Username:   firstNonEmptyEnv("MANIFEST_REGISTRY_USERNAME", "IMAGE_REGISTRY_USERNAME"),
		Password:   firstNonEmptyEnv("MANIFEST_REGISTRY_PASSWORD", "IMAGE_REGISTRY_PASSWORD"),
		PlainHTTP:  firstBoolEnv("MANIFEST_REGISTRY_PLAIN_HTTP", "MANIFEST_OCI_PLAIN_HTTP", "IMAGE_REGISTRY_PLAIN_HTTP"),
	}
	if cfg.Repository == "" {
		cfg.Repository = "manifests"
	}
	if cfg.Registry == "" && cfg.Namespace == "" {
		return model.ManifestRegistryConfig{}, false, nil
	}
	if cfg.Registry == "" {
		return model.ManifestRegistryConfig{}, false, fmt.Errorf("manifest registry config missing registry")
	}
	if cfg.Namespace == "" {
		return model.ManifestRegistryConfig{}, false, fmt.Errorf("manifest registry config missing namespace")
	}
	return cfg, true, nil
}

func firstNonEmptyEnv(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}

func firstBoolEnv(keys ...string) bool {
	for _, key := range keys {
		value := strings.TrimSpace(os.Getenv(key))
		if value == "" {
			continue
		}
		parsed, err := strconv.ParseBool(value)
		if err == nil {
			return parsed
		}
	}
	return false
}
