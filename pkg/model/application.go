package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Application struct {
	BaseModel `bson:",inline"`

	Name               string              `bson:"name" json:"name"`
	ProjectID          *primitive.ObjectID `bson:"project_id,omitempty" json:"project_id,omitempty"`
	ProjectName        string              `bson:"project_name" json:"project_name"`
	RepoURL            string              `bson:"repo_url" json:"repo_url"`
	ActiveManifestName string              `bson:"active_manifest_name" json:"active_manifest_name"`
	ActiveManifestID   *primitive.ObjectID `bson:"active_manifest_id,omitempty" json:"active_manifest_id,omitempty"`
	Replica            *int32              `bson:"replica,omitempty" json:"replica,omitempty"`
	Type               ReleaseType         `bson:"type" json:"type"`
	ConfigMaps         []*ConfigMap        `bson:"config_maps,omitempty" json:"config_maps,omitempty"`
	Service            Service             `bson:"service" json:"service"`
	Internet           Internet            `bson:"internet" json:"internet"`
	Envs               map[string][]EnvVar `bson:"envs,omitempty" json:"envs,omitempty"`
	Status             string              `bson:"status" json:"status"`
}

func (Application) CollectionName() string { return "applications" }
