package service

import (
	"testing"

	"github.com/bsonger/devflow-release-service/pkg/downstream"
	"github.com/bsonger/devflow-release-service/pkg/model"
	"github.com/google/uuid"
)

func TestBuildManifestPrefersDigestAndRendersObjects(t *testing.T) {
	req := &model.CreateManifestRequest{
		ApplicationID: mustUUID("11111111-1111-1111-1111-111111111111"),
		EnvironmentID: mustUUID("22222222-2222-2222-2222-222222222222"),
		ImageID:       mustUUID("33333333-3333-3333-3333-333333333333"),
	}
	image := &model.Image{
		ApplicationID: req.ApplicationID,
		Name:          "demo-api",
		RepoAddress:   "registry.cn-hangzhou.aliyuncs.com/devflow",
		Tag:           "20260411-120000",
		Digest:        "sha256:abc",
	}
	appConfig := &downstream.AppConfig{
		ID:                "cfg-1",
		Name:              "demo-config",
		RenderedConfigMap: map[string]string{"app.yaml": "foo: bar"},
		Files:             []downstream.ManifestFile{{Name: "app.yaml", Content: "foo: bar"}},
	}
	workload := &downstream.WorkloadConfig{
		ID:           "wc-1",
		Name:         "demo-workload",
		Replicas:     2,
		WorkloadType: "deployment",
		Strategy:     "rolling-update",
	}
	services := []downstream.ManifestService{{
		ID:   "svc-1",
		Name: "demo-api",
		Ports: []downstream.ManifestServicePort{{
			Name:        "http",
			ServicePort: 80,
			TargetPort:  8080,
			Protocol:    "TCP",
		}},
	}}
	routes := []downstream.ManifestRoute{{
		ID:          "route-1",
		Name:        "web",
		Host:        "demo.example.com",
		Path:        "/",
		ServiceName: "demo-api",
		ServicePort: 80,
	}}

	got, err := buildManifest(req, image, "demo-api", appConfig, workload, services, routes, "staging")
	if err != nil {
		t.Fatal(err)
	}
	if got.ImageRef != "registry.cn-hangzhou.aliyuncs.com/devflow/demo-api@sha256:abc" {
		t.Fatalf("unexpected image ref %q", got.ImageRef)
	}
	if len(got.RenderedObjects) != 4 {
		t.Fatalf("unexpected rendered object count %d", len(got.RenderedObjects))
	}
	if got.RenderedYAML == "" {
		t.Fatal("expected rendered yaml")
	}
}

func TestBuildManifestRejectsInvalidRouteTarget(t *testing.T) {
	req := &model.CreateManifestRequest{
		ApplicationID: uuid.New(),
		EnvironmentID: uuid.New(),
		ImageID:       uuid.New(),
	}
	image := &model.Image{
		ApplicationID: req.ApplicationID,
		Name:          "demo-api",
		RepoAddress:   "registry.cn-hangzhou.aliyuncs.com/devflow",
		Tag:           "20260411-120000",
	}
	appConfig := &downstream.AppConfig{RenderedConfigMap: map[string]string{"app.yaml": "foo: bar"}}
	workload := &downstream.WorkloadConfig{Replicas: 1, WorkloadType: "deployment"}
	_, err := buildManifest(req, image, "demo-api", appConfig, workload, nil, []downstream.ManifestRoute{{
		Name:        "bad",
		Host:        "demo.example.com",
		Path:        "/",
		ServiceName: "missing",
		ServicePort: 80,
	}}, "staging")
	if err == nil {
		t.Fatal("expected error")
	}
}

func mustUUID(value string) uuid.UUID {
	id, err := uuid.Parse(value)
	if err != nil {
		panic(err)
	}
	return id
}
