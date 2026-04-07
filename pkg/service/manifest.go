package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/bsonger/devflow-common/client/tekton"
	"github.com/bsonger/devflow-release-service/pkg/model"
	"github.com/bsonger/devflow-release-service/pkg/runtime"
	"github.com/bsonger/devflow-release-service/pkg/store"
	"github.com/bsonger/devflow-service-common/loggingx"
	"github.com/google/uuid"
	v1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var ManifestService = &manifestService{}

const (
	tektonNamespace       = "tekton-pipelines"
	tektonBuildPipeline   = "devflow-tekton-image-build"
	tektonPVCGenerateName = "devflow-tekton-image-build"
)

type manifestService struct{}

func (s *manifestService) CreateManifest(ctx context.Context, m *model.Manifest) (uuid.UUID, error) {
	logger := loggingx.LoggerFromContext(ctx)
	logger.Info("create manifest start",
		zap.String("application_id", m.ApplicationID.String()),
		zap.String("branch", m.Branch),
	)

	app, err := ApplicationService.Get(ctx, m.ApplicationID)
	if err != nil {
		logger.Error("get application failed", zap.Error(err))
		return uuid.Nil, err
	}

	m.RepoAddress = app.RepoAddress
	if m.RepoAddress == "" {
		m.RepoAddress = app.RepoURL
	}
	m.Name = model.GenerateManifestVersion(app.Name)
	m.Status = model.ManifestPending
	m.WithCreateDefault()
	if m.Branch == "" {
		m.Branch = "main"
	}

	if runtime.IsIntentMode() {
		if err := s.insert(ctx, m); err != nil {
			return uuid.Nil, err
		}
		intentID, err := IntentService.CreateBuildIntent(ctx, m)
		if err != nil {
			return m.ID, err
		}
		logger.Info("create manifest success in intent mode", zap.String("manifest", m.Name), zap.String("intent_id", intentID.String()))
		return m.ID, nil
	}

	if err := s.submitBuild(ctx, m); err != nil {
		return uuid.Nil, err
	}
	if err := s.insert(ctx, m); err != nil {
		return uuid.Nil, err
	}

	logger.Info("create manifest success", zap.String("manifest", m.Name), zap.String("pipelineRun", m.PipelineID))
	return m.ID, nil
}

func (s *manifestService) insert(ctx context.Context, m *model.Manifest) error {
	stepsJSON, err := marshalJSON(m.Steps, "[]")
	if err != nil {
		return err
	}
	_, err = store.DB().ExecContext(ctx, `
		insert into manifests (
			id, execution_intent_id, application_id, configuration_revision_id, runtime_spec_revision_id, name, branch, repo_address, commit_hash, digest, pipeline_id, steps, status, created_at, updated_at, deleted_at
		) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
	`, m.ID, nullableUUIDPtr(m.ExecutionIntentID), m.ApplicationID, nullableUUIDPtr(m.ConfigurationRevisionID), nullableUUIDPtr(m.RuntimeSpecRevisionID), m.Name, m.Branch, m.RepoAddress, m.CommitHash, m.Digest, m.PipelineID, stepsJSON, m.Status, m.CreatedAt, m.UpdatedAt, m.DeletedAt)
	return err
}

func (s *manifestService) DispatchBuild(ctx context.Context, manifestID uuid.UUID) error {
	manifest, err := s.Get(ctx, manifestID)
	if err != nil {
		return err
	}
	if err := s.submitBuild(ctx, manifest); err != nil {
		return err
	}
	return s.updatePipelineAndSteps(ctx, manifest.ID, manifest.PipelineID, manifest.Steps)
}

func (s *manifestService) submitBuild(ctx context.Context, m *model.Manifest) error {
	logger := loggingx.LoggerFromContext(ctx)

	pvc, err := tekton.CreatePVC(ctx, tektonNamespace, tektonPVCGenerateName, "local-path", "1Gi")
	if err != nil {
		logger.Error("create pvc failed", zap.Error(err))
		return err
	}

	pctx, span := StartServiceSpan(ctx, "Tekton.CreatePipelineRun")
	defer span.End()

	pr := m.GeneratePipelineRun(tektonBuildPipeline, pvc.Name)
	sc := trace.SpanContextFromContext(pctx)
	pr.Annotations = map[string]string{
		model.TraceIDAnnotation: sc.TraceID().String(),
		model.SpanAnnotation:    sc.SpanID().String(),
	}
	pr, err = tekton.CreatePipelineRun(pctx, tektonNamespace, pr)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	if err := tekton.PatchPVCOwner(ctx, pvc, pr); err != nil {
		logger.Warn("patch pvc owner failed", zap.Error(err))
	}
	m.PipelineID = pr.Name

	pipeline, err := tekton.GetPipeline(ctx, pr.Namespace, pr.Spec.PipelineRef.Name)
	if err != nil {
		return err
	}
	m.Steps = BuildStepsFromPipeline(pipeline)
	return nil
}

func (s *manifestService) GetManifest(ctx context.Context, id uuid.UUID) (*model.Manifest, error) {
	return s.Get(ctx, id)
}

func (s *manifestService) Update(ctx context.Context, m *model.Manifest) error {
	current, err := s.Get(ctx, m.ID)
	if err != nil {
		return err
	}
	m.CreatedAt = current.CreatedAt
	m.DeletedAt = current.DeletedAt
	m.WithUpdateDefault()
	return s.updateRow(ctx, m)
}

func (s *manifestService) List(ctx context.Context, filter ManifestListFilter) ([]model.Manifest, error) {
	query := `
		select id, execution_intent_id, application_id, configuration_revision_id, runtime_spec_revision_id, name, branch, repo_address, commit_hash, digest, pipeline_id, steps, status, created_at, updated_at, deleted_at
		from manifests
	`
	clauses := make([]string, 0, 6)
	args := make([]any, 0, 6)
	if !filter.IncludeDeleted {
		clauses = append(clauses, "deleted_at is null")
	}
	if filter.ApplicationID != nil {
		args = append(args, *filter.ApplicationID)
		clauses = append(clauses, placeholderClause("application_id", len(args)))
	}
	if filter.PipelineID != "" {
		args = append(args, filter.PipelineID)
		clauses = append(clauses, placeholderClause("pipeline_id", len(args)))
	}
	if filter.Status != "" {
		args = append(args, filter.Status)
		clauses = append(clauses, placeholderClause("status", len(args)))
	}
	if filter.Branch != "" {
		args = append(args, filter.Branch)
		clauses = append(clauses, placeholderClause("branch", len(args)))
	}
	if filter.Name != "" {
		args = append(args, filter.Name)
		clauses = append(clauses, placeholderClause("name", len(args)))
	}
	if len(clauses) > 0 {
		query += " where " + strings.Join(clauses, " and ")
	}
	query += " order by created_at desc"

	rows, err := store.DB().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]model.Manifest, 0)
	for rows.Next() {
		item, err := scanManifest(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *manifestService) Get(ctx context.Context, id uuid.UUID) (*model.Manifest, error) {
	return scanManifest(store.DB().QueryRowContext(ctx, `
		select id, execution_intent_id, application_id, configuration_revision_id, runtime_spec_revision_id, name, branch, repo_address, commit_hash, digest, pipeline_id, steps, status, created_at, updated_at, deleted_at
		from manifests
		where id = $1 and deleted_at is null
	`, id))
}

func (s *manifestService) AssignPipelineID(ctx context.Context, manifestID uuid.UUID, pipelineID string) error {
	if manifestID == uuid.Nil {
		return errors.New("manifest id cannot be zero")
	}
	result, err := store.DB().ExecContext(ctx, `
		update manifests
		set pipeline_id = $2, updated_at = $3
		where id = $1 and deleted_at is null
	`, manifestID, pipelineID, time.Now())
	if err != nil {
		return err
	}
	return ensureRowsAffected(result)
}

func (s *manifestService) UpdateManifestStatusByID(ctx context.Context, manifestID uuid.UUID, status model.ManifestStatus) error {
	if manifestID == uuid.Nil {
		return errors.New("manifest id cannot be zero")
	}
	current, err := s.Get(ctx, manifestID)
	if err != nil {
		return err
	}
	current.Status = status
	current.UpdatedAt = time.Now()
	return s.updateStatusAndSteps(ctx, current.ID, current.Status, current.Steps, current.PipelineID)
}

func (s *manifestService) UpdateStepStatus(ctx context.Context, pipelineID, taskName string, status model.StepStatus, message string, start, end *time.Time) error {
	manifest, err := s.GetManifestByPipelineID(ctx, pipelineID)
	if err != nil {
		return err
	}
	changed := false
	for i := range manifest.Steps {
		if manifest.Steps[i].TaskName != taskName {
			continue
		}
		if manifest.Steps[i].Status == model.StepFailed || manifest.Steps[i].Status == model.StepSucceeded || manifest.Steps[i].Status == status {
			return nil
		}
		manifest.Steps[i].Status = status
		manifest.Steps[i].Message = message
		if start != nil {
			manifest.Steps[i].StartTime = start
		}
		if end != nil {
			manifest.Steps[i].EndTime = end
		}
		changed = true
	}
	if !changed {
		return nil
	}
	manifest.UpdatedAt = time.Now()
	return s.updateStatusAndSteps(ctx, manifest.ID, manifest.Status, manifest.Steps, manifest.PipelineID)
}

func (s *manifestService) UpdateManifestStatus(ctx context.Context, pipelineID string, status model.ManifestStatus) error {
	manifest, err := s.GetManifestByPipelineID(ctx, pipelineID)
	if err != nil {
		return err
	}
	manifest.Status = status
	manifest.UpdatedAt = time.Now()
	return s.updateStatusAndSteps(ctx, manifest.ID, manifest.Status, manifest.Steps, manifest.PipelineID)
}

func BuildStepsFromPipeline(pipeline *v1.Pipeline) []model.ManifestStep {
	steps := make([]model.ManifestStep, 0, len(pipeline.Spec.Tasks)+len(pipeline.Spec.Finally))
	for _, task := range pipeline.Spec.Tasks {
		steps = append(steps, model.ManifestStep{TaskName: task.Name, Status: model.StepPending})
	}
	for _, task := range pipeline.Spec.Finally {
		steps = append(steps, model.ManifestStep{TaskName: task.Name, Status: model.StepPending})
	}
	return steps
}

func (s *manifestService) BindTaskRun(ctx context.Context, pipelineID, taskName, taskRun string) error {
	manifest, err := s.GetManifestByPipelineID(ctx, pipelineID)
	if err != nil {
		return err
	}
	changed := false
	for i := range manifest.Steps {
		if manifest.Steps[i].TaskName == taskName {
			if manifest.Steps[i].TaskRun == taskRun {
				return nil
			}
			manifest.Steps[i].TaskRun = taskRun
			changed = true
		}
	}
	if !changed {
		return nil
	}
	manifest.UpdatedAt = time.Now()
	return s.updateStatusAndSteps(ctx, manifest.ID, manifest.Status, manifest.Steps, manifest.PipelineID)
}

func (s *manifestService) GetManifestByPipelineID(ctx context.Context, pipelineID string) (*model.Manifest, error) {
	return scanManifest(store.DB().QueryRowContext(ctx, `
		select id, execution_intent_id, application_id, configuration_revision_id, runtime_spec_revision_id, name, branch, repo_address, commit_hash, digest, pipeline_id, steps, status, created_at, updated_at, deleted_at
		from manifests
		where pipeline_id = $1 and deleted_at is null
	`, pipelineID))
}

func (s *manifestService) Patch(ctx context.Context, id uuid.UUID, patch *model.PatchManifestRequest) error {
	if patch == nil || patch.IsEmpty() {
		return nil
	}
	manifest, err := s.Get(ctx, id)
	if err != nil {
		return err
	}
	if patch.Digest != "" {
		manifest.Digest = patch.Digest
	}
	if patch.CommitHash != "" {
		manifest.CommitHash = patch.CommitHash
	}
	manifest.UpdatedAt = time.Now()
	return s.updateRow(ctx, manifest)
}

func (s *manifestService) updatePipelineAndSteps(ctx context.Context, id uuid.UUID, pipelineID string, steps []model.ManifestStep) error {
	stepsJSON, err := marshalJSON(steps, "[]")
	if err != nil {
		return err
	}
	result, err := store.DB().ExecContext(ctx, `
		update manifests
		set pipeline_id = $2, steps = $3, updated_at = $4
		where id = $1 and deleted_at is null
	`, id, pipelineID, stepsJSON, time.Now())
	if err != nil {
		return err
	}
	return ensureRowsAffected(result)
}

func (s *manifestService) updateStatusAndSteps(ctx context.Context, id uuid.UUID, status model.ManifestStatus, steps []model.ManifestStep, pipelineID string) error {
	stepsJSON, err := marshalJSON(steps, "[]")
	if err != nil {
		return err
	}
	result, err := store.DB().ExecContext(ctx, `
		update manifests
		set status = $2, steps = $3, pipeline_id = $4, updated_at = $5
		where id = $1 and deleted_at is null
	`, id, status, stepsJSON, pipelineID, time.Now())
	if err != nil {
		return err
	}
	return ensureRowsAffected(result)
}

func (s *manifestService) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := store.DB().ExecContext(ctx, `
		update manifests
		set deleted_at = $2, updated_at = $2
		where id = $1 and deleted_at is null
	`, id, time.Now())
	if err != nil {
		return err
	}
	return ensureRowsAffected(result)
}

func (s *manifestService) updateRow(ctx context.Context, m *model.Manifest) error {
	stepsJSON, err := marshalJSON(m.Steps, "[]")
	if err != nil {
		return err
	}
	result, err := store.DB().ExecContext(ctx, `
		update manifests
		set execution_intent_id=$2, application_id=$3, configuration_revision_id=$4, runtime_spec_revision_id=$5, name=$6, branch=$7, repo_address=$8, commit_hash=$9, digest=$10, pipeline_id=$11, steps=$12, status=$13, updated_at=$14, deleted_at=$15
		where id=$1 and deleted_at is null
	`, m.ID, nullableUUIDPtr(m.ExecutionIntentID), m.ApplicationID, nullableUUIDPtr(m.ConfigurationRevisionID), nullableUUIDPtr(m.RuntimeSpecRevisionID), m.Name, m.Branch, m.RepoAddress, m.CommitHash, m.Digest, m.PipelineID, stepsJSON, m.Status, m.UpdatedAt, m.DeletedAt)
	if err != nil {
		return err
	}
	return ensureRowsAffected(result)
}
