package downstream

import (
	"context"
	"fmt"
	"net/url"
)

type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ManifestFile struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type AppConfig struct {
	ID                string            `json:"id"`
	ApplicationID     string            `json:"application_id"`
	EnvironmentID     string            `json:"environment_id"`
	Name              string            `json:"name"`
	SourcePath        string            `json:"source_path"`
	Files             []ManifestFile    `json:"files,omitempty"`
	RenderedConfigMap map[string]string `json:"rendered_configmap,omitempty"`
	SourceCommit      string            `json:"source_commit,omitempty"`
}

type renderedConfigMapEnvelope struct {
	Data map[string]string `json:"data,omitempty"`
}

type WorkloadConfig struct {
	ID            string         `json:"id"`
	ApplicationID string         `json:"application_id"`
	EnvironmentID string         `json:"environment_id"`
	Name          string         `json:"name"`
	Replicas      int            `json:"replicas"`
	Resources     map[string]any `json:"resources,omitempty"`
	Probes        map[string]any `json:"probes,omitempty"`
	Env           []EnvVar       `json:"env,omitempty"`
	WorkloadType  string         `json:"workload_type,omitempty"`
	Strategy      string         `json:"strategy,omitempty"`
}

type ConfigManifestClient struct{ *Client }

func NewConfigManifestClient(baseURL string) *ConfigManifestClient {
	return &ConfigManifestClient{Client: newClient(baseURL)}
}

func (c *ConfigManifestClient) FindAppConfig(ctx context.Context, applicationID, environmentID string) (*AppConfig, error) {
	path := fmt.Sprintf("/api/v1/app-configs?application_id=%s&environment_id=%s", url.QueryEscape(applicationID), url.QueryEscape(environmentID))
	var items []AppConfig
	if err := c.getEnvelopeData(ctx, path, &items); err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}
	item := items[0]
	if item.ID == "" {
		return nil, nil
	}
	return c.GetAppConfig(ctx, item.ID)
}

func (c *ConfigManifestClient) GetAppConfig(ctx context.Context, id string) (*AppConfig, error) {
	var item struct {
		ID                string                    `json:"id"`
		ApplicationID     string                    `json:"application_id"`
		EnvironmentID     string                    `json:"environment_id"`
		Name              string                    `json:"name"`
		SourcePath        string                    `json:"source_path"`
		Files             []ManifestFile            `json:"files,omitempty"`
		RenderedConfigMap renderedConfigMapEnvelope `json:"rendered_configmap,omitempty"`
		SourceCommit      string                    `json:"source_commit,omitempty"`
	}
	if err := c.getEnvelopeData(ctx, fmt.Sprintf("/api/v1/app-configs/%s", id), &item); err != nil {
		return nil, err
	}
	return &AppConfig{
		ID:                item.ID,
		ApplicationID:     item.ApplicationID,
		EnvironmentID:     item.EnvironmentID,
		Name:              item.Name,
		SourcePath:        item.SourcePath,
		Files:             item.Files,
		RenderedConfigMap: item.RenderedConfigMap.Data,
		SourceCommit:      item.SourceCommit,
	}, nil
}

func (c *ConfigManifestClient) FindWorkloadConfig(ctx context.Context, applicationID, environmentID string) (*WorkloadConfig, error) {
	path := fmt.Sprintf("/api/v1/workload-configs?application_id=%s&environment_id=%s", url.QueryEscape(applicationID), url.QueryEscape(environmentID))
	var items []WorkloadConfig
	if err := c.getEnvelopeData(ctx, path, &items); err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}
	item := items[0]
	if item.ID == "" {
		return nil, nil
	}
	return c.GetWorkloadConfig(ctx, item.ID)
}

func (c *ConfigManifestClient) GetWorkloadConfig(ctx context.Context, id string) (*WorkloadConfig, error) {
	var item WorkloadConfig
	if err := c.getEnvelopeData(ctx, fmt.Sprintf("/api/v1/workload-configs/%s", id), &item); err != nil {
		return nil, err
	}
	return &item, nil
}
