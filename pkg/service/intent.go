package service

import (
	"context"
	"errors"
	"time"

	"github.com/bsonger/devflow-common/client/logging"
	"github.com/bsonger/devflow-common/client/mongo"
	"github.com/bsonger/devflow-release-service/pkg/model"
	"github.com/bsonger/devflow-release-service/pkg/store"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongoDriver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

var IntentService = &intentService{}

type intentService struct{}

var ErrIntentNotFound = errors.New("intent not found")

func (s *intentService) CreateBuildIntent(ctx context.Context, manifest *model.Manifest) (uuid.UUID, error) {
	intent := &model.Intent{
		Kind:          model.IntentKindBuild,
		Status:        model.IntentPending,
		ResourceType:  "manifest",
		ResourceID:    manifest.ID,
		ApplicationID: manifest.ApplicationID,
		ManifestID:    uuidPtr(manifest.ID),
		RepoAddress:   manifest.RepoAddress,
		Branch:        manifest.Branch,
	}
	intent.WithCreateDefault()
	doc, err := intentToDoc(intent)
	if err != nil {
		return uuid.Nil, err
	}
	if err := mongo.Repo.Create(ctx, doc); err != nil {
		return uuid.Nil, err
	}
	intent.ID = bridgeObjectIDToUUID(doc.ID)
	if err := s.bindIntentToManifest(ctx, manifest.ID, intent.ID); err != nil {
		return intent.ID, err
	}
	manifest.ExecutionIntentID = uuidPtr(intent.ID)
	logging.LoggerWithContext(ctx).Info("build intent created", zap.String("intent_id", intent.ID.String()), zap.String("manifest_id", manifest.ID.String()))
	return intent.ID, nil
}

func (s *intentService) CreateReleaseIntent(ctx context.Context, release *model.Release) (uuid.UUID, error) {
	intent := &model.Intent{
		Kind:          model.IntentKindRelease,
		Status:        model.IntentPending,
		ResourceType:  "release",
		ResourceID:    release.ID,
		ApplicationID: release.ApplicationID,
		ManifestID:    uuidPtr(release.ManifestID),
		ReleaseID:     uuidPtr(release.ID),
		ReleaseType:   release.Type,
		Env:           release.Env,
	}
	intent.WithCreateDefault()
	doc, err := intentToDoc(intent)
	if err != nil {
		return uuid.Nil, err
	}
	if err := mongo.Repo.Create(ctx, doc); err != nil {
		return uuid.Nil, err
	}
	intent.ID = bridgeObjectIDToUUID(doc.ID)
	if err := s.bindIntentToRelease(ctx, release.ID, intent.ID); err != nil {
		return intent.ID, err
	}
	release.ExecutionIntentID = uuidPtr(intent.ID)
	logging.LoggerWithContext(ctx).Info("release intent created", zap.String("intent_id", intent.ID.String()), zap.String("release_id", release.ID.String()))
	return intent.ID, nil
}

func (s *intentService) UpdateStatus(ctx context.Context, id uuid.UUID, status model.IntentStatus, externalRef, message string) error {
	oid, err := bridgeUUIDToObjectID(id)
	if err != nil {
		return err
	}
	return mongo.Repo.UpdateByID(ctx, &intentDoc{}, oid, bson.M{"$set": bson.M{
		"status":       status,
		"external_ref": externalRef,
		"message":      message,
		"last_error":   "",
		"updated_at":   time.Now(),
	}})
}

func (s *intentService) Get(ctx context.Context, id uuid.UUID) (*model.Intent, error) {
	oid, err := bridgeUUIDToObjectID(id)
	if err != nil {
		return nil, err
	}
	doc := &intentDoc{}
	if err := mongo.Repo.FindByID(ctx, doc, oid); err != nil {
		return nil, err
	}
	intent := intentFromDoc(doc)
	return &intent, nil
}

func (s *intentService) List(ctx context.Context, filter primitive.M) ([]*model.Intent, error) {
	var docs []intentDoc
	if err := mongo.Repo.List(ctx, &intentDoc{}, filter, &docs); err != nil {
		return nil, err
	}
	out := make([]*model.Intent, 0, len(docs))
	for i := range docs {
		intent := intentFromDoc(&docs[i])
		out = append(out, &intent)
	}
	return out, nil
}

func (s *intentService) ListPending(ctx context.Context, limit int) ([]model.Intent, error) {
	var docs []intentDoc
	if err := mongo.Repo.List(ctx, &intentDoc{}, bson.M{"status": model.IntentPending}, &docs); err != nil {
		return nil, err
	}
	if limit > 0 && len(docs) > limit {
		docs = docs[:limit]
	}
	out := make([]model.Intent, 0, len(docs))
	for i := range docs {
		out = append(out, intentFromDoc(&docs[i]))
	}
	return out, nil
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
		"$inc": bson.M{"attempt_count": 1},
	}
	opts := options.FindOneAndUpdate().SetSort(bson.D{{Key: "created_at", Value: 1}}).SetReturnDocument(options.After)
	doc := &intentDoc{}
	err := store.Collection(doc.CollectionName()).FindOneAndUpdate(ctx, filter, update, opts).Decode(doc)
	if errors.Is(err, mongoDriver.ErrNoDocuments) {
		return nil, ErrIntentNotFound
	}
	if err != nil {
		return nil, err
	}
	intent := intentFromDoc(doc)
	return &intent, nil
}

func (s *intentService) MarkSubmitted(ctx context.Context, id uuid.UUID, externalRef, message string) error {
	oid, err := bridgeUUIDToObjectID(id)
	if err != nil {
		return err
	}
	now := time.Now()
	return mongo.Repo.UpdateByID(ctx, &intentDoc{}, oid, bson.M{"$set": bson.M{
		"status":           model.IntentRunning,
		"external_ref":     externalRef,
		"message":          message,
		"last_error":       "",
		"updated_at":       now,
		"claimed_by":       "",
		"claimed_at":       nil,
		"lease_expires_at": nil,
	}})
}

func (s *intentService) MarkFailed(ctx context.Context, id uuid.UUID, message string) error {
	oid, err := bridgeUUIDToObjectID(id)
	if err != nil {
		return err
	}
	now := time.Now()
	return mongo.Repo.UpdateByID(ctx, &intentDoc{}, oid, bson.M{"$set": bson.M{
		"status":           model.IntentFailed,
		"message":          message,
		"last_error":       message,
		"updated_at":       now,
		"claimed_by":       "",
		"claimed_at":       nil,
		"lease_expires_at": nil,
	}})
}

func (s *intentService) UpdateStatusByResource(ctx context.Context, kind model.IntentKind, resourceID uuid.UUID, status model.IntentStatus, externalRef, message string) error {
	resourceOID, err := bridgeUUIDToObjectID(resourceID)
	if err != nil {
		return err
	}
	return mongo.Repo.UpdateOne(ctx, &intentDoc{}, bson.M{
		"kind":        kind,
		"resource_id": resourceOID,
	}, bson.M{"$set": bson.M{"status": status, "external_ref": externalRef, "message": message, "updated_at": time.Now()}})
}

func (s *intentService) bindIntentToManifest(ctx context.Context, manifestID, intentID uuid.UUID) error {
	manifestOID, err := bridgeUUIDToObjectID(manifestID)
	if err != nil {
		return err
	}
	intentOID, err := bridgeUUIDToObjectID(intentID)
	if err != nil {
		return err
	}
	return mongo.Repo.UpdateByID(ctx, &manifestDoc{}, manifestOID, bson.M{"$set": bson.M{"execution_intent_id": intentOID, "updated_at": time.Now()}})
}

func (s *intentService) bindIntentToRelease(ctx context.Context, releaseID, intentID uuid.UUID) error {
	releaseOID, err := bridgeUUIDToObjectID(releaseID)
	if err != nil {
		return err
	}
	intentOID, err := bridgeUUIDToObjectID(intentID)
	if err != nil {
		return err
	}
	return mongo.Repo.UpdateByID(ctx, &releaseDoc{}, releaseOID, bson.M{"$set": bson.M{"execution_intent_id": intentOID, "updated_at": time.Now()}})
}

func uuidPtr(id uuid.UUID) *uuid.UUID {
	if id == uuid.Nil {
		return nil
	}
	v := id
	return &v
}
