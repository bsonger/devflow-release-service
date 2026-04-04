package service

import (
	"context"
	"errors"
	"time"

	appv1 "github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	argoutil "github.com/argoproj/argo-cd/v3/util/argo"
	"github.com/bsonger/devflow-common/client/argo"
	"github.com/bsonger/devflow-common/client/logging"
	"github.com/bsonger/devflow-common/client/mongo"
	"github.com/bsonger/devflow-release-service/pkg/model"
	"github.com/bsonger/devflow-release-service/pkg/runtime"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongoDriver "go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var JobService = &jobService{}

type jobService struct{}

//	func NewJobService() *jobService {
//		return &jobService{}
//	}
func (s *jobService) Create(ctx context.Context, job *model.Job) (primitive.ObjectID, error) {
	log := logging.LoggerWithContext(ctx).With(
		zap.String("job.type", job.Type),
		zap.String("manifest.id", job.ManifestID.Hex()),
	)

	log.Info("job create started")

	// ---------- 1️⃣ 获取 Manifest ----------
	manifest, err := ManifestService.Get(ctx, job.ManifestID)
	if err != nil {
		log.Error("get manifest failed", zap.Error(err))
		return primitive.NilObjectID, err
	}

	job.ManifestName = manifest.Name
	job.ApplicationId = manifest.ApplicationId

	// ---------- 2️⃣ 默认值 ----------
	if job.Type == "" {
		job.Type = model.JobUpgrade
	}

	// ---------- 3️⃣ 获取 Application ----------
	app, err := ApplicationService.Get(ctx, manifest.ApplicationId)
	if err != nil {
		log.Error("get application failed",
			zap.String("application.id", manifest.ApplicationId.Hex()),
			zap.Error(err),
		)
		return primitive.NilObjectID, err
	}

	job.ApplicationName = app.Name
	job.ProjectName = app.ProjectName
	job.Env = "prod"

	// ---------- 4️⃣ 初始化 Job ----------
	job.Status = model.JobPending
	if len(job.Steps) == 0 {
		job.Steps = model.DefaultJobSteps(app.Type, job.Type)
	}
	job.WithCreateDefault()

	// ---------- 5️⃣ 落库 ----------
	if err := mongo.Repo.Create(ctx, job); err != nil {
		log.Error("create job record failed", zap.Error(err))
		return primitive.NilObjectID, err
	}

	log = log.With(
		zap.String("job.id", job.ID.Hex()),
		zap.String("application.id", job.ApplicationId.Hex()),
	)

	log.Info("job record created")

	if runtime.IsIntentMode() {
		intentID, err := IntentService.CreateReleaseIntent(ctx, job)
		if err != nil {
			log.Error("create release intent failed", zap.Error(err))
			return job.ID, err
		}

		log.Info("job accepted in intent mode",
			zap.String("intent_id", intentID.Hex()),
		)

		return job.ID, nil
	}

	if err := s.DispatchRelease(ctx, job.ID); err != nil {
		s.handleSyncArgoError(ctx, job, err)
		return job.ID, err
	}

	log.Info("job synced to argo successfully")

	return job.ID, nil
}

func (s *jobService) DispatchRelease(ctx context.Context, jobID primitive.ObjectID) error {
	job, err := s.Get(ctx, jobID)
	if err != nil {
		return err
	}

	log := logging.LoggerWithContext(ctx).With(
		zap.String("job.id", job.ID.Hex()),
	)

	if err := s.updateStatus(ctx, job.ID, model.JobSyncing); err != nil {
		log.Error("update job status to syncing failed", zap.Error(err))
		return err
	}

	job.Status = model.JobSyncing

	log.Info("job status changed",
		zap.String("job.status", string(job.Status)),
	)

	if err := s.syncArgo(ctx, job); err != nil {
		return err
	}

	log.Info("job synced to argo successfully")
	return nil
}

func (s *jobService) handleSyncArgoError(ctx context.Context, job *model.Job, err error) {
	log := logging.LoggerWithContext(ctx).With(
		zap.String("job.id", job.ID.Hex()),
		zap.String("job.type", job.Type),
	)

	log.Error("sync argo failed", zap.Error(err))

	// 1️⃣ 更新状态 → Failed
	if uErr := s.updateStatus(ctx, job.ID, model.JobSyncFailed); uErr != nil {
		log.Error("update job status to failed failed", zap.Error(uErr))
	}
}

func (s *jobService) Get(ctx context.Context, id primitive.ObjectID) (*model.Job, error) {
	log := logging.LoggerWithContext(ctx).With(
		zap.String("job.id", id.Hex()),
		zap.String("operation", "get_job"),
	)

	job := &model.Job{}
	err := mongo.Repo.FindByID(ctx, job, id)
	if err != nil {
		log.Error("get job failed", zap.Error(err))
		return nil, err
	}
	if job.DeletedAt != nil {
		log.Warn("job already deleted")
		return nil, mongoDriver.ErrNoDocuments
	}

	log.Debug("job fetched")
	return job, nil
}

func (s *jobService) Update(ctx context.Context, job *model.Job) error {
	log := logging.LoggerWithContext(ctx).With(
		zap.String("job.id", job.ID.Hex()),
		zap.String("operation", "update_job"),
	)

	current := &model.Job{}
	if err := mongo.Repo.FindByID(ctx, current, job.ID); err != nil {
		log.Error("load job failed", zap.Error(err))
		return err
	}
	if current.DeletedAt != nil {
		log.Warn("update skipped for deleted job")
		return mongoDriver.ErrNoDocuments
	}

	job.CreatedAt = current.CreatedAt
	job.DeletedAt = current.DeletedAt
	job.WithUpdateDefault()

	if err := mongo.Repo.Update(ctx, job); err != nil {
		log.Error("update job failed", zap.Error(err))
		return err
	}

	log.Debug("job updated")
	return nil
}

func (s *jobService) Delete(ctx context.Context, id primitive.ObjectID) error {
	log := logging.LoggerWithContext(ctx).With(
		zap.String("job.id", id.Hex()),
		zap.String("operation", "delete_job"),
	)

	now := time.Now()
	update := primitive.M{
		"$set": primitive.M{
			"deleted_at": now,
			"updated_at": now,
		},
	}

	if err := mongo.Repo.UpdateByID(ctx, &model.Job{}, id, update); err != nil {
		log.Error("delete job failed", zap.Error(err))
		return err
	}

	log.Info("job deleted")
	return nil
}

func (s *jobService) List(ctx context.Context, filter primitive.M) ([]*model.Job, error) {
	log := logging.LoggerWithContext(ctx).With(
		zap.String("operation", "list_jobs"),
		zap.Any("filter", filter),
	)

	var jobs []*model.Job
	if err := mongo.Repo.List(ctx, &model.Job{}, filter, &jobs); err != nil {
		log.Error("list jobs failed", zap.Error(err))
		return nil, err
	}

	log.Debug("list jobs success", zap.Int("count", len(jobs)))
	return jobs, nil
}

func (s *jobService) updateStatus(ctx context.Context, jobID primitive.ObjectID, status model.JobStatus) error {
	filter := bson.M{
		"_id": jobID,
		"status": bson.M{
			"$nin": []model.JobStatus{model.JobSucceeded, model.JobFailed, status},
		},
	}
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}
	return mongo.Repo.UpdateOne(ctx, &model.Job{}, filter, update)
}

func (s *jobService) UpdateStatus(ctx context.Context, jobID primitive.ObjectID, status model.JobStatus) error {
	return s.updateStatus(ctx, jobID, status)
}

func (s *jobService) UpdateStep(ctx context.Context, jobID primitive.ObjectID, stepName string, status model.StepStatus, progress int32, message string, start, end *time.Time) error {
	if stepName == "" {
		return nil
	}

	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}

	job, err := s.Get(ctx, jobID)
	if err != nil {
		return err
	}

	nextSteps := cloneJobSteps(job.Steps)
	currentStep := findJobStep(job.Steps, stepName)
	if currentStep == nil {
		if err := s.createStepIfNotExists(ctx, jobID, stepName, status, progress, message, start, end); err != nil {
			return err
		}

		nextSteps = append(nextSteps, model.JobStep{
			Name:      stepName,
			Progress:  progress,
			Status:    status,
			Message:   message,
			StartTime: start,
			EndTime:   end,
		})
		return s.updateStatusFromSteps(ctx, jobID, job.Type, job.Status, nextSteps)
	}

	if currentStep.Status == model.StepFailed || currentStep.Status == model.StepSucceeded {
		return nil
	}

	update := bson.M{
		"steps.$.status":   status,
		"steps.$.progress": progress,
		"steps.$.message":  message,
		"updated_at":       time.Now(),
	}
	if start != nil {
		update["steps.$.start_time"] = *start
	}
	if end != nil {
		update["steps.$.end_time"] = *end
	}

	filter := bson.M{
		"_id": jobID,
		"steps": bson.M{
			"$elemMatch": bson.M{
				"name": stepName,
				"status": bson.M{
					"$nin": []model.StepStatus{model.StepFailed, model.StepSucceeded},
				},
			},
		},
	}

	if err := mongo.Repo.UpdateOne(ctx, &model.Job{}, filter, bson.M{"$set": update}); err != nil {
		return err
	}

	applyJobStepUpdate(nextSteps, stepName, status, progress, message, start, end)
	return s.updateStatusFromSteps(ctx, jobID, job.Type, job.Status, nextSteps)
}

func (s *jobService) createStepIfNotExists(ctx context.Context, jobID primitive.ObjectID, stepName string, status model.StepStatus, progress int32, message string, start, end *time.Time) error {
	step := model.JobStep{
		Name:      stepName,
		Progress:  progress,
		Status:    status,
		Message:   message,
		StartTime: start,
		EndTime:   end,
	}

	filter := bson.M{
		"_id": jobID,
		"steps": bson.M{
			"$not": bson.M{
				"$elemMatch": bson.M{
					"name": stepName,
				},
			},
		},
	}

	update := bson.M{
		"$push": bson.M{
			"steps": step,
		},
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	return mongo.Repo.UpdateOne(ctx, &model.Job{}, filter, update)
}

func findJobStep(steps []model.JobStep, stepName string) *model.JobStep {
	for _, step := range steps {
		if step.Name == stepName {
			current := step
			return &current
		}
	}
	return nil
}

func cloneJobSteps(steps []model.JobStep) []model.JobStep {
	if len(steps) == 0 {
		return nil
	}
	cloned := make([]model.JobStep, len(steps))
	copy(cloned, steps)
	return cloned
}

func applyJobStepUpdate(steps []model.JobStep, stepName string, status model.StepStatus, progress int32, message string, start, end *time.Time) {
	for i := range steps {
		if steps[i].Name != stepName {
			continue
		}
		steps[i].Status = status
		steps[i].Progress = progress
		steps[i].Message = message
		if start != nil {
			steps[i].StartTime = start
		}
		if end != nil {
			steps[i].EndTime = end
		}
		return
	}
}

func (s *jobService) updateStatusFromSteps(ctx context.Context, jobID primitive.ObjectID, jobType string, currentStatus model.JobStatus, steps []model.JobStep) error {
	nextStatus := model.DeriveJobStatusFromSteps(jobType, currentStatus, steps)
	if nextStatus == currentStatus {
		return nil
	}
	return s.updateStatus(ctx, jobID, nextStatus)
}

func (s *jobService) syncArgo(ctx context.Context, job *model.Job) error {

	log := logging.LoggerWithContext(ctx)
	var err error
	application := job.GenerateApplication()
	// 3.2 获取当前 trace context
	sc := trace.SpanContextFromContext(ctx)
	application.Annotations = map[string]string{
		model.TraceIDAnnotation: sc.TraceID().String(),
		model.SpanAnnotation:    sc.SpanID().String(),
	}
	application.Labels = map[string]string{
		"status":         string(model.JobRunning),
		model.JobIDLabel: job.ID.Hex(),
	}

	switch job.Type {
	case model.JobInstall:
		err = argo.CreateApplication(ctx, application)
		if err == nil {
			err = s.syncArgoApplication(ctx, application.Name)
		}
	case model.JobUpgrade, model.JobRollback:
		err = argo.UpdateApplication(ctx, application)
	default:
		err = errors.New("unknown job type")
	}

	if err != nil {
		log.Error("Argo sync failed",
			zap.String("job_id", job.ID.Hex()),
			zap.String("type", job.Type),
			zap.Error(err),
		)
		return err
	}

	log.Info("Argo sync triggered",
		zap.String("job_id", job.ID.Hex()),
	)
	return nil
}

func (s *jobService) syncArgoApplication(ctx context.Context, appName string) error {
	log := logging.LoggerWithContext(ctx).With(
		zap.String("application.name", appName),
	)

	applications := argo.ArgoCdClient.ArgoprojV1alpha1().Applications("argocd")
	_, err := argoutil.SetAppOperation(applications, appName, &appv1.Operation{
		Sync: &appv1.SyncOperation{},
	})
	if err != nil {
		log.Error("Argo application sync failed", zap.Error(err))
		return err
	}

	log.Info("Argo application sync triggered")
	return nil
}
