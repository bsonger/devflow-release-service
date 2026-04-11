package downstream

import (
	"context"
	"fmt"
)

type ApplicationEnvironment struct {
	ID            string `json:"id"`
	ApplicationID string `json:"application_id"`
}

type OrchestratorManifestClient struct{ *Client }

func NewOrchestratorManifestClient(baseURL string) *OrchestratorManifestClient {
	return &OrchestratorManifestClient{Client: newClient(baseURL)}
}

func (c *OrchestratorManifestClient) GetApplicationEnvironment(ctx context.Context, applicationID, environmentID string) (*ApplicationEnvironment, error) {
	var out ApplicationEnvironment
	if err := c.getEnvelopeData(ctx, fmt.Sprintf("/api/v1/platform/applications/%s/environments/%s", applicationID, environmentID), &out); err != nil {
		return nil, err
	}
	return &out, nil
}
