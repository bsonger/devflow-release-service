package service

import (
	"errors"
	"time"

	commonmodel "github.com/bsonger/devflow-common/model"
	"github.com/bsonger/devflow-release-service/pkg/model"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	errInvalidUUIDBridge = errors.New("invalid bridged uuid")
	bridgeUUIDPrefix     = [4]byte{'d', 'f', 'l', 'w'}
)

type serviceShape struct {
	Ports []model.Port `bson:"ports,omitempty"`
}

type applicationProjection struct {
	commonmodel.BaseModel `bson:",inline"`

	Name        string            `bson:"name"`
	ProjectName string            `bson:"project_name,omitempty"`
	RepoURL     string            `bson:"repo_url,omitempty"`
	RepoAddress string            `bson:"repo_address,omitempty"`
	Replica     *int32            `bson:"replica,omitempty"`
	Type        model.ReleaseType `bson:"type"`
	Service     serviceShape      `bson:"service,omitempty"`
	Internet    model.Internet    `bson:"internet,omitempty"`
}

func (applicationProjection) CollectionName() string { return "applications" }

type manifestDoc struct {
	commonmodel.BaseModel `bson:",inline"`

	ExecutionIntentID *primitive.ObjectID     `bson:"execution_intent_id,omitempty"`
	ApplicationID     primitive.ObjectID      `bson:"application_id"`
	Name              string                  `bson:"name"`
	Branch            string                  `bson:"branch"`
	RepoAddress       string                  `bson:"repo_address,omitempty"`
	RepoURL           string                  `bson:"repo_url,omitempty"`
	CommitHash        string                  `bson:"commit_hash,omitempty"`
	Replica           *int32                  `bson:"replica,omitempty"`
	Digest            string                  `bson:"digest,omitempty"`
	Type              model.ReleaseType       `bson:"type"`
	Services          []model.ManifestService `bson:"services,omitempty"`
	PipelineID        string                  `bson:"pipeline_id,omitempty"`
	Steps             []model.ManifestStep    `bson:"steps,omitempty"`
	Status            model.ManifestStatus    `bson:"status"`
}

func (manifestDoc) CollectionName() string { return "manifests" }

type releaseDoc struct {
	commonmodel.BaseModel `bson:",inline"`

	ExecutionIntentID       *primitive.ObjectID `bson:"execution_intent_id,omitempty"`
	ConfigurationID         *primitive.ObjectID `bson:"configuration_id,omitempty"`
	ConfigurationRevisionID *primitive.ObjectID `bson:"configuration_revision_id,omitempty"`
	ApplicationID           primitive.ObjectID  `bson:"application_id"`
	ManifestID              primitive.ObjectID  `bson:"manifest_id"`
	Env                     string              `bson:"env"`
	Type                    string              `bson:"type"`
	Steps                   []model.ReleaseStep `bson:"steps,omitempty"`
	Status                  model.ReleaseStatus `bson:"status"`
	ExternalRef             string              `bson:"external_ref,omitempty"`
}

func (releaseDoc) CollectionName() string { return "releases" }

type intentDoc struct {
	commonmodel.BaseModel `bson:",inline"`

	Kind           model.IntentKind    `bson:"kind"`
	Status         model.IntentStatus  `bson:"status"`
	ResourceType   string              `bson:"resource_type"`
	ResourceID     primitive.ObjectID  `bson:"resource_id"`
	ApplicationID  primitive.ObjectID  `bson:"application_id"`
	ManifestID     *primitive.ObjectID `bson:"manifest_id,omitempty"`
	ReleaseID      *primitive.ObjectID `bson:"release_id,omitempty"`
	ReleaseType    string              `bson:"release_type,omitempty"`
	Env            string              `bson:"env,omitempty"`
	RepoAddress    string              `bson:"repo_address,omitempty"`
	RepoURL        string              `bson:"repo_url,omitempty"`
	Branch         string              `bson:"branch,omitempty"`
	ExternalRef    string              `bson:"external_ref,omitempty"`
	TraceID        string              `bson:"trace_id,omitempty"`
	Message        string              `bson:"message,omitempty"`
	LastError      string              `bson:"last_error,omitempty"`
	ClaimedBy      string              `bson:"claimed_by,omitempty"`
	ClaimedAt      *time.Time          `bson:"claimed_at,omitempty"`
	LeaseExpiresAt *time.Time          `bson:"lease_expires_at,omitempty"`
	AttemptCount   int                 `bson:"attempt_count,omitempty"`
}

func (intentDoc) CollectionName() string { return "execution_intents" }

func bridgeObjectIDToUUID(id primitive.ObjectID) uuid.UUID {
	var raw [16]byte
	copy(raw[:4], bridgeUUIDPrefix[:])
	copy(raw[4:], id[:])
	return uuid.UUID(raw)
}

func bridgeUUIDToObjectID(id uuid.UUID) (primitive.ObjectID, error) {
	raw := [16]byte(id)
	if raw[0] != bridgeUUIDPrefix[0] || raw[1] != bridgeUUIDPrefix[1] || raw[2] != bridgeUUIDPrefix[2] || raw[3] != bridgeUUIDPrefix[3] {
		return primitive.NilObjectID, errInvalidUUIDBridge
	}
	var oid primitive.ObjectID
	copy(oid[:], raw[4:])
	return oid, nil
}

func BridgeUUIDToObjectID(id uuid.UUID) (primitive.ObjectID, error) {
	return bridgeUUIDToObjectID(id)
}

func BridgeObjectIDToUUID(id primitive.ObjectID) uuid.UUID {
	return bridgeObjectIDToUUID(id)
}

func applicationIDPtr(id *primitive.ObjectID) uuid.UUID {
	if id == nil || id.IsZero() {
		return uuid.Nil
	}
	return bridgeObjectIDToUUID(*id)
}

func manifestFromDoc(doc *manifestDoc) model.Manifest {
	m := model.Manifest{
		BaseModel: model.BaseModel{
			ID:        bridgeObjectIDToUUID(doc.ID),
			CreatedAt: doc.CreatedAt,
			UpdatedAt: doc.UpdatedAt,
			DeletedAt: doc.DeletedAt,
		},
		ApplicationID: bridgeObjectIDToUUID(doc.ApplicationID),
		Name:          doc.Name,
		Branch:        doc.Branch,
		RepoAddress:   doc.RepoAddress,
		CommitHash:    doc.CommitHash,
		Replica:       doc.Replica,
		Digest:        doc.Digest,
		Type:          doc.Type,
		Services:      doc.Services,
		PipelineID:    doc.PipelineID,
		Steps:         doc.Steps,
		Status:        doc.Status,
	}
	if m.RepoAddress == "" {
		m.RepoAddress = doc.RepoURL
	}
	if doc.ExecutionIntentID != nil && !doc.ExecutionIntentID.IsZero() {
		id := bridgeObjectIDToUUID(*doc.ExecutionIntentID)
		m.ExecutionIntentID = &id
	}
	return m
}

func manifestToDoc(m *model.Manifest) (*manifestDoc, error) {
	id := primitive.NewObjectID()
	if m.ID != uuid.Nil {
		if bridged, err := bridgeUUIDToObjectID(m.ID); err == nil {
			id = bridged
		}
	}
	appID, err := bridgeUUIDToObjectID(m.ApplicationID)
	if err != nil {
		return nil, err
	}
	var intentID *primitive.ObjectID
	if m.ExecutionIntentID != nil && *m.ExecutionIntentID != uuid.Nil {
		bridged, err := bridgeUUIDToObjectID(*m.ExecutionIntentID)
		if err != nil {
			return nil, err
		}
		intentID = &bridged
	}
	return &manifestDoc{
		BaseModel: commonmodel.BaseModel{
			ID:        id,
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
			DeletedAt: m.DeletedAt,
		},
		ExecutionIntentID: intentID,
		ApplicationID:     appID,
		Name:              m.Name,
		Branch:            m.Branch,
		RepoAddress:       m.RepoAddress,
		RepoURL:           m.RepoAddress,
		CommitHash:        m.CommitHash,
		Replica:           m.Replica,
		Digest:            m.Digest,
		Type:              m.Type,
		Services:          m.Services,
		PipelineID:        m.PipelineID,
		Steps:             m.Steps,
		Status:            m.Status,
	}, nil
}

func releaseFromDoc(doc *releaseDoc) model.Release {
	r := model.Release{
		BaseModel: model.BaseModel{
			ID:        bridgeObjectIDToUUID(doc.ID),
			CreatedAt: doc.CreatedAt,
			UpdatedAt: doc.UpdatedAt,
			DeletedAt: doc.DeletedAt,
		},
		ApplicationID: bridgeObjectIDToUUID(doc.ApplicationID),
		ManifestID:    bridgeObjectIDToUUID(doc.ManifestID),
		Env:           doc.Env,
		Type:          doc.Type,
		Steps:         doc.Steps,
		Status:        doc.Status,
		ExternalRef:   doc.ExternalRef,
	}
	if doc.ExecutionIntentID != nil && !doc.ExecutionIntentID.IsZero() {
		id := bridgeObjectIDToUUID(*doc.ExecutionIntentID)
		r.ExecutionIntentID = &id
	}
	if doc.ConfigurationID != nil && !doc.ConfigurationID.IsZero() {
		id := bridgeObjectIDToUUID(*doc.ConfigurationID)
		r.ConfigurationID = &id
	}
	if doc.ConfigurationRevisionID != nil && !doc.ConfigurationRevisionID.IsZero() {
		id := bridgeObjectIDToUUID(*doc.ConfigurationRevisionID)
		r.ConfigurationRevisionID = &id
	}
	return r
}

func releaseToDoc(r *model.Release) (*releaseDoc, error) {
	id := primitive.NewObjectID()
	if r.ID != uuid.Nil {
		if bridged, err := bridgeUUIDToObjectID(r.ID); err == nil {
			id = bridged
		}
	}
	appID, err := bridgeUUIDToObjectID(r.ApplicationID)
	if err != nil {
		return nil, err
	}
	manifestID, err := bridgeUUIDToObjectID(r.ManifestID)
	if err != nil {
		return nil, err
	}
	doc := &releaseDoc{
		BaseModel: commonmodel.BaseModel{
			ID:        id,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
			DeletedAt: r.DeletedAt,
		},
		ApplicationID: appID,
		ManifestID:    manifestID,
		Env:           r.Env,
		Type:          r.Type,
		Steps:         r.Steps,
		Status:        r.Status,
		ExternalRef:   r.ExternalRef,
	}
	if r.ExecutionIntentID != nil && *r.ExecutionIntentID != uuid.Nil {
		oid, err := bridgeUUIDToObjectID(*r.ExecutionIntentID)
		if err != nil {
			return nil, err
		}
		doc.ExecutionIntentID = &oid
	}
	if r.ConfigurationID != nil && *r.ConfigurationID != uuid.Nil {
		oid, err := bridgeUUIDToObjectID(*r.ConfigurationID)
		if err != nil {
			return nil, err
		}
		doc.ConfigurationID = &oid
	}
	if r.ConfigurationRevisionID != nil && *r.ConfigurationRevisionID != uuid.Nil {
		oid, err := bridgeUUIDToObjectID(*r.ConfigurationRevisionID)
		if err != nil {
			return nil, err
		}
		doc.ConfigurationRevisionID = &oid
	}
	return doc, nil
}

func intentFromDoc(doc *intentDoc) model.Intent {
	i := model.Intent{
		BaseModel: model.BaseModel{
			ID:        bridgeObjectIDToUUID(doc.ID),
			CreatedAt: doc.CreatedAt,
			UpdatedAt: doc.UpdatedAt,
			DeletedAt: doc.DeletedAt,
		},
		Kind:           doc.Kind,
		Status:         doc.Status,
		ResourceType:   doc.ResourceType,
		ResourceID:     bridgeObjectIDToUUID(doc.ResourceID),
		ApplicationID:  bridgeObjectIDToUUID(doc.ApplicationID),
		ReleaseType:    doc.ReleaseType,
		Env:            doc.Env,
		RepoAddress:    doc.RepoAddress,
		Branch:         doc.Branch,
		ExternalRef:    doc.ExternalRef,
		TraceID:        doc.TraceID,
		Message:        doc.Message,
		LastError:      doc.LastError,
		ClaimedBy:      doc.ClaimedBy,
		ClaimedAt:      doc.ClaimedAt,
		LeaseExpiresAt: doc.LeaseExpiresAt,
		AttemptCount:   doc.AttemptCount,
	}
	if i.RepoAddress == "" {
		i.RepoAddress = doc.RepoURL
	}
	if doc.ManifestID != nil && !doc.ManifestID.IsZero() {
		id := bridgeObjectIDToUUID(*doc.ManifestID)
		i.ManifestID = &id
	}
	if doc.ReleaseID != nil && !doc.ReleaseID.IsZero() {
		id := bridgeObjectIDToUUID(*doc.ReleaseID)
		i.ReleaseID = &id
	}
	return i
}

func intentToDoc(i *model.Intent) (*intentDoc, error) {
	id := primitive.NewObjectID()
	if i.ID != uuid.Nil {
		if bridged, err := bridgeUUIDToObjectID(i.ID); err == nil {
			id = bridged
		}
	}
	resourceID, err := bridgeUUIDToObjectID(i.ResourceID)
	if err != nil {
		return nil, err
	}
	appID, err := bridgeUUIDToObjectID(i.ApplicationID)
	if err != nil {
		return nil, err
	}
	doc := &intentDoc{
		BaseModel: commonmodel.BaseModel{
			ID:        id,
			CreatedAt: i.CreatedAt,
			UpdatedAt: i.UpdatedAt,
			DeletedAt: i.DeletedAt,
		},
		Kind:           i.Kind,
		Status:         i.Status,
		ResourceType:   i.ResourceType,
		ResourceID:     resourceID,
		ApplicationID:  appID,
		ReleaseType:    i.ReleaseType,
		Env:            i.Env,
		RepoAddress:    i.RepoAddress,
		RepoURL:        i.RepoAddress,
		Branch:         i.Branch,
		ExternalRef:    i.ExternalRef,
		TraceID:        i.TraceID,
		Message:        i.Message,
		LastError:      i.LastError,
		ClaimedBy:      i.ClaimedBy,
		ClaimedAt:      i.ClaimedAt,
		LeaseExpiresAt: i.LeaseExpiresAt,
		AttemptCount:   i.AttemptCount,
	}
	if i.ManifestID != nil && *i.ManifestID != uuid.Nil {
		oid, err := bridgeUUIDToObjectID(*i.ManifestID)
		if err != nil {
			return nil, err
		}
		doc.ManifestID = &oid
	}
	if i.ReleaseID != nil && *i.ReleaseID != uuid.Nil {
		oid, err := bridgeUUIDToObjectID(*i.ReleaseID)
		if err != nil {
			return nil, err
		}
		doc.ReleaseID = &oid
	}
	return doc, nil
}
