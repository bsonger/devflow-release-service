package downstream

import (
	"context"
	"fmt"
)

type AppProject struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AppApplication struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	Name      string `json:"name"`
}

type AppEnvironment struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	ClusterID string `json:"cluster_id"`
}

type AppCluster struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Server             string `json:"server"`
	OnboardingReady    bool   `json:"onboarding_ready"`
	OnboardingError    string `json:"onboarding_error,omitempty"`
	OnboardingCheckedAt string `json:"onboarding_checked_at,omitempty"`
}

type AppManifestClient struct{ *Client }

func NewAppManifestClient(baseURL string) *AppManifestClient {
	return &AppManifestClient{Client: newClient(baseURL)}
}

func (c *AppManifestClient) GetApplication(ctx context.Context, id string) (*AppApplication, error) {
	var out AppApplication
	if err := c.getEnvelopeData(ctx, fmt.Sprintf("/api/v1/applications/%s", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *AppManifestClient) GetProject(ctx context.Context, id string) (*AppProject, error) {
	var out AppProject
	if err := c.getEnvelopeData(ctx, fmt.Sprintf("/api/v1/projects/%s", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *AppManifestClient) GetEnvironment(ctx context.Context, id string) (*AppEnvironment, error) {
	var out AppEnvironment
	if err := c.getEnvelopeData(ctx, fmt.Sprintf("/api/v1/environments/%s", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *AppManifestClient) GetCluster(ctx context.Context, id string) (*AppCluster, error) {
	var out AppCluster
	if err := c.getEnvelopeData(ctx, fmt.Sprintf("/api/v1/clusters/%s", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}
