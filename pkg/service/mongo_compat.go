package service

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/bsonger/devflow-release-service/pkg/model"
	"github.com/google/uuid"
)

type applicationProjection struct {
	ID          uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
	Name        string
	ProjectName string
	RepoAddress string
	RepoURL     string
}

func scanImage(scanner interface {
	Scan(dest ...any) error
}) (*model.Image, error) {
	var (
		item                    model.Image
		executionIntent         sql.NullString
		configurationRevisionID sql.NullString
		runtimeSpecRevisionID   sql.NullString
		stepsBytes              []byte
		deletedAt               sql.NullTime
	)

	if err := scanner.Scan(
		&item.ID,
		&executionIntent,
		&item.ApplicationID,
		&configurationRevisionID,
		&runtimeSpecRevisionID,
		&item.Name,
		&item.Branch,
		&item.RepoAddress,
		&item.CommitHash,
		&item.Digest,
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

func scanRelease(scanner interface {
	Scan(dest ...any) error
}) (*model.Release, error) {
	var (
		item            model.Release
		executionIntent sql.NullString
		stepsBytes      []byte
		deletedAt       sql.NullTime
	)

	if err := scanner.Scan(
		&item.ID,
		&executionIntent,
		&item.ApplicationID,
		&item.ImageID,
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
