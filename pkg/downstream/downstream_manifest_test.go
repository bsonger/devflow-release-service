package downstream

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestManifestBindingClientUsesApplicationEnvironmentEndpoint(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/platform/applications/app-1/environments/env-1" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		_, _ = io.WriteString(w, `{"data":{"id":"ae-1","application_id":"app-1"}}`)
	}))
	defer ts.Close()

	client := NewOrchestratorManifestClient(ts.URL)
	got, err := client.GetApplicationEnvironment(context.Background(), "app-1", "env-1")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "ae-1" || got.ApplicationID != "app-1" {
		t.Fatalf("unexpected payload %+v", got)
	}
}

func TestManifestBindingClientAcceptsBareJSONResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/platform/applications/app-1/environments/staging" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		_, _ = io.WriteString(w, `{"id":"ae-1","application_id":"app-1"}`)
	}))
	defer ts.Close()

	client := NewOrchestratorManifestClient(ts.URL)
	got, err := client.GetApplicationEnvironment(context.Background(), "app-1", "staging")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "ae-1" || got.ApplicationID != "app-1" {
		t.Fatalf("unexpected payload %+v", got)
	}
}

func TestNetworkManifestClientTreatsNullEnvelopeDataAsEmptyList(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/applications/app-1/routes" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		_, _ = io.WriteString(w, `{"data":null,"pagination":{"total":0}}`)
	}))
	defer ts.Close()

	client := NewNetworkManifestClient(ts.URL)
	got, err := client.ListRoutes(context.Background(), "app-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("routes = %+v, want empty", got)
	}
}
