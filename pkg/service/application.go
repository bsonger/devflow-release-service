package service

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/bsonger/devflow-release-service/pkg/model"
	"github.com/bsonger/devflow-release-service/pkg/store"
	"github.com/bsonger/devflow-service-common/loggingx"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var ApplicationService = NewApplicationService()

type applicationService struct{}

func NewApplicationService() *applicationService {
	return &applicationService{}
}

func (s *applicationService) Get(ctx context.Context, id uuid.UUID) (*applicationProjection, error) {
	log := loggingx.LoggerWithContext(ctx).With(
		zap.String("operation", "get_application"),
		zap.String("application_id", id.String()),
	)

	row := store.DB().QueryRowContext(ctx, `
		select
			a.id,
			a.name,
			coalesce(p.name, ''),
			a.repo_address,
			a.replica,
			a.type,
			coalesce(s.internet, ''),
			coalesce(s.ports, '[]'::jsonb),
			a.created_at,
			a.updated_at,
			a.deleted_at
		from applications a
		left join projects p on p.id = a.project_id and p.deleted_at is null
		left join lateral (
			select internet, ports
			from services
			where application_id = a.id and deleted_at is null
			order by created_at asc
			limit 1
		) s on true
		where a.id = $1 and a.deleted_at is null
	`, id)

	app, err := scanApplicationProjection(row)
	if err != nil {
		log.Error("get application failed", zap.Error(err))
		return nil, err
	}

	log.Debug("application fetched", zap.String("application_name", app.Name))
	return app, nil
}

func scanApplicationProjection(scanner interface {
	Scan(dest ...any) error
}) (*applicationProjection, error) {
	var (
		app        applicationProjection
		replica    sql.NullInt32
		internet   sql.NullString
		portsBytes []byte
		deletedAt  sql.NullTime
	)

	if err := scanner.Scan(
		&app.ID,
		&app.Name,
		&app.ProjectName,
		&app.RepoAddress,
		&replica,
		&app.Type,
		&internet,
		&portsBytes,
		&app.CreatedAt,
		&app.UpdatedAt,
		&deletedAt,
	); err != nil {
		return nil, err
	}

	if replica.Valid {
		value := replica.Int32
		app.Replica = &value
	}
	if internet.Valid {
		app.Internet = model.Internet(internet.String)
	}
	if len(portsBytes) > 0 {
		if err := json.Unmarshal(portsBytes, &app.Service.Ports); err != nil {
			return nil, err
		}
	}
	if deletedAt.Valid {
		app.DeletedAt = &deletedAt.Time
	}

	return &app, nil
}
