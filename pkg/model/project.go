package model

type ProjectStatus string

const (
	ProjectActive   ProjectStatus = "active"
	ProjectArchived ProjectStatus = "archived"
)

type Project struct {
	BaseModel `bson:",inline"`

	Name        string            `bson:"name" json:"name"`
	Key         string            `bson:"key" json:"key"`
	Description string            `bson:"description,omitempty" json:"description,omitempty"`
	Namespace   string            `bson:"namespace,omitempty" json:"namespace,omitempty"`
	Owner       string            `bson:"owner,omitempty" json:"owner,omitempty"`
	Labels      map[string]string `bson:"labels,omitempty" json:"labels,omitempty"`
	Status      ProjectStatus     `bson:"status" json:"status"`
}

func (p *Project) ApplyDefaults() {
	if p.Status == "" {
		p.Status = ProjectActive
	}
	if p.Namespace == "" {
		p.Namespace = p.Name
	}
}

func (Project) CollectionName() string { return "projects" }
