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

var ReleaseService = &releaseService{}

type releaseService struct{}

func populateReleaseDefaults(release *model.Release, manifest *model.Manifest, app *model.Application) {
	release.ManifestName = manifest.Name
	release.ApplicationId = manifest.ApplicationId
	if release.Type == "" {
		release.Type = model.ReleaseUpgrade
	}
	release.ApplicationName = app.Name
	release.ProjectName = app.ProjectName
	if release.Env == "" {
		release.Env = "prod"
	}
	release.Status = model.ReleasePending
	if len(release.Steps) == 0 {
		release.Steps = model.DefaultReleaseSteps(app.Type, release.Type)
	}
}

func (s *releaseService) Create(ctx context.Context, release *model.Release) (primitive.ObjectID, error) {
	log := logging.LoggerWithContext(ctx).With(
		zap.String("release.type", release.Type),
		zap.String("manifest.id", release.ManifestID.Hex()),
	)

	log.Info("release create started")

	manifest, err := ManifestService.Get(ctx, release.ManifestID)
	if err != nil {
		log.Error("get manifest failed", zap.Error(err))
		return primitive.NilObjectID, err
	}

	app, err := ApplicationService.Get(ctx, manifest.ApplicationId)
	if err != nil {
		log.Error("get application failed",
			zap.String("application.id", manifest.ApplicationId.Hex()),
			zap.Error(err),
		)
		return primitive.NilObjectID, err
	}

	populateReleaseDefaults(release, manifest, app)
	release.WithCreateDefault()

	if err := mongo.Repo.Create(ctx, release); err != nil {
		log.Error("create release record failed", zap.Error(err))
		return primitive.NilObjectID, err
	}

	log = log.With(
		zap.String("release.id", release.ID.Hex()),
		zap.String("application.id", release.ApplicationId.Hex()),
	)

	log.Info("release record created")

	if runtime.IsIntentMode() {
		intentID, err := IntentService.CreateReleaseIntent(ctx, release)
		if err != nil {
			log.Error("create release intent failed", zap.Error(err))
			return release.ID, err
		}

		log.Info("release accepted in intent mode",
			zap.String("intent_id", intentID.Hex()),
		)

		return release.ID, nil
	}

	if err := s.DispatchRelease(ctx, release.ID); err != nil {
		s.handleSyncArgoError(ctx, release, err)
		return release.ID, err
	}

	log.Info("release synced to argo successfully")

	return release.ID, nil
}

func (s *releaseService) DispatchRelease(ctx context.Context, releaseID primitive.ObjectID) error {
	release, err := s.Get(ctx, releaseID)
	if err != nil {
		return err
	}

	log := logging.LoggerWithContext(ctx).With(
		zap.String("release.id", release.ID.Hex()),
	)

	if err := s.updateStatus(ctx, release.ID, model.ReleaseSyncing); err != nil {
		log.Error("update release status to syncing failed", zap.Error(err))
		return err
	}

	release.Status = model.ReleaseSyncing

	log.Info("release status changed",
		zap.String("release.status", string(release.Status)),
	)

	if err := s.syncArgo(ctx, release); err != nil {
		return err
	}

	log.Info("release synced to argo successfully")
	return nil
}

func (s *releaseService) handleSyncArgoError(ctx context.Context, release *model.Release, err error) {
	log := logging.LoggerWithContext(ctx).With(
		zap.String("release.id", release.ID.Hex()),
		zap.String("release.type", release.Type),
	)

	log.Error("sync argo failed", zap.Error(err))

	if uErr := s.updateStatus(ctx, release.ID, model.ReleaseSyncFailed); uErr != nil {
		log.Error("update release status to failed failed", zap.Error(uErr))
	}
}

func (s *releaseService) Get(ctx context.Context, id primitive.ObjectID) (*model.Release, error) {
	log := logging.LoggerWithContext(ctx).With(
		zap.String("release.id", id.Hex()),
		zap.String("operation", "get_release"),
	)

	release := &model.Release{}
	err := mongo.Repo.FindByID(ctx, release, id)
	if err != nil {
		log.Error("get release failed", zap.Error(err))
		return nil, err
	}
	if release.DeletedAt != nil {
		log.Warn("release already deleted")
		return nil, mongoDriver.ErrNoDocuments
	}

	log.Debug("release fetched")
	return release, nil
}

func (s *releaseService) Update(ctx context.Context, release *model.Release) error {
	log := logging.LoggerWithContext(ctx).With(
		zap.String("release.id", release.ID.Hex()),
		zap.String("operation", "update_release"),
	)

	current := &model.Release{}
	if err := mongo.Repo.FindByID(ctx, current, release.ID); err != nil {
		log.Error("load release failed", zap.Error(err))
		return err
	}
	if current.DeletedAt != nil {
		log.Warn("update skipped for deleted release")
		return mongoDriver.ErrNoDocuments
	}

	release.CreatedAt = current.CreatedAt
	release.DeletedAt = current.DeletedAt
	release.WithUpdateDefault()

	if err := mongo.Repo.Update(ctx, release); err != nil {
		log.Error("update release failed", zap.Error(err))
		return err
	}

	log.Debug("release updated")
	return nil
}

func (s *releaseService) Delete(ctx context.Context, id primitive.ObjectID) error {
	log := logging.LoggerWithContext(ctx).With(
		zap.String("release.id", id.Hex()),
		zap.String("operation", "delete_release"),
	)

	now := time.Now()
	update := primitive.M{
		"$set": primitive.M{
			"deleted_at": now,
			"updated_at": now,
		},
	}

	if err := mongo.Repo.UpdateByID(ctx, &model.Release{}, id, update); err != nil {
		log.Error("delete release failed", zap.Error(err))
		return err
	}

	log.Info("release deleted")
	return nil
}

func (s *releaseService) List(ctx context.Context, filter primitive.M) ([]*model.Release, error) {
	log := logging.LoggerWithContext(ctx).With(
		zap.String("operation", "list_releases"),
		zap.Any("filter", filter),
	)

	var releases []*model.Release
	if err := mongo.Repo.List(ctx, &model.Release{}, filter, &releases); err != nil {
		log.Error("list releases failed", zap.Error(err))
		return nil, err
	}

	log.Debug("list releases success", zap.Int("count", len(releases)))
	return releases, nil
}

func (s *releaseService) updateStatus(ctx context.Context, releaseID primitive.ObjectID, status model.ReleaseStatus) error {
	filter := bson.M{
		"_id": releaseID,
		"status": bson.M{
			"$nin": []model.ReleaseStatus{model.ReleaseSucceeded, model.ReleaseFailed, status},
		},
	}
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}
	return mongo.Repo.UpdateOne(ctx, &model.Release{}, filter, update)
}

func (s *releaseService) UpdateStatus(ctx context.Context, releaseID primitive.ObjectID, status model.ReleaseStatus) error {
	return s.updateStatus(ctx, releaseID, status)
}

func (s *releaseService) UpdateStep(ctx context.Context, releaseID primitive.ObjectID, stepName string, status model.StepStatus, progress int32, message string, start, end *time.Time) error {
	if stepName == "" {
		return nil
	}

	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}

	release, err := s.Get(ctx, releaseID)
	if err != nil {
		return err
	}

	nextSteps := cloneReleaseSteps(release.Steps)
	currentStep := findReleaseStep(release.Steps, stepName)
	if currentStep == nil {
		if err := s.createStepIfNotExists(ctx, releaseID, stepName, status, progress, message, start, end); err != nil {
			return err
		}

		nextSteps = append(nextSteps, model.ReleaseStep{
			Name:      stepName,
			Progress:  progress,
			Status:    status,
			Message:   message,
			StartTime: start,
			EndTime:   end,
		})
		return s.updateStatusFromSteps(ctx, releaseID, release.Type, release.Status, nextSteps)
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
		"_id": releaseID,
		"steps": bson.M{
			"$elemMatch": bson.M{
				"name": stepName,
				"status": bson.M{
					"$nin": []model.StepStatus{model.StepFailed, model.StepSucceeded},
				},
			},
		},
	}

	if err := mongo.Repo.UpdateOne(ctx, &model.Release{}, filter, bson.M{"$set": update}); err != nil {
		return err
	}

	applyReleaseStepUpdate(nextSteps, stepName, status, progress, message, start, end)
	return s.updateStatusFromSteps(ctx, releaseID, release.Type, release.Status, nextSteps)
}

func (s *releaseService) createStepIfNotExists(ctx context.Context, releaseID primitive.ObjectID, stepName string, status model.StepStatus, progress int32, message string, start, end *time.Time) error {
	step := model.ReleaseStep{
		Name:      stepName,
		Progress:  progress,
		Status:    status,
		Message:   message,
		StartTime: start,
		EndTime:   end,
	}

	filter := bson.M{
		"_id": releaseID,
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

	return mongo.Repo.UpdateOne(ctx, &model.Release{}, filter, update)
}

func findReleaseStep(steps []model.ReleaseStep, stepName string) *model.ReleaseStep {
	for _, step := range steps {
		if step.Name == stepName {
			current := step
			return &current
		}
	}
	return nil
}

func cloneReleaseSteps(steps []model.ReleaseStep) []model.ReleaseStep {
	if len(steps) == 0 {
		return nil
	}
	cloned := make([]model.ReleaseStep, len(steps))
	copy(cloned, steps)
	return cloned
}

func applyReleaseStepUpdate(steps []model.ReleaseStep, stepName string, status model.StepStatus, progress int32, message string, start, end *time.Time) {
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

func (s *releaseService) updateStatusFromSteps(ctx context.Context, releaseID primitive.ObjectID, releaseAction string, currentStatus model.ReleaseStatus, steps []model.ReleaseStep) error {
	nextStatus := model.DeriveReleaseStatusFromSteps(releaseAction, currentStatus, steps)
	if nextStatus == currentStatus {
		return nil
	}
	return s.updateStatus(ctx, releaseID, nextStatus)
}

func (s *releaseService) syncArgo(ctx context.Context, release *model.Release) error {
	log := logging.LoggerWithContext(ctx)
	var err error
	application := release.GenerateApplication()
	sc := trace.SpanContextFromContext(ctx)
	application.Annotations = map[string]string{
		model.TraceIDAnnotation: sc.TraceID().String(),
		model.SpanAnnotation:    sc.SpanID().String(),
	}
	application.Labels = map[string]string{
		"status":             string(model.ReleaseRunning),
		model.ReleaseIDLabel: release.ID.Hex(),
	}

	switch release.Type {
	case model.ReleaseInstall:
		err = argo.CreateApplication(ctx, application)
		if err == nil {
			err = s.syncArgoApplication(ctx, application.Name)
		}
	case model.ReleaseUpgrade, model.ReleaseRollback:
		err = argo.UpdateApplication(ctx, application)
	default:
		err = errors.New("unknown release type")
	}

	if err != nil {
		log.Error("argo sync failed",
			zap.String("release_id", release.ID.Hex()),
			zap.String("type", release.Type),
			zap.Error(err),
		)
		return err
	}

	log.Info("argo sync triggered",
		zap.String("release_id", release.ID.Hex()),
	)
	return nil
}

func (s *releaseService) syncArgoApplication(ctx context.Context, appName string) error {
	log := logging.LoggerWithContext(ctx).With(
		zap.String("application.name", appName),
	)

	applications := argo.ArgoCdClient.ArgoprojV1alpha1().Applications("argocd")
	_, err := argoutil.SetAppOperation(applications, appName, &appv1.Operation{
		Sync: &appv1.SyncOperation{},
	})
	if err != nil {
		log.Error("argo application sync failed", zap.Error(err))
		return err
	}

	log.Info("argo application sync triggered")
	return nil
}
