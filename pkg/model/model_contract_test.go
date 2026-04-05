package model

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
)

func TestManifestContract(t *testing.T) {
	typ := reflect.TypeOf(Manifest{})
	for _, field := range []string{"ExecutionIntentID", "ApplicationID", "ConfigurationRevisionID", "RuntimeSpecRevisionID", "Name", "Branch", "RepoAddress", "PipelineID", "Status"} {
		f, ok := typ.FieldByName(field)
		if !ok {
			t.Fatalf("Manifest missing field %s", field)
		}
		if field == "ApplicationID" && f.Type != reflect.TypeOf(uuid.UUID{}) {
			t.Fatalf("Manifest.ApplicationID type = %v, want uuid.UUID", f.Type)
		}
	}
	for _, removed := range []string{"ApplicationName", "GitRepo", "ConfigMaps", "Service", "Internet", "Envs", "Replica", "Type", "Services"} {
		if _, ok := typ.FieldByName(removed); ok {
			t.Fatalf("Manifest should not expose legacy field %s", removed)
		}
	}
}

func TestReleaseContract(t *testing.T) {
	typ := reflect.TypeOf(Release{})
	for _, field := range []string{"ExecutionIntentID", "ApplicationID", "ManifestID", "Env", "Type", "Status"} {
		f, ok := typ.FieldByName(field)
		if !ok {
			t.Fatalf("Release missing field %s", field)
		}
		if field == "ManifestID" && f.Type != reflect.TypeOf(uuid.UUID{}) {
			t.Fatalf("Release.ManifestID type = %v, want uuid.UUID", f.Type)
		}
	}
}

func TestIntentContract(t *testing.T) {
	typ := reflect.TypeOf(Intent{})
	for _, field := range []string{"ResourceID", "TraceID", "Message", "LastError", "ClaimedBy", "AttemptCount"} {
		if _, ok := typ.FieldByName(field); !ok {
			t.Fatalf("Intent missing field %s", field)
		}
	}
}

func TestBaseModelWithCreateDefault(t *testing.T) {
	var base BaseModel
	base.WithCreateDefault()

	if base.ID == uuid.Nil {
		t.Fatal("BaseModel.WithCreateDefault should assign a UUID")
	}
	if base.CreatedAt.IsZero() || base.UpdatedAt.IsZero() {
		t.Fatal("BaseModel.WithCreateDefault should set timestamps")
	}
}
