package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bsonger/devflow-release-service/pkg/model"
	"github.com/bsonger/devflow-release-service/pkg/service"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestBuildIntentFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	applicationID := service.BridgeObjectIDToUUID(primitive.NewObjectID())
	manifestID := service.BridgeObjectIDToUUID(primitive.NewObjectID())
	resourceID := service.BridgeObjectIDToUUID(primitive.NewObjectID())
	applicationOID, _ := service.BridgeUUIDToObjectID(applicationID)
	manifestOID, _ := service.BridgeUUIDToObjectID(manifestID)
	resourceOID, _ := service.BridgeUUIDToObjectID(resourceID)

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

	if got := filter["kind"]; got != model.IntentKindBuild {
		t.Fatalf("unexpected kind: got %#v want %#v", got, model.IntentKindBuild)
	}
	if got := filter["status"]; got != model.IntentPending {
		t.Fatalf("unexpected status: got %#v want %#v", got, model.IntentPending)
	}
	if got := filter["application_id"]; got != applicationOID {
		t.Fatalf("unexpected application_id: got %#v want %#v", got, applicationOID)
	}
	if got := filter["manifest_id"]; got != manifestOID {
		t.Fatalf("unexpected manifest_id: got %#v want %#v", got, manifestOID)
	}
	if got := filter["resource_id"]; got != resourceOID {
		t.Fatalf("unexpected resource_id: got %#v want %#v", got, resourceOID)
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

	releaseID := service.BridgeObjectIDToUUID(primitive.NewObjectID())
	releaseOID, _ := service.BridgeUUIDToObjectID(releaseID)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/intents?release_id="+releaseID.String()+"&release_type=Upgrade", nil)

	filter, err := buildIntentFilter(ctx)
	if err != nil {
		t.Fatalf("buildIntentFilter returned error: %v", err)
	}

	gotReleaseID, ok := filter["release_id"]
	if !ok {
		t.Fatalf("expected release_id filter to be set")
	}
	if got := filter["release_type"]; got != "Upgrade" {
		t.Fatalf("unexpected release_type filter: got %#v want %q", got, "Upgrade")
	}
	if gotReleaseID != releaseOID {
		t.Fatalf("unexpected release_id filter: got %#v want %#v", gotReleaseID, releaseOID)
	}
}
