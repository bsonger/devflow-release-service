package service

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/bsonger/devflow-release-service/pkg/model"
	"github.com/google/uuid"
)

type serviceShape struct {
	Ports []model.Port `json:"ports"`
}

type applicationProjection struct {
	ID          uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
	Name        string
	ProjectName string
	RepoAddress string
	RepoURL     string
	Replica     *int32
	Type        model.ReleaseType
	Service     serviceShape
	Internet    model.Internet
}

func scanManifest(scanner interface {
	Scan(dest ...any) error
}) (*model.Manifest, error) {
	var (
		item            model.Manifest
		executionIntent sql.NullString
		replica         sql.NullInt32
		servicesBytes   []byte
		stepsBytes      []byte
		deletedAt       sql.NullTime
	)

	if err := scanner.Scan(
		&item.ID,
		&executionIntent,
		&item.ApplicationID,
		&item.Name,
		&item.Branch,
		&item.RepoAddress,
		&item.CommitHash,
		&replica,
		&item.Digest,
		&item.Type,
		&servicesBytes,
		&item.PipelineID,
		&stepsBytes,
		&item.Status,
		&item.CreatedAt,
		&item.UpdatedAt,
		&deletedAt,
	); err != nil {
		return nil, err
	}

	intentID, err := parseNullUUID(executionIntent)
	if err != nil {
		return nil, err
	}
	item.ExecutionIntentID = intentID
	if replica.Valid {
		value := replica.Int32
		item.Replica = &value
	}
	if len(servicesBytes) > 0 {
		if err := json.Unmarshal(servicesBytes, &item.Services); err != nil {
			return nil, err
		}
	}
	if len(stepsBytes) > 0 {
		if err := json.Unmarshal(stepsBytes, &item.Steps); err != nil {
			return nil, err
		}
	}
	if deletedAt.Valid {
		item.DeletedAt = &deletedAt.Time
	}
	return &item, nil
}

func scanRelease(scanner interface {
	Scan(dest ...any) error
}) (*model.Release, error) {
	var (
		item                    model.Release
		executionIntent         sql.NullString
		configurationID         sql.NullString
		configurationRevisionID sql.NullString
		runtimeSpecRevisionID   sql.NullString
		stepsBytes              []byte
		deletedAt               sql.NullTime
	)

	if err := scanner.Scan(
		&item.ID,
		&executionIntent,
		&configurationID,
		&configurationRevisionID,
		&runtimeSpecRevisionID,
		&item.ApplicationID,
		&item.ManifestID,
		&item.Env,
		&item.Type,
		&stepsBytes,
		&item.Status,
		&item.ExternalRef,
		&item.CreatedAt,
		&item.UpdatedAt,
		&deletedAt,
	); err != nil {
		return nil, err
	}

	var err error
	item.ExecutionIntentID, err = parseNullUUID(executionIntent)
	if err != nil {
		return nil, err
	}
	item.ConfigurationID, err = parseNullUUID(configurationID)
	if err != nil {
		return nil, err
	}
	item.ConfigurationRevisionID, err = parseNullUUID(configurationRevisionID)
	if err != nil {
		return nil, err
	}
	item.RuntimeSpecRevisionID, err = parseNullUUID(runtimeSpecRevisionID)
	if err != nil {
		return nil, err
	}
	if len(stepsBytes) > 0 {
		if err := json.Unmarshal(stepsBytes, &item.Steps); err != nil {
			return nil, err
		}
	}
	if deletedAt.Valid {
		item.DeletedAt = &deletedAt.Time
	}
	return &item, nil
}

func scanIntent(scanner interface {
	Scan(dest ...any) error
}) (*model.Intent, error) {
	var (
		item           model.Intent
		manifestID     sql.NullString
		releaseID      sql.NullString
		claimedAt      sql.NullTime
		leaseExpiresAt sql.NullTime
		deletedAt      sql.NullTime
	)

	if err := scanner.Scan(
		&item.ID,
		&item.Kind,
		&item.Status,
		&item.ResourceType,
		&item.ResourceID,
		&item.ApplicationID,
		&manifestID,
		&releaseID,
		&item.ReleaseType,
		&item.Env,
		&item.RepoAddress,
		&item.Branch,
		&item.ExternalRef,
		&item.TraceID,
		&item.Message,
		&item.LastError,
		&item.ClaimedBy,
		&claimedAt,
		&leaseExpiresAt,
		&item.AttemptCount,
		&item.CreatedAt,
		&item.UpdatedAt,
		&deletedAt,
	); err != nil {
		return nil, err
	}

	var err error
	item.ManifestID, err = parseNullUUID(manifestID)
	if err != nil {
		return nil, err
	}
	item.ReleaseID, err = parseNullUUID(releaseID)
	if err != nil {
		return nil, err
	}
	if claimedAt.Valid {
		item.ClaimedAt = &claimedAt.Time
	}
	if leaseExpiresAt.Valid {
		item.LeaseExpiresAt = &leaseExpiresAt.Time
	}
	if deletedAt.Valid {
		item.DeletedAt = &deletedAt.Time
	}
	return &item, nil
}
