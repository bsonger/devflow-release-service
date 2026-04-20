package service

import (
	"context"
	"errors"
	"testing"

	"github.com/bsonger/devflow-release-service/pkg/downstream"
)

type fakeBindingReader struct {
	binding *downstream.ApplicationEnvironment
	err     error
}

func (f fakeBindingReader) GetApplicationEnvironment(context.Context, string, string) (*downstream.ApplicationEnvironment, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.binding, nil
}

type fakeOwnerReader struct {
	application *downstream.AppApplication
	project     *downstream.AppProject
	environment *downstream.AppEnvironment
	cluster     *downstream.AppCluster

	applicationErr error
	projectErr     error
	environmentErr error
	clusterErr     error
}

func (f fakeOwnerReader) GetApplication(context.Context, string) (*downstream.AppApplication, error) {
	if f.applicationErr != nil {
		return nil, f.applicationErr
	}
	return f.application, nil
}

func (f fakeOwnerReader) GetProject(context.Context, string) (*downstream.AppProject, error) {
	if f.projectErr != nil {
		return nil, f.projectErr
	}
	return f.project, nil
}

func (f fakeOwnerReader) GetEnvironment(context.Context, string) (*downstream.AppEnvironment, error) {
	if f.environmentErr != nil {
		return nil, f.environmentErr
	}
	return f.environment, nil
}

func (f fakeOwnerReader) GetCluster(context.Context, string) (*downstream.AppCluster, error) {
	if f.clusterErr != nil {
		return nil, f.clusterErr
	}
	return f.cluster, nil
}

func TestResolveDeployTargetProductionNamespaceAndClusterServer(t *testing.T) {
	resolver := &deployTargetResolver{
		bindingReader: fakeBindingReader{binding: &downstream.ApplicationEnvironment{ID: "ae-1", ApplicationID: "app-1"}},
		ownerReader: fakeOwnerReader{
			application: &downstream.AppApplication{ID: "app-1", ProjectID: "proj-1", Name: "checkout"},
			project:     &downstream.AppProject{ID: "proj-1", Name: "Checkout_API"},
			environment: &downstream.AppEnvironment{ID: "env-1", Name: "Production", ClusterID: "cluster-1"},
			cluster:     &downstream.AppCluster{ID: "cluster-1", Name: "prod", Server: "https://k8s.prod.example.com/", OnboardingReady: true, OnboardingCheckedAt: "2026-04-19T00:00:00Z"},
		},
	}

	target, err := resolver.Resolve(context.Background(), "app-1", "env-1")
	if err != nil {
		t.Fatalf("Resolve error = %v", err)
	}
	if target.Namespace != "checkout-api" {
		t.Fatalf("namespace = %q, want checkout-api", target.Namespace)
	}
	if target.DestinationServer != "https://k8s.prod.example.com" {
		t.Fatalf("destination server = %q", target.DestinationServer)
	}
}

func TestResolveDeployTargetNonProductionNamespaceNormalization(t *testing.T) {
	resolver := &deployTargetResolver{
		bindingReader: fakeBindingReader{binding: &downstream.ApplicationEnvironment{ID: "ae-1", ApplicationID: "app-1"}},
		ownerReader: fakeOwnerReader{
			application: &downstream.AppApplication{ID: "app-1", ProjectID: "proj-1", Name: "checkout"},
			project:     &downstream.AppProject{ID: "proj-1", Name: "Checkout Service"},
			environment: &downstream.AppEnvironment{ID: "env-1", Name: "QA_Stage", ClusterID: "cluster-1"},
			cluster:     &downstream.AppCluster{ID: "cluster-1", Name: "qa", Server: "https://k8s.qa.example.com", OnboardingReady: true, OnboardingCheckedAt: "2026-04-19T00:00:00Z"},
		},
	}

	target, err := resolver.Resolve(context.Background(), "app-1", "env-1")
	if err != nil {
		t.Fatalf("Resolve error = %v", err)
	}
	if target.Namespace != "checkout-service-qa-stage" {
		t.Fatalf("namespace = %q, want checkout-service-qa-stage", target.Namespace)
	}
}

func TestResolveDeployTargetProductionUUIDEnvironmentStillUsesProjectNamespace(t *testing.T) {
	resolver := &deployTargetResolver{
		bindingReader: fakeBindingReader{binding: &downstream.ApplicationEnvironment{
			ID:            "360696d1-4255-4065-8030-04ee56de8b12",
			ApplicationID: "e712e979-3728-444f-9808-20ea6fc655a5",
			Environment: downstream.ApplicationEnvironmentRef{
				ID:   "13e18088-ae0a-427c-9f0e-3b0ae6bef13f",
				Name: "production",
			},
		}},
		ownerReader: fakeOwnerReader{
			application: &downstream.AppApplication{ID: "e712e979-3728-444f-9808-20ea6fc655a5", ProjectID: "5355c518-97d3-45ac-9070-229960c2eaec", Name: "devflow-app-service"},
			project:     &downstream.AppProject{ID: "5355c518-97d3-45ac-9070-229960c2eaec", Name: "devflow"},
			environment: &downstream.AppEnvironment{ID: "13e18088-ae0a-427c-9f0e-3b0ae6bef13f", Name: "production", ClusterID: "58472672-15ec-4616-a3a0-4475d0f841dc"},
			cluster:     &downstream.AppCluster{ID: "58472672-15ec-4616-a3a0-4475d0f841dc", Name: "shared-production", Server: "https://kubernetes.default.svc", OnboardingReady: true, OnboardingCheckedAt: "2026-04-20T00:24:53Z"},
		},
	}

	target, err := resolver.Resolve(context.Background(), "e712e979-3728-444f-9808-20ea6fc655a5", "13e18088-ae0a-427c-9f0e-3b0ae6bef13f")
	if err != nil {
		t.Fatalf("Resolve error = %v", err)
	}
	if target.Namespace != "devflow" {
		t.Fatalf("namespace = %q, want devflow", target.Namespace)
	}
	if target.EnvironmentName != "production" {
		t.Fatalf("environment name = %q, want production", target.EnvironmentName)
	}
}

func TestResolveDeployTargetReturnsMissingBindingError(t *testing.T) {
	resolver := &deployTargetResolver{
		bindingReader: fakeBindingReader{err: errors.New("downstream request failed: 404 Not Found")},
		ownerReader:   fakeOwnerReader{},
	}

	_, err := resolver.Resolve(context.Background(), "app-1", "env-1")
	if !errors.Is(err, ErrDeployTargetBindingMissing) {
		t.Fatalf("error = %v, want %v", err, ErrDeployTargetBindingMissing)
	}
}

func TestResolveDeployTargetReturnsMalformedEnvironmentError(t *testing.T) {
	resolver := &deployTargetResolver{
		bindingReader: fakeBindingReader{binding: &downstream.ApplicationEnvironment{ID: "ae-1", ApplicationID: "app-1"}},
		ownerReader: fakeOwnerReader{
			application: &downstream.AppApplication{ID: "app-1", ProjectID: "proj-1"},
			project:     &downstream.AppProject{ID: "proj-1", Name: "checkout"},
			environment: &downstream.AppEnvironment{ID: "env-1", Name: "staging", ClusterID: ""},
		},
	}

	_, err := resolver.Resolve(context.Background(), "app-1", "env-1")
	if !errors.Is(err, ErrDeployTargetEnvironmentMetadataMalformed) {
		t.Fatalf("error = %v, want %v", err, ErrDeployTargetEnvironmentMetadataMalformed)
	}
}

func TestResolveDeployTargetReturnsInvalidClusterServerError(t *testing.T) {
	resolver := &deployTargetResolver{
		bindingReader: fakeBindingReader{binding: &downstream.ApplicationEnvironment{ID: "ae-1", ApplicationID: "app-1"}},
		ownerReader: fakeOwnerReader{
			application: &downstream.AppApplication{ID: "app-1", ProjectID: "proj-1"},
			project:     &downstream.AppProject{ID: "proj-1", Name: "checkout"},
			environment: &downstream.AppEnvironment{ID: "env-1", Name: "staging", ClusterID: "cluster-1"},
			cluster:     &downstream.AppCluster{ID: "cluster-1", Name: "staging", Server: "kubernetes.default.svc", OnboardingReady: true, OnboardingCheckedAt: "2026-04-19T00:00:00Z"},
		},
	}

	_, err := resolver.Resolve(context.Background(), "app-1", "env-1")
	if !errors.Is(err, ErrDeployTargetClusterServerInvalid) {
		t.Fatalf("error = %v, want %v", err, ErrDeployTargetClusterServerInvalid)
	}
}

func TestDeriveNamespaceProductionSpecialCase(t *testing.T) {
	namespace, err := deriveNamespace("Payments_API", "production")
	if err != nil {
		t.Fatalf("deriveNamespace error = %v", err)
	}
	if namespace != "payments-api" {
		t.Fatalf("namespace = %q, want payments-api", namespace)
	}
}

func TestResolveDeployTargetBlocksUnreadyClusterWithOnboardingError(t *testing.T) {
	resolver := &deployTargetResolver{
		bindingReader: fakeBindingReader{binding: &downstream.ApplicationEnvironment{ID: "ae-1", ApplicationID: "app-1"}},
		ownerReader: fakeOwnerReader{
			application: &downstream.AppApplication{ID: "app-1", ProjectID: "proj-1", Name: "checkout"},
			project:     &downstream.AppProject{ID: "proj-1", Name: "Checkout"},
			environment: &downstream.AppEnvironment{ID: "env-1", Name: "staging", ClusterID: "cluster-1"},
			cluster:     &downstream.AppCluster{ID: "cluster-1", Name: "staging", Server: "https://k8s.staging.example.com", OnboardingReady: false, OnboardingError: "secret upsert failed: timeout", OnboardingCheckedAt: "2026-04-18T12:00:00Z"},
		},
	}

	_, err := resolver.Resolve(context.Background(), "app-1", "env-1")
	if !errors.Is(err, ErrDeployTargetClusterNotReady) {
		t.Fatalf("error = %v, want ErrDeployTargetClusterNotReady", err)
	}
}

func TestResolveDeployTargetBlocksUnreadyClusterWithoutOnboardingError(t *testing.T) {
	resolver := &deployTargetResolver{
		bindingReader: fakeBindingReader{binding: &downstream.ApplicationEnvironment{ID: "ae-1", ApplicationID: "app-1"}},
		ownerReader: fakeOwnerReader{
			application: &downstream.AppApplication{ID: "app-1", ProjectID: "proj-1", Name: "checkout"},
			project:     &downstream.AppProject{ID: "proj-1", Name: "Checkout"},
			environment: &downstream.AppEnvironment{ID: "env-1", Name: "staging", ClusterID: "cluster-1"},
			cluster:     &downstream.AppCluster{ID: "cluster-1", Name: "staging", Server: "https://k8s.staging.example.com", OnboardingReady: false, OnboardingError: "", OnboardingCheckedAt: "2026-04-18T12:00:00Z"},
		},
	}

	_, err := resolver.Resolve(context.Background(), "app-1", "env-1")
	if !errors.Is(err, ErrDeployTargetClusterNotReady) {
		t.Fatalf("error = %v, want ErrDeployTargetClusterNotReady", err)
	}
	// Should use "unknown onboarding failure" fallback reason
	if err.Error() == "" {
		t.Fatalf("expected non-empty error message")
	}
}

func TestResolveDeployTargetBlocksUncheckedClusterAsMalformed(t *testing.T) {
	resolver := &deployTargetResolver{
		bindingReader: fakeBindingReader{binding: &downstream.ApplicationEnvironment{ID: "ae-1", ApplicationID: "app-1"}},
		ownerReader: fakeOwnerReader{
			application: &downstream.AppApplication{ID: "app-1", ProjectID: "proj-1", Name: "checkout"},
			project:     &downstream.AppProject{ID: "proj-1", Name: "Checkout"},
			environment: &downstream.AppEnvironment{ID: "env-1", Name: "staging", ClusterID: "cluster-1"},
			cluster:     &downstream.AppCluster{ID: "cluster-1", Name: "staging", Server: "https://k8s.staging.example.com", OnboardingReady: false, OnboardingError: "", OnboardingCheckedAt: ""},
		},
	}

	_, err := resolver.Resolve(context.Background(), "app-1", "env-1")
	if !errors.Is(err, ErrDeployTargetClusterReadinessMalformed) {
		t.Fatalf("error = %v, want ErrDeployTargetClusterReadinessMalformed", err)
	}
}

func TestResolveDeployTargetClusterLookupFailed(t *testing.T) {
	resolver := &deployTargetResolver{
		bindingReader: fakeBindingReader{binding: &downstream.ApplicationEnvironment{ID: "ae-1", ApplicationID: "app-1"}},
		ownerReader: fakeOwnerReader{
			application: &downstream.AppApplication{ID: "app-1", ProjectID: "proj-1", Name: "checkout"},
			project:     &downstream.AppProject{ID: "proj-1", Name: "Checkout"},
			environment: &downstream.AppEnvironment{ID: "env-1", Name: "staging", ClusterID: "cluster-1"},
			clusterErr:  errors.New("downstream request failed: 500 Internal Server Error"),
		},
	}

	_, err := resolver.Resolve(context.Background(), "app-1", "env-1")
	if !errors.Is(err, ErrDeployTargetClusterLookupFailed) {
		t.Fatalf("error = %v, want ErrDeployTargetClusterLookupFailed", err)
	}
}

func TestResolveDeployTargetClusterMetadataMissingOn404(t *testing.T) {
	resolver := &deployTargetResolver{
		bindingReader: fakeBindingReader{binding: &downstream.ApplicationEnvironment{ID: "ae-1", ApplicationID: "app-1"}},
		ownerReader: fakeOwnerReader{
			application: &downstream.AppApplication{ID: "app-1", ProjectID: "proj-1", Name: "checkout"},
			project:     &downstream.AppProject{ID: "proj-1", Name: "Checkout"},
			environment: &downstream.AppEnvironment{ID: "env-1", Name: "staging", ClusterID: "cluster-missing"},
			clusterErr:  errors.New("downstream request failed: 404 Not Found"),
		},
	}

	_, err := resolver.Resolve(context.Background(), "app-1", "env-1")
	if !errors.Is(err, ErrDeployTargetClusterMetadataMissing) {
		t.Fatalf("error = %v, want ErrDeployTargetClusterMetadataMissing", err)
	}
}

func TestResolveDeployTargetClusterMetadataMalformedOnNilCluster(t *testing.T) {
	resolver := &deployTargetResolver{
		bindingReader: fakeBindingReader{binding: &downstream.ApplicationEnvironment{ID: "ae-1", ApplicationID: "app-1"}},
		ownerReader: fakeOwnerReader{
			application: &downstream.AppApplication{ID: "app-1", ProjectID: "proj-1", Name: "checkout"},
			project:     &downstream.AppProject{ID: "proj-1", Name: "Checkout"},
			environment: &downstream.AppEnvironment{ID: "env-1", Name: "staging", ClusterID: "cluster-1"},
			cluster:     nil,
		},
	}

	_, err := resolver.Resolve(context.Background(), "app-1", "env-1")
	if !errors.Is(err, ErrDeployTargetClusterMetadataMalformed) {
		t.Fatalf("error = %v, want ErrDeployTargetClusterMetadataMalformed", err)
	}
}

func TestResolveDeployTargetClusterMetadataMalformedOnEmptyID(t *testing.T) {
	resolver := &deployTargetResolver{
		bindingReader: fakeBindingReader{binding: &downstream.ApplicationEnvironment{ID: "ae-1", ApplicationID: "app-1"}},
		ownerReader: fakeOwnerReader{
			application: &downstream.AppApplication{ID: "app-1", ProjectID: "proj-1", Name: "checkout"},
			project:     &downstream.AppProject{ID: "proj-1", Name: "Checkout"},
			environment: &downstream.AppEnvironment{ID: "env-1", Name: "staging", ClusterID: "cluster-1"},
			cluster:     &downstream.AppCluster{ID: "", Name: "staging", Server: "https://k8s.staging.example.com"},
		},
	}

	_, err := resolver.Resolve(context.Background(), "app-1", "env-1")
	if !errors.Is(err, ErrDeployTargetClusterMetadataMalformed) {
		t.Fatalf("error = %v, want ErrDeployTargetClusterMetadataMalformed", err)
	}
}

func TestResolveDeployTargetReadyClusterResolvesSuccessfully(t *testing.T) {
	resolver := &deployTargetResolver{
		bindingReader: fakeBindingReader{binding: &downstream.ApplicationEnvironment{ID: "ae-1", ApplicationID: "app-1"}},
		ownerReader: fakeOwnerReader{
			application: &downstream.AppApplication{ID: "app-1", ProjectID: "proj-1", Name: "checkout"},
			project:     &downstream.AppProject{ID: "proj-1", Name: "Checkout"},
			environment: &downstream.AppEnvironment{ID: "env-1", Name: "staging", ClusterID: "cluster-1"},
			cluster:     &downstream.AppCluster{ID: "cluster-1", Name: "staging", Server: "https://k8s.staging.example.com", OnboardingReady: true, OnboardingCheckedAt: "2026-04-19T00:00:00Z"},
		},
	}

	target, err := resolver.Resolve(context.Background(), "app-1", "env-1")
	if err != nil {
		t.Fatalf("Resolve error = %v", err)
	}
	if target.Namespace != "checkout-staging" {
		t.Fatalf("namespace = %q, want checkout-staging", target.Namespace)
	}
	if target.DestinationServer != "https://k8s.staging.example.com" {
		t.Fatalf("destination server = %q, want https://k8s.staging.example.com", target.DestinationServer)
	}
	if target.ClusterID != "cluster-1" {
		t.Fatalf("cluster_id = %q, want cluster-1", target.ClusterID)
	}
}

func TestResolveDeployTargetEmptyApplicationID(t *testing.T) {
	resolver := &deployTargetResolver{
		bindingReader: fakeBindingReader{},
		ownerReader:   fakeOwnerReader{},
	}
	_, err := resolver.Resolve(context.Background(), "", "env-1")
	if !errors.Is(err, ErrDeployTargetApplicationIDRequired) {
		t.Fatalf("error = %v, want ErrDeployTargetApplicationIDRequired", err)
	}
}

func TestResolveDeployTargetEmptyEnvironmentID(t *testing.T) {
	resolver := &deployTargetResolver{
		bindingReader: fakeBindingReader{},
		ownerReader:   fakeOwnerReader{},
	}
	_, err := resolver.Resolve(context.Background(), "app-1", "")
	if !errors.Is(err, ErrDeployTargetEnvironmentIDRequired) {
		t.Fatalf("error = %v, want ErrDeployTargetEnvironmentIDRequired", err)
	}
}

func TestDeriveNamespaceRequiresNames(t *testing.T) {
	if _, err := deriveNamespace("", "staging"); !errors.Is(err, ErrNamespaceProjectNameRequired) {
		t.Fatalf("error = %v, want %v", err, ErrNamespaceProjectNameRequired)
	}
	if _, err := deriveNamespace("checkout", ""); !errors.Is(err, ErrNamespaceEnvironmentNameRequired) {
		t.Fatalf("error = %v, want %v", err, ErrNamespaceEnvironmentNameRequired)
	}
}
