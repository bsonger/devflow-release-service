package model

type Application struct {
	BaseModel `bson:",inline"`

	Name        string              `bson:"name" json:"name"`
	ProjectName string              `bson:"project_name" json:"project_name"`
	RepoURL     string              `bson:"repo_url" json:"repo_url"`
	Replica     *int32              `bson:"replica,omitempty" json:"replica,omitempty"`
	Type        ReleaseType         `bson:"type" json:"type"`
	ConfigMaps  []*ConfigMap        `bson:"config_maps,omitempty" json:"config_maps,omitempty"`
	Service     Service             `bson:"service" json:"service"`
	Internet    Internet            `bson:"internet" json:"internet"`
	Envs        map[string][]EnvVar `bson:"envs,omitempty" json:"envs,omitempty"`
	Status      string              `bson:"status" json:"status"`
}

func (Application) CollectionName() string { return "applications" }
