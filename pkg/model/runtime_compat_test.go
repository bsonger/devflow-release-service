package model

import (
	"testing"
	"time"
)

func TestGeneratePipelineRunUsesProvidedPipelineName(t *testing.T) {
	image := &Image{}
	run := image.GeneratePipelineRun(
		"devflow-tekton-image-build",
		"pvc-1",
		ImageRegistryConfig{Registry: "registry.cn-hangzhou.aliyuncs.com", Namespace: "devflow"},
		ImageTarget{Name: "portal-api", Tag: "20260408-130500"},
	)
	if run.GenerateName != "devflow-tekton-image-build-run-" {
		t.Fatalf("GenerateName = %q", run.GenerateName)
	}
	if run.Spec.PipelineRef == nil || run.Spec.PipelineRef.Name != "devflow-tekton-image-build" {
		t.Fatalf("PipelineRef = %+v", run.Spec.PipelineRef)
	}
}

func TestGeneratePipelineRunCarriesImageIdentifierMetadata(t *testing.T) {
	image := &Image{}
	image.WithCreateDefault()
	run := image.GeneratePipelineRun(
		"devflow-tekton-image-build",
		"pvc-1",
		ImageRegistryConfig{Registry: "registry.cn-hangzhou.aliyuncs.com", Namespace: "devflow"},
		ImageTarget{Name: "portal-api", Tag: "20260408-130500"},
	)
	if got := run.Labels["devflow.image/id"]; got != image.ID.String() {
		t.Fatalf("label image id = %q want %q", got, image.ID.String())
	}
	if got := run.Annotations["devflow.image/id"]; got != image.ID.String() {
		t.Fatalf("annotation image id = %q want %q", got, image.ID.String())
	}
}

func TestGeneratePipelineRunParamsUsesConfigDrivenRegistryAndTarget(t *testing.T) {
	image := &Image{RepoAddress: "git@example.com/portal-api.git", Branch: "feature/login"}
	image.WithCreateDefault()
	target, err := BuildImageTarget(
		ImageRegistryConfig{Registry: "registry.cn-hangzhou.aliyuncs.com", Namespace: "devflow"},
		"Portal API",
		image.Branch,
		time.Date(2026, 4, 8, 13, 5, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("BuildImageTarget returned error: %v", err)
	}
	params := image.GeneratePipelineRunParams(
		ImageRegistryConfig{Registry: "registry.cn-hangzhou.aliyuncs.com", Namespace: "devflow"},
		target,
	)

	got := map[string]string{}
	for _, param := range params {
		got[param.Name] = param.Value.StringVal
	}
	if got["image-registry"] != "registry.cn-hangzhou.aliyuncs.com/devflow" {
		t.Fatalf("image-registry = %q", got["image-registry"])
	}
	if got["name"] != "portal-api-feature-login" {
		t.Fatalf("name = %q", got["name"])
	}
	if got["image-tag"] != "20260408-130500" {
		t.Fatalf("image-tag = %q", got["image-tag"])
	}
}
