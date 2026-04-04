package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IntentKind string

const (
	IntentKindBuild   IntentKind = "build"
	IntentKindRelease IntentKind = "release"
)

type IntentStatus string

const (
	IntentPending   IntentStatus = "Pending"
	IntentRunning   IntentStatus = "Running"
	IntentSucceeded IntentStatus = "Succeeded"
	IntentFailed    IntentStatus = "Failed"
)

type Intent struct {
	BaseModel `bson:",inline"`

	Kind            IntentKind          `bson:"kind" json:"kind"`
	Status          IntentStatus        `bson:"status" json:"status"`
	ResourceType    string              `bson:"resource_type" json:"resource_type"`
	ResourceID      primitive.ObjectID  `bson:"resource_id" json:"resource_id"`
	ApplicationID   primitive.ObjectID  `bson:"application_id" json:"application_id"`
	ApplicationName string              `bson:"application_name" json:"application_name"`
	ManifestID      *primitive.ObjectID `bson:"manifest_id,omitempty" json:"manifest_id,omitempty"`
	ManifestName    string              `bson:"manifest_name,omitempty" json:"manifest_name,omitempty"`
	JobID           *primitive.ObjectID `bson:"job_id,omitempty" json:"job_id,omitempty"`
	JobType         string              `bson:"job_type,omitempty" json:"job_type,omitempty"`
	Env             string              `bson:"env,omitempty" json:"env,omitempty"`
	RepoURL         string              `bson:"repo_url,omitempty" json:"repo_url,omitempty"`
	Branch          string              `bson:"branch,omitempty" json:"branch,omitempty"`
	ExternalRef     string              `bson:"external_ref,omitempty" json:"external_ref,omitempty"`
	Message         string              `bson:"message,omitempty" json:"message,omitempty"`
	LastError       string              `bson:"last_error,omitempty" json:"last_error,omitempty"`
	ClaimedBy       string              `bson:"claimed_by,omitempty" json:"claimed_by,omitempty"`
	ClaimedAt       *time.Time          `bson:"claimed_at,omitempty" json:"claimed_at,omitempty"`
	LeaseExpiresAt  *time.Time          `bson:"lease_expires_at,omitempty" json:"lease_expires_at,omitempty"`
	AttemptCount    int                 `bson:"attempt_count,omitempty" json:"attempt_count,omitempty"`
}

func (Intent) CollectionName() string { return "execution_intents" }
