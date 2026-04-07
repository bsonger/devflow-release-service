package model

import "testing"

func TestGeneratePipelineRunUsesProvidedPipelineName(t *testing.T) {
	manifest := &Manifest{}
	run := manifest.GeneratePipelineRun("devflow-tekton-image-build", "pvc-1")
	if run.GenerateName != "devflow-tekton-image-build-run-" {
		t.Fatalf("GenerateName = %q", run.GenerateName)
	}
	if run.Spec.PipelineRef == nil || run.Spec.PipelineRef.Name != "devflow-tekton-image-build" {
		t.Fatalf("PipelineRef = %+v", run.Spec.PipelineRef)
	}
}

func TestGeneratePipelineRunCarriesManifestIdentifierMetadata(t *testing.T) {
	manifest := &Manifest{}
	manifest.WithCreateDefault()
	run := manifest.GeneratePipelineRun("devflow-tekton-image-build", "pvc-1")
	if got := run.Labels["devflow.manifest/id"]; got != manifest.ID.String() {
		t.Fatalf("label manifest id = %q want %q", got, manifest.ID.String())
	}
	if got := run.Annotations["devflow.manifest/id"]; got != manifest.ID.String() {
		t.Fatalf("annotation manifest id = %q want %q", got, manifest.ID.String())
	}
}
