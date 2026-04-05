package model

import (
	"time"

	"github.com/google/uuid"
)

type Intent struct {
	BaseModel

	Kind           IntentKind   `json:"kind" db:"kind"`
	Status         IntentStatus `json:"status" db:"status"`
	ResourceType   string       `json:"resource_type" db:"resource_type"`
	ResourceID     uuid.UUID    `json:"resource_id" db:"resource_id"`
	ApplicationID  uuid.UUID    `json:"application_id" db:"application_id"`
	ManifestID     *uuid.UUID   `json:"manifest_id,omitempty" db:"manifest_id"`
	ReleaseID      *uuid.UUID   `json:"release_id,omitempty" db:"release_id"`
	ReleaseType    string       `json:"release_type,omitempty" db:"release_type"`
	Env            string       `json:"env,omitempty" db:"env"`
	RepoAddress    string       `json:"repo_address,omitempty" db:"repo_address"`
	Branch         string       `json:"branch,omitempty" db:"branch"`
	ExternalRef    string       `json:"external_ref,omitempty" db:"external_ref"`
	TraceID        string       `json:"trace_id,omitempty" db:"trace_id"`
	Message        string       `json:"message,omitempty" db:"message"`
	LastError      string       `json:"last_error,omitempty" db:"last_error"`
	ClaimedBy      string       `json:"claimed_by,omitempty" db:"claimed_by"`
	ClaimedAt      *time.Time   `json:"claimed_at,omitempty" db:"claimed_at"`
	LeaseExpiresAt *time.Time   `json:"lease_expires_at,omitempty" db:"lease_expires_at"`
	AttemptCount   int          `json:"attempt_count" db:"attempt_count"`
}

func (Intent) CollectionName() string { return "execution_intents" }
