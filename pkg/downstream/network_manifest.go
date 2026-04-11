package downstream

import (
	"context"
	"fmt"
)

type ManifestServicePort struct {
	Name        string `json:"name,omitempty"`
	ServicePort int    `json:"service_port"`
	TargetPort  int    `json:"target_port"`
	Protocol    string `json:"protocol,omitempty"`
}

type ManifestService struct {
	ID    string                `json:"id"`
	Name  string                `json:"name"`
	Ports []ManifestServicePort `json:"ports,omitempty"`
}

type ManifestRoute struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Host        string `json:"host"`
	Path        string `json:"path"`
	ServiceName string `json:"service_name"`
	ServicePort int    `json:"service_port"`
}

type NetworkManifestClient struct{ *Client }

func NewNetworkManifestClient(baseURL string) *NetworkManifestClient {
	return &NetworkManifestClient{Client: newClient(baseURL)}
}

func (c *NetworkManifestClient) ListServices(ctx context.Context, applicationID string) ([]ManifestService, error) {
	var out []ManifestService
	if err := c.getEnvelopeData(ctx, fmt.Sprintf("/api/v1/applications/%s/services", applicationID), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *NetworkManifestClient) ListRoutes(ctx context.Context, applicationID string) ([]ManifestRoute, error) {
	var out []ManifestRoute
	if err := c.getEnvelopeData(ctx, fmt.Sprintf("/api/v1/applications/%s/routes", applicationID), &out); err != nil {
		return nil, err
	}
	return out, nil
}
