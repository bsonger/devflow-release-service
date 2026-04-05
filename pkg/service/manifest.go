package service

import (
	"context"
	"errors"
	"time"

	"github.com/bsonger/devflow-common/client/logging"
	"github.com/bsonger/devflow-common/client/mongo"
	"github.com/bsonger/devflow-common/client/tekton"
	"github.com/bsonger/devflow-release-service/pkg/model"
	"github.com/bsonger/devflow-release-service/pkg/runtime"
	"github.com/google/uuid"
	v1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var ManifestService = &manifestService{}

const namespace = "tekton-pipelines"

type manifestService struct{}

func (s *manifestService) CreateManifest(ctx context.Context, m *model.Manifest) (uuid.UUID, error) {
	logger := logging.LoggerFromContext(ctx)
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
	m.Replica = app.Replica
	m.Type = app.Type
	if len(m.Services) == 0 && len(app.Service.Ports) > 0 {
		m.Services = []model.ManifestService{{
			Name:     app.Name,
			Internet: app.Internet,
			Ports:    app.Service.Ports,
		}}
	}
	m.Name = model.GenerateManifestVersion(app.Name)
	m.Status = model.ManifestPending
	m.WithCreateDefault()
	if m.Branch == "" {
		m.Branch = "main"
	}

	if runtime.IsIntentMode() {
		doc, err := manifestToDoc(m)
		if err != nil {
			return uuid.Nil, err
		}
		if err := mongo.Repo.Create(ctx, doc); err != nil {
			return uuid.Nil, err
		}
		m.ID = bridgeObjectIDToUUID(doc.ID)
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
	doc, err := manifestToDoc(m)
	if err != nil {
		return uuid.Nil, err
	}
	if err := mongo.Repo.Create(ctx, doc); err != nil {
		return uuid.Nil, err
	}
	m.ID = bridgeObjectIDToUUID(doc.ID)

	logger.Info("create manifest success", zap.String("manifest", m.Name), zap.String("pipelineRun", m.PipelineID))
	return m.ID, nil
}

func (s *manifestService) DispatchBuild(ctx context.Context, manifestID uuid.UUID) error {
	manifest, err := s.Get(ctx, manifestID)
	if err != nil {
		return err
	}
	if err := s.submitBuild(ctx, manifest); err != nil {
		return err
	}
	oid, err := bridgeUUIDToObjectID(manifestID)
	if err != nil {
		return err
	}
	return mongo.Repo.UpdateByID(ctx, &manifestDoc{}, oid, bson.M{
		"$set": bson.M{
			"pipeline_id": manifest.PipelineID,
			"steps":       manifest.Steps,
			"updated_at":  time.Now(),
		},
	})
}

func (s *manifestService) submitBuild(ctx context.Context, m *model.Manifest) error {
	logger := logging.LoggerFromContext(ctx)

	pvc, err := tekton.CreatePVC(ctx, namespace, "devflow-ci", "local-path", "1Gi")
	if err != nil {
		logger.Error("create pvc failed", zap.Error(err))
		return err
	}

	pctx, span := StartServiceSpan(ctx, "Tekton.CreatePipelineRun")
	defer span.End()

	pr := m.GeneratePipelineRun("devflow-ci", pvc.Name)
	sc := trace.SpanContextFromContext(pctx)
	pr.Annotations = map[string]string{
		model.TraceIDAnnotation: sc.TraceID().String(),
		model.SpanAnnotation:    sc.SpanID().String(),
	}
	pr, err = tekton.CreatePipelineRun(pctx, namespace, pr)
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
	doc, err := manifestToDoc(m)
	if err != nil {
		return err
	}
	oid, err := bridgeUUIDToObjectID(m.ID)
	if err != nil {
		return err
	}
	doc.ID = oid
	return mongo.Repo.Update(ctx, doc)
}

func (s *manifestService) List(ctx context.Context, filter primitive.M) ([]model.Manifest, error) {
	var docs []manifestDoc
	if err := mongo.Repo.List(ctx, &manifestDoc{}, filter, &docs); err != nil {
		return nil, err
	}
	out := make([]model.Manifest, 0, len(docs))
	for i := range docs {
		out = append(out, manifestFromDoc(&docs[i]))
	}
	return out, nil
}

func (s *manifestService) Get(ctx context.Context, id uuid.UUID) (*model.Manifest, error) {
	oid, err := bridgeUUIDToObjectID(id)
	if err != nil {
		return nil, err
	}
	doc := &manifestDoc{}
	if err := mongo.Repo.FindByID(ctx, doc, oid); err != nil {
		return nil, err
	}
	m := manifestFromDoc(doc)
	return &m, nil
}

func (s *manifestService) AssignPipelineID(ctx context.Context, manifestID uuid.UUID, pipelineID string) error {
	if manifestID == uuid.Nil {
		return errors.New("manifest id cannot be zero")
	}
	oid, err := bridgeUUIDToObjectID(manifestID)
	if err != nil {
		return err
	}
	return mongo.Repo.UpdateByID(ctx, &manifestDoc{}, oid, bson.M{"$set": bson.M{"pipeline_id": pipelineID, "updated_at": time.Now()}})
}

func (s *manifestService) UpdateManifestStatusByID(ctx context.Context, manifestID uuid.UUID, status model.ManifestStatus) error {
	if manifestID == uuid.Nil {
		return errors.New("manifest id cannot be zero")
	}
	oid, err := bridgeUUIDToObjectID(manifestID)
	if err != nil {
		return err
	}
	return mongo.Repo.UpdateByID(ctx, &manifestDoc{}, oid, bson.M{"$set": bson.M{"status": status, "updated_at": time.Now()}})
}

func (s *manifestService) UpdateStepStatus(ctx context.Context, pipelineID, taskName string, status model.StepStatus, message string, start, end *time.Time) error {
	update := bson.M{"steps.$.status": status, "steps.$.message": message, "updated_at": time.Now()}
	if start != nil {
		update["steps.$.start_time"] = *start
	}
	if end != nil {
		update["steps.$.end_time"] = *end
	}
	filter := bson.M{
		"pipeline_id": pipelineID,
		"steps": bson.M{
			"$elemMatch": bson.M{
				"task_name": taskName,
				"status":    bson.M{"$nin": []model.StepStatus{model.StepFailed, model.StepSucceeded, status}},
			},
		},
	}
	return mongo.Repo.UpdateOne(ctx, &manifestDoc{}, filter, bson.M{"$set": update})
}

func (s *manifestService) UpdateManifestStatus(ctx context.Context, pipelineID string, status model.ManifestStatus) error {
	return mongo.Repo.UpdateOne(ctx, &manifestDoc{}, bson.M{
		"pipeline_id": pipelineID,
		"status":      bson.M{"$nin": []model.ManifestStatus{model.ManifestFailed, model.ManifestSucceeded, status}},
	}, bson.M{"$set": bson.M{"status": status, "updated_at": time.Now()}})
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
	return mongo.Repo.UpdateOne(ctx, &manifestDoc{}, bson.M{
		"pipeline_id": pipelineID,
		"steps": bson.M{
			"$elemMatch": bson.M{
				"task_name": taskName,
				"task_run":  bson.M{"$exists": false},
				"status":    bson.M{"$nin": []model.StepStatus{model.StepFailed, model.StepSucceeded}},
			},
		},
	}, bson.M{"$set": bson.M{"steps.$.task_run": taskRun, "updated_at": time.Now()}})
}

func (s *manifestService) GetManifestByPipelineID(ctx context.Context, pipelineID string) (*model.Manifest, error) {
	var doc manifestDoc
	if err := mongo.Repo.FindOne(ctx, &doc, bson.M{"pipeline_id": pipelineID}); err != nil {
		return nil, err
	}
	m := manifestFromDoc(&doc)
	return &m, nil
}

func (s *manifestService) Patch(ctx context.Context, id uuid.UUID, patch *model.PatchManifestRequest) error {
	oid, err := bridgeUUIDToObjectID(id)
	if err != nil {
		return err
	}
	set := bson.M{}
	if patch.Digest != "" {
		set["digest"] = patch.Digest
	}
	if patch.CommitHash != "" {
		set["commit_hash"] = patch.CommitHash
	}
	if len(set) == 0 {
		return nil
	}
	set["updated_at"] = time.Now()
	return mongo.Repo.UpdateOne(ctx, &manifestDoc{}, bson.M{"_id": oid}, bson.M{"$set": set})
}
