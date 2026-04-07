package service

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type ManifestListFilter struct {
	IncludeDeleted bool
	ApplicationID  *uuid.UUID
	PipelineID     string
	Status         string
	Branch         string
	Name           string
}

type ReleaseListFilter struct {
	IncludeDeleted bool
	ApplicationID  *uuid.UUID
	ManifestID     *uuid.UUID
	Status         string
	Type           string
}

type IntentListFilter struct {
	Kind         string
	Status       string
	ResourceType string
	ClaimedBy    string
	ResourceID   *uuid.UUID
}

func marshalJSON(value any, empty string) ([]byte, error) {
	if value == nil {
		return []byte(empty), nil
	}
	return json.Marshal(value)
}

func nullableUUID(id uuid.UUID) any {
	if id == uuid.Nil {
		return nil
	}
	return id
}

func nullableUUIDPtr(id *uuid.UUID) any {
	if id == nil || *id == uuid.Nil {
		return nil
	}
	return *id
}

func nullableTimePtr(t *time.Time) any {
	if t == nil {
		return nil
	}
	return *t
}

func placeholderClause(column string, position int) string {
	return column + " = $" + strconv.Itoa(position)
}

func parseNullUUID(value sql.NullString) (*uuid.UUID, error) {
	if !value.Valid || value.String == "" {
		return nil, nil
	}
	parsed, err := uuid.Parse(value.String)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func parseRequiredUUID(value sql.NullString) (uuid.UUID, error) {
	if !value.Valid || value.String == "" {
		return uuid.Nil, nil
	}
	return uuid.Parse(value.String)
}

func ensureRowsAffected(result sql.Result) error {
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}
