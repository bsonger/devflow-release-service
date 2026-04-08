package service

import (
	"context"
	"testing"

	"github.com/bsonger/devflow-release-service/pkg/model"
	"github.com/google/uuid"
)

func TestPopulateReleaseDefaultsPreservesProvidedEnv(t *testing.T) {
	imageID := uuid.New()
	appID := uuid.New()
	release := &model.Release{ImageID: imageID, Env: "staging"}
	image := &model.Image{BaseModel: model.BaseModel{ID: imageID}, Name: "demo-main", ApplicationID: appID}

	populateReleaseDefaults(release, image, "prod")

	if release.Env != "staging" {
		t.Fatalf("got env %s want staging", release.Env)
	}
}

func TestPopulateReleaseDefaultsFallsBackToProd(t *testing.T) {
	imageID := uuid.New()
	appID := uuid.New()
	release := &model.Release{ImageID: imageID}
	image := &model.Image{BaseModel: model.BaseModel{ID: imageID}, Name: "demo-main", ApplicationID: appID}

	populateReleaseDefaults(release, image, "prod")

	if release.Env != "prod" {
		t.Fatalf("got env %s want prod", release.Env)
	}
	if release.Type != model.ReleaseUpgrade {
		t.Fatalf("got type %s want %s", release.Type, model.ReleaseUpgrade)
	}
}

func TestResolveReleaseEnvironmentRequiresImageRuntimeSpecRevisionID(t *testing.T) {
	svc := &releaseService{}
	image := &model.Image{ApplicationID: uuid.New()}
	release := &model.Release{
		ImageID: uuid.New(),
		Env:     "staging",
	}

	_, err := svc.resolveReleaseEnvironment(context.Background(), release, image)
	if err == nil {
		t.Fatalf("expected error when image runtime_spec_revision_id is missing")
	}
}
