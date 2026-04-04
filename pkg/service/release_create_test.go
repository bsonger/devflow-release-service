package service

import (
	"testing"

	"github.com/bsonger/devflow-release-service/pkg/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestPopulateReleaseDefaultsPreservesProvidedEnv(t *testing.T) {
	manifestID := primitive.NewObjectID()
	appID := primitive.NewObjectID()
	cfgID := primitive.NewObjectID()
	release := &model.Release{ManifestID: manifestID, ConfigurationID: &cfgID, Env: "staging"}
	manifest := &model.Manifest{BaseModel: model.BaseModel{ID: manifestID}, Name: "demo-main", ApplicationId: appID}
	app := &model.Application{Name: "demo", ProjectName: "proj", Type: model.Normal}

	populateReleaseDefaults(release, manifest, app)

	if release.Env != "staging" {
		t.Fatalf("got env %s want staging", release.Env)
	}
	if release.ConfigurationID == nil || *release.ConfigurationID != cfgID {
		t.Fatalf("configuration_id was not preserved")
	}
}

func TestPopulateReleaseDefaultsFallsBackToProd(t *testing.T) {
	manifestID := primitive.NewObjectID()
	appID := primitive.NewObjectID()
	release := &model.Release{ManifestID: manifestID}
	manifest := &model.Manifest{BaseModel: model.BaseModel{ID: manifestID}, Name: "demo-main", ApplicationId: appID}
	app := &model.Application{Name: "demo", ProjectName: "proj", Type: model.Normal}

	populateReleaseDefaults(release, manifest, app)

	if release.Env != "prod" {
		t.Fatalf("got env %s want prod", release.Env)
	}
	if release.Type != model.ReleaseUpgrade {
		t.Fatalf("got type %s want %s", release.Type, model.ReleaseUpgrade)
	}
}
