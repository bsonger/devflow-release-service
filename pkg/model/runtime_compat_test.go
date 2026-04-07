package model

import (
	"testing"

	"github.com/google/uuid"
)

func TestGeneratePipelineRunUsesProvidedPipelineName(t *testing.T) {
	manifest := &Manifest{BaseModel: BaseModel{ID: uuid.New()}, RepoAddress: "git@github.com:bsonger/devflow.git", Branch: "main", Name: "devflow-main"}
	run := manifest.GeneratePipelineRun("devflow-tekton-image-build", "pvc-1")

	if run.Spec.PipelineRef == nil || run.Spec.PipelineRef.Name != "devflow-tekton-image-build" {
		t.Fatalf("unexpected pipeline ref: %+v", run.Spec.PipelineRef)
	}
	if run.GenerateName != "devflow-tekton-image-build-run-" {
		t.Fatalf("unexpected generateName: %s", run.GenerateName)
	}
}
