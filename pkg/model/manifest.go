package model

import "github.com/google/uuid"

type Manifest struct {
	BaseModel

	ExecutionIntentID       *uuid.UUID     `json:"execution_intent_id,omitempty" db:"execution_intent_id"`
	ApplicationID           uuid.UUID      `json:"application_id" db:"application_id"`
	ConfigurationRevisionID *uuid.UUID     `json:"configuration_revision_id,omitempty" db:"configuration_revision_id"`
	RuntimeSpecRevisionID   *uuid.UUID     `json:"runtime_spec_revision_id,omitempty" db:"runtime_spec_revision_id"`
	Name                    string         `json:"name" db:"name"`
	Branch                  string         `json:"branch" db:"branch"`
	RepoAddress             string         `json:"repo_address" db:"repo_address"`
	CommitHash              string         `json:"commit_hash,omitempty" db:"commit_hash"`
	Digest                  string         `json:"digest,omitempty" db:"digest"`
	PipelineID              string         `json:"pipeline_id,omitempty" db:"pipeline_id"`
	Steps                   []ManifestStep `json:"steps" db:"steps"`
	Status                  ManifestStatus `json:"status" db:"status"`
}

type PatchManifestRequest struct {
	CommitHash string `json:"commit_hash,omitempty"`
	Digest     string `json:"digest,omitempty"`
}

func (r *PatchManifestRequest) IsEmpty() bool {
	return r.CommitHash == "" && r.Digest == ""
}

func (m *Manifest) CollectionName() string { return "manifests" }
