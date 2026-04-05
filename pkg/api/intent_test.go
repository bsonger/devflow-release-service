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

	resourceID := uuid.New()

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet,
		"/api/v1/intents?kind=build&status=Pending&resource_id="+resourceID.String()+"&claimed_by=worker-1",
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
	if filter.ResourceID == nil || *filter.ResourceID != resourceID {
		t.Fatalf("unexpected resource_id: got %#v want %#v", filter.ResourceID, resourceID)
	}
	if got := filter.ClaimedBy; got != "worker-1" {
		t.Fatalf("unexpected claimed_by: got %#v want %q", got, "worker-1")
	}
}

func TestBuildIntentFilterInvalidObjectID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/intents?resource_id=invalid-id", nil)

	_, err := buildIntentFilter(ctx)
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	if err.Error() != "invalid resource_id" {
		t.Fatalf("unexpected error: got %q want %q", err.Error(), "invalid resource_id")
	}
}
