package model

import "testing"

func TestProjectApplyDefaults(t *testing.T) {
	project := &Project{Name: "dev-platform"}
	project.ApplyDefaults()

	if project.Status != ProjectActive {
		t.Fatalf("expected default status %q, got %q", ProjectActive, project.Status)
	}
	if project.Namespace != "dev-platform" {
		t.Fatalf("expected namespace to default to project name, got %q", project.Namespace)
	}
}

func TestProjectCollectionName(t *testing.T) {
	if got := (Project{}).CollectionName(); got != "projects" {
		t.Fatalf("expected projects collection, got %q", got)
	}
}
