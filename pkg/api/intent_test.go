package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bsonger/devflow-release-service/pkg/model"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestBuildIntentFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	applicationID := primitive.NewObjectID()
	manifestID := primitive.NewObjectID()
	resourceID := primitive.NewObjectID()

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet,
		"/api/v1/intents?kind=build&status=Pending&application_id="+applicationID.Hex()+"&manifest_id="+manifestID.Hex()+"&resource_id="+resourceID.Hex()+"&claimed_by=worker-1&branch=main",
		nil,
	)

	filter, err := buildIntentFilter(ctx)
	if err != nil {
		t.Fatalf("buildIntentFilter returned error: %v", err)
	}

	if got := filter["kind"]; got != model.IntentKindBuild {
		t.Fatalf("unexpected kind: got %#v want %#v", got, model.IntentKindBuild)
	}
	if got := filter["status"]; got != model.IntentPending {
		t.Fatalf("unexpected status: got %#v want %#v", got, model.IntentPending)
	}
	if got := filter["application_id"]; got != applicationID {
		t.Fatalf("unexpected application_id: got %#v want %#v", got, applicationID)
	}
	if got := filter["manifest_id"]; got != manifestID {
		t.Fatalf("unexpected manifest_id: got %#v want %#v", got, manifestID)
	}
	if got := filter["resource_id"]; got != resourceID {
		t.Fatalf("unexpected resource_id: got %#v want %#v", got, resourceID)
	}
	if got := filter["claimed_by"]; got != "worker-1" {
		t.Fatalf("unexpected claimed_by: got %#v want %q", got, "worker-1")
	}
	if got := filter["branch"]; got != "main" {
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

	releaseID := primitive.NewObjectID()

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/intents?release_id="+releaseID.Hex()+"&release_type=Upgrade", nil)

	filter, err := buildIntentFilter(ctx)
	if err != nil {
		t.Fatalf("buildIntentFilter returned error: %v", err)
	}

	if got := filter["job_id"]; got != releaseID {
		t.Fatalf("unexpected job_id filter: got %#v want %#v", got, releaseID)
	}
	if got := filter["job_type"]; got != "Upgrade" {
		t.Fatalf("unexpected job_type filter: got %#v want %q", got, "Upgrade")
	}
}
