package api

import "github.com/google/uuid"

type CreateResponse struct {
	ID                string `json:"id"`
	ExecutionIntentID string `json:"execution_intent_id,omitempty"`
}

func newCreateResponse(id uuid.UUID, intentID *uuid.UUID) CreateResponse {
	resp := CreateResponse{ID: id.String()}
	if intentID != nil && *intentID != uuid.Nil {
		resp.ExecutionIntentID = intentID.String()
	}
	return resp
}
