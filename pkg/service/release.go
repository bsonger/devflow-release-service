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
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongoDriver "go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ReleaseService = &releaseService{}

type releaseService struct{}

func populateReleaseDefaults(release *model.Release, manifest *model.Manifest, app *applicationProjection) {
	release.ApplicationID = manifest.ApplicationID
	if release.Type == "" {
		release.Type = model.ReleaseUpgrade
	}
	if release.Env == "" {
		release.Env = "prod"
	}
	release.Status = model.ReleasePending
	if len(release.Steps) == 0 {
		release.Steps = model.DefaultReleaseSteps(app.Type, release.Type)
	}
}

func (s *releaseService) Create(ctx context.Context, release *model.Release) (uuid.UUID, error) {
	log := logging.LoggerWithContext(ctx).With(zap.String("release.type", release.Type), zap.String("manifest.id", release.ManifestID.String()))

	manifest, err := ManifestService.Get(ctx, release.ManifestID)
	if err != nil {
		return uuid.Nil, err
	}
	app, err := ApplicationService.Get(ctx, manifest.ApplicationID)
	if err != nil {
		return uuid.Nil, err
	}

	populateReleaseDefaults(release, manifest, app)
	release.WithCreateDefault()
	doc, err := releaseToDoc(release)
	if err != nil {
		return uuid.Nil, err
	}
	if err := mongo.Repo.Create(ctx, doc); err != nil {
		return uuid.Nil, err
	}
	release.ID = bridgeObjectIDToUUID(doc.ID)

	log = log.With(zap.String("release.id", release.ID.String()), zap.String("application.id", release.ApplicationID.String()))

	if runtime.IsIntentMode() {
		intentID, err := IntentService.CreateReleaseIntent(ctx, release)
		if err != nil {
			return release.ID, err
		}
		log.Info("release accepted in intent mode", zap.String("intent_id", intentID.String()))
		return release.ID, nil
	}

	if err := s.DispatchRelease(ctx, release.ID); err != nil {
		s.handleSyncArgoError(ctx, release, err)
		return release.ID, err
	}
	return release.ID, nil
}

func (s *releaseService) DispatchRelease(ctx context.Context, releaseID uuid.UUID) error {
	release, err := s.Get(ctx, releaseID)
	if err != nil {
		return err
	}
	if err := s.updateStatus(ctx, release.ID, model.ReleaseSyncing); err != nil {
		return err
	}
	release.Status = model.ReleaseSyncing
	return s.syncArgo(ctx, release)
}

func (s *releaseService) handleSyncArgoError(ctx context.Context, release *model.Release, err error) {
	log := logging.LoggerWithContext(ctx).With(zap.String("release.id", release.ID.String()), zap.String("release.type", release.Type))
	log.Error("sync argo failed", zap.Error(err))
	_ = s.updateStatus(ctx, release.ID, model.ReleaseSyncFailed)
}

func (s *releaseService) Get(ctx context.Context, id uuid.UUID) (*model.Release, error) {
	oid, err := bridgeUUIDToObjectID(id)
	if err != nil {
		return nil, err
	}
	doc := &releaseDoc{}
	if err := mongo.Repo.FindByID(ctx, doc, oid); err != nil {
		return nil, err
	}
	if doc.DeletedAt != nil {
		return nil, mongoDriver.ErrNoDocuments
	}
	r := releaseFromDoc(doc)
	return &r, nil
}

func (s *releaseService) Update(ctx context.Context, release *model.Release) error {
	current, err := s.Get(ctx, release.ID)
	if err != nil {
		return err
	}
	release.CreatedAt = current.CreatedAt
	release.DeletedAt = current.DeletedAt
	release.WithUpdateDefault()
	doc, err := releaseToDoc(release)
	if err != nil {
		return err
	}
	oid, err := bridgeUUIDToObjectID(release.ID)
	if err != nil {
		return err
	}
	doc.ID = oid
	return mongo.Repo.Update(ctx, doc)
}

func (s *releaseService) Delete(ctx context.Context, id uuid.UUID) error {
	oid, err := bridgeUUIDToObjectID(id)
	if err != nil {
		return err
	}
	return mongo.Repo.UpdateByID(ctx, &releaseDoc{}, oid, primitive.M{"$set": primitive.M{"deleted_at": time.Now(), "updated_at": time.Now()}})
}

func (s *releaseService) List(ctx context.Context, filter primitive.M) ([]*model.Release, error) {
	var docs []releaseDoc
	if err := mongo.Repo.List(ctx, &releaseDoc{}, filter, &docs); err != nil {
		return nil, err
	}
	out := make([]*model.Release, 0, len(docs))
	for i := range docs {
		item := releaseFromDoc(&docs[i])
		out = append(out, &item)
	}
	return out, nil
}

func (s *releaseService) updateStatus(ctx context.Context, releaseID uuid.UUID, status model.ReleaseStatus) error {
	oid, err := bridgeUUIDToObjectID(releaseID)
	if err != nil {
		return err
	}
	return mongo.Repo.UpdateOne(ctx, &releaseDoc{}, bson.M{
		"_id":    oid,
		"status": bson.M{"$nin": []model.ReleaseStatus{model.ReleaseSucceeded, model.ReleaseFailed, model.ReleaseRolledBack, model.ReleaseSyncFailed, status}},
	}, bson.M{"$set": bson.M{"status": status, "updated_at": time.Now()}})
}

func (s *releaseService) UpdateStatus(ctx context.Context, releaseID uuid.UUID, status model.ReleaseStatus) error {
	return s.updateStatus(ctx, releaseID, status)
}

func (s *releaseService) UpdateStep(ctx context.Context, releaseID uuid.UUID, stepName string, status model.StepStatus, progress int32, message string, start, end *time.Time) error {
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
		nextSteps = append(nextSteps, model.ReleaseStep{Name: stepName, Progress: progress, Status: status, Message: message, StartTime: start, EndTime: end})
		return s.updateStatusFromSteps(ctx, releaseID, release.Type, release.Status, nextSteps)
	}
	if currentStep.Status == model.StepFailed || currentStep.Status == model.StepSucceeded {
		return nil
	}
	oid, err := bridgeUUIDToObjectID(releaseID)
	if err != nil {
		return err
	}
	update := bson.M{"steps.$.status": status, "steps.$.progress": progress, "steps.$.message": message, "updated_at": time.Now()}
	if start != nil {
		update["steps.$.start_time"] = *start
	}
	if end != nil {
		update["steps.$.end_time"] = *end
	}
	if err := mongo.Repo.UpdateOne(ctx, &releaseDoc{}, bson.M{
		"_id":   oid,
		"steps": bson.M{"$elemMatch": bson.M{"name": stepName, "status": bson.M{"$nin": []model.StepStatus{model.StepFailed, model.StepSucceeded}}}},
	}, bson.M{"$set": update}); err != nil {
		return err
	}
	applyReleaseStepUpdate(nextSteps, stepName, status, progress, message, start, end)
	return s.updateStatusFromSteps(ctx, releaseID, release.Type, release.Status, nextSteps)
}

func (s *releaseService) createStepIfNotExists(ctx context.Context, releaseID uuid.UUID, stepName string, status model.StepStatus, progress int32, message string, start, end *time.Time) error {
	oid, err := bridgeUUIDToObjectID(releaseID)
	if err != nil {
		return err
	}
	step := model.ReleaseStep{Name: stepName, Progress: progress, Status: status, Message: message, StartTime: start, EndTime: end}
	return mongo.Repo.UpdateOne(ctx, &releaseDoc{}, bson.M{
		"_id":   oid,
		"steps": bson.M{"$not": bson.M{"$elemMatch": bson.M{"name": stepName}}},
	}, bson.M{"$push": bson.M{"steps": step}, "$set": bson.M{"updated_at": time.Now()}})
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

func (s *releaseService) updateStatusFromSteps(ctx context.Context, releaseID uuid.UUID, releaseAction string, currentStatus model.ReleaseStatus, steps []model.ReleaseStep) error {
	nextStatus := model.DeriveReleaseStatusFromSteps(releaseAction, currentStatus, steps)
	if nextStatus == currentStatus {
		return nil
	}
	return s.updateStatus(ctx, releaseID, nextStatus)
}

func (s *releaseService) syncArgo(ctx context.Context, release *model.Release) error {
	log := logging.LoggerWithContext(ctx)
	app, err := ApplicationService.Get(ctx, release.ApplicationID)
	if err != nil {
		return err
	}
	manifestRepo := model.GetConfigRepo()
	manifestID := release.ManifestID.String()
	releaseID := release.ID.String()
	name := app.Name
	if name == "" {
		name = release.ApplicationID.String()
	}
	namespace := app.ProjectName
	if namespace == "" {
		namespace = "default"
	}

	application := &appv1.Application{
		TypeMeta:   metav1.TypeMeta{Kind: "Application", APIVersion: "argoproj.io/v1alpha1"},
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: appv1.ApplicationSpec{
			Project: "app",
			Source: &appv1.ApplicationSource{
				RepoURL: manifestRepo.Address,
				Path:    "./",
				Plugin: &appv1.ApplicationSourcePlugin{
					Name: "plugin",
					Parameters: []appv1.ApplicationSourcePluginParameter{
						{Name: "env", String_: &release.Env},
						{Name: "manifest-id", String_: &manifestID},
						{Name: "release-id", String_: &releaseID},
					},
				},
			},
			Destination: appv1.ApplicationDestination{Server: "https://kubernetes.default.svc", Namespace: namespace},
		},
	}
	sc := trace.SpanContextFromContext(ctx)
	application.Annotations = map[string]string{
		model.TraceIDAnnotation: sc.TraceID().String(),
		model.SpanAnnotation:    sc.SpanID().String(),
	}
	application.Labels = map[string]string{"status": string(model.ReleaseRunning), model.ReleaseIDLabel: release.ID.String()}

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
		log.Error("argo sync failed", zap.String("release_id", release.ID.String()), zap.String("type", release.Type), zap.Error(err))
		return err
	}
	return nil
}

func (s *releaseService) syncArgoApplication(ctx context.Context, appName string) error {
	applications := argo.ArgoCdClient.ArgoprojV1alpha1().Applications("argocd")
	_, err := argoutil.SetAppOperation(applications, appName, &appv1.Operation{Sync: &appv1.SyncOperation{}})
	return err
}
