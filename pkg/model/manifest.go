package model

import (
	"github.com/google/uuid"
)

type Manifest struct {
	BaseModel

	ApplicationID          uuid.UUID                `json:"application_id" db:"application_id"`
	EnvironmentID          uuid.UUID                `json:"environment_id" db:"environment_id"`
	ImageID                uuid.UUID                `json:"image_id" db:"image_id"`
	ImageRef               string                   `json:"image_ref" db:"image_ref"`
	ServicesSnapshot       []ManifestService        `json:"services_snapshot" db:"services_snapshot"`
	RoutesSnapshot         []ManifestRoute          `json:"routes_snapshot" db:"routes_snapshot"`
	AppConfigSnapshot      ManifestAppConfig        `json:"app_config_snapshot" db:"app_config_snapshot"`
	WorkloadConfigSnapshot ManifestWorkloadConfig   `json:"workload_config_snapshot" db:"workload_config_snapshot"`
	RenderedObjects        []ManifestRenderedObject `json:"rendered_objects" db:"rendered_objects"`
	RenderedYAML           string                   `json:"rendered_yaml" db:"rendered_yaml"`
	Status                 ManifestStatus           `json:"status" db:"status"`
}

type CreateManifestRequest struct {
	ApplicationID uuid.UUID `json:"application_id"`
	EnvironmentID uuid.UUID `json:"environment_id"`
	ImageID       uuid.UUID `json:"image_id"`
}

type ManifestListFilter struct {
	ApplicationID  *uuid.UUID
	EnvironmentID  *uuid.UUID
	ImageID        *uuid.UUID
	IncludeDeleted bool
}

type ManifestServicePort struct {
	Name        string `json:"name,omitempty"`
	ServicePort int    `json:"service_port"`
	TargetPort  int    `json:"target_port"`
	Protocol    string `json:"protocol,omitempty"`
}

type ManifestService struct {
	ID    string                `json:"id,omitempty"`
	Name  string                `json:"name"`
	Ports []ManifestServicePort `json:"ports,omitempty"`
}

type ManifestRoute struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Host        string `json:"host"`
	Path        string `json:"path"`
	ServiceName string `json:"service_name"`
	ServicePort int    `json:"service_port"`
}

type ManifestFile struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type ManifestAppConfig struct {
	ID           string            `json:"id,omitempty"`
	Name         string            `json:"name,omitempty"`
	Files        []ManifestFile    `json:"files,omitempty"`
	Data         map[string]string `json:"data,omitempty"`
	SourcePath   string            `json:"source_path,omitempty"`
	RevisionID   string            `json:"revision_id,omitempty"`
	SourceCommit string            `json:"source_commit,omitempty"`
}

type ManifestWorkloadConfig struct {
	ID           string         `json:"id,omitempty"`
	Name         string         `json:"name,omitempty"`
	Replicas     int            `json:"replicas"`
	Resources    map[string]any `json:"resources,omitempty"`
	Probes       map[string]any `json:"probes,omitempty"`
	Env          []EnvVar       `json:"env,omitempty"`
	WorkloadType string         `json:"workload_type,omitempty"`
	Strategy     string         `json:"strategy,omitempty"`
}

type ManifestRenderedObject struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	YAML      string `json:"yaml"`
}

func (m *Manifest) CollectionName() string { return "manifests" }
