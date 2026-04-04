package api

import (
	"github.com/bsonger/devflow-service-common/httpx"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CreateResponse struct {
	ID                string `json:"id"`
	ExecutionIntentID string `json:"execution_intent_id,omitempty"`
}

func newCreateResponse(id primitive.ObjectID, intentID *primitive.ObjectID) CreateResponse {
	resp := httpx.NewCreateResponse(id, intentID)
	return CreateResponse{
		ID:                resp.ID,
		ExecutionIntentID: resp.ExecutionIntentID,
	}
}
