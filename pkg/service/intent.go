package service

import (
	"context"
	"errors"
	"time"

	"github.com/bsonger/devflow-common/client/logging"
	"github.com/bsonger/devflow-common/client/mongo"
	"github.com/bsonger/devflow-release-service/pkg/model"
	"github.com/bsonger/devflow-release-service/pkg/store"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongoDriver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

var IntentService = &intentService{}

type intentService struct{}

var ErrIntentNotFound = errors.New("intent not found")

func (s *intentService) CreateBuildIntent(ctx context.Context, manifest *model.Manifest) (primitive.ObjectID, error) {
	intent := &model.Intent{
		Kind:            model.IntentKindBuild,
		Status:          model.IntentPending,
		ResourceType:    "manifest",
		ResourceID:      manifest.ID,
		ApplicationID:   manifest.ApplicationId,
		ApplicationName: manifest.ApplicationName,
		ManifestID:      objectIDPtr(manifest.ID),
		ManifestName:    manifest.Name,
		RepoURL:         manifest.GitRepo,
		Branch:          manifest.Branch,
	}
	intent.WithCreateDefault()

	if err := mongo.Repo.Create(ctx, intent); err != nil {
		return primitive.NilObjectID, err
	}

	if err := s.bindIntentToManifest(ctx, manifest.ID, intent.ID); err != nil {
		return intent.ID, err
	}
	manifest.ExecutionIntentID = objectIDPtr(intent.ID)

	logging.LoggerWithContext(ctx).Info("build intent created",
		zap.String("intent_id", intent.ID.Hex()),
		zap.String("manifest_id", manifest.ID.Hex()),
	)

	return intent.ID, nil
}

func (s *intentService) CreateReleaseIntent(ctx context.Context, release *model.Release) (primitive.ObjectID, error) {
	intent := &model.Intent{
		Kind:            model.IntentKindRelease,
		Status:          model.IntentPending,
		ResourceType:    "release",
		ResourceID:      release.ID,
		ApplicationID:   release.ApplicationId,
		ApplicationName: release.ApplicationName,
		ManifestID:      objectIDPtr(release.ManifestID),
		ManifestName:    release.ManifestName,
		ReleaseID:       objectIDPtr(release.ID),
		ReleaseType:     release.Type,
		Env:             release.Env,
	}
	intent.WithCreateDefault()

	if err := mongo.Repo.Create(ctx, intent); err != nil {
		return primitive.NilObjectID, err
	}

	if err := s.bindIntentToRelease(ctx, release.ID, intent.ID); err != nil {
		return intent.ID, err
	}
	release.ExecutionIntentID = objectIDPtr(intent.ID)

	logging.LoggerWithContext(ctx).Info("release intent created",
		zap.String("intent_id", intent.ID.Hex()),
		zap.String("release_id", release.ID.Hex()),
	)

	return intent.ID, nil
}

func (s *intentService) UpdateStatus(ctx context.Context, id primitive.ObjectID, status model.IntentStatus, externalRef, message string) error {
	update := bson.M{
		"$set": bson.M{
			"status":       status,
			"external_ref": externalRef,
			"message":      message,
			"last_error":   "",
			"updated_at":   time.Now(),
		},
	}

	return mongo.Repo.UpdateByID(ctx, &model.Intent{}, id, update)
}

func (s *intentService) Get(ctx context.Context, id primitive.ObjectID) (*model.Intent, error) {
	intent := &model.Intent{}
	if err := mongo.Repo.FindByID(ctx, intent, id); err != nil {
		return nil, err
	}
	return intent, nil
}

func (s *intentService) List(ctx context.Context, filter primitive.M) ([]*model.Intent, error) {
	var intents []*model.Intent
	if err := mongo.Repo.List(ctx, &model.Intent{}, filter, &intents); err != nil {
		return nil, err
	}
	return intents, nil
}

func (s *intentService) ListPending(ctx context.Context, limit int) ([]model.Intent, error) {
	filter := bson.M{
		"status": model.IntentPending,
	}

	var intents []model.Intent
	if err := mongo.Repo.List(ctx, &model.Intent{}, filter, &intents); err != nil {
		return nil, err
	}

	if limit > 0 && len(intents) > limit {
		intents = intents[:limit]
	}

	return intents, nil
}

func (s *intentService) ClaimNextPending(ctx context.Context, workerID string, leaseDuration time.Duration) (*model.Intent, error) {
	now := time.Now()
	leaseExpiresAt := now.Add(leaseDuration)

	filter := bson.M{
		"status": model.IntentPending,
		"$or": []bson.M{
			{"claimed_by": bson.M{"$exists": false}},
			{"claimed_by": ""},
			{"lease_expires_at": bson.M{"$lt": now}},
		},
	}

	update := bson.M{
		"$set": bson.M{
			"claimed_by":       workerID,
			"claimed_at":       now,
			"lease_expires_at": leaseExpiresAt,
			"updated_at":       now,
			"message":          "claimed by worker",
		},
		"$inc": bson.M{
			"attempt_count": 1,
		},
	}

	opts := options.FindOneAndUpdate().
		SetSort(bson.D{{Key: "created_at", Value: 1}}).
		SetReturnDocument(options.After)

	intent := &model.Intent{}
	err := store.Collection(intent.CollectionName()).
		FindOneAndUpdate(ctx, filter, update, opts).
		Decode(intent)
	if errors.Is(err, mongoDriver.ErrNoDocuments) {
		return nil, ErrIntentNotFound
	}
	if err != nil {
		return nil, err
	}
	return intent, nil
}

func (s *intentService) MarkSubmitted(ctx context.Context, id primitive.ObjectID, externalRef, message string) error {
	now := time.Now()
	return mongo.Repo.UpdateByID(ctx, &model.Intent{}, id, bson.M{
		"$set": bson.M{
			"status":           model.IntentRunning,
			"external_ref":     externalRef,
			"message":          message,
			"last_error":       "",
			"updated_at":       now,
			"claimed_by":       "",
			"claimed_at":       nil,
			"lease_expires_at": nil,
		},
	})
}

func (s *intentService) MarkFailed(ctx context.Context, id primitive.ObjectID, message string) error {
	now := time.Now()
	return mongo.Repo.UpdateByID(ctx, &model.Intent{}, id, bson.M{
		"$set": bson.M{
			"status":           model.IntentFailed,
			"message":          message,
			"last_error":       message,
			"updated_at":       now,
			"claimed_by":       "",
			"claimed_at":       nil,
			"lease_expires_at": nil,
		},
	})
}

func (s *intentService) UpdateStatusByResource(ctx context.Context, kind model.IntentKind, resourceID primitive.ObjectID, status model.IntentStatus, externalRef, message string) error {
	filter := bson.M{
		"kind":        kind,
		"resource_id": resourceID,
	}
	update := bson.M{
		"$set": bson.M{
			"status":       status,
			"external_ref": externalRef,
			"message":      message,
			"updated_at":   time.Now(),
		},
	}

	return mongo.Repo.UpdateOne(ctx, &model.Intent{}, filter, update)
}

func (s *intentService) bindIntentToManifest(ctx context.Context, manifestID, intentID primitive.ObjectID) error {
	return mongo.Repo.UpdateByID(ctx, &model.Manifest{}, manifestID, bson.M{
		"$set": bson.M{
			"execution_intent_id": intentID,
			"updated_at":          time.Now(),
		},
	})
}

func (s *intentService) bindIntentToRelease(ctx context.Context, releaseID, intentID primitive.ObjectID) error {
	return mongo.Repo.UpdateByID(ctx, &model.Release{}, releaseID, bson.M{
		"$set": bson.M{
			"execution_intent_id": intentID,
			"updated_at":          time.Now(),
		},
	})
}

func objectIDPtr(id primitive.ObjectID) *primitive.ObjectID {
	if id.IsZero() {
		return nil
	}
	v := id
	return &v
}
