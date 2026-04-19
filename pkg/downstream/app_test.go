package downstream

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAppManifestClientReadsApplicationProjectEnvironmentCluster(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/applications/app-1":
			_, _ = io.WriteString(w, `{"data":{"id":"app-1","project_id":"proj-1","name":"checkout"}}`)
		case "/api/v1/projects/proj-1":
			_, _ = io.WriteString(w, `{"data":{"id":"proj-1","name":"checkout"}}`)
		case "/api/v1/environments/env-1":
			_, _ = io.WriteString(w, `{"data":{"id":"env-1","name":"staging","cluster_id":"cluster-1"}}`)
		case "/api/v1/clusters/cluster-1":
			_, _ = io.WriteString(w, `{"data":{"id":"cluster-1","name":"staging","server":"https://k8s.staging.example.com","onboarding_ready":true,"onboarding_error":"","onboarding_checked_at":"2026-04-19T00:00:00Z"}}`)
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer ts.Close()

	client := NewAppManifestClient(ts.URL)
	app, err := client.GetApplication(context.Background(), "app-1")
	if err != nil {
		t.Fatal(err)
	}
	if app.ProjectID != "proj-1" {
		t.Fatalf("application project_id = %q", app.ProjectID)
	}
	project, err := client.GetProject(context.Background(), "proj-1")
	if err != nil {
		t.Fatal(err)
	}
	if project.Name != "checkout" {
		t.Fatalf("project name = %q", project.Name)
	}
	env, err := client.GetEnvironment(context.Background(), "env-1")
	if err != nil {
		t.Fatal(err)
	}
	if env.ClusterID != "cluster-1" {
		t.Fatalf("environment cluster_id = %q", env.ClusterID)
	}
	cluster, err := client.GetCluster(context.Background(), "cluster-1")
	if err != nil {
		t.Fatal(err)
	}
	if cluster.Server != "https://k8s.staging.example.com" {
		t.Fatalf("cluster server = %q", cluster.Server)
	}
	if !cluster.OnboardingReady {
		t.Fatalf("cluster onboarding_ready = false, want true")
	}
	if cluster.OnboardingCheckedAt != "2026-04-19T00:00:00Z" {
		t.Fatalf("cluster onboarding_checked_at = %q, want 2026-04-19T00:00:00Z", cluster.OnboardingCheckedAt)
	}
}

func TestAppManifestClientParsesUnreadyClusterDiagnostics(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"data":{"id":"cluster-2","name":"dev","server":"https://k8s.dev.example.com","onboarding_ready":false,"onboarding_error":"secret upsert failed: timeout","onboarding_checked_at":"2026-04-18T12:00:00Z"}}`)
	}))
	defer ts.Close()

	client := NewAppManifestClient(ts.URL)
	cluster, err := client.GetCluster(context.Background(), "cluster-2")
	if err != nil {
		t.Fatal(err)
	}
	if cluster.OnboardingReady {
		t.Fatalf("onboarding_ready = true, want false")
	}
	if cluster.OnboardingError != "secret upsert failed: timeout" {
		t.Fatalf("onboarding_error = %q", cluster.OnboardingError)
	}
}

func TestAppManifestClientParsesClusterWithMissingReadinessFields(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"data":{"id":"cluster-3","name":"legacy","server":"https://k8s.legacy.example.com"}}`)
	}))
	defer ts.Close()

	client := NewAppManifestClient(ts.URL)
	cluster, err := client.GetCluster(context.Background(), "cluster-3")
	if err != nil {
		t.Fatal(err)
	}
	if cluster.OnboardingReady {
		t.Fatalf("onboarding_ready = true for legacy cluster without readiness fields, want false (zero value)")
	}
}

func TestAppManifestClientAcceptsBareJSONResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/environments/env-1" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		_, _ = io.WriteString(w, `{"id":"env-1","name":"staging","cluster_id":"cluster-1"}`)
	}))
	defer ts.Close()

	client := NewAppManifestClient(ts.URL)
	env, err := client.GetEnvironment(context.Background(), "env-1")
	if err != nil {
		t.Fatal(err)
	}
	if env.ID != "env-1" || env.ClusterID != "cluster-1" {
		t.Fatalf("unexpected env payload %+v", env)
	}
}

func TestAppManifestClientClusterNotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"error":{"code":"not_found","message":"cluster not found"}}`)
	}))
	defer ts.Close()

	client := NewAppManifestClient(ts.URL)
	_, err := client.GetCluster(context.Background(), "cluster-missing")
	if err == nil {
		t.Fatal("expected error for 404 cluster response")
	}
}

func TestAppManifestClientClusterInternalError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, `{"error":{"code":"internal","message":"database error"}}`)
	}))
	defer ts.Close()

	client := NewAppManifestClient(ts.URL)
	_, err := client.GetCluster(context.Background(), "cluster-err")
	if err == nil {
		t.Fatal("expected error for 500 cluster response")
	}
}
