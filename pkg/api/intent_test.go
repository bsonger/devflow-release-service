package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bsonger/devflow-release-service/pkg/model"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestBuildIntentFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	applicationID := uuid.New()
	manifestID := uuid.New()
	resourceID := uuid.New()

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet,
		"/api/v1/intents?kind=build&status=Pending&application_id="+applicationID.String()+"&manifest_id="+manifestID.String()+"&resource_id="+resourceID.String()+"&claimed_by=worker-1&branch=main",
		nil,
	)

	filter, err := buildIntentFilter(ctx)
	if err != nil {
		t.Fatalf("buildIntentFilter returned error: %v", err)
	}

	if got := filter.Kind; got != string(model.IntentKindBuild) {
		t.Fatalf("unexpected kind: got %#v want %#v", got, model.IntentKindBuild)
	}
	if got := filter.Status; got != string(model.IntentPending) {
		t.Fatalf("unexpected status: got %#v want %#v", got, model.IntentPending)
	}
	if filter.ApplicationID == nil || *filter.ApplicationID != applicationID {
		t.Fatalf("unexpected application_id: got %#v want %#v", filter.ApplicationID, applicationID)
	}
	if filter.ManifestID == nil || *filter.ManifestID != manifestID {
		t.Fatalf("unexpected manifest_id: got %#v want %#v", filter.ManifestID, manifestID)
	}
	if filter.ResourceID == nil || *filter.ResourceID != resourceID {
		t.Fatalf("unexpected resource_id: got %#v want %#v", filter.ResourceID, resourceID)
	}
	if got := filter.ClaimedBy; got != "worker-1" {
		t.Fatalf("unexpected claimed_by: got %#v want %q", got, "worker-1")
	}
	if got := filter.Branch; got != "main" {
		t.Fatalf("unexpected branch: got %#v want %q", got, "main")
	}
}

func TestBuildIntentFilterInvalidObjectID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/intents?application_id=invalid-id", nil)

	_, err := buildIntentFilter(ctx)
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	if err.Error() != "invalid application_id" {
		t.Fatalf("unexpected error: got %q want %q", err.Error(), "invalid application_id")
	}
}

func TestBuildIntentFilterReleaseIDAlias(t *testing.T) {
	gin.SetMode(gin.TestMode)

	releaseID := uuid.New()

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/intents?release_id="+releaseID.String()+"&release_type=Upgrade", nil)

	filter, err := buildIntentFilter(ctx)
	if err != nil {
		t.Fatalf("buildIntentFilter returned error: %v", err)
	}

	if filter.ReleaseID == nil {
		t.Fatalf("expected release_id filter to be set")
	}
	if got := filter.ReleaseType; got != "Upgrade" {
		t.Fatalf("unexpected release_type filter: got %#v want %q", got, "Upgrade")
	}
	if *filter.ReleaseID != releaseID {
		t.Fatalf("unexpected release_id filter: got %#v want %#v", *filter.ReleaseID, releaseID)
	}
}
