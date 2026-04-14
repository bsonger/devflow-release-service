package service

import (
	"context"
	"strings"
	"testing"

	appv1 "github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	"github.com/bsonger/devflow-release-service/pkg/model"
	"github.com/google/uuid"
)

func TestBuildArgoApplicationUsesManifestOCIArtifactWhenPresent(t *testing.T) {
	release := &model.Release{
		BaseModel:     model.BaseModel{ID: uuid.New()},
		ApplicationID: uuid.New(),
		ManifestID:    uuid.New(),
		ImageID:       uuid.New(),
		Env:           "staging",
	}
	manifest := &model.Manifest{
		BaseModel:          model.BaseModel{ID: release.ManifestID},
		ArtifactRepository: "zot.zot.svc.cluster.local:5000/devflow/manifests/demo-api/staging",
		ArtifactTag:        release.ManifestID.String(),
		ArtifactDigest:     "sha256:abc123",
	}

	app := buildArgoApplication(release, manifest, &applicationProjection{
		Name:        "demo-api",
		ProjectName: "devflow-staging",
	})

	if app.Spec.Source == nil {
		t.Fatal("expected source")
	}
	if app.Spec.Source.RepoURL != "oci://"+manifest.ArtifactRepository {
		t.Fatalf("RepoURL = %q", app.Spec.Source.RepoURL)
	}
	if app.Spec.Source.TargetRevision != manifest.ArtifactDigest {
		t.Fatalf("TargetRevision = %q", app.Spec.Source.TargetRevision)
	}
	if app.Spec.Source.Path != "." {
		t.Fatalf("Path = %q", app.Spec.Source.Path)
	}
	if app.Spec.Source.Plugin != nil {
		t.Fatalf("expected plugin source to be cleared, got %+v", app.Spec.Source.Plugin)
	}
	if app.Spec.Destination.Namespace != "devflow-staging" {
		t.Fatalf("namespace = %q", app.Spec.Destination.Namespace)
	}
}

func TestBuildArgoApplicationFallsBackToRepoPluginWithoutArtifact(t *testing.T) {
	model.InitConfigRepo(&model.Repo{Address: "https://github.com/bsonger/manifests.git", Path: "manifests"})
	release := &model.Release{
		BaseModel:     model.BaseModel{ID: uuid.New()},
		ApplicationID: uuid.New(),
		ManifestID:    uuid.New(),
		ImageID:       uuid.New(),
		Env:           "prod",
	}
	manifest := &model.Manifest{BaseModel: model.BaseModel{ID: release.ManifestID}}

	app := buildArgoApplication(release, manifest, &applicationProjection{
		Name:        "demo-api",
		ProjectName: "devflow-prod",
	})

	if app.Spec.Source == nil || app.Spec.Source.Plugin == nil {
		t.Fatal("expected repo plugin fallback source")
	}
	if app.Spec.Source.RepoURL != model.GetConfigRepo().Address {
		t.Fatalf("RepoURL = %q", app.Spec.Source.RepoURL)
	}
	if app.Spec.Source.Path != "./" {
		t.Fatalf("Path = %q", app.Spec.Source.Path)
	}
	if len(app.Spec.Source.Plugin.Parameters) != 4 {
		t.Fatalf("plugin parameters = %d", len(app.Spec.Source.Plugin.Parameters))
	}
}

func TestBuildOCIApplicationSourceUsesTagWhenDigestMissing(t *testing.T) {
	source := buildOCIApplicationSource(&model.Manifest{
		ArtifactRepository: "registry.example.com/devflow/manifests/demo/prod",
		ArtifactTag:        "manifest-tag",
	})
	if source == nil {
		t.Fatal("expected source")
	}
	if source.TargetRevision != "manifest-tag" {
		t.Fatalf("TargetRevision = %q", source.TargetRevision)
	}
	if !strings.HasPrefix(source.RepoURL, "oci://registry.example.com/") {
		t.Fatalf("RepoURL = %q", source.RepoURL)
	}
}

func TestBuildOCIApplicationSourceReturnsNilWithoutArtifactRepository(t *testing.T) {
	if source := buildOCIApplicationSource(&model.Manifest{}); source != nil {
		t.Fatalf("expected nil source, got %+v", source)
	}
}

func TestBuildArgoApplicationUsesEnvironmentNamespaceWhenProjectNameEmpty(t *testing.T) {
	release := &model.Release{
		BaseModel:     model.BaseModel{ID: uuid.New()},
		ApplicationID: uuid.New(),
		ManifestID:    uuid.New(),
		ImageID:       uuid.New(),
		Env:           "staging",
	}

	app := buildArgoApplication(release, &model.Manifest{
		BaseModel:          model.BaseModel{ID: release.ManifestID},
		EnvironmentID:      "staging",
		ArtifactRepository: "zot.zot.svc.cluster.local:5000/devflow/manifests/demo-api/staging",
		ArtifactDigest:     "sha256:abc123",
	}, &applicationProjection{
		Name: "demo-api",
	})

	if app.Spec.Destination.Namespace != "devflow-staging" {
		t.Fatalf("namespace = %q, want devflow-staging", app.Spec.Destination.Namespace)
	}
}


func TestApplyReleaseApplicationTriggersSyncAfterUpgrade(t *testing.T) {
	app := &appv1.Application{}
	updated := false
	synced := false
	err := applyReleaseApplication(context.Background(), model.ReleaseUpgrade, app,
		func(context.Context, *appv1.Application) error {
			t.Fatal("create should not be called for upgrade")
			return nil
		},
		func(_ context.Context, got *appv1.Application) error {
			updated = got == app
			return nil
		},
		func(_ context.Context, name string) error {
			synced = true
			return nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if !updated {
		t.Fatal("expected update to be called")
	}
	if !synced {
		t.Fatal("expected sync to be called")
	}
}

func TestApplyReleaseApplicationTriggersSyncAfterInstall(t *testing.T) {
	app := &appv1.Application{}
	created := false
	synced := false
	err := applyReleaseApplication(context.Background(), model.ReleaseInstall, app,
		func(_ context.Context, got *appv1.Application) error {
			created = got == app
			return nil
		},
		func(context.Context, *appv1.Application) error {
			t.Fatal("update should not be called for install")
			return nil
		},
		func(_ context.Context, name string) error {
			synced = true
			return nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if !created {
		t.Fatal("expected create to be called")
	}
	if !synced {
		t.Fatal("expected sync to be called")
	}
}

var _ *appv1.Application
