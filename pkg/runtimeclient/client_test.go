package runtimeclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

func TestClientGetRuntimeSpecRevision(t *testing.T) {
	revisionID := uuid.New()
	runtimeSpecID := uuid.New()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/runtime-spec-revisions/"+revisionID.String() {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"` + revisionID.String() + `","runtime_spec_id":"` + runtimeSpecID.String() + `"}`))
	}))
	defer server.Close()

	client := New(server.URL)
	got, err := client.GetRuntimeSpecRevision(context.Background(), revisionID)
	if err != nil {
		t.Fatalf("GetRuntimeSpecRevision: %v", err)
	}
	if got.ID != revisionID || got.RuntimeSpecID != runtimeSpecID {
		t.Fatalf("unexpected revision payload: %+v", got)
	}
}
