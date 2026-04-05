package service

import (
	"context"
	"testing"

	"github.com/bsonger/devflow-release-service/pkg/model"
	"github.com/google/uuid"
)

func TestPopulateReleaseDefaultsPreservesProvidedEnv(t *testing.T) {
	manifestID := uuid.New()
	appID := uuid.New()
	release := &model.Release{ManifestID: manifestID, Env: "staging"}
	manifest := &model.Manifest{BaseModel: model.BaseModel{ID: manifestID}, Name: "demo-main", ApplicationID: appID}

	populateReleaseDefaults(release, manifest, "prod")

	if release.Env != "staging" {
		t.Fatalf("got env %s want staging", release.Env)
	}
}

func TestPopulateReleaseDefaultsFallsBackToProd(t *testing.T) {
	manifestID := uuid.New()
	appID := uuid.New()
	release := &model.Release{ManifestID: manifestID}
	manifest := &model.Manifest{BaseModel: model.BaseModel{ID: manifestID}, Name: "demo-main", ApplicationID: appID}

	populateReleaseDefaults(release, manifest, "prod")

	if release.Env != "prod" {
		t.Fatalf("got env %s want prod", release.Env)
	}
	if release.Type != model.ReleaseUpgrade {
		t.Fatalf("got type %s want %s", release.Type, model.ReleaseUpgrade)
	}
}

func TestResolveReleaseEnvironmentRequiresManifestRuntimeSpecRevisionID(t *testing.T) {
	svc := &releaseService{}
	manifest := &model.Manifest{ApplicationID: uuid.New()}
	release := &model.Release{
		ManifestID: uuid.New(),
		Env:        "staging",
	}

	_, err := svc.resolveReleaseEnvironment(context.Background(), release, manifest)
	if err == nil {
		t.Fatalf("expected error when manifest runtime_spec_revision_id is missing")
	}
}
