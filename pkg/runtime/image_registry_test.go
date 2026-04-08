package runtime

import "testing"

func TestImageRegistryConfigFromEnvRequiresRegistryAndNamespace(t *testing.T) {
	t.Setenv("IMAGE_REGISTRY", "")
	t.Setenv("IMAGE_REGISTRY_NAMESPACE", "")

	_, err := ImageRegistryConfigFromEnv()
	if err == nil {
		t.Fatal("expected error when registry config is missing")
	}
}

func TestImageRegistryConfigFromEnvReadsValues(t *testing.T) {
	t.Setenv("IMAGE_REGISTRY", "registry.cn-hangzhou.aliyuncs.com")
	t.Setenv("IMAGE_REGISTRY_NAMESPACE", "devflow")
	t.Setenv("IMAGE_REGISTRY_USERNAME", "user")
	t.Setenv("IMAGE_REGISTRY_PASSWORD", "pass")

	cfg, err := ImageRegistryConfigFromEnv()
	if err != nil {
		t.Fatalf("ImageRegistryConfigFromEnv returned error: %v", err)
	}
	if cfg.Repository() != "registry.cn-hangzhou.aliyuncs.com/devflow" {
		t.Fatalf("Repository = %q", cfg.Repository())
	}
	if cfg.Username != "user" || cfg.Password != "pass" {
		t.Fatalf("unexpected credentials: %+v", cfg)
	}
}
