package runtime

import "testing"

func TestManifestRegistryConfigFromEnvFallsBackToImageRegistry(t *testing.T) {
	t.Setenv("IMAGE_REGISTRY", "registry.example.com")
	t.Setenv("IMAGE_REGISTRY_NAMESPACE", "devflow")
	t.Setenv("IMAGE_REGISTRY_USERNAME", "image-user")
	t.Setenv("IMAGE_REGISTRY_PASSWORD", "image-pass")
	t.Setenv("IMAGE_REGISTRY_PLAIN_HTTP", "true")

	cfg, enabled, err := ManifestRegistryConfigFromEnv()
	if err != nil {
		t.Fatalf("ManifestRegistryConfigFromEnv() error = %v", err)
	}
	if !enabled {
		t.Fatal("expected manifest registry publishing to be enabled")
	}
	if cfg.Registry != "registry.example.com" || cfg.Namespace != "devflow" {
		t.Fatalf("unexpected cfg %+v", cfg)
	}
	if cfg.Repository != "manifests" {
		t.Fatalf("cfg.Repository = %q, want manifests", cfg.Repository)
	}
	if cfg.Username != "image-user" || cfg.Password != "image-pass" {
		t.Fatalf("unexpected credentials %+v", cfg)
	}
	if !cfg.PlainHTTP {
		t.Fatalf("expected PlainHTTP to be true, got %+v", cfg)
	}
}

func TestManifestRegistryConfigFromEnvRejectsPartialConfig(t *testing.T) {
	t.Setenv("MANIFEST_REGISTRY", "registry.example.com")

	_, enabled, err := ManifestRegistryConfigFromEnv()
	if err == nil {
		t.Fatal("expected error")
	}
	if enabled {
		t.Fatal("expected publishing to stay disabled on partial config")
	}
}
